//go:build windows

package state

import (
	"fmt"
	"runtime"

	"golang.org/x/sys/windows"
)

func protectStateDirectory(path string) error {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return fmt.Errorf("read account for data_dir %q: %w", path, err)
	}
	var pin runtime.Pinner
	pin.Pin(user.User.Sid) // TrusteeValue 保存 SID 的地址，构造 ACL 时保持该地址有效。
	defer pin.Unpin()
	acl, err := windows.ACLFromEntries([]windows.EXPLICIT_ACCESS{{
		AccessPermissions: windows.GENERIC_ALL,
		AccessMode:        windows.SET_ACCESS,
		Inheritance:       windows.OBJECT_INHERIT_ACE | windows.CONTAINER_INHERIT_ACE,
		Trustee:           windows.TRUSTEE{TrusteeForm: windows.TRUSTEE_IS_SID, TrusteeType: windows.TRUSTEE_IS_USER, TrusteeValue: windows.TrusteeValueFromSID(user.User.Sid)},
	}}, nil)
	if err != nil {
		return fmt.Errorf("create data_dir ACL %q: %w", path, err)
	}
	// 关闭父目录 ACL 继承；子文件和目录只继承当前账号的完整控制权限。
	if err := windows.SetNamedSecurityInfo(path, windows.SE_FILE_OBJECT, windows.DACL_SECURITY_INFORMATION|windows.PROTECTED_DACL_SECURITY_INFORMATION, nil, nil, acl, nil); err != nil {
		return fmt.Errorf("protect data_dir %q: %w", path, err)
	}
	return nil
}

// Windows 文件由受保护的 data_dir 继承 DACL，Unix mode 位不表示有效访问权限。
func checkStatePermissions(string) error { return nil }
