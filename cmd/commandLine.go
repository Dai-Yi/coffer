package cmd

import (
	"coffer/container"
	"coffer/log"
	"coffer/subsys"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	help            bool
	version         bool
	interactive     bool
	background      bool
	dataPersistence string
	cpuShare        string
	memory          string
	cpuset_cpus     string
	cpuset_mems     string
	name            string
)

func init() {
	flag.BoolVar(&help, "h", false, "") //不用flag自带usage
	flag.BoolVar(&help, "help", false, "")
	flag.BoolVar(&version, "v", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&interactive, "i", false, "")
	flag.BoolVar(&background, "b", false, "")
	flag.StringVar(&dataPersistence, "p", "", "")
	flag.StringVar(&cpuShare, "cpushare", "0", "")
	flag.StringVar(&memory, "memory", "", "")
	flag.StringVar(&cpuset_cpus, "cpuset-cpus", "0", "")
	flag.StringVar(&cpuset_mems, "cpuset-mems", "0", "")
	flag.StringVar(&name, "name", "", "")
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
	commit		Create a new image from a container's changes
Run 'coffer COMMAND -h' for more information on a command
`)
}
func runUsage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer run [OPTIONS] IMAGE [COMMAND] [ARG...]
Run a command in a new container
Options:
	-i			Make the app interactive:attach STDIN,STDOUT,STDERR
	-p			Bind mount a data volume(Data Persistence)
	-b			Run container in background(CANNOT be used with -i at the same time)
	-cpushare		CPU shares (relative weight)
	-memory			Memory limit
	-cpuset-cpus		CPUs in which to allow execution
	-cpuset-mems		MEMs in which to allow execution
	-name				Assign a name to the container
`)
}
func commitUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer commit CONTAINER
Create a new image from a container's changes`)
}
func CMDControl() {
	var command []string
	if len(os.Args) <= 1 { //未输入参数
		log.Logout("ERROR", "Missing Command, enter -h or -help to show usage")
		return
	} else {
		flag.Parse()          //第一次解析，解析help、version参数
		if flag.NArg() >= 1 { //可能有命令,run,commit
			argument := os.Args[1] //保存命令
			os.Args = os.Args[1:]  //删除阻碍解析的coffer命令
			flag.Parse()           //第二次解析，解析命令参数
			if strings.EqualFold(argument, "run") {
				if background && interactive {
					log.Logout("ERROR", "Application interaction(-i) and background running(-b) cannot be used at the same time")
					return
				}
				if flag.NArg() >= 1 { //有待运行程序
					command = flag.Args() //排除run参数
					runCommand(command)
				} else { //run后没有可执行程序
					if help { //run help
						flag.Usage = runUsage
					} else {
						fmt.Println("\"coffer run\" requires at least 1 argument.\nSee 'coffer run -help'.")
						log.Logout("ERROR", "Error command:No executable commands")
						return
					}
				}
			} else if argument == "INiTcoNtaInER" { //内部命令，禁止外部调用
				initCommand()
			} else if strings.EqualFold(argument, "commit") {
				if flag.NArg() == 1 { //有镜像名称
					command = flag.Args()
					commitCommand(command[0])
				} else { //commit后没有镜像名称
					if help { //commit help
						flag.Usage = commitUsage
					} else {
						fmt.Println("\"coffer commit\" requires at least 1 argument.\nSee 'coffer commit -help'.")
						log.Logout("ERROR", "Error command:No executable commands")
						return
					}
				}
			} else {
				log.Logout("ERROR", "Invalid command")
				return
			}
		} else { //没有命令，只有flag参数
			if version {
				fmt.Println("Version：coffer/1.0.0")
			} else {
				log.Logout("ERROR", "Invalid command")
				return
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
	if err := run(interactive, background, name, dataPersistence, commands, resConfig); err != nil {
		log.Logout("ERROR", "Run image error,", err.Error())
		return
	}
}
func commitCommand(image string) {
	log.Logout("INFO", "Commit", image)
	if err := commitContainer(image); err != nil {
		log.Logout("ERROR", "Commit container error:", err.Error())
		return
	}
}
func initCommand() {
	if err := container.InitializeContainer(); err != nil {
		log.Logout("ERROR", "Initialize container error:", err.Error())
		container.GracefulExit()
	}
}
