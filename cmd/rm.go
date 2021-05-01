package cmd

import (
	"coffer/container"
	"fmt"
	"os"
)

//删除容器
func rmContainer(containerName string) error {
	//根据容器名获取容器信息对象
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("get container %s info error %v", containerName, err)
	}
	//只删除停止状态的容器
	if containerInfo.Status != container.STOP {
		return fmt.Errorf("couldn't remove running container")
	}
	//找到对应存储容器信息的文件路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	//移除所有信息
	if err := os.RemoveAll(dirURL); err != nil {
		return fmt.Errorf("remove file %s error %v", dirURL, err)
	}
	container.DeleteWorkSpace(container.RootURL, container.MntURL, containerInfo.Volume)
	return nil
}
