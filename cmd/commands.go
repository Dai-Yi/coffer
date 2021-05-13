package cmd

import (
	"coffer/container"
	"coffer/log"
	"coffer/net"
	"coffer/subsys"
	"flag"
	"fmt"
	"os"
)

type command interface { //指令接口,所有指令都需要实现以下接口
	usage()                                          //使用说明
	execute(nonFlagNum int, argument []string) error //需要执行的具体操作
}

type runCommand struct{}

func (*runCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer run [OPTIONS] IMAGE [COMMAND]
Run a command in a new container.

Options:
	-i			Make the app interactive:attach STDIN,STDOUT,STDERR
	-p			Bind mount a data volume(Data Persistence)
	-b			Run container in background(CANNOT be used with -i at the same time)
	-cpushare		CPU shares (relative weight)
	-memory			Memory limit
	-cpuset-cpus		CPUs in which to allow execution
	-cpuset-mems		MEMs in which to allow execution
	-name			Assign a name to the container
	-e			Set environment variables(can be used multiple times at one time)
`)
}
func (*runCommand) execute(nonFlagNum int, argument []string) error {
	if background && interactive { //后台运行与可交互不能同时执行
		return fmt.Errorf("application interaction(-i) and background running(-b) cannot be used at the same time")
	}
	if nonFlagNum < 2 { //run后没有可执行程序和命令
		fmt.Printf("\"coffer run\" requires two parameters: image name and command.\nSee 'coffer run -help'.\n")
		return fmt.Errorf("error command:No image name or executable commands")
	}
	imageName := argument[0]
	cmdArray := argument[1:]
	resConfig := &subsys.ResourceConfig{
		MemoryLimit: memory,
		CpuShare:    cpuShare,
		Cpuset: &subsys.CpuSet{
			Cpus: cpuset_cpus,
			Mems: cpuset_mems,
		},
	}
	env := os.Getenv(container.ENV_RUN)      //获取环境变量判断当前是否为后台进程
	if !interactive && env != "background" { //如果需要后台运行但当前并非后台进程则转换为后台进程
		if err := transform(); err != nil {
			return fmt.Errorf("transform self into background error")
		}
		return nil
	} //到这里时，肯定是前台运行或后台守护进程已启动
	setProcessName("coffer")
	//如果已经转换为后台进程或需要交互则正常运行
	if err := run(interactive, dataPersistence, containerName, imageName, network,
		cmdArray, environment.String(), portmapping.String(), resConfig); err != nil {
		return fmt.Errorf("run image error->%v", err)
	}
	return nil
}

type initCommand struct{}

func (*initCommand) usage() {}
func (*initCommand) execute(nonFlagNum int, argument []string) error {
	if err := container.InitializeContainer(); err != nil {
		return fmt.Errorf("initialize container error->%v", err)
	}
	return nil
}

type commitCommand struct{}

func (*commitCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer commit CONTAINER IMAGE
Create a new image from a container.

`)
}
func (*commitCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum < 2 { //commit后缺少容器名称或镜像名称
		fmt.Println("\"coffer commit\" requires two parameters: container name and image name.\nSee 'coffer commit -help'.")
		return fmt.Errorf("error command:No container or image specified")
	}
	containerName := argument[0]
	imageName := argument[1]
	if err := commitContainer(containerName, imageName); err != nil {
		return fmt.Errorf("commit container error->%v", err)
	} else {
		log.Logout("INFO", "Commit", containerName, "to", imageName, "succeeded")
	}
	return nil
}

type psCommand struct{}

func (*psCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer ps
List containers.

`)
}
func (*psCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum >= 1 { //有非flag参数
		fmt.Println("there are redundant parameters.\nSee 'coffer ps -help'.")
		return fmt.Errorf("error command:Redundant commands")
	}
	if err := printContainers(); err != nil {
		return fmt.Errorf("print containers error->%v", err.Error())
	}
	return nil
}

type logCommand struct{}

func (*logCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer log CONTAINER
Print log of a container.

`)
}
func (*logCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum != 1 { //log后没有容器名称或参数过多
		fmt.Printf("\"coffer log\" requires only 1 argument.\nSee 'coffer log -help'.\n")
		return fmt.Errorf("error command:No executable commands or redundant commands")
	}
	if err := logContainer(argument[0]); err != nil {
		return fmt.Errorf("log container error->%v", err)
	}
	return nil
}

type execCommand struct{}

func (*execCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer exec CONTAINER COMMAND
Run a command in a container running in the background.

`)
}
func (*execCommand) execute(nonFlagNum int, argument []string) error {
	//若已经指定了环境变量,说明C代码已经运行,直接返回以免重复调用
	if os.Getenv(ENV_EXEC_PID) != "" {
		log.Logout("INFO", "pid callback,pid:", os.Getegid())
		return nil
	}
	if nonFlagNum < 2 { //exec后缺少容器名或命令
		fmt.Println("\"coffer exec\" requires two parameters: container and command.\nSee 'coffer exec -help'.")
		return fmt.Errorf("error command:No container specified or executable commands")
	}
	container := argument[0] //容器名是第一个
	cmdArray := argument[1:] //容器名后为命令
	if err := execContainer(container, cmdArray); err != nil {
		return fmt.Errorf("exec container error->%v", err)
	}
	return nil
}

type stopCommand struct{}

func (*stopCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer stop CONTAINER
Stop the running container.

`)
}
func (*stopCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum != 1 { //stop后没有容器名称或参数过多
		fmt.Printf("\"coffer stop\" requires only 1 argument.\nSee 'coffer stop -help'.\n")
		return fmt.Errorf("error command:No executable commands or redundant commands")
	}
	if err := stopContainer(argument[0]); err != nil {
		return fmt.Errorf("stop container error->%v", err)
	} else {
		log.Logout("INFO", "Container stopped succeeded")
	}
	return nil
}

type rmCommand struct{}

func (*rmCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer rm CONTAINER
Remove the stopped container.

`)
}
func (*rmCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum != 1 { //rm后没有容器名称或参数过多
		fmt.Printf("\"coffer rm\" requires at least 1 argument.\nSee 'coffer rm -help'.\n")
		return fmt.Errorf("error command:No executable commands")
	}
	if err := rmContainer(argument[0]); err != nil {
		return fmt.Errorf("remove container error->%v", err)
	} else {
		log.Logout("INFO", "Remove container succeeded")
	}
	return nil
}

type networkCommand struct{}

func (*networkCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer network COMMAND
Manage network.

Commands:
	create			Create a container network
	list			List container network
	remove			Remove container network
Run 'coffer network COMMAND -help' for more information on a command.
`)
}
func (*networkCommand) execute(_ int, _ []string) error {
	if len(os.Args) <= 1 { //未输入参数
		return fmt.Errorf("missing Command, enter -h or -help to show network usage")
	}
	netArgument := os.Args[1] //保存命令
	os.Args = os.Args[1:]     //删除阻碍解析的network命令
	flag.Parse()              //network存在子命令，需要多解析一次
	switch netArgument {
	case "create":
		if flag.NArg() >= 1 { //有网络名
			if err := net.Init(); err != nil { //初始化网络
				return fmt.Errorf("initialize network error->%v", err)
			}
			err := net.CreateNetwork(driver, subnet, flag.Args()[0]) //创建网络
			if err != nil {
				return fmt.Errorf("create network error->%v", err)
			}
			return nil
		} else { //create后没有网络名
			if !help {
				fmt.Println("requires at least 1 argument.\nSee 'coffer network create -help'.")
				return fmt.Errorf("error command:No network name")
			}
			flag.Usage = networkCreateUsage
		}
	case "list":
		if flag.NArg() >= 1 { //有非flag参数
			fmt.Println("there are redundant parameters.\nSee 'coffer network list -help'.")
			return fmt.Errorf("error command:Redundant commands")
		} else {
			if help {
				flag.Usage = networkListUsage
			} else {
				if err := net.Init(); err != nil { //初始化网络
					return fmt.Errorf("initialize network error->%v", err)
				}
				if err := net.ListNetwork(); err != nil { //显示网络列表
					return fmt.Errorf("list network error->%v", err)
				}
			}
		}
	case "remove":
		if flag.NArg() >= 1 { //有待运行程序
			if err := net.Init(); err != nil { //初始化网络
				return fmt.Errorf("initialize network error->%v", err)
			}
			if err := net.DeleteNetwork(flag.Args()[0]); err != nil {
				return fmt.Errorf("remove network error->%v", err)
			}
		} else { //remove后没有可执行程序
			if !help {
				fmt.Println("requires at least 1 argument.\nSee 'coffer network remove -help'.")
				return fmt.Errorf("error command:No network name")
			}
			flag.Usage = networkRemoveUsage
		}
	}
	if help {
		flag.Usage()
	}
	return nil
}
func networkCreateUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer network create [OPTIONS] NETWORK
Create container network.

Options:
	-driver			Driver to manage the Network (default "bridge")
	-subnet			Subnet in CIDR format that represents a network segment
`)
}
func networkListUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer network list
List container network.

`)
}

func networkRemoveUsage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer network remove NETWORK
Remove container network.

`)
}
