package main

import (
	"coffer/cmd"
	"time"
)

func main() {
	cmd.CMDControl()
	time.Sleep(time.Duration(1) * time.Second) //延时函数拖延时间让资源回收完毕
}
