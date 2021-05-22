package net

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct { //Bridge网络驱动
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (bridgeDriver *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	//获取网段的字符串中的网关IP地址和网络IP段
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	//初始化网络对象
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  bridgeDriver.Name(),
	}
	//配置网桥
	err := bridgeDriver.initBridge(n)
	if err != nil {
		return nil, fmt.Errorf("init bridge error->%v", err)
	}
	//返回配置好的网络
	return n, err
}

func (bridgeDriver *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("find bridge error->%v", err)
	}
	return netlink.LinkDel(br)
}

//连接一个网络和网络端点
func (bridgeDriver *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	//获取网络名,即Bridge接口名
	bridgeName := network.Name
	//通过接口名获取到Bridge接口的对象和接口属性
	bridge, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("find bridge error->%v", err)
	}
	//创建Veth接口的配置
	linkAttr := netlink.NewLinkAttrs()
	//Linux接口名有限制,所以名字取endpoint ID的后五位
	linkAttr.Name = "veth-" + endpoint.ID[:5]
	//设置Veth的一端挂载到网络对应的Bridge上
	linkAttr.MasterIndex = bridge.Attrs().Index
	//创建Veth对象,通过PeerName配置Veth在容器一端的接口名
	//默认网络配置名为eth0
	endpoint.Device = netlink.Veth{
		LinkAttrs: linkAttr,
		PeerName:  "eth-" + endpoint.ID[:5],
	}
	//创建Veth接口
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("add dndpoint device error->%v", err)
	}
	//设置Veth启动
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("add endpoint device error->%v", err)
	}
	return nil
}

//初始化Bridge设备
func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	//创建Bridge虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("add bridge %s error->%v", bridgeName, err)
	}
	//设置Bridge设备的地址和路由
	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("assigning address %s on bridge %s error->%v", gatewayIP, bridgeName, err)
	}
	//启动Bridge设备
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("set bridge %s up error->%v", bridgeName, err)
	}
	//设置iptables的SNAT规则
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return fmt.Errorf("set iptables for %s error->%v", bridgeName, err)
	}
	return nil
}

// 删除Bridge设备
func (d *BridgeNetworkDriver) DeleteBridge(n *Network) error {
	//网络名即为Linux Bridge设备名
	bridgeName := n.Name
	//找到网络对应的设备
	l, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("get link with name %s error->%v", bridgeName, err)
	}
	//删除网络对应的Bridge设备
	if err := netlink.LinkDel(l); err != nil {
		return fmt.Errorf("remove bridge %s error->%v", bridgeName, err)
	}
	return nil
}

//创建Linux Bridge设备
func createBridgeInterface(bridgeName string) error {
	//先检查释放已经存在了Bridge设备
	_, err := net.InterfaceByName(bridgeName)
	//如果已经存在了或者错误则返回
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return nil
	}
	//初始化一个netlink的Link对象
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	//使用Link的属性创建netlink的Bridge对象
	br := &netlink.Bridge{LinkAttrs: la}
	//创建Bridge虚拟网络设备
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("create bridge %s error->%v", bridgeName, err)
	}
	return nil
}

//设置网络接口为UP状态
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("retrieving %s error->%v", iface.Attrs().Name, err)
	}
	//设置接口状态为UP状态
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("enabling %s interface error->%v", interfaceName, err)
	}
	return nil
}

//设置网络接口的IP地址
func setInterfaceIP(name string, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		//找到需要设置的网络接口
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge link from netlink, Run \"ip link\" to troubleshoot the error->%v", err)
	}
	//返回值包含网段信息和ip
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return fmt.Errorf("prase ip error->%v", err)
	}
	//给网络接口配置地址
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0, Broadcast: nil}
	return netlink.AddrAdd(iface, addr)
}

////通过命令的方式来配置防火墙
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	//从网桥上流入和流出的包都允许转发，-A加入新规则，-i匹配从这块网卡流入的数据，-o匹配从这块网卡流出的数据
	iptablesCmd1 := "-P FORWARD ACCEPT"
	cmd1 := exec.Command("iptables", strings.Split(iptablesCmd1, " ")...)
	//执行iptables命令
	output1, err1 := cmd1.Output()
	if err1 != nil {
		return fmt.Errorf("iptables Output:%v", output1)
	}
	//只要是从网桥上出来的包，都对其做源IP的转换。-s源地址,!表示该ip除外，-o指定出口网卡 -j指定动作
	iptablesCmd2 := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE",
		subnet.String(), bridgeName)
	cmd2 := exec.Command("iptables", strings.Split(iptablesCmd2, " ")...)
	//执行iptables命令配置SNAT规则
	output2, err2 := cmd2.Output()
	if err2 != nil {
		return fmt.Errorf("iptables Output:%v", output2)
	}
	return nil
}
