package ws

// See https://github.com/eranyanay/1m-go-websockets/blob/master/4_optimize_gobwas/epoll.go

import (
	"net"
	"reflect"
)

func websocketFD(conn net.Conn) int {
	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")

	return int(pfdVal.FieldByName("Sysfd").Int())
}
