package subsys

import (
	"bufio"
	"coffer/utils"
	"fmt"
	"os"
	"path"
	"strings"
)

type ResourceConfig struct {
	MemoryLimit string  //内存
	CpuShare    string  //CPU时间片权重,默认值是1024
	Cpuset      *CpuSet //CPU和内存节点分配给一组任务
}
type CpuSet struct {
	Cpus string //CPU列表,默认关闭状态(0)
	Mems string //内存节点列表,默认关闭状态(0)
}
type Subsystem interface {
	Name() string                               //返回subsystem的名字
	Set(path string, res *ResourceConfig) error //CGroup限制资源
	Apply(path string, pid int) error           //进程添加到CGroup
	Remove(path string) error                   //将进程移出CGroup
}

func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open("/proc/self/mountinfo") //os.open用于读
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f) //bufio将文件读取到内存中
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ") //切割字符串
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return fields[4]
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return ""
	}
	return ""
}
func GetCgroupPath(subsystem string, cgroupPath string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)
	if !utils.PathExists(path.Join(cgroupRoot, cgroupPath)) && autoCreate { //若不存在且需要创建
		if err := os.Mkdir(path.Join(cgroupRoot, cgroupPath), 0755); err == nil {
		} else {
			return "", fmt.Errorf("create cgroup error->%v", err)
		}
	}
	return path.Join(cgroupRoot, cgroupPath), nil
}
