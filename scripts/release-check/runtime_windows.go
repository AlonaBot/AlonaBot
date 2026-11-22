package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	fileSecurityObject      = 1 // SE_FILE_OBJECT
	daclSecurityInformation = 4 // DACL_SECURITY_INFORMATION
	aclSizeInformation      = 2 // ACL_SIZE_INFORMATION
	accessAllowedACE        = 0
	inheritOnlyACE          = 0x08
	fileReadData            = 0x00000001
	genericRead             = 0x80000000
	genericAll              = 0x10000000
)

var (
	generateConsoleCtrlEvent = syscall.NewLazyDLL("kernel32.dll").NewProc("GenerateConsoleCtrlEvent")
	advapi32                 = syscall.NewLazyDLL("advapi32.dll")
	getNamedSecurityInfo     = advapi32.NewProc("GetNamedSecurityInfoW")
	getACLInformation        = advapi32.NewProc("GetAclInformation")
	getACE                   = advapi32.NewProc("GetAce")
)

func prepareRuntime(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}

func interruptRuntime(process *os.Process) error {
	// CTRL_BREAK_EVENT 定向本次 Core 进程组，由 Core 的信号处理正常退出。
	ok, _, err := generateConsoleCtrlEvent.Call(syscall.CTRL_BREAK_EVENT, uintptr(process.Pid))
	if ok == 0 {
		return err
	}
	return nil
}

func checkTokenPermissions(path string) (err error) {
	// 1. Windows mode 位没有 Unix 0600 语义，读取文件继承到的实际 DACL。
	name, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	var descriptor, dacl unsafe.Pointer
	code, _, _ := getNamedSecurityInfo.Call(uintptr(unsafe.Pointer(name)), fileSecurityObject, daclSecurityInformation,
		0, 0, uintptr(unsafe.Pointer(&dacl)), 0, uintptr(unsafe.Pointer(&descriptor)))
	if code != 0 {
		return fmt.Errorf("read first token DACL: %w", syscall.Errno(code))
	}
	defer func() {
		_, freeErr := syscall.LocalFree(syscall.Handle(descriptor))
		err = errors.Join(err, freeErr)
	}()
	if dacl == nil {
		return errors.New("first token has an unrestricted null DACL")
	}
	var size struct{ Count, BytesInUse, BytesFree uint32 }
	ok, _, aclErr := getACLInformation.Call(uintptr(dacl), uintptr(unsafe.Pointer(&size)), unsafe.Sizeof(size), aclSizeInformation)
	if ok == 0 {
		return fmt.Errorf("inspect first token DACL: %w", aclErr)
	}
	account, err := syscall.OpenCurrentProcessToken()
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, account.Close()) }()
	user, err := account.GetTokenUser()
	if err != nil {
		return err
	}
	userSID, err := user.User.Sid.String()
	if err != nil {
		return err
	}
	// 2. Core 的受保护 data_dir 只授予运行账号完整控制；子文件沿用这一许可。
	readable := false
	for index := uint32(0); index < size.Count; index++ {
		var entry unsafe.Pointer
		ok, _, aceErr := getACE.Call(uintptr(dacl), uintptr(index), uintptr(unsafe.Pointer(&entry)))
		if ok == 0 {
			return fmt.Errorf("read first token ACE %d: %w", index, aceErr)
		}
		ace := (*struct {
			Type, Flags uint8
			Size        uint16
			Mask        uint32
			SIDStart    uint32
		})(entry)
		if ace.Type != accessAllowedACE || ace.Flags&inheritOnlyACE != 0 {
			return fmt.Errorf("first token ACE %d does not match Core account-only DACL: type %d, flags %#x", index, ace.Type, ace.Flags)
		}
		sid, err := (*syscall.SID)(unsafe.Pointer(&ace.SIDStart)).String()
		if err != nil {
			return err
		}
		if sid != userSID {
			return fmt.Errorf("first token ACE %d grants access to another account: %s", index, sid)
		}
		readable = readable || ace.Mask&(fileReadData|genericRead|genericAll) != 0
	}
	if !readable {
		return errors.New("first token DACL does not grant read access to the running account")
	}
	return nil
}
