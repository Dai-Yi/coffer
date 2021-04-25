package cgroups

import (
	"coffer/log"
	"coffer/subsys"
	"fmt"
)

type CgroupManager struct {
	CgroupPath string
	Resource   *subsys.ResourceConfig
}

func (c *CgroupManager) Apply(pid int) error { //应用
	for _, subsystem := range subsys.SubsystemsList {
		if err := subsystem.Apply(c.CgroupPath, pid); err != nil { //调用每个subsystem的apply方法
			return fmt.Errorf("apply cgroup error:%v", err)
		}
	}
	return nil
}
func (c *CgroupManager) Destroy() error { //删
	for _, subsystem := range subsys.SubsystemsList {
		if err := subsystem.Remove(c.CgroupPath); err != nil { //调用每个subsystem的remove方法
			return fmt.Errorf("remove cgroup error:%v", err)
		}
	}
	return nil
}
func (c *CgroupManager) Set(res *subsys.ResourceConfig) error { //改
	for _, subsystem := range subsys.SubsystemsList {
		if err := subsystem.Set(c.CgroupPath, res); err != nil { //调用每个subsystem的set方法
			return fmt.Errorf("set cgroup error:%v", err)
		}
	}
	return nil
}
func (c *CgroupManager) Name() { //查
	var temp []string
	for _, subsystem := range subsys.SubsystemsList {
		temp = append(temp, subsystem.Name()) //调用每个subsystem的Name方法
	}
	log.Logout("INFO", "subsystems name:", temp)
}
