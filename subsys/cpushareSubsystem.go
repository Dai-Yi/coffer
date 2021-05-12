package subsys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type CpushareSubsystem struct {
}

//将进程加入到cgroup中
func (s *CpushareSubsystem) Apply(cgroupPath string, pid int) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		//进程pid写入cgroup下task文件中
		if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "tasks"),
			[]byte(strconv.Itoa(pid)), 0644); err != nil {
			return fmt.Errorf("set cgroup proc error->%v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup error->%v", err)
	}
}

//删除对应cgroup
func (s *CpushareSubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath) //删除对应目录即删除对应cgroup
	} else {
		return err
	}
}

//设置cgroupPath对应的cgroup的cpu资源限制
func (s *CpushareSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuShare != "" {
			if err := ioutil.WriteFile(path.Join(subsysCgroupPath, "cpu.shares"),
				[]byte(res.CpuShare), 0644); err != nil {
				return fmt.Errorf("set cgroup cpu share error->%v", err)
			}
		}
		return nil
	} else {
		return err
	}
}

//返回cgroup名字
func (s *CpushareSubsystem) Name() string {
	return "cpu"
}
