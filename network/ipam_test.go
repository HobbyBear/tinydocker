package network

import (
	"fmt"
	"net"
	"testing"
)

func TestAlloc(t *testing.T) {
	subnet := "192.168.0.0/24"
	ip, err := IpAmfs.AllocIp(subnet)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ip.To4().String())
	ip, err = IpAmfs.AllocIp(subnet)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ip.To4().String())
	//err = IpAmfs.ReleaseIp(subnet, ip)
	//if err != nil {
	//	t.Fatal(err)
	//}
	ip, err = IpAmfs.AllocIp(subnet)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ip.To4().String())
}

func TestBitMap_BitClean(t *testing.T) {
	ip, cidr, err := net.ParseCIDR("192.168.0.1/24")
	if err != nil {
		t.Fatal(err)
	}
	ip = ip.To4()
	ones, total := cidr.Mask.Size()
	fmt.Println(total - ones)
	fmt.Println(1 << (total - ones))
	//bitM := InitBitMap(10)
	//bitM.BitSet(5)
	//fmt.Println(bitM.BitExist(5))
	//bitM.BitClean(5)
	//fmt.Println(bitM.BitExist(5))
}
