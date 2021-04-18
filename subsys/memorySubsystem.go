package subsys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type MemorySubsystem struct {
}

//将进程加入到cgroup中
func (s *MemorySubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		//进程pid写入cgroup下task文件中
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc fail,%v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error:%v", cgroupPath, err)
	}
}

//删除对应cgroup
func (s *MemorySubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath) //删除对应目录即删除对应cgroup
	} else {
		return err
	}
}

//设置cgroupPath对应的cgroup的内存资源限制
func (s *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			//将限制写入到memory.limit_in_bytes即可实现限制内存
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"),
				[]byte(res.MemoryLimit), 0644); err != nil {
				return fmt.Errorf("set cgroup memory fail,%v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

//返回cgroup名字
func (s *MemorySubsystem) Name() string {
	return "memory"
}
