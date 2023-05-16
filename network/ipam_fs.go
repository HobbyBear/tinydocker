package network

import (
	"encoding/binary"
	"encoding/json"
	"net"
	"os"
	"tinydocker/config"
)

type ipAmFs struct {
	subnets map[string]*bitMap
	path    string
}

var IpAmfs = &ipAmFs{
	subnets: make(map[string]*bitMap),
	path:    config.IpAmStorageFsPath,
}

func (ipamfs *ipAmFs) AllocIp(subnet string) (net.IP, error) {
	if err := ipamfs.loadConf(); err != nil {
		return nil, err
	}
	ip, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	ip = ip.To4()
	ones, total := cidr.Mask.Size()
	bitmap := ipamfs.subnets[subnet]
	if bitmap == nil || bitmap.Bitmap == nil {
		bitmap = InitBitMap(2 << (total - ones))
		bitmap.BitSet(1)
		ipamfs.subnets[subnet] = bitmap
	}

	for pos := 0; pos < (total - ones); pos++ {
		if bitmap.BitExist(pos) {
			continue
		}
		bitmap.BitSet(pos)
		for setCnt := 4; setCnt >= 1; setCnt-- {
			[]byte(ip)[4-setCnt] += uint8(pos >> (8 * (setCnt - 1)))
		}
		ip[3] += 1
		break
	}
	err = ipamfs.sync()
	if err != nil {
		return nil, err
	}
	return ip, nil
}

func (ipamfs *ipAmFs) ReleaseIp(subnet string, ip net.IP) error {
	if err := ipamfs.loadConf(); err != nil {
		return err
	}
	bitmap := ipamfs.subnets[subnet]
	if bitmap == nil {
		return nil
	}
	_, cidr, err := net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	pos := getIPIndex(ip, cidr.Mask)
	bitmap.BitClean(pos - 1)
	return ipamfs.sync()
}

func getIPIndex(ip net.IP, mask net.IPMask) int {
	ipInt := ipToUint32(ip)
	firstIP := ipToUint32(ip.Mask(mask))
	return int(ipInt - firstIP)
}
func ipToUint32(ip net.IP) uint32 {
	if ip == nil {
		return 0
	}
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip)
}

func (ipamfs *ipAmFs) loadConf() error {
	if _, err := os.Stat(ipamfs.path); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}
	data, err := os.ReadFile(ipamfs.path)
	if err != nil {
		return err
	}
	if len(ipamfs.subnets) == 0 {
		ipamfs.subnets = make(map[string]*bitMap)
	}
	if len(data) == 0 {
		return nil
	}
	err = json.Unmarshal(data, &ipamfs.subnets)
	if err != nil {
		return err
	}
	return nil
}

func (ipamfs *ipAmFs) sync() error {
	if _, err := os.Stat(ipamfs.path); err != nil {
		if os.IsNotExist(err) {
			os.Create(ipamfs.path)
		} else {
			return err
		}
	}
	data, err := json.Marshal(ipamfs.subnets)
	if err != nil {
		return err
	}
	err = os.WriteFile(ipamfs.path, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
