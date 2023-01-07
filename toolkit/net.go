package toolkit

import (
	"net"
	"strconv"
	"strings"
	"unsafe"
)

var ips map[string]string // netInterfaceName ==> IP

func initNetinterfaceIpMap() error {
	ips = make(map[string]string)
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}
	for _, i := range interfaces {
		byName, err := net.InterfaceByName(i.Name)
		if err != nil {
			return err
		}
		addresses, err := byName.Addrs()
		for _, v := range addresses {
			if strings.Count(v.String(), ":") >= 2 {
				continue
			}
			strs := strings.Split(v.String(), "/")
			if len(strs) < 1 {
				continue
			}
			ips[byName.Name] = strs[0]
		}
	}
	return nil
}
func GetIpByNetCardName(name string) string {
	if ips == nil {
		return ""
	}
	return ips[name]
}

func Ipv4TOUint32(Ip string) (uint32, bool) {
	strs := strings.Split(Ip, ".")
	if len(strs) != 4 {
		return 0, false
	}
	var ret uint32
	ptr := (*[4]uint8)(unsafe.Pointer(&ret))
	for k, v := range strs {
		val, _ := strconv.ParseUint(v, 10, 8)
		ptr[k] = uint8(val)
	}
	return ret, true
}
