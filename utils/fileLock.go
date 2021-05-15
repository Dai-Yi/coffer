package utils

import (
	"fmt"
	"os"
	"syscall"
)

//文件锁
//类型:LOCK_SH共享锁,LOCK_EX排他锁
//模式:LOCK_NB非阻塞模式,默认阻塞模式
func Lock(file *os.File) error {
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("lock file %v failed->%v", file.Name(), err)
	}
	return nil
}

// LOCK_UN解锁
func UnLock(file *os.File) error {
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
		return fmt.Errorf("unlock file %v failed->%v", file.Name(), err)
	}
	return nil
}
