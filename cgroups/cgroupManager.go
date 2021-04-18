package cgroups

import (
	"coffer/log"
	"coffer/subsys"
)

type CgroupManager struct {
	CgroupPath string
	Resource   *subsys.ResourceConfig
}

func (c *CgroupManager) Apply(pid int) error { //应用
	for _, subsystem := range subsys.SubsystemsList {
		if err := subsystem.Apply(c.CgroupPath, pid); err != nil { //调用每个subsystem的apply方法
			log.Logout("ERROR", "Apply cgroup error:"+err.Error())
		}
	}
	return nil
}
func (c *CgroupManager) Destroy() error { //删
	for _, subsystem := range subsys.SubsystemsList {
		if err := subsystem.Remove(c.CgroupPath); err != nil { //调用每个subsystem的remove方法
			log.Logout("ERROR", "Remove cgroup error:"+err.Error())
		}
	}
	return nil
}
func (c *CgroupManager) Set(res *subsys.ResourceConfig) error { //改
	for _, subsystem := range subsys.SubsystemsList {
		if err := subsystem.Set(c.CgroupPath, res); err != nil { //调用每个subsystem的set方法
			log.Logout("ERROR", "Set cgroup error:"+err.Error())
		}
	}
	return nil
}
func (c *CgroupManager) Name() error { //查
	var temp []string
	for _, subsystem := range subsys.SubsystemsList {
		temp = append(temp, subsystem.Name()) //调用每个subsystem的Name方法
	}
	log.Logout("INFO", []string{"subsystems name:"}, temp)
	return nil
}
