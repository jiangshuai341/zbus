package socket

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var listenerBacklogMaxSize = maxListenerBacklog()

func maxListenerBacklog() int {
	fd, err := os.Open("/proc/sys/net/core/somaxconn")
	if err != nil {
		return syscall.SOMAXCONN
	}
	defer fd.Close()
	rd := bufio.NewReader(fd)
	line, err := rd.ReadString('\n')
	if err != nil {
		return syscall.SOMAXCONN
	}
	f := strings.Fields(line)
	n, err := strconv.Atoi(f[0])
	if err != nil || n == 0 {
		return syscall.SOMAXCONN
	}
	if n > 1<<16-1 {
		n = 1<<16 - 1
	}
	return n
}

//The default buffer size is 8 KB. The maximum size is 8 MB (8096 KB)

func Dup(fd int) (int, string, error) {
	syscall.ForkLock.RLock()
	defer syscall.ForkLock.RUnlock()
	newFD, err := syscall.Dup(fd)
	if err != nil {
		return -1, "dup", err
	}
	syscall.CloseOnExec(newFD)
	return newFD, "", nil
}

// createSockFD 默认 非阻塞 子进程不继承
func createSockFD(family, sotype, proto int) (int, error) {
	return syscall.Socket(family, sotype|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, proto)
}
