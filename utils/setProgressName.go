package utils

import (
	"syscall"
	"unsafe"
)

func SetProcessName(name string) error { //设置进程名称
	bytes := append([]byte(name), 0)
	ptr := unsafe.Pointer(&bytes[0])
	if _, _, errno := syscall.RawSyscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(ptr), 0, 0, 0, 0); errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}
