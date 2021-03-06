package cmd

import (
	"coffer/container"
	"coffer/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"syscall"
)

func stopContainer(containerName string) error {
	//根据容器名获取对应的容器进程PID
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		return fmt.Errorf("get contaienr pid by name %s error->%v", containerName, err)
	}
	//将string类型的PID转换为int类型
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		return fmt.Errorf("conver pid from string to int error->%v", err)
	}
	//发送kill信号给容器进程
	syscall.Kill(pidInt, syscall.SIGTERM)
	//根据容器名获取对应的容器信息对象
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("get container %s info error->%v", containerName, err)
	}
	//更改容器状态为STOP,设置PID为空
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	//修改后的信息重新序列化为json字符串
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		return fmt.Errorf("json marshal %s error->%v", containerName, err)
	}
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigFile
	//写入修改后的数据覆盖容器原来的容器信息
	configFile, err := os.OpenFile(configFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file %s error->%v", configFile.Name(), err)
	}
	defer configFile.Close()
	utils.Lock(configFile)
	_, err = configFile.Write(newContentBytes)
	utils.UnLock(configFile)
	if err != nil {
		return fmt.Errorf("write file %s error->%v", configFile.Name(), err)
	}
	return nil
}

//根据容器名获取对应的容器信息对象
func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	//存放容器信息的路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigFile
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("read config file error->%v", err)
	}
	var containerInfo container.ContainerInfo
	//将容器信息字符串反序列化为容器信息对象
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, fmt.Errorf("json unmarshal %s error->%v", containerName, err)
	}
	return &containerInfo, nil
}
