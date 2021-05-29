package container

import (
	"coffer/utils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func NewProcess(tty bool, volume string, environment []string, containerName string, imageName string) (*exec.Cmd, *os.File, error) { //创建容器进程
	readPipe, writePipe, err := os.Pipe() //创建管道用于传递命令给容器
	if err != nil {                       //管道创建失败
		return nil, nil, fmt.Errorf("new pipe error->%v", err)
	}
	cmd := exec.Command("/proc/self/exe", "INiTcoNtaInER") //调用自身，创建容器进程
	cmd.SysProcAttr = &syscall.SysProcAttr{                //使用namespace隔离
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET |
			syscall.CLONE_NEWIPC,
	}
	if tty { //如果要交互，显示容器运行信息
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} //运行过程中产生的日志输出到log文件
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName)
	if !utils.PathExists(dirURL) {
		if err := os.MkdirAll(dirURL, 0644); err != nil {
			return nil, nil, fmt.Errorf("container process mkdir error->%v", err)
		}
	}
	stdLogFilePath := dirURL + ContainerLogFile
	stdLogFile, err := os.Create(stdLogFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("container process create log file error->%v", err)
	}
	cmd.Stdout = stdLogFile
	cmd.Stderr = stdLogFile
	cmd.ExtraFiles = []*os.File{readPipe}          //附加管道文件读取端，使容器能够读取管道传入的命令
	cmd.Env = append(os.Environ(), environment...) //将环境变量添加上用户自定义环境变量
	cmd.Dir = fmt.Sprintf(MntURL, containerName)
	if err := NewWorkSpace(containerName, imageName, volume); err != nil {
		return nil, nil, fmt.Errorf("create new work space error->%v", err)
	}
	return cmd, writePipe, nil
}

func PipeReceive() (string, error) {
	pipe := os.NewFile(uintptr(3), "pipe") //从文件描述符获取管道
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		return "", fmt.Errorf("init read pipe error->%v", err)
	}
	return string(msg), nil
}
func InitializeContainer() error { //容器内部初始化
	tempList, err := PipeReceive() //从管道读取到命令
	if err != nil {
		return fmt.Errorf("receive command error->%v", err)
	}
	if len(tempList) == 0 {
		return fmt.Errorf("run container get user command error->command list is empty")
	}
	cmdList := strings.Split(tempList, " ")
	if err := setMount(); err != nil {
		return fmt.Errorf("set mount error->%v", err)
	}
	path, err := exec.LookPath(cmdList[0])
	if err != nil {
		return fmt.Errorf("exec look path error->%v", err)
	}
	if err := syscall.Exec(path, cmdList[0:], os.Environ()); err != nil { //Exec使程序进程覆盖容器进程
		return fmt.Errorf("exec command error->%v,", err)
	}
	return nil
}
func setMount() error {
	//设置挂载为私有，不影响其他命名空间
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	pwd, err := os.Getwd() // 获取当前工作目录的根路径
	if err != nil {
		return fmt.Errorf("get current location error->%v", err)
	}
	utils.Logout("INFO", "Current location:", pwd)
	if err = changeRoot(pwd); err != nil {
		return fmt.Errorf("change root mount error->%v", err)
	}
	//挂载proc文件系统
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	//tmpfs是一种基于内存的文件系统,使用RAM或swap分区来储存
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
	return nil
}
func changeRoot(root string) error { //更改根目录
	//为了使老root和新root不在同一个文件系统下，重新mount一次root
	//bind mount是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error->%v", err)
	}
	oldRoot := filepath.Join(root, ".pivot_root") //存储old_root
	if err := os.Mkdir(oldRoot, 0777); err != nil {
		return fmt.Errorf("make old_root file error->%v", err)
	}
	//使当前进程所在mount namespace内的所有进程的root mount移到put_old，然后将new_root作为新的root mount
	if err := syscall.PivotRoot(root, oldRoot); err != nil {
		return fmt.Errorf("pivot_root error->%v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / error->%v", err)
	}
	oldRoot = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(oldRoot, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir error->%v", err)
	}
	// 删除临时文件夹
	return os.Remove(oldRoot)
}
