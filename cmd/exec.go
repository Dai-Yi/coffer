package cmd

import (
	"coffer/container"
	"coffer/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

//控制c代码中setns系统调用
const ENV_EXEC_PID = "cofferPID"
const ENV_EXEC_CMD = "cofferCMD"

func execContainer(containerName string, comArray []string) error {
	//根据传递过来的容器名称获取容器进程对应的PID
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		return fmt.Errorf("exec container getContainerPidByName %s error %v", containerName, err)
	}
	//以空格为分隔符拼接成一个字符串
	cmdStr := strings.Join(comArray, " ")
	log.Logout("INFO", "container pid", pid)
	log.Logout("INFO", "command", cmdStr)
	//调用自己,创建出一个子进程
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//使用已经启用的环境变量和命令
	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("exec container %s error %v", containerName, err)
	}
	return nil
}

//根据容器名获取对应容器PID
func getContainerPidByName(containerName string) (string, error) {
	//先拼接出存储容器信息的路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirURL + container.ConfigFile
	//读取该路径下的文件内容
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo container.ContainerInfo
	//将文件内容反序列化为容器信息对象,返回对应PID
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}
