package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
	"tinydocker/log"
)

type bridgeDriver struct {
}

func (b *bridgeDriver) Name() string {
	return "bridge"
}

var BridgeDriver = &bridgeDriver{}

func truncate(maxlen int, str string) string {
	if len(str) <= maxlen {
		return str
	}
	return str[:maxlen]
}

func createBridge(networkName string, interfaceIp *net.IPNet) (string, error) {
	bridgeName := truncate(15, fmt.Sprintf("br-%s", networkName))
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName
	br := &netlink.Bridge{LinkAttrs: la}
	if err := netlink.LinkAdd(br); err != nil {
		return "", fmt.Errorf("bridge creation failed for bridge %s: %s", bridgeName, err)
	}
	addr := &netlink.Addr{IPNet: interfaceIp, Peer: interfaceIp, Label: "", Flags: 0, Scope: 0}
	if err := netlink.AddrAdd(br, addr); err != nil {
		return "", fmt.Errorf("bridge add addr fail %s", err)
	}

	if err := netlink.LinkSetUp(br); err != nil {
		return "", fmt.Errorf("error enabling interface for %s: %v", bridgeName, err)
	}
	return bridgeName, nil
}

// like this ip 192.167.0.100/24
func genInterfaceIp(rawIpWithRange string) (*net.IPNet, error) {
	ipNet, err := netlink.ParseIPNet(rawIpWithRange)
	if err != nil {
		return nil, fmt.Errorf("parse ip fail ip=%+s err=%s", rawIpWithRange, err)
	}
	return ipNet, nil
}

func setSNat(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("set snat fail %s", err)
	}
	return nil
}

func (b *bridgeDriver) CreateNetwork(networkName string, subnet string, networkType networktype) error {

	if networkType != BridgeNetworkType {
		return fmt.Errorf("support bridge network type now ")
	}

	// 检查网络命名是否存在
	if err := NetMgr.LoadConf(); err != nil {
		return fmt.Errorf("netMgr loadConf fail %s", err)
	}
	if _, ok := NetMgr.Storage[networkName]; ok {
		log.Info("exist default network ,will not create new network ")
		return nil
	}

	// 创建网桥
	interfaceIp, err := genInterfaceIp(subnet)
	if err != nil {
		return fmt.Errorf("genInterfaceIp err=%s", err)
	}
	bridgeName, err := createBridge(networkName, interfaceIp)
	if err != nil {
		return fmt.Errorf("createBridge err=%s", err)
	}

	_, cidr, _ := net.ParseCIDR(subnet)

	err = setSNat(bridgeName, cidr)
	if err != nil {
		log.Error("%s", err)
	}
	NetMgr.Storage[networkName] = &NetConf{
		NetworkName: networkName,
		IpRange:     cidr,
		Driver:      BridgeNetworkType.String(),
		BridgeName:  bridgeName,
		BridgeIp:    interfaceIp,
	}
	return NetMgr.Sync()
}

func (b *bridgeDriver) DeleteNetwork(name string) error {
	//TODO implement me
	panic("implement me")
}

func (b *bridgeDriver) CrateVeth(networkName string) (*netlink.Veth, *NetConf, error) {
	// 检查网络命名是否存在
	if err := NetMgr.LoadConf(); err != nil {
		return nil, nil, fmt.Errorf("netMgr loadConf fail %s", err)
	}
	networkConf, ok := NetMgr.Storage[networkName]
	if !ok {
		return nil, nil, fmt.Errorf("name %s network is invalid", networkName)
	}
	br, err := netlink.LinkByName(networkConf.BridgeName)
	if err != nil {
		return nil, nil, fmt.Errorf("link by name fail err=%s", err)
	}
	la := netlink.NewLinkAttrs()
	vethname := truncate(15, "veth-"+strconv.Itoa(10+int(rand.Int31n(10)))+"-"+networkConf.NetworkName)
	la.Name = vethname
	la.MasterIndex = br.Attrs().Index
	// 创建veth设备
	vethLink := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  truncate(15, "cif-"+vethname),
	}
	if err := netlink.LinkAdd(vethLink); err != nil {
		return nil, nil, fmt.Errorf("veth creation failed for bridge %s: %s", networkName, err)
	}

	if err := netlink.LinkSetUp(vethLink); err != nil {
		return nil, nil, fmt.Errorf("error enabling interface for %s: %v", networkName, err)
	}
	return vethLink, networkConf, nil
}

func (b *bridgeDriver) setContainerIp(peerName string, pid int, containerIp net.IP, gateway *net.IPNet) error {
	peerLink, err := netlink.LinkByName(peerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	loLink, err := netlink.LinkByName("lo")
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}

	defer enterContainerNetns(&peerLink, pid)()

	containerVethInterfaceIP := *gateway
	containerVethInterfaceIP.IP = containerIp
	if err = setInterfaceIP(peerName, containerVethInterfaceIP.String()); err != nil {
		return fmt.Errorf("%v,%s", containerIp, err)
	}

	if err := netlink.LinkSetUp(peerLink); err != nil {
		return fmt.Errorf("netlink.LinkSetUp fail  name=%s err=%s", peerName, err)
	}

	if err := netlink.LinkSetUp(loLink); err != nil {
		return fmt.Errorf("netlink.LinkSetUp fail  name=%s err=%s", peerName, err)
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        gateway.IP,
		Dst:       cidr,
	}
	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("router add fail %s", err)
	}

	return nil
}

func enterContainerNetns(vethLink *netlink.Link, pid int) func() {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%d/ns/net", pid), os.O_RDONLY, 0)
	if err != nil {
		fmt.Println(fmt.Errorf("error get container net namespace, %v", err))
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	// 修改veth peer 另外一端移到容器的namespace中
	if err = netlink.LinkSetNsFd(*vethLink, int(nsFD)); err != nil {
		log.Error("error set link netns , %v", err)
	}

	// 获取当前的网络namespace
	origns, err := netns.Get()
	if err != nil {
		log.Error("error get current netns, %v", err)
	}

	// 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		log.Error("error set netns, %v", err)
	}
	return func() {
		netns.Set(origns)
		origns.Close()
		runtime.UnlockOSThread()
		f.Close()
	}
}

// Set the IP addr of a netlink interface
func setInterfaceIP(name string, rawIP string) error {
	retries := 2
	var iface netlink.Link
	var err error
	for i := 0; i < retries; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		fmt.Println(fmt.Errorf("error retrieving new bridge netlink link [ %s ]... retrying", name))
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("Abandoning retrieving the new bridge link from netlink, Run [ ip link ] to troubleshoot the error: %v", err)
	}
	ipNet, err := netlink.ParseIPNet(rawIP)
	if err != nil {
		return err
	}
	addr := &netlink.Addr{IPNet: ipNet, Peer: ipNet, Label: "", Flags: 0, Scope: 0}
	return netlink.AddrAdd(iface, addr)
}
