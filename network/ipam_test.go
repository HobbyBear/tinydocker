package network

import (
	"fmt"
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
	err = IpAmfs.ReleaseIp(subnet, ip)
	if err != nil {
		t.Fatal(err)
	}
	ip, err = IpAmfs.AllocIp(subnet)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ip.To4().String())
}

func TestBitMap_BitClean(t *testing.T) {
	bitM := InitBitMap(10)
	bitM.BitSet(5)
	fmt.Println(bitM.BitExist(5))
	bitM.BitClean(5)
	fmt.Println(bitM.BitExist(5))
}
