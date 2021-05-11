package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type env []string //自定义env类型
func (e *env) String() string { //Value.String接口实现
	r := []string{}
	for _, s := range *e {
		r = append(r, s)
	}
	return strings.Join(r, ",")
}
func (e *env) Set(s string) error { //Value.Set接口实现
	*e = append(*e, s)
	return nil
}

var (
	instructionOperation = map[string]command{ //指令集，实现新指令需添加到之内
		"run":           &runCommand{},
		"INiTcoNtaInER": &initCommand{},
		"commit":        &commitCommand{},
		"ps":            &psCommand{},
		"log":           &logCommand{},
		"exec":          &execCommand{},
		"stop":          &stopCommand{},
		"rm":            &rmCommand{},
		"network":       &networkCommand{},
	}
	environment     flag.Value //自定义Value类型
	help            bool
	version         bool
	interactive     bool
	background      bool
	dataPersistence string
	cpuShare        string
	memory          string
	cpuset_cpus     string
	cpuset_mems     string
	containerName   string

	driver string
	subnet string
)

func init() {
	flag.Usage = defaultusage
	environment = &env{} //实现flag.Value接口
	flag.Var(environment, "e", "")
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
	flag.StringVar(&containerName, "name", "", "")
	//network子命令参数
	flag.StringVar(&driver, "driver", "", "")
	flag.StringVar(&subnet, "subnet", "", "")
}
func defaultusage() {
	fmt.Fprintf(os.Stderr, `Usage:	coffer [OPTIONS] COMMAND
	coffer [ -h | -help | -v | -version ]
Options:
	-h,-help		Print usage
	-v,-version		Print version information
Management Commands:
	network			Container network commands
Commands:
	run			Run a command in a new container
	commit		Create a new image from a container
	ps			List containers
	log			Print log of a container
	exec		Run a command in a running container
	stop		Stop the running container
	rm			Remove the stopped container
Run 'coffer COMMAND -help' for more information on a command
`)
}

func CMDControl() {
	if len(os.Args) <= 1 { //未输入参数
		log.SetPrefix("[ERROR]")
		log.Println("Missing Command, enter -h or -help to show usage")
		return
	}
	flag.Parse()          //第一次解析，解析help、version参数
	if flag.NArg() >= 1 { //可能有命令,run,commit等
		argument := os.Args[1] //保存命令
		os.Args = os.Args[1:]  //删除阻碍解析的coffer命令
		flag.Parse()           //第二次解析，解析命令参数
		cmd, ok := instructionOperation[argument]
		if !ok { //若没有找到相应指令
			log.SetPrefix("[ERROR]")
			log.Println("Invalid command")
			return
		}
		if help {
			flag.Usage = cmd.usage
		} else {
			if err := cmd.execute(flag.NArg(), flag.Args()); err != nil { //执行指令
				log.SetPrefix("[ERROR]")
				log.Println(err.Error())
				return
			}
		}
	} else { //没有命令，只有flag参数
		if version {
			fmt.Println("Version：coffer/1.0.0")
			return
		}
	} //无论有没有命令,出现help则显示帮助
	if !help {
		log.SetPrefix("[ERROR]")
		log.Println("Invalid command")
		return
	}
	flag.Usage()
}
