package proc

import (
	"coffer/log"
	"os"
	"os/exec"
	"syscall"
)

func Run(tty bool, command string) { //run命令
	newContainer := createContainerProcess(tty, command) //首先创建容器进程
	if err := newContainer.Start(); err != nil {         //运行容器进程
		log.Logout("ERROR", err.Error())
	}
	newContainer.Wait()
	os.Exit(-1)
}
func createContainerProcess(tty bool, command string) *exec.Cmd { //创建容器进程
	args := []string{"init", command}
	cmd := exec.Command("/proc/self/exe", args...) //调用自身创建子进程(容器进程),同时传入init命令
	cmd.SysProcAttr = &syscall.SysProcAttr{        //使用namespace隔离
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}
	if tty { //如果需要，显示容器运行信息
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}

func InitializeContainer(command string) error { //容器内部初始化
	log.Logout("INFO", command)
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	argv := []string{command}
	if err := syscall.Exec(command, argv, os.Environ()); err != nil {
		log.Logout("ERROR", err.Error())
	}
	return nil
}
