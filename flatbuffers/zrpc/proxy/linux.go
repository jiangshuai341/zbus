//go:build linux

package proxy

import (
	"github.com/jiangshuai341/zbus/toolkit"
	"github.com/jiangshuai341/zbus/znet/socket"
	"os"
	"strconv"
)

func addListenUDS() error {
	port, err := toolkit.GetFreePort()
	if err != nil {
		return os.NewSyscallError("[addListenUDS] GetFreePort", err)
	}
	udsFd, _, err := socket.UDSSocket("0.0.0.0:"+strconv.Itoa(port), true)
	if err != nil {
		return os.NewSyscallError("[addListenUDS] Create UDS Socket", err)
	}
	err = accepter.AddListen(udsFd)
	if err != nil {
		return os.NewSyscallError("[addListenUDS] accepter AddListen", err)
	}
	return nil
}
