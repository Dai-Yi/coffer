package cmd

import (
	"coffer/log"
	"coffer/proc"
	"flag"
	"os"
	"strings"
)

var (
	help        bool
	version     bool
	showProcess bool
)

func init() {
	flag.BoolVar(&help, "h", false, "显示帮助") //使用方法
	flag.BoolVar(&help, "help", false, "显示帮助")
	flag.BoolVar(&version, "v", false, "显示当前版本") //版本
	flag.BoolVar(&version, "version", false, "显示当前版本")
	flag.BoolVar(&showProcess, "s", false, "运行进程时显示运行结果")
}
func Monitor() {
	if len(os.Args) <= 1 { //未输入参数
		log.Logout("WARN", "未输入参数 使用-h或-help查看帮助")
	} else {
		flag.Parse()          //解析
		if flag.NArg() >= 1 { //有非flag参数
			nonFlagArgument := flag.Args()              //将非flag参数储存起来
			os.Args = os.Args[flag.NArg()-1:]           //删除阻碍解析的非flag参数(少删一个：Parse会跳过第一个)
			flag.Parse()                                //重新解析
			for i := 0; i < len(nonFlagArgument); i++ { //遍历非参数命令
				if strings.EqualFold(nonFlagArgument[i], "run") { //run命令
					log.Logout("INFO", "running "+nonFlagArgument[i+1])
					proc.Run(showProcess, nonFlagArgument[i+1])
				}
				if nonFlagArgument[i] == "init" { //init是内部命令，禁止外部调用
					log.Logout("INFO", "initializing "+nonFlagArgument[i+1])
					proc.InitializeContainer(nonFlagArgument[i+1])
				}
			}
		} else {
			if help {
				flag.PrintDefaults()
			}
			if version {
				log.Logout("INFO", "当前版本：1.0.0")
			}
		}

	}

}
