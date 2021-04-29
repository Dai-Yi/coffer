package cmd

import (
	"coffer/cgroups"
	"coffer/container"
	"coffer/subsys"
	"fmt"
	"os"
	"strings"
)

func sendCommand(cmdList []string, writePipe *os.File) {
	command := strings.Join(cmdList, " ")
	writePipe.WriteString(command) //命令写入管道
	writePipe.Close()              //关闭写入端
}
func run(tty bool, background bool, volume string, name string, cmdList []string, res *subsys.ResourceConfig) error { //run命令
	containerProcess, writePipe := container.NewProcess(tty, volume) //首先创建容器进程和管道
	if containerProcess == nil {                                     //容器创建失败
		return fmt.Errorf("create new container error")
	}
	if err := containerProcess.Start(); err != nil { //运行容器进程
		return fmt.Errorf("container start error,%v", err)
	}
	container.Monitor(volume)
	containerID, err := container.GenerateInfo(containerProcess.Process.Pid, cmdList, name, volume)
	if err != nil {
		return fmt.Errorf("generate container information error,%v", err)
	}
	//创建cgroup manager，并通过set和apply设置资源限制
	cgroupManager := cgroups.CgroupManager{CgroupPath: "cofferCgroup"}
	if err := cgroupManager.Set(res); err != nil { //设置容器限制
		container.GracefulExit()
		return err
	}
	//将容器进程加入到各个子系统
	if err := cgroupManager.Apply(containerProcess.Process.Pid); err != nil {
		container.GracefulExit()
		return err
	}
	sendCommand(cmdList, writePipe) //传递命令给容器
	if tty {
		containerProcess.Wait() //容器进程等待容器内进程结束
		container.DeleteContainerInfo(containerID)
		defer container.GracefulExit()
		cgroupManager.Destroy() //运行完后销毁cgroup manager
	}
	return nil
}
