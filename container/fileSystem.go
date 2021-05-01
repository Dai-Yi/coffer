package container

import (
	"coffer/log"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func NewWorkSpace(containerName string, volume string) error {
	if err := createReadOnlyLayer(); err != nil {
		return fmt.Errorf("create readonly layer error,%v", err)
	}
	if err := createWriteLayer(containerName); err != nil {
		return fmt.Errorf("create write layer error,%v", err)
	}
	if err := createMountPoint(containerName); err != nil {
		return fmt.Errorf("create mount point error,%v", err)
	}
	if volume != "" { //判断是否需要挂载数据卷
		volumeURLs := strings.Split(volume, ":") //根据冒号拆分为宿主机目录和虚拟容器目录
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if err := mountVolume(volumeURLs, containerName); err != nil {
				return fmt.Errorf("mount volume error,%v", err)
			}
			log.Logout("INFO", "New work space volume URLs: ", volumeURLs)
		} else {
			return fmt.Errorf("volume parameter input error")
		}
	}
	return nil
}

//判断文件路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		return !os.IsNotExist(err)
	}
}

//创建只读层
func createReadOnlyLayer() error {
	imageURL := RootURL + "busybox.tar"
	program := RootURL + "busybox/"
	if !PathExists(program) {
		if err := os.MkdirAll(program, 0622); err != nil {
			return fmt.Errorf("mkdir error,%v", err)
		}
		if _, err := exec.Command("tar", "-xvf", imageURL, "-C", program).CombinedOutput(); err != nil {
			return fmt.Errorf("untar dir error,%v", err)
		}
	}
	return nil
}

//创建writeLayer文件夹作为容器可写层
func createWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if !PathExists(writeURL) {
		if err := os.MkdirAll(writeURL, 0777); err != nil {
			return fmt.Errorf("mkdir write layer dir error,%v", err)
		}
	}
	return nil
}

//创建mnt文件夹作为挂载点
func createMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	if !PathExists(mntURL) {
		if err := os.MkdirAll(mntURL, 0777); err != nil {
			return fmt.Errorf("mkdir mountpoint dir error,%v", err.Error())
		}
	}
	tmpWriteLayer := fmt.Sprintf(WriteLayerURL, containerName)
	tmpImageLocation := RootURL + "busybox"
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	//把writeLayer目录和busybox目录mount到mnt
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mount wirte layer and image to mnt error,%v", err)
	}
	return nil
}

//挂载数据卷
func mountVolume(volumeURLs []string, containerName string) error {
	//创建宿主机文件目录
	hostURL := volumeURLs[0]
	if !PathExists(hostURL) {
		if err := os.Mkdir(hostURL, 0777); err != nil {
			return fmt.Errorf("mkdir host dir error,%v", err)
		}
	}
	//在容器文件系统中创建挂载点
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerVolumeURL := mntURL + volumeURLs[1]
	if !PathExists(containerVolumeURL) {
		if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
			return fmt.Errorf("mkdir container volume dir error,%v", err)
		}
	}
	dirs := "dirs=" + hostURL
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run mount command error,%v", err)
	}
	return nil
}

//删除对应数据卷、可写层和挂载点
func DeleteWorkSpace(volume string, contaienrName string) error {
	if volume != "" {
		volumeURLs := strings.Split(volume, ":")
		if len(volumeURLs) == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			if err := deleteVolume(volumeURLs, contaienrName); err != nil {
				return fmt.Errorf("delete volume error,%v", err)
			}
		}
	}
	if err := deleteMountPoint(contaienrName); err != nil {
		return fmt.Errorf("delete mount point error,%v", err)
	}
	if err := deleteWriteLayer(contaienrName); err != nil {
		return fmt.Errorf("delete write layer error,%v", err)
	}
	return nil
}

//卸载容器中数据卷挂载点的文件系统
func deleteVolume(volumeURLs []string, containerName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	containerURL := mntURL + volumeURLs[1]
	cmd := exec.Command("umount", containerURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run umount volume commnad error,%v", err)
	}
	return nil
}

//卸载容器文件系统挂载点
func deleteMountPoint(containerName string) error {
	mntURL := fmt.Sprintf(MntURL, containerName)
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run umount mnt command error,%v", err)
	}
	if PathExists(mntURL) {
		if err := os.RemoveAll(mntURL); err != nil {
			return fmt.Errorf("remove mount point dir error,%v", err)
		}
	}
	return nil
}

//删除可写层
func deleteWriteLayer(containerName string) error {
	writeURL := fmt.Sprintf(WriteLayerURL, containerName)
	if PathExists(writeURL) {
		if err := os.RemoveAll(writeURL); err != nil {
			return fmt.Errorf("remove write layer dir error,%v", err)
		}
	}
	return nil
}
