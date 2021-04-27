package container

import (
	"coffer/log"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/coffer/%s/"
	ConfigName          string = "containerConfig.json"
	//ContainerLogFile    string = "container.log"
	RootURL       string = "/root"
	MntURL        string = "/root/mnt/%s"
	WriteLayerURL string = "/root/writeLayer/%s"
)

type containerInfo struct {
	Pid         string `json:"pid"`        //容器的init进程在宿主机上的 PID
	Id          string `json:"id"`         //容器Id
	Name        string `json:"name"`       //容器名
	Command     string `json:"command"`    //容器内init运行命令
	CreatedTime string `json:"createTime"` //创建时间
	Status      string `json:"status"`     //容器的状态
	Volume      string `json:"volume"`     //容器的数据卷
	//PortMapping []string `json:"portmapping"` //端口映射
}

func idGenerator() string { //ID生成器
	rand.Seed(time.Now().UnixNano()) //以纳秒时间戳为种子
	id := make([]byte, 10)           //十位ID
	for i := range id {
		id[i] = byte(rand.Intn(10)) //产生0-9的伪随机数
	}
	return string(id)
}

//生成容器信息
func GenerateInfo(containerPID int, commandArray []string,
	containerName string, volume string) (string, error) {
	id := idGenerator() //生成10位id
	if containerName == "" {
		containerName = id
	}
	containerInformation := &containerInfo{
		Id:          id,
		Pid:         strconv.Itoa(containerPID),
		Command:     strings.Join(commandArray, ""),           //容器所执行的指令
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"), //生成创建时间,ps:必须是这个时间
		Status:      RUNNING,
		Name:        containerName,
		Volume:      volume,
	}
	//将容器信息化为字符串
	jsonBytes, err := json.Marshal(containerInformation)
	if err != nil {
		return "", fmt.Errorf("record container info error,%v", err)
	}
	jsonString := string(jsonBytes)
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName) //string拼接成路径
	if !pathExists(dirURL) {                                  //如果路径不存在则创建
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			return "", fmt.Errorf("mkdir error,%v", err)
		}
	}
	//创建json文件
	jsonfile := dirURL + "/" + ConfigName
	file, err := os.Create(jsonfile)
	if err != nil {
		return "", fmt.Errorf("create json file error,%v", err)
	}
	defer file.Close()
	//将json化之后的数据写入文件
	if _, err := file.WriteString(jsonString); err != nil {
		return "", fmt.Errorf("write file error,%v", err)
	}
	return id, nil
}

//删除容器信息
func DeleteContainerInfo(containerId string) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerId) //拼接路径
	if err := os.RemoveAll(dirURL); err != nil {
		log.Logout("ERROR", "Remove dir error ", err)
	}
}
