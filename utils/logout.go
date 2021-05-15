package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
)

const CofferLogFile string = "/var/run/coffer/cofferLogout.log"

func Logout(level string, information ...interface{}) {
	//打开日志文件
	logFile, err := os.OpenFile(CofferLogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	//延迟关闭
	defer logFile.Close()

	//设置日志同时输出到log文件和终端
	writers := []io.Writer{logFile, os.Stdout}
	fileAndStdoutWriter := io.MultiWriter(writers...) //省略号是将writers切片打散
	file, line := printCallerPosition()
	fileline := fmt.Sprintf("%s:%d", file, line)
	//创建自定义logger
	logger := log.New(fileAndStdoutWriter, "["+level+"] "+fileline+" ", log.LstdFlags)
	Lock(logFile) //加锁
	switch level {
	case "PANIC":
		logger.Panicln(information...) //panic将终止函数并“抛出异常”
	default:
		logger.Println(information...)
	}
	UnLock(logFile) //解锁
}
func printCallerPosition() (string, int) {
	pc, _, _, _ := runtime.Caller(2) //调用堆栈的第2层函数
	file := runtime.FuncForPC(pc).Name()
	_, line := runtime.FuncForPC(pc).FileLine(pc)
	return file, line
}
