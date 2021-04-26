package cmd

import (
	"coffer/cgroups"
	"coffer/cntr"
	"coffer/log"
	"coffer/subsys"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func sendCommand(cmdList []string, writePipe *os.File) {
	command := strings.Join(cmdList, " ")
	writePipe.WriteString(command) //命令写入管道
	writePipe.Close()              //关闭写入端
}
func run(tty bool, background bool, volume string, cmdList []string, res *subsys.ResourceConfig) error { //run命令
	newContainer, writePipe := createContainerProcess(tty, volume) //首先创建容器进程和管道
	if newContainer == nil {                                       //容器创建失败
		return fmt.Errorf("create new container error")
	}
	if err := newContainer.Start(); err != nil { //运行容器进程
		return fmt.Errorf("container start error,%v", err)
	}
	cntr.Monitor(volume)
	//创建cgroup manager，并通过set和apply设置资源限制
	cgroupManager := cgroups.CgroupManager{CgroupPath: "cofferCgroup"}

	if err := cgroupManager.Set(res); err != nil { //设置容器限制
		cntr.GracefulExit()
		return err
	}
	//将容器进程加入到各个子系统
	if err := cgroupManager.Apply(newContainer.Process.Pid); err != nil {
		cntr.GracefulExit()
		return err
	}
	sendCommand(cmdList, writePipe) //传递命令给容器
	if tty {
		newContainer.Wait() //容器进程等待容器内进程结束
		defer cntr.GracefulExit()
	}
	cgroupManager.Destroy() //运行完后销毁cgroup manager
	return nil
}
func createContainerProcess(tty bool, volume string) (*exec.Cmd, *os.File) { //创建容器进程
	readPipe, writePipe, err := os.Pipe() //创建管道用于传递命令给容器
	if err != nil {                       //管道创建失败
		log.Logout("ERROR", "New pipe error "+err.Error())
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "INiTcoNtaInER") //调用自身，创建容器进程
	cmd.SysProcAttr = &syscall.SysProcAttr{                //使用namespace隔离
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
		// Setpgid: true,//开启之后可以kill组进程，但有bug，bash无法使用
	}
	if tty { //如果需要，显示容器运行信息
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe} //附加管道文件读取端，使容器能够读取管道传入的命令
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	cntr.NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, writePipe
}
