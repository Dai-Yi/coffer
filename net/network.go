package net

import (
	"coffer/container"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"text/tabwriter"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

var (
	defaultNetworkPath = "/var/run/mydocker/network/network/"
	drivers            = map[string]NetworkDriver{}
	networks           = map[string]*Network{}
)

type Endpoint struct {
	ID          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	Network     *Network
	PortMapping []string
}

type Network struct {
	Name    string     //网络名
	IpRange *net.IPNet //地址段
	Driver  string     //网络驱动名
}

type NetworkDriver interface {
	Name() string                                         //驱动名
	Create(subnet string, name string) (*Network, error)  //创建网络
	Delete(network Network) error                         //删除网络
	Connect(network *Network, endpoint *Endpoint) error   //连接容器网络端点到网络
	Disconnect(network Network, endpoint *Endpoint) error //从网络上移除容器网络端点
}

//创建网络
func CreateNetwork(driver, subnet, name string) error {
	//将网段字符串转换成net.IPNet对象
	_, cidr, _ := net.ParseCIDR(subnet)
	//通过IPAM分配网关IP,获取到网段中第一个IP作为网关IP
	ip, err := ipAllocator.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip
	//调用指定的网络驱动创建网络,
	nw, err := drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}
	//保存网络信息
	return nw.dump(defaultNetworkPath)
}

func ListNetwork() error {
	//使用tabwriter显示网络
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	//遍历网络信息
	for _, nw := range networks {
		fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	//输出到标准输出
	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush error %v", err)
	}
	return nil
}

//删除网络
func DeleteNetwork(networkName string) error {
	//查找网络是否存在
	nw, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}
	//释放网络网关的IP
	if err := ipAllocator.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("error Remove Network gateway ip: %s", err)
	}
	//调用网络驱动删除网络创建的设备与配置
	if err := drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("error Remove Network DriverError: %s", err)
	}
	//从网络的配置目录中删除该网络对应的配置文件
	return nw.remove(defaultNetworkPath)
}

//保存网络配置
func (nw *Network) dump(dumpPath string) error {
	if !container.PathExists(dumpPath) { //检查保存的目录是否存在,不存在则创建
		if err := os.MkdirAll(dumpPath, 0644); err != nil {
			return err
		}
	}
	//保存的文件名为网络名
	nwPath := path.Join(dumpPath, nw.Name)
	//打开文件用于写入,参数为:存在内容则清空,只写入,不存在则创建
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error：%v", err)
	}
	defer nwFile.Close()
	//json序列化网络对象到json字符串
	nwJson, err := json.Marshal(nw)
	if err != nil {
		return fmt.Errorf("error：%v", err)
	}
	//写入网络配置
	_, err = nwFile.Write(nwJson)
	if err != nil {
		return fmt.Errorf("error：%v", err)
	}
	return nil
}

//删除网络对应的配置文件
func (nw *Network) remove(dumpPath string) error {
	path := path.Join(dumpPath, nw.Name)
	if container.PathExists(path) {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

//读取网络配置
func (nw *Network) load(dumpPath string) error {
	//打开配置文件
	nwConfigFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer nwConfigFile.Close()
	//从配置文件中读取网络配置的json字符串
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}
	//json字符串反序列化为网络
	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		return fmt.Errorf("error load nw info,%v", err)
	}
	return nil
}

func Init() error {
	//加载网络驱动
	var bridgeDriver = BridgeNetworkDriver{}
	drivers[bridgeDriver.Name()] = &bridgeDriver
	//判断网络的配置目录是否存在,不存在则创建
	if _, err := os.Stat(defaultNetworkPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(defaultNetworkPath, 0644)
		} else {
			return err
		}
	}
	//检查网络配置目录中的所有文件
	filepath.Walk(defaultNetworkPath, func(nwPath string, info os.FileInfo, err error) error {
		if strings.HasSuffix(nwPath, "/") { //如果是目录则跳过
			return nil
		}
		_, nwName := path.Split(nwPath) //加载文件名作为网络名
		nw := &Network{
			Name: nwName,
		}
		//加载网络配置信息
		if err := nw.load(nwPath); err != nil {
			return fmt.Errorf("error load network: %s", err)
		}
		networks[nwName] = nw
		return nil
	})
	log.SetPrefix("[INFO]")
	log.Println("networks:", networks)
	return nil
}

//将容器的网络端点加入到容器的网络空间中
func enterContainerNetns(enLink *netlink.Link, cinfo *container.ContainerInfo) func() {
	//找到容器的net Namespace
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cinfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Println("Error get container net namespace:", err.Error())
	}
	//取到文件描述符
	nsFD := f.Fd()
	//锁定当前程序所执行的线程
	runtime.LockOSThread()
	// 修改veth peer 另外一端移到容器的namespace中
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		log.SetPrefix("[ERROR]")
		log.Println("Error set link netns:", err.Error())
	}
	// 获取当前的网络namespace
	origns, err := netns.Get()
	if err != nil {
		log.SetPrefix("[ERROR]")
		log.Println("Error get current netns:", err.Error())
	}
	// 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.SetPrefix("[ERROR]")
		log.Println("Error set netns", err.Error())
	}
	//返回之前net Namespace的函数
	return func() {
		netns.Set(origns)        //恢复到上面获取到的Net Namespace
		origns.Close()           //关闭Namespace文件
		runtime.UnlockOSThread() //取消对当前程序的线程锁定
		f.Close()                //关闭Namespace文件
	}
}

//配置容器网络端点的地址和路由
func configEndpointIpAddressAndRoute(ep *Endpoint, cinfo *container.ContainerInfo) error {
	//通过网络端点中Veth的另一端
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	//将容器的网络端点加入到容器的网络空间中
	defer enterContainerNetns(&peerLink, cinfo)()
	//获取到容器的IP地址及网段,用于配置容器内部接口地址
	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress
	//设置容器内Veth端点的IP
	if err = setInterfaceIP(ep.Device.PeerName, interfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}
	//启动容器内的Veth端点
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}
	//设置网络接口为UP状态
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}
	//设置容器内的外部请求都通过容器内的Veth端点访问
	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	//构建要添加的路由数据
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP,
		Dst:       cidr,
	}
	//添加路由到容器的网络空间
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return err
	}
	return nil
}

//配置端口映射
func configPortMapping(ep *Endpoint, cinfo *container.ContainerInfo) error {
	//遍历容器端口映射列表
	for _, pm := range ep.PortMapping {
		//分隔成宿主机的端口和容器的端口
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			log.SetPrefix("ERROR")
			log.Println("Port mapping format error:", pm)
			continue
		}
		//调用命令配置iptables
		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		//执行iptables命令,添加端口映射转发规则
		output, err := cmd.Output()
		if err != nil {
			log.SetPrefix("ERROR")
			log.Println("Iptables Output", output)
			continue
		}
	}
	return nil
}

//连接容器到之前创建的网络
func Connect(networkName string, cinfo *container.ContainerInfo) error {
	//从network字典中取到网络的配置信息
	network, ok := networks[networkName]
	if !ok {
		return fmt.Errorf("no such network: %s", networkName)
	}
	// 从网络的IP段中分配容器IP地址
	ip, err := ipAllocator.Allocate(network.IpRange)
	if err != nil {
		return err
	}
	// 创建网络端点
	ep := &Endpoint{
		ID:          fmt.Sprintf("%s-%s", cinfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cinfo.PortMapping,
	}
	// 调用网络驱动挂载和配置网络端点
	if err = drivers[network.Driver].Connect(network, ep); err != nil {
		return err
	}
	// 到容器的namespace配置容器网络设备IP地址
	if err = configEndpointIpAddressAndRoute(ep, cinfo); err != nil {
		return err
	}
	//配置容器到宿主机的端口映射
	return configPortMapping(ep, cinfo)
}

func Disconnect(networkName string, cinfo *container.ContainerInfo) error {
	return nil
}
