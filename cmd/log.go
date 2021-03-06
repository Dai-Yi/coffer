package cmd

import (
	"coffer/container"
	"fmt"
	"io/ioutil"
	"os"
)

func logContainer(containerName string) error {
	dirURL := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFileLocation := dirURL + container.ContainerLogFile
	file, err := os.Open(logFileLocation) //os.Open用于读
	if err != nil {
		return fmt.Errorf("log container open file %s error->%v", logFileLocation, err)
	}
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("log container read file %s error->%v", logFileLocation, err)
	}
	fmt.Fprint(os.Stdout, string(content))
	return nil
}
