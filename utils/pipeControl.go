package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

//给父进程的管道发送信息
func PipeSendToParent(msgList interface{}) {
	var msgStr string
	val1, ok := msgList.([]string) //断言是否为[]string类型
	if ok {                        //如果是则转换为string类型
		msgStr = strings.Join(val1, " ")
	} else { //如果不是[]string则为string类型
		val2 := msgList.(string) //断言是否为string类型
		msgStr = val2
	}
	writePipe := os.NewFile(uintptr(3), "pipe") //从文件描述符获取管道
	Lock(writePipe)                             //加锁
	writePipe.WriteString(msgStr)               //若失败则告诉父进程失败
	UnLock(writePipe)                           //解锁
	writePipe.Close()                           //关闭写入端
}

//给子进程的管道发送信息
func PipeSendToChild(msgList interface{}, writePipe *os.File) {
	var msgStr string
	val1, ok := msgList.([]string) //断言是否为[]string类型
	if ok {                        //如果是则转换为string类型
		msgStr = strings.Join(val1, " ")
	} else { //如果不是[]string则为string类型
		val2 := msgList.(string) //断言是否为string类型
		msgStr = val2
	}
	Lock(writePipe)               //加锁
	writePipe.WriteString(msgStr) //命令写入管道
	UnLock(writePipe)             //解锁
	writePipe.Close()             //关闭写入端
}

//父进程接收信息
func PipeReceiveFromParent(readPipe *os.File) (string, error) {
	msg, err := ioutil.ReadAll(readPipe)
	if err != nil {
		return "", fmt.Errorf("init read pipe error->%v", err)
	}
	return string(msg), nil
}

//子进程接收信息
func PipeReceiveFromChild() (string, error) {
	pipe := os.NewFile(uintptr(3), "pipe") //从文件描述符获取管道
	//用这种方式是因为在创建进程后会初始化新的管道，只有这种方式才能访问父进程的管道
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		return "", fmt.Errorf("init read pipe error->%v", err)
	}
	return string(msg), nil
}
