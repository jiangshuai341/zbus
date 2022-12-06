package socket

import (
	"github.com/jiangshuai341/zbus/errors"
	"net"
	"os"
	"syscall"
)

//UDS : Unix domain socket 又叫 IPC(inter-process communication 进程间通信) socket，用于实现同一主机上的进程间通信

func udsSocket(addr string, listen bool, sockOpts ...Option) (fd int, netAddr net.Addr, err error) {
	var (
		family int
		sa     syscall.Sockaddr
	)
	if sa, family, netAddr, err = GetUnixSocket(addr); err != nil {
		err = os.NewSyscallError("socket", err)
		return
	}
	if fd, err = createSockFD(family, syscall.SOCK_STREAM, 0); err != nil {
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

	for _, sockOpt := range sockOpts {
		if err = sockOpt.SetSockOpt(fd, sockOpt.Opt); err != nil {
			return
		}
	}

	if listen {
		if err = os.NewSyscallError("bind", syscall.Bind(fd, sa)); err != nil {
			return
		}
		err = os.NewSyscallError("listen", syscall.Listen(fd, listenerBacklogMaxSize))
	} else {
		err = os.NewSyscallError("connect", syscall.Connect(fd, sa))
	}
	return
}

func GetUnixSocket(addr string) (sa syscall.Sockaddr, family int, unixAddr *net.UnixAddr, err error) {
	unixAddr, err = net.ResolveUnixAddr(string(UDS), addr)
	if err != nil {
		return
	}
	switch unixAddr.Network() {
	case string(UDS):
		sa = &syscall.SockaddrUnix{Name: unixAddr.Name}
		family = syscall.AF_UNIX
	default:
		err = errors.ErrUnsupportedUDSProtocol
	}
	return
}
