package utils

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

func Logout(level string, information ...interface{}) {
	file, line := printCallerPosition()
	fileline := fmt.Sprintf("%s:%d", file, line)
	//创建自定义logger
	logger := log.New(os.Stdout, "["+level+"] "+fileline+" ", log.LstdFlags)
	switch level {
	case "PANIC":
		logger.Panicln(information...) //panic将终止函数并“抛出异常”
	default:
		logger.Println(information...)
	}
}
func printCallerPosition() (string, int) {
	pc, _, _, _ := runtime.Caller(2) //调用堆栈的第2层函数
	file := runtime.FuncForPC(pc).Name()
	_, line := runtime.FuncForPC(pc).FileLine(pc)
	return file, line
}
