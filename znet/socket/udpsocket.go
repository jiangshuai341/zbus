package socket

import (
	"errors"
	"net"
	"os"
	"syscall"
)

func udpSocket(protoType ProtoType, addr string, sockOpts ...Option) (fd int, netAddr net.Addr, err error) {
	var (
		family   int
		ipv6only bool
		sa       syscall.Sockaddr
	)
	if sa, family, netAddr, ipv6only, err = getUDPSockAddr(protoType, addr); err != nil {
		return
	}
	if fd, err = createSockFD(family, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP); err != nil {
		err = os.NewSyscallError("socket", err)
		return
	}

	defer func() {
		if err != nil {
			if err, ok := err.(*os.SyscallError); ok && err.Err == syscall.EINPROGRESS {
				return
			}
			_ = syscall.Close(fd)
		}
	}()

	if family == syscall.AF_INET6 && ipv6only {
		if err = setIPv6Only(fd, 1); err != nil {
			return
		}
	}
	if err = os.NewSyscallError("setsockopt", syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_BROADCAST, 1)); err != nil {
		return
	}

	for _, sockOpt := range sockOpts {
		if err = sockOpt.SetSockOpt(fd, sockOpt.Opt); err != nil {
			return
		}
	}

	//UDP connect 没什么意义 关闭选项降低系统复杂度
	//err = os.NewSyscallError("connnect", syscall.Connect(fd, sa))
	err = os.NewSyscallError("bind", syscall.Bind(fd, sa))

	return
}

func getUDPSockAddr(protoType ProtoType, addr string) (sa syscall.Sockaddr, family int, udpAddr *net.UDPAddr, ipv6only bool, err error) {
	udpAddr, err = net.ResolveUDPAddr(string(protoType), addr)
	if err != nil {
		return
	}
	udpVersion, err := determineUDPProto(protoType, udpAddr)
	if err != nil {
		return
	}

	switch udpVersion {
	case UDP4:
		family = syscall.AF_INET
		sa, err = ipToSockaddr(family, udpAddr.IP, udpAddr.Port, "")
	case UDP6:
		ipv6only = true
		fallthrough
	case UDP:
		family = syscall.AF_INET6
		sa, err = ipToSockaddr(family, udpAddr.IP, udpAddr.Port, udpAddr.Zone)
	default:
		err = errors.New("only udp/udp4/udp6 are supported")
	}
	return
}
func determineUDPProto(protoType ProtoType, addr *net.UDPAddr) (ProtoType, error) {
	if addr.IP.To4() != nil {
		return UDP4, nil
	}
	if addr.IP.To16() != nil {
		return UDP6, nil
	}
	switch protoType {
	case UDP, UDP4, UDP6:
		return protoType, nil
	}
	return "", errors.New("only udp/udp4/udp6 are supported")
}
