package cmd

import (
	"coffer/container"
	"coffer/log"
	"coffer/subsys"
	"fmt"
	"os"
)

type command interface { //指令接口,所有指令都需要实现以下接口
	usage()                                          //使用说明
	execute(nonFlagNum int, argument []string) error //需要执行的具体操作
}

type runCommand struct{}

func (run *runCommand) usage() {
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
	-e			Set environment variables(can be used multiple times at one time)
`)
}
func (run1 *runCommand) execute(nonFlagNum int, argument []string) error {
	if background && interactive { //后台运行与可交互不能同时执行
		return fmt.Errorf("application interaction(-i) and background running(-b) cannot be used at the same time")
	}
	if nonFlagNum < 1 { //run后没有可执行程序
		fmt.Printf("\"coffer run\" requires at least 1 argument.\nSee 'coffer run -help'.\n")
		return fmt.Errorf("error command:No executable commands")
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
	if err := run(interactive, dataPersistence,
		containerName, imageName, cmdArray, environment.String(), resConfig); err != nil {
		return fmt.Errorf("run image error,%v", err)
	}
	return nil
}

type initCommand struct{}

func (init *initCommand) usage() {}
func (init *initCommand) execute(nonFlagNum int, argument []string) error {
	if err := container.InitializeContainer(); err != nil {
		return fmt.Errorf("initialize container error:%v", err)
	}
	return nil
}

type commitCommand struct{}

func (commit *commitCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer commit CONTAINER IMAGE
Create a new image from a container
`)
}
func (commit *commitCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum < 2 { //commit后缺少容器名称或镜像名称
		fmt.Println("\"coffer exec\" requires two parameters: container name and image name.\nSee 'coffer commit -help'.")
		return fmt.Errorf("error command:No container or image specified")
	}
	containerName := argument[0]
	imageName := argument[1]
	if err := commitContainer(containerName, imageName); err != nil {
		return fmt.Errorf("commit container error:%v", err)
	} else {
		log.Logout("INFO", "Commit ", containerName, " to ", imageName, " succeeded")
	}
	return nil
}

type psCommand struct{}

func (ps *psCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer ps
List containers
`)
}
func (ps *psCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum >= 1 { //有非flag参数
		fmt.Println("there are redundant parameters.\nSee 'coffer ps -help'.")
		return fmt.Errorf("error command:Redundant commands")
	}
	if err := ListContainers(); err != nil {
		return fmt.Errorf("list container error:%v", err.Error())
	}
	return nil
}

type logCommand struct{}

func (log *logCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer log CONTAINER
Print log of a container
`)
}
func (log1 *logCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum != 1 { //log后没有容器名称或参数过多
		fmt.Printf("\"coffer run\" requires only 1 argument.\nSee 'coffer log -help'.\n")
		return fmt.Errorf("error command:No executable commands or redundant commands")
	}
	if err := LogContainer(argument[0]); err != nil {
		return fmt.Errorf("log container error:%v", err)
	}
	return nil
}

type execCommand struct{}

func (exec *execCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer exec CONTAINER COMMAND
Run a command in a running container
`)
}
func (exec *execCommand) execute(nonFlagNum int, argument []string) error {
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
		return fmt.Errorf("exec container error:%v", err)
	}
	return nil
}

type stopCommand struct{}

func (stop *stopCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer stop CONTAINER
Stop the running container
`)
}
func (stop *stopCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum != 1 { //stop后没有容器名称或参数过多
		fmt.Printf("\"coffer run\" requires only 1 argument.\nSee 'coffer stop -help'.\n")
		return fmt.Errorf("error command:No executable commands or redundant commands")
	}
	if err := stopContainer(argument[0]); err != nil {
		return fmt.Errorf("stop container error:%v", err)
	} else {
		log.Logout("INFO", "Container stopped succeeded")
	}
	return nil
}

type rmCommand struct{}

func (rm *rmCommand) usage() {
	fmt.Fprintf(os.Stderr, `Usage:  coffer rm CONTAINER
Remove the stopped container
`)
}
func (rm *rmCommand) execute(nonFlagNum int, argument []string) error {
	if nonFlagNum != 1 { //rm后没有容器名称或参数过多
		fmt.Printf("\"coffer run\" requires at least 1 argument.\nSee 'coffer rm -help'.\n")
		return fmt.Errorf("error command:No executable commands")
	}
	if err := rmContainer(argument[0]); err != nil {
		return fmt.Errorf("remove container error:%v", err)
	} else {
		log.Logout("INFO", "Remove container succeeded")
	}
	return nil
}
