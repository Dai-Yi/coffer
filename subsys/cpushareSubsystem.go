package subsys

import (
	"coffer/utils"
	"fmt"
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
		cpushareTaskFile, err := os.OpenFile(path.Join(subsysCgroupPath, "tasks"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("open file %s error->%v", cpushareTaskFile.Name(), err)
		}
		defer cpushareTaskFile.Close()
		utils.Lock(cpushareTaskFile)
		_, err = cpushareTaskFile.Write([]byte(strconv.Itoa(pid)))
		utils.UnLock(cpushareTaskFile)
		if err != nil {
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
		return fmt.Errorf("get cgroup path error->%v", err)
	}
}

//设置cgroupPath对应的cgroup的cpu资源限制
func (s *CpushareSubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.CpuShare != "" {
			cpusharesFile, err := os.OpenFile(path.Join(subsysCgroupPath, "cpu.shares"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("open file %s error->%v", cpusharesFile.Name(), err)
			}
			defer cpusharesFile.Close()
			utils.Lock(cpusharesFile)
			_, err = cpusharesFile.Write([]byte(res.CpuShare))
			utils.UnLock(cpusharesFile)
			if err != nil {
				return fmt.Errorf("set cgroup cpu share error->%v", err)
			}
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup path error->%v", err)
	}
}

//返回cgroup名字
func (s *CpushareSubsystem) Name() string {
	return "cpu"
}
