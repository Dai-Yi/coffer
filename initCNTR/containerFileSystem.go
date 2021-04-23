package initCNTR

import (
	"coffer/log"
	"os"
	"os/exec"
)

func NewWorkSpace(rootURL string, mntURL string) {
	createReadOnlyLayer(rootURL)
	createWriteLayer(rootURL)
	createMountPoint(rootURL, mntURL)
}
func pathExists(path string) (bool, error) { //判断文件路径是否存在
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func createReadOnlyLayer(rootURL string) { //创建只读层
	program := rootURL + "busybox/"
	imageUrl := rootURL + "busybox.tar"
	exist, err := pathExists(program)
	if err != nil {
		log.Logout("ERROR", "Fail to judge whether dir ", program, " exists", err)
	}
	if !exist {
		if err := os.MkdirAll(program, 0622); err != nil {
			log.Logout("ERROR", "Mkdir ", program, " error", err)
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", program).CombinedOutput(); err != nil {
			log.Logout("ERROR", "Untar dir ", program, " error", err)
		}
	}
}

//创建writeLayer文件夹作为容器可写层
func createWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Logout("ERROR", "Mkdir write layer dir ", writeURL, " error", err)
	}
}

//创建mnt文件夹作为挂载点
func createMountPoint(rootURL string, mntURL string) {
	if err := os.MkdirAll(mntURL, 0777); err != nil {
		log.Logout("ERROR", "Mkdir mountpoint dir ", mntURL, " error", err)
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

//删除容器时删除对应可写层和挂载点
func DeleteWorkSpace(rootURL string, mntURL string) {
	deleteMountPoint(rootURL, mntURL)
	deleteWriteLayer(rootURL)
}
func deleteMountPoint(rootURL string, mntURL string) {
	cmd := exec.Command("umount", mntURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Logout("ERROR", err.Error())
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Logout("ERROR", "Remove mountpoint dir ", mntURL, " error", err)
	}
}
func deleteWriteLayer(rootURL string) {
	writeURL := rootURL + "writeLayer/"
	if err := os.RemoveAll(writeURL); err != nil {
		log.Logout("ERROR", "Remove writeLayer dir ", writeURL, " error", err)
	}
}
