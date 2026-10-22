// 本文件测试协议版本与能力协商。
package validate

import (
	"testing"

	pb "github.com/AlonaBot/AlonaBot/contracts/pb/alonabotv1"
)

// caps 把 name, version 成对参数转为能力列表。
func caps(pairs ...any) []*pb.ProtocolCapability {
	var out []*pb.ProtocolCapability
	for i := 0; i < len(pairs); i += 2 {
		out = append(out, &pb.ProtocolCapability{Name: pairs[i].(string), Version: uint32(pairs[i+1].(int))})
	}
	return out
}

// Minor 取较小值、Major 必须一致、能力按 name+version 精确匹配，双方必需能力都要在交集中。
func TestNegotiate(t *testing.T) {
	core := Offer{Major: 1, Minor: 2, Supported: caps("send", 1, "media", 1, "bus", 1)}
	cases := []struct {
		// name 是子测试名。
		name string
		// peer 是对端的声明。
		peer Offer
		// wantErr 为 true 表示协商应失败。
		wantErr bool
		// wantMinor 是期望的协商次版本。
		wantMinor uint32
		// wantInter 是期望的交集大小。
		wantInter int
	}{
		{"Minor 取较小值", Offer{Major: 1, Minor: 0, Supported: caps("send", 1)}, false, 0, 1},
		{"对端 Minor 更高时取本端", Offer{Major: 1, Minor: 5, Supported: caps("send", 1, "bus", 1)}, false, 2, 2},
		{"Major 不一致", Offer{Major: 2, Minor: 0}, true, 0, 0},
		{"对端必需能力不在交集", Offer{Major: 1, Supported: caps("send", 1, "agent", 1), Required: caps("agent", 1)}, true, 0, 0},
		{"能力版本不同不算匹配", Offer{Major: 1, Supported: caps("media", 2), Required: caps("media", 2)}, true, 0, 0},
		{"必需能力在交集中", Offer{Major: 1, Supported: caps("media", 1), Required: caps("media", 1)}, false, 0, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Negotiate(core, c.peer)
			if (err != nil) != c.wantErr {
				t.Fatalf("err = %v，wantErr=%v", err, c.wantErr)
			}
			if err != nil {
				if code(err) != pb.ErrorCode_ERROR_CODE_FAILED_PRECONDITION {
					t.Fatalf("协商失败应为 FAILED_PRECONDITION，实际 %v", err)
				}
				return
			}
			if got.Minor != c.wantMinor || len(got.Intersection) != c.wantInter {
				t.Fatalf("got minor=%d inter=%d，期望 %d/%d", got.Minor, len(got.Intersection), c.wantMinor, c.wantInter)
			}
		})
	}
	// 本端自身的必需能力也必须被对端支持。
	strict := Offer{Major: 1, Supported: caps("send", 1), Required: caps("send", 1)}
	if _, err := Negotiate(strict, Offer{Major: 1}); err == nil {
		t.Fatal("本端必需能力缺失时应失败")
	}
}
