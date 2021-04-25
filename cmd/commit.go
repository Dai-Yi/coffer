package cmd

import (
	"fmt"
	"os/exec"
)

func commitContainer(imageName string) error {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		return fmt.Errorf("tar folder error,%v", err)
	}
	return nil
}
