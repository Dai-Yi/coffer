package net

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/vishvananda/netlink"
)

type BridgeNetworkDriver struct {
}

func (d *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	//获取网段的字符串中的网关IP地址和网络IP段
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	//初始化网络对象
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}
	//配置网桥
	err := d.initBridge(n)
	if err != nil {
		return nil, fmt.Errorf("error init bridge: %v", err)
	}
	//返回配置好的网络
	return n, err
}

func (d *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	return netlink.LinkDel(br)
}

//连接一个网络和网络端点
func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	//获取网络名,即Bridge接口名
	bridgeName := network.Name
	//通过接口名获取到Bridge接口的对象和接口属性
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}
	//创建Veth接口的配置
	la := netlink.NewLinkAttrs()
	//Linux接口名有限制,所以名字去endpoint ID的前五位
	la.Name = endpoint.ID[:5]
	//设置Veth的一端挂载到网络对应的Bridge上
	la.MasterIndex = br.Attrs().Index
	//创建Veth对象,通过PeerName配置Veth另外一端的接口名
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.ID[:5],
	}
	//创建Veth接口
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	//设置Veth启动
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	return nil
}

func (d *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
}

//初始化Bridge设备
func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	//创建Bridge虚拟设备
	bridgeName := n.Name
	if err := createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("error add bridge： %s, Error: %v", bridgeName, err)
	}
	//设置Bridge设备的地址和路由
	gatewayIP := *n.IpRange
	gatewayIP.IP = n.IpRange.IP
	if err := setInterfaceIP(bridgeName, gatewayIP.String()); err != nil {
		return fmt.Errorf("error assigning address: %s on bridge: %s with an error of: %v", gatewayIP, bridgeName, err)
	}
	//启动Bridge设备
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("error set bridge up: %s, Error: %v", bridgeName, err)
	}
	//设置iptables的SNAT规则
	if err := setupIPTables(bridgeName, n.IpRange); err != nil {
		return fmt.Errorf("error setting iptables for %s: %v", bridgeName, err)
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
		return fmt.Errorf("getting link with name %s failed: %v", bridgeName, err)
	}
	//删除网络对应的Bridge设备
	if err := netlink.LinkDel(l); err != nil {
		return fmt.Errorf("failed to remove bridge interface %s delete: %v", bridgeName, err)
	}
	return nil
}

//创建Linux Bridge设备
func createBridgeInterface(bridgeName string) error {
	//先检查释放已经存在了Bridge设备
	_, err := net.InterfaceByName(bridgeName)
	//如果已经存在了或者错误则返回
	if err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return err
	}
	//初始化一个netlink的Link对象
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	//使用Link的属性创建netlink的Bridge对象
	br := &netlink.Bridge{LinkAttrs: la}
	//创建Bridge虚拟网络设备
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge creation failed for bridge %s: %v", bridgeName, err)
	}
	return nil
}

//设置网络接口为UP状态
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [ %s ]: %v", iface.Attrs().Name, err)
	}
	//设置接口状态为UP状态
	if err := netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error enabling interface for %s: %v", interfaceName, err)
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
		// log.Debugf("error retrieving new bridge netlink link [ %s ]... retrying", name)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v", err)
	}
	//返回值包含网段信息和ip
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	//给网络接口配置地址
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0, Broadcast: nil}
	return netlink.AddrAdd(iface, addr)
}

//设置iptables对应Bridge的MASQUERADE规则
func setupIPTables(bridgeName string, subnet *net.IPNet) error {
	//通过命令的方式来配置
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	//执行iptables命令配置SNAT规则
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("iptables Output, %v", output)
	}
	return nil
}
