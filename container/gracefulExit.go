package container

import (
	"coffer/log"
	"os"
	"os/signal"
	"syscall"
)

func GracefulExit() { //优雅退出
	pgid, err := syscall.Getpgid(syscall.Getpid()) //获取进程组id
	if err != nil {                                //无法获取进程组id
		log.Logout("ERROR", "Abnormal Exit", err.Error()) //异常退出
		os.Exit(1)
	}
	//负数表示发送进程组信号
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil { //主动发送退出信号
		log.Logout("ERROR", "Kill process error", err.Error())
	}
}
func Monitor(volume string) { //优雅退出
	c := make(chan os.Signal) //信号通道
	// 监听退出信号
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() { //匿名goroutine并发函数
		for s := range c {
			switch s {
			case syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT:
				DeleteWorkSpace("/root/", "/root/mnt/", volume)
				log.Logout("INFO", "Container closed")
				os.Exit(0)
			default:
			}
		}
	}()
}
