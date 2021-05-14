package utils

import "os"

//判断文件路径是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	} else {
		return !os.IsNotExist(err)
	}
}
