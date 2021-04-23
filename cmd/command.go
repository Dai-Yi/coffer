package cmd

import (
	"coffer/initCNTR"
	"coffer/log"
	"coffer/proc"
	"coffer/subsys"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	help        bool
	version     bool
	showProcess bool
	cpuShare    string
	memory      string
	cpuset_cpus string
	cpuset_mems string
)

func init() {
	flag.BoolVar(&help, "h", false, "") //不用flag自带usage
	flag.BoolVar(&help, "help", false, "")
	flag.BoolVar(&version, "v", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&showProcess, "s", false, "")
	flag.StringVar(&cpuShare, "cpushare", "0", "")
	flag.StringVar(&memory, "memory", "", "")
	flag.StringVar(&cpuset_cpus, "cpuset-cpus", "0", "")
	flag.StringVar(&cpuset_mems, "cpuset-mems", "0", "")
	flag.Usage = usage
}
func usage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer [OPTIONS] COMMAND
	coffer [ -h | -help | -v | -version ]
Options:
	-h,-help		Print usage
	-v,-version		Print version information
Commands:
	run			Run a command in a new container
`)
}
func runUsage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer run [OPTIONS] IMAGE [COMMAND] [ARG...]
Run a command in a new container
Options:
	-s			Attach STDIN,STDOUT,STDERR
	-cpushare		CPU shares (relative weight)
	-memory			Memory limit
	-cpuset-cpus		CPUs in which to allow execution
	-cpuset-mems		MEMs in which to allow execution
`)
}
func Monitor() {
	var image []string
	if len(os.Args) <= 1 { //未输入参数
		log.Logout("ERROR", "Missing Command, enter -h or -help to show usage")
		return
	} else {
		flag.Parse()          //第一次解析，解析help、version参数
		if flag.NArg() >= 1 { //可能有命令,run
			argument := os.Args[1] //保存命令
			os.Args = os.Args[1:]  //删除阻碍解析的coffer命令
			flag.Parse()           //第二次解析，解析命令参数
			if strings.EqualFold(argument, "run") {
				if flag.NArg() >= 1 { //有待运行程序
					image = flag.Args() //排除run参数
					runCommand(image)
				} else { //run后没有可执行程序
					if help { //run help
						flag.Usage = runUsage
					} else {
						fmt.Println("\"coffer run\" requires at least 1 argument.\nSee 'coffer run -help'.")
						log.Logout("ERROR", "Error command:No executable commands")
					}
				}
			} else if argument == "INiTcoNtaInER" { //内部命令，禁止外部调用
				initCNTR.InitializeContainer()
			} else {
				log.Logout("ERROR", "Invalid command")
			}
		} else { //没有命令，只有flag参数
			if version {
				fmt.Println("Version：coffer/1.0.0")
			} else {
				log.Logout("ERROR", "Invalid command")
			}
		}
		if help {
			flag.Usage()
		}
	}
}
func runCommand(commands []string) {
	log.Logout("INFO", "Run", commands)
	resConfig := &subsys.ResourceConfig{
		MemoryLimit: memory,
		CpuShare:    cpuShare,
		Cpuset: &subsys.CpuSet{
			Cpus: cpuset_cpus,
			Mems: cpuset_mems,
		}}
	proc.Run(showProcess, commands, resConfig) //传递coffer run之后的命令
}
