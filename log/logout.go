package log

import (
	"io"
	"log"
	"os"
)

func Logout(level string, information ...interface{}) {
	//打开日志文件
	f, err := os.OpenFile("./cofferLogout.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//延迟关闭
	defer f.Close()
	//设置日志同时输出到log文件和终端
	writers := []io.Writer{f, os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...)                     //省略号是将writers切片打散
	logger := log.New(fileAndStdoutWriter, "["+level+"] ", log.LstdFlags) //创建logger
	switch level {
	case "INFO":
		logger.Println(information...)
	case "ERROR":
		logger.Fatalln(information...) //fatal将直接终止程序
	case "PANIC":
		logger.Panicln(information...) //panic将终止函数并“抛出异常”
	default:
		logger.Println(information...)
	}
}
