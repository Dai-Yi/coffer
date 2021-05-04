package cmd

import (
	"coffer/container"
	"fmt"
)

//删除容器
func rmContainer(containerName string) error {
	//根据容器名获取容器信息对象
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		return fmt.Errorf("get container %s error %v", containerName, err)
	}
	//只删除停止状态的容器
	if containerInfo.Status != container.STOP {
		return fmt.Errorf("couldn't remove running container")
	}
	container.DeleteInfo(containerName)
	container.DeleteWorkSpace(containerInfo.Volume, containerName)
	return nil
}
