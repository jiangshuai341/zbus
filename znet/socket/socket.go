package socket

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

type ProtoType string

const (
	TCP4 ProtoType = "tcp4"
	TCP6 ProtoType = "tcp6"
	TCP  ProtoType = "tcp"

	UDP  ProtoType = "udp"  // 同时 ipv4 ipv6
	UDP4 ProtoType = "udp4" // ipv4 only
	UDP6 ProtoType = "udp6" // ipv6 only

	UDS ProtoType = "unix" // IPC
)

// AutoListen
// `tcp://192.168.0.10:9851`
// `unix://socket`.
func AutoListen(url string, sockOpts ...Option) (fd int, err error) {
	network, addr := ParseProtoAddr(url)
	switch network {
	case TCP, TCP4, TCP6:
		fd, _, err = tcpSocket(network, addr, true, sockOpts...)
	case UDP, UDP4, UDP6:
		fd, _, err = udpSocket(network, addr, sockOpts...)
	case UDS:
		fd, _, err = udsSocket(addr, true, sockOpts...)
	}
	return
}

// AutoConnect
// `tcp://0.0.0.0:9851`
// `udp://0.0.0.0:9851`
// `unix:///tmp/temp.sock`.
func AutoConnect(url string, sockOpts ...Option) (fd int, err error) {
	network, addr := ParseProtoAddr(url)
	switch network {
	case TCP, TCP4, TCP6:
		fd, _, err = tcpSocket(network, addr, false, sockOpts...)
	case UDP, UDP4, UDP6:
		err = errors.New("udp not need connect")
	case UDS:
		fd, _, err = udsSocket(addr, false, sockOpts...)
	}
	return
}

func SockaddrToTCPOrUnixAddr(sa syscall.Sockaddr) net.Addr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		ip := sockaddrInet4ToIP(sa)
		return &net.TCPAddr{IP: ip, Port: sa.Port}
	case *syscall.SockaddrInet6:
		ip, zone := sockaddrInet6ToIPAndZone(sa)
		return &net.TCPAddr{IP: ip, Port: sa.Port, Zone: zone}
	case *syscall.SockaddrUnix:
		return &net.UnixAddr{Name: sa.Name, Net: "unix"}
	}
	return nil
}

func SockaddrToUDPAddr(sa syscall.Sockaddr) net.Addr {
	switch sa := sa.(type) {
	case *syscall.SockaddrInet4:
		ip := sockaddrInet4ToIP(sa)
		return &net.UDPAddr{IP: ip, Port: sa.Port}
	case *syscall.SockaddrInet6:
		ip, zone := sockaddrInet6ToIPAndZone(sa)
		return &net.UDPAddr{IP: ip, Port: sa.Port, Zone: zone}
	}
	return nil
}

func ParseProtoAddr(addr string) (network ProtoType, address string) {
	network = "tcp"
	address = strings.ToLower(addr)
	if strings.Contains(address, "://") {
		pair := strings.Split(address, "://")
		network = ProtoType(pair[0])
		address = pair[1]
	}
	return
}
