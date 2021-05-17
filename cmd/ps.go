package cmd

import (
	"coffer/container"
	"coffer/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"text/tabwriter"
)

func printContainers() error {
	containers, err := listContainers()
	if err != nil {
		return fmt.Errorf("list containers error->%v", err)
	}
	//打印控制台信息
	writer := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	//输出控制台信息列表
	fmt.Fprint(writer, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(writer, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	//刷新标准输出流缓冲区,打印容器列表
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("flush error->%v", err)
	}
	return nil
}
func listContainers() ([]*container.ContainerInfo, error) {
	var containers []*container.ContainerInfo
	//获取容器信息文件路径
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirURL = dirURL[:len(dirURL)-1]
	//读取该目录下所有文件
	if !utils.PathExists(dirURL) {
		return nil, fmt.Errorf("no container created")
	}
	files, err := ioutil.ReadDir(dirURL)
	if err != nil {
		return nil, fmt.Errorf("read dir error->%v", err)
	}
	//遍历文件
	for _, file := range files {
		if file.Name() == "network" { //忽略network文件
			continue
		}
		//根据容器配置文件获取对应信息,转换为容器信息对象
		tmpContainer, err := getContainerInfo(file)
		if err != nil { //有读取不出来的就跳过
			utils.Logout("ERROR", "Get container info error->", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	return containers, nil
}

func getContainerInfo(file os.FileInfo) (*container.ContainerInfo, error) {
	//获取文件名
	containerName := file.Name()
	//根据文件名生成文件绝对路径
	configFileDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFileDir = configFileDir + container.ConfigFile
	//读取json文件内的容器信息
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		return nil, fmt.Errorf("read file error->%v", err)
	}
	var containerInfo container.ContainerInfo
	//容器信息反序列化为容器信息对象
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		return nil, fmt.Errorf("json unmarshal error->%v", err)
	}
	return &containerInfo, nil
}
