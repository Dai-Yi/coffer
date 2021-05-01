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
	commit		Save the contents of the running container as a image
	ps			List containers
	log			Print log of a container
	exec		Run a command in a running container
	stop		Stop the running container
	rm			Remove the stopped container
Run 'coffer COMMAND -h' for more information on a command
`)
}
func runUsage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer run [OPTIONS] IMAGE [COMMAND]
Run a command in a new container
Options:
	-i			Make the app interactive:attach STDIN,STDOUT,STDERR
	-p			Bind mount a data volume(Data Persistence)
	-b			Run container in background(CANNOT be used with -i at the same time)
	-cpushare		CPU shares (relative weight)
	-memory			Memory limit
	-cpuset-cpus		CPUs in which to allow execution
	-cpuset-mems		MEMs in which to allow execution
	-name			Assign a name to the container
`)
}
func commitUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer commit IMAGE
Save the contents of the running container as a image
`)
}
func psUsage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer ps
List containers
`)
}
func logUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer log CONTAINER
Print log of a container
`)
}
func execUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer exec CONTAINER COMMAND
Run a command in a running container
`)
}
func stopUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer stop CONTAINER
Stop the running container
`)
}
func rmUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer rm CONTAINER
Remove the stopped container
`)
}
func CMDControl() {
	if len(os.Args) <= 1 { //未输入参数
		log.Logout("ERROR", "Missing Command, enter -h or -help to show usage")
		return
	} else {
		flag.Parse()          //第一次解析，解析help、version参数
		if flag.NArg() >= 1 { //可能有命令,run,commit
			argument := os.Args[1]                  //保存命令
			os.Args = os.Args[1:]                   //删除阻碍解析的coffer命令
			flag.Parse()                            //第二次解析，解析命令参数
			if strings.EqualFold(argument, "run") { //run指令
				if background && interactive {
					log.Logout("ERROR", "Application interaction(-i) and background running(-b) cannot be used at the same time")
					return
				}
				if flag.NArg() >= 1 { //有待运行程序
					runCommand(flag.Args()) //排除run参数
				} else { //run后没有可执行程序
					helpAndErrorHandle("run", runUsage)
				}
			} else if argument == "INiTcoNtaInER" { //内部命令，禁止外部调用
				initCommand()
			} else if strings.EqualFold(argument, "commit") { //commit指令
				if flag.NArg() == 1 { //有镜像名称
					commitCommand(flag.Args()[0])
				} else { //commit后没有镜像名称
					helpAndErrorHandle("commit", commitUsage)
				}
			} else if strings.EqualFold(argument, "ps") { //ps 指令
				if flag.NArg() >= 1 { //有非flag参数
					fmt.Println("there are redundant parameters.\nSee 'coffer ps -help'.")
					log.Logout("ERROR", "Error command:Redundant commands")
					return
				} else {
					if help { //ps help
						flag.Usage = psUsage
					} else {
						psCommand()
					}
				}
			} else if strings.EqualFold(argument, "log") { //log指令
				if flag.NArg() == 1 { //有容器名称
					logCommand(flag.Args()[0])
				} else { //log后没有容器名称
					helpAndErrorHandle("log", logUsage)
				}
			} else if strings.EqualFold(argument, "exec") { //exec指令
				//若已经指定了环境变量,说明C代码已经运行,直接返回以免重复调用
				if os.Getenv(ENV_EXEC_PID) != "" {
					log.Logout("INFO", "pid callback,pid:", os.Getegid())
					return
				}
				if flag.NArg() >= 2 { //有容器名和命令
					execCommand(flag.Args())
				} else { //exec后缺少容器名或命令
					helpAndErrorHandle("exec", execUsage)
				}
			} else if strings.EqualFold(argument, "stop") { //stop指令
				if flag.NArg() == 1 { //有容器名称
					stopCommand(flag.Args()[0])
				} else { //stop后没有容器名称
					helpAndErrorHandle("stop", stopUsage)
				}
			} else if strings.EqualFold(argument, "rm") { //rm指令{
				if flag.NArg() == 1 { //有容器名称
					rmCommand(flag.Args()[0])
				} else { //rm后没有容器名称
					helpAndErrorHandle("rm", rmUsage)
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
func helpAndErrorHandle(action string, usage func()) {
	if help {
		flag.Usage = usage
	} else {
		switch action {
		case "run", "commit", "log", "stop", "rm":
			fmt.Printf("\"coffer run\" requires at least 1 argument.\nSee 'coffer %v -help'.\n", action)
			log.Logout("ERROR", "Error command:No executable commands")
		case "exec":
			fmt.Println("\"coffer exec\" requires two parameters: container and command.\nSee 'coffer exec -help'.")
			log.Logout("ERROR", "Error command:No container specified or executable commands")
		}
		os.Exit(0)
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
	if err := run(interactive, background, dataPersistence, name, commands, resConfig); err != nil {
		log.Logout("ERROR", "Run image error,", err.Error())
		os.Exit(1)
	}
}
func commitCommand(image string) {
	log.Logout("INFO", "Commit", image)
	if err := commitContainer(image); err != nil {
		log.Logout("ERROR", "Commit container error:", err.Error())
		os.Exit(1)
	}
}
func initCommand() {
	if err := container.InitializeContainer(); err != nil {
		log.Logout("ERROR", "Initialize container error:", err.Error())
		//container.GracefulExit()
		os.Exit(1)
	}
}
func psCommand() {
	if err := ListContainers(); err != nil {
		log.Logout("ERROR", "List container error:", err.Error())
		os.Exit(1)
	}
}
func logCommand(container string) {
	if err := LogContainer(container); err != nil {
		log.Logout("ERROR", "log container error:", err.Error())
		os.Exit(1)
	}
}
func execCommand(commands []string) {
	container := commands[0] //容器名是第一个
	cmdArray := commands[1:] //容器名后为命令
	if err := execContainer(container, cmdArray); err != nil {
		log.Logout("ERROR", "exec container error:", err.Error())
		os.Exit(1)
	}
}
func stopCommand(container string) {
	if err := stopContainer(container); err != nil {
		log.Logout("ERROR", "stop container error:", err.Error())
		os.Exit(1)
	}
}
func rmCommand(container string) {
	if err := rmContainer(container); err != nil {
		log.Logout("ERROR", "remove container error:", err.Error())
		os.Exit(1)
	}
}
