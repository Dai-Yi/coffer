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
func runUsage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer run [OPTIONS]
Options:
	-s			Attach STDIN,STDOUT,STDERR
`)
}
func Monitor() {
	if len(os.Args) <= 1 { //未输入参数
		log.Logout("ERROR", "Missing Command, enter -h or -help to show usage")
		return
	} else {
		flag.Parse()          //解析
		if flag.NArg() >= 1 { //有非flag参数
			argument := flag.Args()                    //将参数储存起来
			os.Args = os.Args[flag.NArg()-1:]          //删除阻碍解析的非flag参数
			flag.Parse()                               //重新解析
			if strings.EqualFold(argument[0], "run") { //run命令
				runCommand(argument[1:])
			} else if argument[0] == "INiTcoNtaInER" { //内部命令，禁止外部调用
				proc.InitializeContainer()
			} else {
				log.Logout("ERROR", "Invalid command")
			}
		} else if version {
			fmt.Println("Version：coffer/1.0.0")
		}
		if help {
			flag.Usage()
		}
	}
}

func runCommand(commands []string) {
	if len(commands) < 1 { //没有可执行命令
		fmt.Println("\"coffer run\" requires at least 1 argument.\nSee 'coffer run -help'.")
		log.Logout("ERROR", "Error command:No executable commands")
	} else {
		if help { //run help
			flag.Usage = runUsage
		} else {
			log.Logout("INFO", "Run "+commands[0])
			proc.Run(showProcess, commands) //传递coffer run之后的命令
		}
	}
}
