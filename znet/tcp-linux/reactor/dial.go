package reactor

import "github.com/jiangshuai341/zbus/znet/socket"

func Dial(addr string) *Connection {
	fd, _, err := socket.TCPSocket(socket.TCP, addr, false)
	if err != nil {
		log.Errorf("[Dial] Create TCPSocket Failed Err:%s", err.Error())
		return nil
	}
	conn, err := newTCPConn(fd)
	if err != nil {
		log.Errorf("[Dial] newTCPConn Failed Err:%s", err.Error())
		return nil
	}
	return conn
}
