package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"google.golang.org/protobuf/proto"
	"net"
	"os"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/contracts/validate"
	"github.com/AlonaBot/AlonaBot/internal/wire"
)

// runBadOutput 以真实协议端点注入响应 schema 错误，绕过 SDK 正常输出的保证。
func runBadOutput() error {
	var conn *wire.Conn
	if os.Getenv(wire.EnvTransport) == wire.TransportStdio {
		conn = wire.NewConn(os.Stdin, os.Stdout, os.Stdin)
	} else {
		stream, err := net.Dial("tcp", os.Getenv(wire.EnvCoreAddr))
		if err != nil {
			return err
		}
		conn = wire.NewStreamConn(stream)
	}
	defer conn.Close()
	id := os.Getenv(wire.EnvPluginID)
	secret, err := hex.DecodeString(os.Getenv(wire.EnvSecret))
	if err != nil {
		return err
	}
	digest, err := hex.DecodeString(os.Getenv(wire.EnvBuildDigest))
	if err != nil {
		return err
	}
	send := func(sid []byte, minor uint32, kind pb.FrameKind, message proto.Message) error {
		env, err := wire.ControlEnvelope(sid, minor, id, kind, message)
		if err != nil {
			return err
		}
		return conn.WriteFrame(env)
	}
	if err := send(nil, 0, pb.FrameKind_FRAME_KIND_HELLO, &pb.Hello{PluginId: id, Secret: secret, BuildDigest: digest, SupportedMajor: validate.ProtocolMajor, SupportedMinor: validate.ProtocolMinor}); err != nil {
		return err
	}
	receive := func() (any, error) {
		env, err := conn.ReadFrame()
		if err != nil {
			return nil, err
		}
		return validate.DecodeControl(env)
	}
	decoded, err := receive()
	if err != nil {
		return err
	}
	ack, ok := decoded.(*pb.HelloAck)
	if !ok {
		return fmt.Errorf("expected HelloAck, got %T", decoded)
	}
	sid, minor := ack.GetSessionId(), ack.GetChosenMinor()
	registration := &pb.Register{CandidateConfigJson: "{}", CandidateConfigVersion: 1, Routes: []*pb.RouteDescriptor{{RouteId: "svc.bad", Kind: pb.RouteKind_ROUTE_KIND_SERVICE, InputSchemaId: inputSchema, InputSchemaVersion: schemaVersion, OutputSchemaId: outputSchema, OutputSchemaVersion: schemaVersion, Spec: &pb.RouteDescriptor_Service{Service: &pb.ServiceSpec{ServiceName: "bad", Access: pb.ServiceAccess_SERVICE_ACCESS_PUBLIC, Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON}}}}}
	registration.Migration = &pb.MigrationState{Mode: pb.MigrationMode_MIGRATION_MODE_NONE}
	if err := send(sid, minor, pb.FrameKind_FRAME_KIND_REGISTER, registration); err != nil {
		return err
	}
	if decoded, err = receive(); err != nil {
		return err
	}
	if _, ok := decoded.(*pb.Prepared); !ok {
		return fmt.Errorf("expected Prepared, got %T", decoded)
	}
	decoded, err = receive()
	if err != nil {
		return err
	}
	activation, ok := decoded.(*pb.Activate)
	if !ok {
		return fmt.Errorf("expected Activate, got %T", decoded)
	}
	if err := send(sid, minor, pb.FrameKind_FRAME_KIND_ACTIVATED, &pb.Activated{GenerationId: activation.GenerationId, Ready: true}); err != nil {
		return err
	}
	stopped := make(chan struct{})
	session := wire.NewSession(conn, wire.Options{Local: pb.InitiatorSide_INITIATOR_SIDE_PLUGIN, SessionID: sid, ProtocolMinor: minor, SourcePluginID: id,
		Handler: func(context.Context, *wire.Inbound) wire.Response {
			return wire.Response{Payload: wire.Payload{Encoding: pb.PayloadEncoding_PAYLOAD_ENCODING_JSON, SchemaID: "wrong.output.schema", SchemaVersion: schemaVersion, Bytes: []byte(`{}`)}}
		},
		Control: func(env *pb.Envelope) error {
			if env.Kind == pb.FrameKind_FRAME_KIND_STOPPING {
				close(stopped)
			}
			return nil
		},
	})
	defer session.Close()
	// STOPPING 或断链结束辅助进程，主 goroutine 不占业务执行器。
	select {
	case <-stopped:
		return nil
	case <-session.Done():
		return session.Err()
	}
}
