package socket

import (
	"github.com/jiangshuai341/zbus/toolkit"
	"github.com/jiangshuai341/zbus/zpool"
	"net"
	"syscall"
)

// IP层操作

func setIPv6Only(fd, ipv6only int) error {
	return syscall.SetsockoptInt(fd, syscall.IPPROTO_IPV6, syscall.IPV6_V6ONLY, ipv6only)
}

func ipToSockaddr(family int, ip net.IP, port int, zone string) (syscall.Sockaddr, error) {
	switch family {
	case syscall.AF_INET:
		sa, err := ipToSockAddrInet4(ip, port)
		if err != nil {
			return nil, err
		}
		return &sa, nil
	case syscall.AF_INET6:
		sa, err := ipToSockAddrInet6(ip, port, zone)
		if err != nil {
			return nil, err
		}
		return &sa, nil
	}
	return nil, &net.AddrError{Err: "invalid address family", Addr: ip.String()}
}
func ipToSockAddrInet4(ip net.IP, port int) (syscall.SockaddrInet4, error) {
	if len(ip) == 0 {
		ip = net.IPv4zero
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return syscall.SockaddrInet4{}, &net.AddrError{
			Err: "non ipv4 address", Addr: ip.String(),
		}
	}
	sa := syscall.SockaddrInet4{Port: port}
	copy(sa.Addr[:], ip4)
	return sa, nil
}
func ipToSockAddrInet6(ip net.IP, port int, zone string) (syscall.SockaddrInet6, error) {
	if len(ip) == 0 || ip.Equal(net.IPv4zero) {
		ip = net.IPv6zero
	}
	ip6 := ip.To16()
	if ip6 == nil {
		return syscall.SockaddrInet6{}, &net.AddrError{
			Err: "non ipv6 address", Addr: ip.String(),
		}
	}
	sa := syscall.SockaddrInet6{Port: port}

	copy(sa.Addr[:], ip6)
	iface, err := net.InterfaceByName(zone)
	if err != nil {
		return sa, nil
	}
	sa.ZoneId = uint32(iface.Index)
	return sa, nil
}

var ipv4InIPv6Prefix = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff}

func sockaddrInet4ToIP(sa *syscall.SockaddrInet4) net.IP {
	ip := zpool.Get2(16)
	// ipv4InIPv6Prefix
	copy(ip[0:12], ipv4InIPv6Prefix)
	copy(ip[12:16], sa.Addr[:])
	return ip
}

func sockaddrInet6ToIPAndZone(sa *syscall.SockaddrInet6) (net.IP, string) {
	ip := zpool.Get2(16)
	copy(ip, sa.Addr[:])
	return ip, ip6ZoneToString(int(sa.ZoneId))
}

func ip6ZoneToString(zone int) string {
	if zone == 0 {
		return ""
	}
	if ifi, err := net.InterfaceByIndex(zone); err == nil {
		return ifi.Name
	}
	return int2decimal(uint(zone))
}

func int2decimal(i uint) string {
	if i == 0 {
		return "0"
	}
	b := zpool.Get2(32)
	bp := len(b)
	for ; i > 0; i /= 10 {
		bp--
		b[bp] = byte(i%10) + '0'
	}
	return toolkit.BytesToString(b[bp:])
}
