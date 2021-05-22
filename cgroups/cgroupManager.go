package cgroups

import (
	"coffer/subsys"
	"coffer/utils"
	"fmt"
)

type CgroupManager struct {
	cgroupPath    string
	subsystemList []subsys.Subsystem
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		cgroupPath: path,
		subsystemList: []subsys.Subsystem{ //子系统列表
			&subsys.CpusetSubsystem{},   //Cpuset子系统
			&subsys.MemorySubsystem{},   //Memory子系统
			&subsys.CpushareSubsystem{}, //Cpushare子系统
		},
	}
}
func (c *CgroupManager) Apply(pid int) error { //应用
	for _, subsystem := range c.subsystemList {
		if err := subsystem.Apply(c.cgroupPath, pid); err != nil { //调用每个subsystem的apply方法
			return fmt.Errorf("apply cgroup error->%v", err)
		}
	}
	return nil
}
func (c *CgroupManager) Destroy() { //删
	for _, subsystem := range c.subsystemList {
		if err := subsystem.Remove(c.cgroupPath); err != nil { //调用每个subsystem的remove方法
			utils.Logout("ERROR", "remove cgroup error->", err.Error())
		}
	}
}
func (c *CgroupManager) Set(res *subsys.ResourceConfig) error { //改
	for _, subsystem := range c.subsystemList {
		if err := subsystem.Set(c.cgroupPath, res); err != nil { //调用每个subsystem的set方法
			return fmt.Errorf("set cgroup error->%v", err)
		}
	}
	return nil
}
