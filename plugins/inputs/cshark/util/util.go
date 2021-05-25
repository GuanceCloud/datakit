package util

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/cshark/protocol"
	"net"
	"strings"
)

func IsSupport(pro string) bool {
	if len(pro) == 0 {
		return true
	}

	pro = strings.ToUpper(pro)

	if flag, ok := protocol.ProtocolDic[pro]; ok {
		if flag {
			return true
		}
	}

	return false
}

func IsIP(ipstr string) bool {
	address := net.ParseIP(ipstr)
	if address != nil {
		return true
	}

	return false
}
