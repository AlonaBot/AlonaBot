package validate

import (
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

func cmdRoute(id string) *pb.RouteDescriptor {
	return &pb.RouteDescriptor{RouteId: id, Kind: pb.RouteKind_ROUTE_KIND_COMMAND,
		Spec: &pb.RouteDescriptor_Command{Command: &pb.CommandSpec{Match: pb.CommandMatch_COMMAND_MATCH_PATH, Path: []string{id}}}}
}

func TestRegister(t *testing.T) {
	ok := &pb.Register{Routes: []*pb.RouteDescriptor{cmdRoute("a"), cmdRoute("b")}, CandidateConfigJson: `{"x":1}`,
		HttpListener: &pb.HttpListener{Host: "127.0.0.1", Port: 8080}}
	if err := Register(ok); err != nil {
		t.Fatalf("合法注册应通过: %v", err)
	}
	cases := map[string]struct {
		r    *pb.Register
		code pb.ErrorCode
	}{
		"重复 route_id": {&pb.Register{Routes: []*pb.RouteDescriptor{cmdRoute("a"), cmdRoute("a")}, CandidateConfigJson: "{}"}, pb.ErrorCode_ERROR_CODE_ALREADY_EXISTS},
		"缺少 route_id": {&pb.Register{Routes: []*pb.RouteDescriptor{cmdRoute("")}, CandidateConfigJson: "{}"}, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		"kind 与描述不符": {&pb.Register{Routes: []*pb.RouteDescriptor{{RouteId: "a", Kind: pb.RouteKind_ROUTE_KIND_HTTP,
			Spec: &pb.RouteDescriptor_Command{Command: &pb.CommandSpec{}}}}, CandidateConfigJson: "{}"}, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		"kind 未设置":   {&pb.Register{Routes: []*pb.RouteDescriptor{{RouteId: "a"}}, CandidateConfigJson: "{}"}, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		"配置不是对象":     {&pb.Register{CandidateConfigJson: "[1]"}, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		"配置为空":       {&pb.Register{}, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		"非 loopback": {&pb.Register{CandidateConfigJson: "{}", HttpListener: &pb.HttpListener{Host: "0.0.0.0", Port: 80}}, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
		"端口为 0":      {&pb.Register{CandidateConfigJson: "{}", HttpListener: &pb.HttpListener{Host: "::1"}}, pb.ErrorCode_ERROR_CODE_INVALID_ARGUMENT},
	}
	for name, c := range cases {
		if err := Register(c.r); code(err) != c.code {
			t.Errorf("%s: 错误码 %v，期望 %v（%v）", name, code(err), c.code, err)
		}
	}
}
