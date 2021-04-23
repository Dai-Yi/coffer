package initCNTR

import (
	"coffer/log"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

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
	setMount()
	path, err := exec.LookPath(cmdList[0])
	if err != nil {
		log.Logout("ERROR", "Exec loop path error,"+err.Error())
	}
	if err := syscall.Exec(path, cmdList[0:], os.Environ()); err != nil { //Exec覆盖容器进程
		log.Logout("ERROR", err.Error())
	}
	return nil
}
func setMount() {
	//设置挂载为私有，不影响其他命名空间
	syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, "")
	pwd, err := os.Getwd() // 获取当前工作目录的根路径
	if err != nil {
		log.Logout("ERROR", "Get current location error", err)
		return
	}
	log.Logout("INFO", "Current location:", pwd)
	if err = changeRoot(pwd); err != nil {
		log.Logout("ERROR", "Change root mount error:"+err.Error())
	}
	//挂载proc文件系统
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	//tmpfs是一种基于内存的文件系统,使用RAM或swap分区来储存
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}
func changeRoot(root string) error { //更改根目录
	//为了使老root和新root不在同一个文件系统下，重新mount一次root
	//bind mount是把相同的内容换了一个挂载点的挂载方法
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount rootfs to itself error: %v", err)
	}
	oldRoot := filepath.Join(root, ".pivot_root") //存储old_root
	if err := os.Mkdir(oldRoot, 0777); err != nil {
		return err
	}
	//使当前进程所在mount namespace内的所有进程的root mount移到put_old，然后将new_root作为新的root mount
	if err := syscall.PivotRoot(root, oldRoot); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}
	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}
	oldRoot = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(oldRoot, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(oldRoot)
}
