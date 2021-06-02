package cmd

import (
	"coffer/container"
	"coffer/utils"
	"fmt"
	"os/exec"
)

func commitContainer(containerName string, imageName string) error {
	mntURL := fmt.Sprintf(container.MntURL, containerName)
	if !utils.PathExists(containerName) {
		return fmt.Errorf("container named %s not exist", containerName)
	}
	imageTar := container.RootURL + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		return fmt.Errorf("tar folder error->%v", err)
	}
	return nil
}
