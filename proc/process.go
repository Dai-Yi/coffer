package proc

import (
	"coffer/cgroups"
	"coffer/initCNTR"
	"coffer/log"
	"coffer/subsys"
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

func Run(tty bool, volume string, cmdList []string, res *subsys.ResourceConfig) { //run命令
	newContainer, writePipe := createContainerProcess(tty, volume) //首先创建容器进程和管道
	if newContainer == nil {                                       //容器创建失败
		log.Logout("ERROR", "Create new container error")
		return
	}
	if err := newContainer.Start(); err != nil { //运行容器进程
		log.Logout("ERROR", err.Error())
	}
	//创建cgroup manager，并通过set和apply设置资源限制
	cgroupManager := cgroups.CgroupManager{CgroupPath: "cofferCgroup"}
	defer cgroupManager.Destroy()                 //函数执行完后销毁cgroup manager
	cgroupManager.Set(res)                        //设置容器限制
	cgroupManager.Apply(newContainer.Process.Pid) //将容器进程加入到各个子系统
	sendCommand(cmdList, writePipe)               //传递命令给容器
	newContainer.Wait()
	log.Logout("INFO", "Container closed")
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	defer initCNTR.DeleteWorkSpace(rootURL, mntURL, volume)
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
	}
	if tty { //如果需要，显示容器运行信息
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	cmd.ExtraFiles = []*os.File{readPipe} //附加管道文件读取端，使容器能够读取管道传入的命令
	mntURL := "/root/mnt/"
	rootURL := "/root/"
	initCNTR.NewWorkSpace(rootURL, mntURL, volume)
	cmd.Dir = mntURL
	return cmd, writePipe
}
