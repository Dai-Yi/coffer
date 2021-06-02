package net

import (
	"coffer/utils"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

type IPAM struct { //IP Adress Management,用于网络IP地址的分配和释放
	SubnetAllocatorPath string //分配文件存放位置
	//网段和位图算法的数组map,key是主机号0-255,value为是否分配0为未分配，1为已分配，可为bool
	Subnets *map[string]string
}

//加载网段地址分配信息
func (ipam *IPAM) load() error {
	//若不存在则说明之前没有分配，不需要加载
	if !utils.PathExists(ipam.SubnetAllocatorPath) {
		return nil
	}
	//打开并读取存储文件
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath) //os.open用于读
	if err != nil {
		return fmt.Errorf("open subnet config file error->%v", err)
	}
	defer subnetConfigFile.Close()
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return fmt.Errorf("read subnet config file error->%v", err)
	}
	//将文件中的内容反序列化出IP的分配信息
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		return fmt.Errorf("json unmarshal error->%v", err)
	}
	return nil
}

//存储网段地址分配信息
func (ipam *IPAM) store() error {
	//检查存储文件所在文件夹是否存在,如果不存在则创建
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath) //分隔路径
	if !utils.PathExists(ipamConfigFileDir) {
		if err := os.MkdirAll(ipamConfigFileDir, 0644); err != nil {
			return fmt.Errorf("make dir %s error->%v", ipamConfigFileDir, err)
		}
	}
	//打开存储文件
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath,
		os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open file %s error->%v", subnetConfigFile.Name(), err)
	}
	defer subnetConfigFile.Close()
	//序列化ipam对象到json字符串
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return fmt.Errorf("marshal ipam error->%v", err)
	}
	utils.Lock(subnetConfigFile)
	//将序列化后的json字符串写入到配置文件
	_, err = subnetConfigFile.Write(ipamConfigJson)
	utils.UnLock(subnetConfigFile)
	if err != nil {
		return fmt.Errorf("write file %s error->%v", subnetConfigFile.Name(), err)
	}
	return nil
}

//从指定网段分配IP地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	//存放网段中地址分配信息的数组赋值为空
	ipam.Subnets = &map[string]string{}
	//从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		return nil, fmt.Errorf("load ipam error->%v", err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	//size函数返回网段的子网掩码的总长度和网段前面的固定位长度
	ones, size := subnet.Mask.Size()
	//若没有分配过这个网段,则初始化网段的分配配置
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		//用0填满这个网段的配置
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-ones))
	}
	//遍历网段的位图数组
	for c := range (*ipam.Subnets)[subnet.String()] {
		//找到数组中为0的项和数组序号,即可分配的IP
		if (*ipam.Subnets)[subnet.String()][c] == '0' {
			ipalloc := []byte((*ipam.Subnets)[subnet.String()])
			ipalloc[c] = '1'
			(*ipam.Subnets)[subnet.String()] = string(ipalloc)
			//此IP为初始IP
			ip = subnet.IP
			for t := uint(4); t > 0; t -= 1 {
				[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
			}
			//由于此处IP从1开始分配,最后再加1
			ip[3] += 1
			break
		}
	}
	//将分配结果保存到文件中
	if err := ipam.store(); err != nil {
		return nil, fmt.Errorf("store ipam error->%v", err)
	}
	return ip, nil
}

//从指定网段中释放IP地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	//加载网段的分配信息
	err := ipam.load()
	if err != nil {
		return fmt.Errorf("load ipam error->%v", err)
	}
	//计算IP地址在网段位图数组中的索引位置
	c := 0
	//将IP地址转换成4字节的表示方式
	releaseIP := ipaddr.To4()
	//由于IP从1开始分配,索引应减1
	releaseIP[3] -= 1
	for t := uint(4); t > 0; t -= 1 {
		c += int(releaseIP[t-1]-subnet.IP[t-1]) << ((4 - t) * 8)
	}
	//将分配的位图数组中索引位置的值置为0
	ipalloc := []byte((*ipam.Subnets)[subnet.String()])
	ipalloc[c] = '0'
	(*ipam.Subnets)[subnet.String()] = string(ipalloc)
	//保存释放掉IP后的网段IP分配信息
	if err := ipam.store(); err != nil {
		return fmt.Errorf("store ipam error->%v", err)
	}
	return nil
}
