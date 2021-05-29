package cmd

import (
	"coffer/cgroups"
	"coffer/container"
	"coffer/net"
	"coffer/subsys"
	"coffer/utils"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const ENV_RUN_SIGN string = "RUN_BACKGROUND"

func pipeSend(msgList interface{}, writePipe *os.File) {
	var msgStr string
	val1, ok := msgList.([]string) //断言是否为[]string类型
	if ok {                        //如果是则转换为string类型
		msgStr = strings.Join(val1, " ")
	} else { //如果不是[]string则为string类型
		val2 := msgList.(string) //断言是否为string类型
		msgStr = val2
	}
	utils.Lock(writePipe)         //加锁
	writePipe.WriteString(msgStr) //命令写入管道
	utils.UnLock(writePipe)       //解锁
	writePipe.Close()             //关闭写入端
}
func duplicateQuery(id string, Name string) (string, error) {
	var newID string
	containers, _ := listContainers()
	if containers == nil { //如果未创建过容器则直接创建
		return id, nil
	}
	for _, item := range containers {
		if item.Name == Name { //名字重复则直接返回错误
			return "", fmt.Errorf("container named %s has existed,please rename", Name)
		}
		if item.Id == id {
			newID, _ = duplicateQuery(idGenerator(), Name) //若id重复则递归生成新id,直到不重复
		} else {
			newID = id
		}
	}
	return newID, nil
}

func run(tty bool, volume string, containerName string, imageName string, network string,
	cmdList []string, environment string, portmapping string, res *subsys.ResourceConfig) error { //run命令
	containerID := idGenerator() //生成10位id
	if containerName == "" {
		containerName = containerID
	}
	containerID, err := duplicateQuery(containerID, containerName) //查询id或name是否重复
	if err != nil {
		return fmt.Errorf("query contianer id and contianer name error->%v", err)
	}
	env := strings.Split(environment, ",") //根据空格拆分为多个用户定义的环境变量
	//创建容器进程和管道
	containerProcess, writePipe, err := container.NewProcess(tty, volume, env, containerName, imageName)
	if err != nil { //容器创建失败
		return fmt.Errorf("create new container error->%v", err)
	}
	if err := containerProcess.Start(); err != nil { //运行容器进程
		return fmt.Errorf("container start error->%v", err)
	}
	port := strings.Split(portmapping, ",")
	if portmapping == "" { //若portmapping为空则port赋为空
		port = nil
	}
	containerInfo := container.ContainerInfo{
		Id:          containerID,
		Pid:         strconv.Itoa(containerProcess.Process.Pid),
		Command:     strings.Join(cmdList, ""),                //容器所执行的指令
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"), //生成创建时间,ps:必须是这个时间
		Status:      container.RUNNING,
		Name:        containerName,
		Volume:      volume,
		PortMapping: port,
	}
	if err := container.StoreInfo(containerInfo); err != nil { //储存容器信息
		return fmt.Errorf("store container information error->%v", err)
	}
	//创建cgroup manager，并通过set和apply设置资源限制
	cgroupManager := cgroups.NewCgroupManager(containerID)
	if err := cgroupManager.Set(res); err != nil { //设置容器限制
		return fmt.Errorf("set cgroup manager error->%v", err)
	}
	//将容器进程加入到各个子系统
	if err := cgroupManager.Apply(containerProcess.Process.Pid); err != nil {
		return fmt.Errorf("apply cgroup manager error->%v", err)
	}
	if network != "" { //配置网络信息
		net.Init()                                                   //初始化网络
		if err := net.Connect(network, &containerInfo); err != nil { //尝试进行网络连接
			return fmt.Errorf("connect network error->%v", err)
		}
	}
	pipeSend(cmdList, writePipe) //传递命令给容器
	containerProcess.Wait()      //后台进程等待容器内进程结束
	if !tty {                    //若非前台运行方式
		utils.Logout("INFO", "Container background running")
		return nil
	}
	defer cgroupManager.Destroy() //运行完后销毁cgroup manager
	container.DeleteInfo(containerName)
	container.DeleteWorkSpace(volume, containerName)
	utils.Logout("INFO", "Container closed")
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
func transform() error { //转换为后台运行
	temp := []string{"coffer"}
	os.Args = append(temp, os.Args...)
	cmd := exec.Command("/proc/self/exe", os.Args...) //调用自身来创建子进程,参数不变
	cmd.Args = os.Args
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=background", ENV_RUN_SIGN)) //添加用于判断的环境变量
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start background process error->%v", err)
	}
	return nil
}
