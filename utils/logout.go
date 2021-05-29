package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

const (
//cofferLogPath     string = "/var/run/coffer/"
//cofferLogFileName string = "cofferLogout.log"
)

func Logout(level string, information ...interface{}) {
	// if !PathExists(cofferLogPath) {
	// 	if err := os.MkdirAll(cofferLogPath, 0644); err != nil {
	// 		log.Fatal("container process mkdir error->", err.Error())
	// 	}
	// }
	// cofferLogFilePath := cofferLogPath + cofferLogFileName
	// cofferLogFile, err := os.Create(cofferLogFilePath)
	// if err != nil {
	// 	log.Fatal("container process create log file error->", err.Error())
	// }
	// //延迟关闭
	// defer cofferLogFile.Close()

	//设置日志同时输出到log文件和终端
	//writers := []io.Writer{cofferLogFile,os.Stdout}
	//fileAndStdoutWriter := io.MultiWriter(writers...) //省略号是将writers切片打散
	file, line := printCallerPosition()
	fileline := fmt.Sprintf("%s:%d", file, line)
	//创建自定义logger
	logger := log.New(os.Stdout, "["+level+"] "+fileline+" ", log.LstdFlags)
	//Lock(cofferLogFile) //加锁
	switch level {
	case "PANIC":
		logger.Panicln(information...) //panic将终止函数并“抛出异常”
	default:
		logger.Println(information...)
	}
	//UnLock(cofferLogFile) //解锁
}
func printCallerPosition() (string, int) {
	pc, _, _, _ := runtime.Caller(2) //调用堆栈的第2层函数
	file := runtime.FuncForPC(pc).Name()
	_, line := runtime.FuncForPC(pc).FileLine(pc)
	return file, line
}
