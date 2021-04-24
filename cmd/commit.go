package cmd

import (
	"coffer/log"
	"os/exec"
)

func commitContainer(imageName string) {
	mntURL := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntURL, ".").CombinedOutput(); err != nil {
		log.Logout("ERROR", "Tar folder error", err.Error())
	}
}
