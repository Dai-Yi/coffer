package proc

import (
	"coffer/cgroups"
	"coffer/log"
	"coffer/subsys"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Run(tty bool, cmdList []string, res *subsys.ResourceConfig) { //run命令
	newContainer, writePipe := createContainerProcess(tty) //首先创建容器进程和管道
	if newContainer == nil {                               //容器创建失败
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
}
func sendCommand(cmdList []string, writePipe *os.File) {
	command := strings.Join(cmdList, " ")
	writePipe.WriteString(command) //命令写入管道
	writePipe.Close()              //关闭写入端
}
func receiveCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe") //从文件描述符获取管道
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Logout("ERROR", "Init read pipe error "+err.Error())
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
func InitializeContainer() error { //容器内部初始化
	cmdList := receiveCommand() //从管道读取到命令
	if len(cmdList) == 0 {
		log.Logout("ERROR", "Run container get user command error,command list is empty")
		return nil
	}
	//设置挂载为私有，不影响其他命名空间
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	path, err := exec.LookPath(cmdList[0])
	if err != nil {
		log.Logout("ERROR", "Exec loop path error,"+err.Error())
	}
	if err := syscall.Exec(path, cmdList[0:], os.Environ()); err != nil { //Exec覆盖容器进程
		log.Logout("ERROR", err.Error())
	}
	return nil
}
func createContainerProcess(tty bool) (*exec.Cmd, *os.File) { //创建容器进程
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
	cmd.ExtraFiles = []*os.File{readPipe} //附加管道文件读取端，使容器能够读取管道传入的命令
	if tty {                              //如果需要，显示容器运行信息
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd, writePipe
}
