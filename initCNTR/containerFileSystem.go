package initCNTR

import (
	"coffer/log"
	"os"
	"os/exec"
	"strings"
)

func NewWorkSpace(rootURL string, mntURL string, volume string) {
	createReadOnlyLayer(rootURL)
	createWriteLayer(rootURL)
	createMountPoint(rootURL, mntURL)
	if volume != "" { //判断是否需要挂载数据卷
		volumeURLs := strings.Split(volume, ":") //根据冒号拆分为宿主机目录和虚拟容器目录
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			mountVolume(rootURL, mntURL, volumeURLs)
			log.Logout("INFO", "New work space volume URLs: ", volumeURLs)
		} else {
			log.Logout("ERROR", "Volume parameter input error")
		}
	}
}

//挂载数据卷
func mountVolume(rootURL string, mntURL string, volumeURLs []string) {
	//创建宿主机文件目录
	hostURL := volumeURLs[0]
	if !pathExists(hostURL) {
		if err := os.Mkdir(hostURL, 0777); err != nil {
			log.Logout("ERROR", "Mkdir host dir error", err.Error())
		}
	}
	//在容器文件系统中创建挂载点
	containerVolumeURL := mntURL + volumeURLs[1]
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Logout("ERROR", "Mkdir container dir error", err.Error())
	}
	dirs := "dirs=" + hostURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Logout("ERROR", "Mount volume failed", err.Error())
	}
}

//判断文件路径是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		return !os.IsNotExist(err)
	}
}

//创建只读层
func createReadOnlyLayer(rootURL string) {
	program := rootURL + "busybox/"
	imageURL := rootURL + "busybox.tar"
	if !pathExists(program) {
		if err := os.MkdirAll(program, 0622); err != nil {
			log.Logout("ERROR", "Mkdir error ", err)
		}
		if _, err := exec.Command("tar", "-xvf", imageURL, "-C", program).CombinedOutput(); err != nil {
			log.Logout("ERROR", "Untar dir error", err.Error())
		}
	}
}

//创建writeLayer文件夹作为容器可写层
func createWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Logout("ERROR", "Mkdir write layer dir error", err.Error())
	}
}

//创建mnt文件夹作为挂载点
func createMountPoint(rootURL string, mntURL string) {
	if err := os.MkdirAll(mntURL, 0777); err != nil {
		log.Logout("ERROR", "Mkdir mountpoint dir error", err.Error())
	}
	//把writeLayer目录和busybox目录mount到mnt
	dirs := "dirs=" + rootURL + "writeLayer:" + rootURL + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Logout("ERROR", err.Error())
	}
}

//删除对应数据卷、可写层和挂载点
func DeleteWorkSpace(rootURL string, mntURL string, volume string) {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			deleteVolume(mntURL, volumeURLs)
		}
	}
	deleteMountPoint(rootURL, mntURL)
	deleteWriteLayer(rootURL)
}

//卸载容器中数据卷挂载点的文件系统
func deleteVolume(mntURL string, volumeURLs []string) {
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Logout("ERROR", "Umount volume failed", err.Error())
	}
}

//卸载容器文件系统挂载点
func deleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Logout("ERROR", "Umount mount point failed", err.Error())
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Logout("ERROR", "Remove mountpoint dir error", err.Error())
	}
}

//删除可写层
func deleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Logout("ERROR", "Remove writeLayer dir error", err.Error())
	}
}
