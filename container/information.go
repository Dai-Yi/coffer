package container

import (
	"coffer/log"
	"encoding/json"
	"fmt"
	"os"
)

const (
	ENV_RUN             string = "RUN_BACKGROUND"
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/coffer/%s/"
	ConfigFile          string = "containerConfig.json"
	ContainerLogFile    string = "container.log"
	RootURL             string = "/root/"
	MntURL              string = "/root/mnt/%s/"
	WriteLayerURL       string = "/root/writeLayer/%s/"
)

type ContainerInfo struct {
	Pid         string   `json:"pid"`         //容器的init进程在宿主机上的 PID
	Id          string   `json:"id"`          //容器Id
	Name        string   `json:"name"`        //容器名
	Command     string   `json:"command"`     //容器内init运行命令
	CreatedTime string   `json:"createTime"`  //创建时间
	Status      string   `json:"status"`      //容器的状态
	Volume      string   `json:"volume"`      //容器的数据卷
	PortMapping []string `json:"portmapping"` //端口映射
}

//储存容器信息
func StoreInfo(c ContainerInfo) error {
	//将容器信息化为字符串
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("record container info error->%v", err)
	}
	jsonString := string(jsonBytes)
	dirURL := fmt.Sprintf(DefaultInfoLocation, c.Name) //string拼接成路径
	if !PathExists(dirURL) {                           //如果路径不存在则创建
		if err := os.MkdirAll(dirURL, 0622); err != nil {
			return fmt.Errorf("mkdir error->%v", err)
		}
	}
	//创建json文件
	jsonfile := dirURL + ConfigFile
	file, err := os.Create(jsonfile)
	if err != nil {
		return fmt.Errorf("create json file error->%v", err)
	}
	defer file.Close()
	//将json化之后的数据写入文件
	if _, err := file.WriteString(jsonString); err != nil {
		return fmt.Errorf("write file error->%v", err)
	}
	return nil
}

//删除容器信息
func DeleteInfo(containerName string) {
	dirURL := fmt.Sprintf(DefaultInfoLocation, containerName) //拼接路径
	if PathExists(dirURL) {                                   //如果没有就不用删了
		if err := os.RemoveAll(dirURL); err != nil {
			log.Logout("ERROR", "Remove dir error->", err)
		}
	}
}
