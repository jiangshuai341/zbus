package socket

import (
	"errors"
	"net"
	"os"
	"syscall"
)

func tcpSocket(protoType ProtoType, addr string, listen bool, sockOpts ...Option) (fd int, netAddr net.Addr, err error) {
	var (
		family   int
		ipv6only bool
		sa       syscall.Sockaddr
	)
	if sa, family, netAddr, ipv6only, err = getTCPSockAddr(protoType, addr); err != nil {
		return
	}
	if fd, err = createSockFD(family, syscall.SOCK_STREAM, syscall.IPPROTO_TCP); err != nil {
		err = os.NewSyscallError("tcpSocket create", err)
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
	for _, socketOpt := range sockOpts {
		if err = socketOpt.SetSockOpt(fd, socketOpt.Opt); err != nil {
			return
		}
	}
	if listen {
		if err = os.NewSyscallError("tcpSocket bind", syscall.Bind(fd, sa)); err != nil {
			return
		}
		err = os.NewSyscallError("tcpSocket listen", syscall.Listen(fd, listenerBacklogMaxSize))
	} else {
		err = os.NewSyscallError("tcpSocket connect", syscall.Connect(fd, sa))
	}
	return
}

func getTCPSockAddr(protoType ProtoType, addr string) (sa syscall.Sockaddr, family int, tcpAddr *net.TCPAddr, ipv6only bool, err error) {
	tcpAddr, err = net.ResolveTCPAddr(string(protoType), addr)
	if err != nil {
		return
	}
	tcpVersion, err := determineTCPProto(protoType, tcpAddr)

	switch tcpVersion {
	case TCP4:
		family = syscall.AF_INET
		sa, err = ipToSockaddr(family, tcpAddr.IP, tcpAddr.Port, "")
	case TCP6:
		ipv6only = true
		fallthrough
	case TCP:
		family = syscall.AF_INET6
		sa, err = ipToSockaddr(family, tcpAddr.IP, tcpAddr.Port, tcpAddr.Zone)
	default:
		err = errors.New("only tcp/tcp4/tcp6 are supported")
	}
	return
}

func determineTCPProto(protoType ProtoType, addr *net.TCPAddr) (ProtoType, error) {
	if addr.IP.To4() != nil {
		return TCP4, nil
	}
	if addr.IP.To16() != nil {
		return TCP6, nil
	}
	switch protoType {
	case TCP4, TCP6, TCP:
		return protoType, nil
	}
	return "", errors.New("only tcp/tcp4/tcp6 are supported")
}
