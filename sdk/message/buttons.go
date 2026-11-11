package message

import (
	"encoding/json"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
	"github.com/AlonaBot/AlonaBot/sdk/internal/messagecodec"
	"google.golang.org/protobuf/proto"
)

// Button 指定动作名、显示标签与 JSON 编码的值。
type Button = messagecodec.Button

type buttons struct {
	messagecodec.Failure
	rows [][]Button
}

// Buttons 按行构造按钮键盘，编码失败随片段留到发送入口报告。
func Buttons(rows ...[]Button) Segment {
	var err error
	for _, row := range rows {
		for _, button := range row {
			if failure := messagecodec.ButtonError(button); err == nil {
				err = failure
			}
		}
	}
	return buttons{Failure: messagecodec.NewFailure(err), rows: rows}
}

func (b buttons) toProto() *pb.Segment {
	segment := &pb.ButtonsSegment{}
	for _, row := range b.rows {
		out := &pb.ButtonRow{}
		for _, button := range row {
			out.Buttons = append(out.Buttons, &pb.Button{ActionId: button.ActionID, Label: button.Label, Value: proto.String(button.Value)})
		}
		segment.Rows = append(segment.Rows, out)
	}
	return &pb.Segment{Data: &pb.Segment_Buttons{Buttons: segment}}
}

type extension struct {
	messagecodec.Failure
	payload *pb.ExtensionSegment
}

// Extension 按目标 Adapter 的声明构造版本化的 JSON 片段。
func Extension(adapterID, name string, version uint32, payload any) Segment {
	encoded, err := json.Marshal(payload)
	return extension{Failure: messagecodec.NewFailure(err), payload: &pb.ExtensionSegment{
		AdapterId: adapterID, Name: name, Version: version, PayloadJson: string(encoded),
	}}
}

func (e extension) toProto() *pb.Segment {
	return &pb.Segment{Data: &pb.Segment_Extension{Extension: e.payload}}
}
