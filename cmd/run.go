package cmd

import (
	"coffer/cgroups"
	"coffer/container"
	"coffer/log"
	"coffer/subsys"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func sendCommand(cmdList []string, writePipe *os.File) {
	command := strings.Join(cmdList, " ")
	writePipe.WriteString(command) //命令写入管道
	writePipe.Close()              //关闭写入端
}
func run(tty bool, volume string, containerName string, imageName string,
	cmdList []string, environment string, res *subsys.ResourceConfig) error { //run命令
	id := idGenerator() //生成10位id
	if containerName == "" {
		containerName = id
	}
	env := strings.Split(environment, ",") //根据空格拆分为多个用户定义的环境变量
	//创建容器进程和管道
	containerProcess, writePipe := container.NewProcess(tty, volume, env, containerName, imageName)
	if containerProcess == nil { //容器创建失败
		return fmt.Errorf("create new container error")
	}
	if err := containerProcess.Start(); err != nil { //运行容器进程
		return fmt.Errorf("container start error,%v", err)
	}
	// container.Monitor(volume)
	c := container.ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerProcess.Process.Pid),
		Command:     strings.Join(cmdList, ""),                //容器所执行的指令
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"), //生成创建时间,ps:必须是这个时间
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
	}
	if err := container.StoreInfo(c); err != nil {
		return fmt.Errorf("generate container information error,%v", err)
	}
	//创建cgroup manager，并通过set和apply设置资源限制
	cgroupManager := cgroups.CgroupManager{CgroupPath: "cofferCgroup"}
	defer cgroupManager.Destroy()                  //运行完后销毁cgroup manager
	if err := cgroupManager.Set(res); err != nil { //设置容器限制
		//container.GracefulExit()
		return err
	}
	//将容器进程加入到各个子系统
	if err := cgroupManager.Apply(containerProcess.Process.Pid); err != nil {
		//container.GracefulExit()
		return err
	}
	sendCommand(cmdList, writePipe) //传递命令给容器
	if tty {
		containerProcess.Wait() //容器进程等待容器内进程结束
		container.DeleteInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
		log.Logout("INFO", "Container closed")
		os.Exit(0)
		// defer container.GracefulExit()
	}
	return nil
}

func idGenerator() string { //ID生成器
	rand.Seed(time.Now().UnixNano()) //以纳秒时间戳为种子
	id := make([]byte, 10)           //十位ID
	for i := range id {
		id[i] = byte(rand.Intn(10) + 48) //产生0-9的伪随机数
	}
	temp := string(id)
	return temp
}
