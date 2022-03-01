// Package ipdb define the interface for ip location or isp.
package ipdb

import (
	"fmt"
	"strconv"
	"strings"
)

type IPdb interface {
	Init(dataDir string, config map[string]string)
	Geo(ip string) (*IPdbRecord, error)
	SearchIsp(ip string) string
}

type IPdbRecord struct {
	Country   string
	Region    string
	City      string
	Isp       string
	Latitude  float32
	Longitude float32
	Timezone  string
	Areacode  string
}

func ParseIPCIDR(ipCidr string) (string, error) {
	var err error
	var cidrLen int64 = 32

	ipCidrs := strings.Split(ipCidr, "/")
	if len(ipCidrs) == 2 {
		cidrLen, err = strconv.ParseInt(ipCidrs[1], 10, 8)
		if err != nil {
			return "", err
		}
	}

	ipBytes := strings.Split(ipCidrs[0], ".")
	if len(ipBytes) != 4 {
		return "", fmt.Errorf("invalid ip address")
	}
	ipBitStr := ""
	for _, ipByteStr := range ipBytes {
		ip, err := strconv.ParseInt(ipByteStr, 10, 16)
		if err != nil {
			return "", err
		}
		if ip >= int64(len(BitConvTemplate)) {
			return "", fmt.Errorf("invalid ip number: %s", ipByteStr)
		}
		if cidrLen >= 8 {
			ipBitStr += BitConvTemplate[ip]
		} else {
			ipBitStr += BitConvTemplate[ip][0:cidrLen]
		}
		cidrLen -= 8
		if cidrLen <= 0 {
			break
		}
	}
	return ipBitStr, nil
}
