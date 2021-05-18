package subsys

import (
	"coffer/utils"
	"fmt"
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
		memoryTaskFile, err := os.OpenFile(path.Join(subsysCgroupPath, "tasks"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("open file %s error->%v", memoryTaskFile.Name(), err)
		}
		defer memoryTaskFile.Close()
		utils.Lock(memoryTaskFile)
		_, err = memoryTaskFile.Write([]byte(strconv.Itoa(pid)))
		utils.UnLock(memoryTaskFile)
		if err != nil {
			return fmt.Errorf("set cgroup proc error->%v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error->%v", cgroupPath, err)
	}
}

//删除对应cgroup
func (s *MemorySubsystem) Remove(cgroupPath string) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		return os.RemoveAll(subsysCgroupPath) //删除对应目录即删除对应cgroup
	} else {
		return fmt.Errorf("get cgroup path error->%v", err)
	}
}

//设置cgroupPath对应的cgroup的内存资源限制
func (s *MemorySubsystem) Set(cgroupPath string, res *ResourceConfig) error {
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, true); err == nil {
		if res.MemoryLimit != "" {
			//将限制写入到memory.limit_in_bytes即可实现限制内存
			memoryFile, err := os.OpenFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("open file %s error->%v", memoryFile.Name(), err)
			}
			defer memoryFile.Close()
			utils.Lock(memoryFile)
			_, err = memoryFile.Write([]byte(res.MemoryLimit))
			utils.UnLock(memoryFile)
			if err != nil {
				return fmt.Errorf("set cgroup memory error->%v", err)
			}
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup path error->%v", err)
	}
}

//返回cgroup名字
func (s *MemorySubsystem) Name() string {
	return "memory"
}
