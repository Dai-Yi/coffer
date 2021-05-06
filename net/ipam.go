package net

import (
	"coffer/container"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

const ipamDefaultAllocatorPath = "/var/run/mydocker/network/ipam/subnet.json"

type IPAM struct { //存放IP地址分配信息
	SubnetAllocatorPath string //分配文件存放位置
	//网段和位图算法的数组map,key是网段,value是分配的位图数组
	Subnets *map[string]string
}

//初始化一个IPAM对象
var ipAllocator = &IPAM{
	SubnetAllocatorPath: ipamDefaultAllocatorPath,
}

//加载网段地址分配信息
func (ipam *IPAM) load() error {
	if !container.PathExists(ipam.SubnetAllocatorPath) {
		return nil
	}
	//打开并读取存储文件
	subnetConfigFile, err := os.Open(ipam.SubnetAllocatorPath)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()
	subnetJson := make([]byte, 2000)
	n, err := subnetConfigFile.Read(subnetJson)
	if err != nil {
		return err
	}
	//将文件中的内容反序列化出IP的分配信息
	err = json.Unmarshal(subnetJson[:n], ipam.Subnets)
	if err != nil {
		return fmt.Errorf("error dump allocation info, %v", err)
	}
	return nil
}

//存储网段地址分配信息
func (ipam *IPAM) dump() error {
	//检查存储文件所在文件夹是否存在,如果不存在则创建
	ipamConfigFileDir, _ := path.Split(ipam.SubnetAllocatorPath)
	if !container.PathExists(ipamConfigFileDir) {
		if err := os.MkdirAll(ipamConfigFileDir, 0644); err != nil {
			return err
		}
	}
	//打开存储文件
	subnetConfigFile, err := os.OpenFile(ipam.SubnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer subnetConfigFile.Close()
	//序列化ipam对象到json字符串
	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}
	//将序列化后的json字符串写入到配置文件
	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

//在网段中分配一个可用的IP地址
func (ipam *IPAM) Allocate(subnet *net.IPNet) (ip net.IP, err error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = &map[string]string{}
	// 从文件中加载已经分配的网段信息
	err = ipam.load()
	if err != nil {
		return nil, fmt.Errorf("error dump allocation info, %v", err)
	}
	_, subnet, _ = net.ParseCIDR(subnet.String())
	//size函数返回网段的子网掩码的总长度和网段前面的固定位长度
	one, size := subnet.Mask.Size()
	//若没有分配过这个网段,则初始化网段的分配配置
	if _, exist := (*ipam.Subnets)[subnet.String()]; !exist {
		//用0填满这个网段的配置
		(*ipam.Subnets)[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
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
	ipam.dump()
	return
}

//释放IP地址
func (ipam *IPAM) Release(subnet *net.IPNet, ipaddr *net.IP) error {
	ipam.Subnets = &map[string]string{}
	//加载网段的分配信息
	_, subnet, _ = net.ParseCIDR(subnet.String())
	err := ipam.load()
	if err != nil {
		return fmt.Errorf("error dump allocation info, %v", err)
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
	ipam.dump()
	return nil
}
