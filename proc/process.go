package proc

import (
	"coffer/log"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Run(tty bool, cmdList []string) { //run命令, res *subsys.ResourceConfig
	log.Logout("DEBUG", "Creating Container Process")
	newContainer, writePipe := createContainerProcess(tty) //首先创建容器进程
	if newContainer == nil {
		log.Logout("ERROR", "Create new container error")
		return
	}
	if err := newContainer.Start(); err != nil { //运行容器进程
		log.Logout("ERROR", err.Error())
	}
	sendInitCommand(cmdList, writePipe) //初始化容器
	newContainer.Wait()
}
func sendInitCommand(cmdList []string, writePipe *os.File) {
	command := strings.Join(cmdList, " ")
	log.Logout("INFO", "command is "+command)
	writePipe.WriteString(command)
	writePipe.Close()
}
func createContainerProcess(tty bool) (*exec.Cmd, *os.File) { //创建容器进程
	readPipe, writePipe, err := os.Pipe()
	if err != nil {
		log.Logout("ERROR", "New pipe error "+err.Error())
		return nil, nil
	}
	cmd := exec.Command("/proc/self/exe", "INiTcoNtaInER")
	cmd.SysProcAttr = &syscall.SysProcAttr{ //使用namespace隔离
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
	cmd.ExtraFiles = []*os.File{readPipe} //传入管道文件读取端
	return cmd, writePipe
}
func readCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe") //从文件描述符获取管道
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Logout("ERROR", "init read pipe error "+err.Error())
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}
func InitializeContainer() error { //容器内部初始化
	cmdList := readCommand()
	if len(cmdList) == 0 {
		log.Logout("ERROR", "Run container get user command error,command list is empty")
		return nil
	}
	log.Logout("DEBUG", "Initializing Container")
	//设置挂载为私有，不影响其他命名空间
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	path, err := exec.LookPath(cmdList[0])
	if err != nil {
		log.Logout("ERROR", "Exec loop path error"+err.Error())
	}
	log.Logout("INFO", "Find path"+path)
	if err := syscall.Exec(path, cmdList[0:], os.Environ()); err != nil { //Exec覆盖容器进程
		log.Logout("ERROR", err.Error())
	}
	return nil
}
