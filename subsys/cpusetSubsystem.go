package subsys

import (
	"coffer/utils"
	"fmt"
	"os"
	"path"
	"strconv"
)

type CpusetSubsystem struct {
}

//将进程加入到cgroup中
func (s *CpusetSubsystem) Apply(cgroupPath string, pid int) error { //将进程添加到cgroup
	if subsysCgroupPath, err := GetCgroupPath(s.Name(), cgroupPath, false); err == nil {
		//坑：在cpuset添加tasks之前cpuset.cpus和cpuset.mems需要配置
		cpusetTaskFile, err := os.OpenFile(path.Join(subsysCgroupPath, "tasks"),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("open %v error->%v", cpusetTaskFile.Name(), err)
		}
		defer cpusetTaskFile.Close()
		utils.Lock(cpusetTaskFile)
		_, err = cpusetTaskFile.Write([]byte(strconv.Itoa(pid)))
		utils.UnLock(cpusetTaskFile)
		if err != nil {
			return fmt.Errorf("set cgroup proc error->%v", err)
		}
		return nil
	} else {
		return fmt.Errorf("get cgroup %s error->%v", cgroupPath, err)
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
			cpusetCpusFile, err := os.OpenFile(path.Join(subsysCgroupPath, "cpuset.cpus"),
				os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("open %v error->%v", cpusetCpusFile.Name(), err)
			}
			defer cpusetCpusFile.Close()
			utils.Lock(cpusetCpusFile)
			_, err = cpusetCpusFile.Write([]byte(res.Cpuset.Cpus))
			utils.UnLock(cpusetCpusFile)
			if err != nil {
				return fmt.Errorf("set cgroup cpuset error->%v", err)
			}
		}
		if res.Cpuset.Mems != "" {
			cpusetMemsFile, err := os.OpenFile(path.Join(subsysCgroupPath, "cpuset.mems"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("open %v error->%v", cpusetMemsFile.Name(), err)
			}
			defer cpusetMemsFile.Close()
			utils.Lock(cpusetMemsFile)
			_, err = cpusetMemsFile.Write([]byte(res.Cpuset.Mems))
			utils.UnLock(cpusetMemsFile)
			if err != nil {
				return fmt.Errorf("set cgroup cpuset error->%v", err)
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
