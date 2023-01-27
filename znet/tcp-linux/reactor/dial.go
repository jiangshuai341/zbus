package reactor

import "github.com/jiangshuai341/zbus/znet/socket"

func Dial(url string) *Connection {
	fd, err := socket.AutoConnect(url)
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
