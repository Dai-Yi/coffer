package cmd

import (
	"coffer/log"
	"coffer/proc"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	help        bool
	version     bool
	showProcess bool
)

func init() {
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "")
	flag.BoolVar(&version, "v", false, "") //版本
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&showProcess, "s", false, "")
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer COMMAND [OPTIONS]
	coffer [ -h | -help | -v | -version ]
Commands:
	run			Run a command in a new container
Options:
	-h,-help		Print usage
	-v,-version		Print version information
`)
}
func Monitor() {
	if len(os.Args) <= 1 { //未输入参数
		log.Logout("ERROR", "Missing Command, enter -h or -help to show usage")
		return
	} else {
		flag.Parse()          //解析
		if flag.NArg() >= 1 { //有非flag参数
			nonFlagArgument := flag.Args()              //将非flag参数储存起来
			os.Args = os.Args[flag.NArg()-1:]           //删除阻碍解析的非flag参数(少删一个：Parse会跳过第一个)
			flag.Parse()                                //重新解析
			for i := 0; i < len(nonFlagArgument); i++ { //遍历非参数命令
				if strings.EqualFold(nonFlagArgument[i], "run") { //run命令
					log.Logout("INFO", "Run "+nonFlagArgument[i+1])
					proc.Run(showProcess, nonFlagArgument[i+1:])
				} else if nonFlagArgument[i] == "INiTcoNtaInER" { //内部命令，禁止外部调用
					proc.InitializeContainer()
				} else {
					log.Logout("ERROR", "invalid command")
					os.Exit(1)
				}
			}
		} else {
			if help {
				flag.Usage()
			}
			if version {
				log.Logout("INFO", "Version：coffer/1.0.0")
			}
		}
	}
}
