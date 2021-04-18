package subsys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpusetSubsystem struct {
}

//将进程加入到cgroup中
func (s *CpusetSubsystem) Apply(cgroupPath string, pid int) error { //将进程添加到cgroup
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		//坑：在添加tasks之前cpuset.cpus和cpuset.mems需要配置
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail,%v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error: %v", cgroupPath, err)
	}
}

//删除对应cgroup
func (s *CpusetSubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath) //删除对应目录即删除对应cgroup
	} else {
		return err
	}
}

//设置cgroupPath对应的cgroup的cpu核心资源限制
func (s *CpusetSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.Cpuset.Cpus != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.cpus"),
				[]byte(res.Cpuset.Cpus), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset fail,%v", err)
			}
		}
		if res.Cpuset.Mems != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpuset.mems"),
				[]byte(res.Cpuset.Mems), 0644); err != nil {
				return fmt.Errorf("set cgroup cpuset fail,%v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

//返回cgroup名字
func (s *CpusetSubsystem) Name() string {
	return "cpuset"
}
