//go:build linux
// +build linux

package l4log

import "strings"

type TCPFlag uint8

const (
	TCPFIN TCPFlag = 1 << iota
	TCPSYN
	TCPRST
	TCPPSH
	TCPACK
	TCPURG
	TCPECE
	TCPCWR
)

func (f TCPFlag) String() string {
	var builder strings.Builder

	if f&TCPFIN != 0 {
		builder.WriteString("FIN|")
	}
	if f&TCPSYN != 0 {
		builder.WriteString("SYN|")
	}
	if f&TCPRST != 0 {
		builder.WriteString("RST|")
	}
	if f&TCPPSH != 0 {
		builder.WriteString("PSH|")
	}
	if f&TCPACK != 0 {
		builder.WriteString("ACK|")
	}
	if f&TCPURG != 0 {
		builder.WriteString("URG|")
	}
	if f&TCPECE != 0 {
		builder.WriteString("ECE|")
	}
	if f&TCPCWR != 0 {
		builder.WriteString("CWR|")
	}

	str := builder.String()
	if len(str) > 0 {
		str = str[:len(str)-1] // remove the last '|'
	}

	return str
}

func (f TCPFlag) HasFlag(flag TCPFlag) bool {
	return f&flag != 0
}

func (f *TCPFlag) AppendTCPFlag(flag TCPFlag) {
	*f |= flag
}

func (f *TCPFlag) ReplaceTCPFlag(flag TCPFlag) {
	*f = flag
}
