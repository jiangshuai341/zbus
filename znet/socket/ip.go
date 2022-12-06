package socket

import (
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
