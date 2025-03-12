// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package plval

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
)

var _ipdb ipdb.IPdb

const (
	IPInfoUnknow = "unknown"

	IPTypeRemoteAddr = "remote-addr"
	IPTypeIllegal    = "illegal"
	IPTypePrivate    = "private"
	IPTypePublic     = "public"

	LocateStatusGEOSuccess = "success"
	LocateStatusGEOFailure = "failure"
	LocateStatusGEONil     = "nil"
)

// Geo get ip info from global IPDB.
func Geo(ip string) (*ipdb.IPdbRecord, error) {
	if _ipdb != nil {
		return _ipdb.Geo(ip)
	}
	return nil, fmt.Errorf("ipdb not ready")
}

// GetIPDB return global configured IPDB instance.
func GetIPDB() (ipdb.IPdb, bool) {
	if _ipdb == nil {
		return nil, false
	}
	return _ipdb, true
}

// SearchISP query IP's ISP info.
func SearchISP(ip string) string {
	if _ipdb != nil {
		return _ipdb.SearchIsp(ip)
	}
	return "unknown"
}

// IIPQuerier defines interface for IP info querier.
type IIPQuerier interface {
	GetSourceIP(req *http.Request) (string, string)
	Query(string) (*ipdb.IPdbRecord, error)
}

// ipQuerierImpl implement IIPQuerier.
type ipQuerierImpl struct {
	originIPHeader string
}

func (x *ipQuerierImpl) GetSourceIP(req *http.Request) (srcIP, ipType string) {
	return RequestSourceIP(req, x.originIPHeader)
}

func (x *ipQuerierImpl) Query(ip string) (*ipdb.IPdbRecord, error) { return geoTags(ip) }

// NewIPQuerier implement IIPQuerier, it query IP info based on global IPDB.
func NewIPQuerier(header string) IIPQuerier {
	return &ipQuerierImpl{originIPHeader: header}
}

// RequestSourceIP get req source IP and checking it's IP type.
func RequestSourceIP(req *http.Request, httpHeader string) (srcIP, ipType string) {
	if httpHeader != "" {
		srcIP = req.Header.Get(httpHeader)

		l.Debugf("get source ip from %s: %s", httpHeader, srcIP)

		if srcIP == "" {
			l.Debugf("get source IP failed, request headers: %+#v", req.Header)
		}
	}

	if srcIP == "" {
		for _, header := range []string{"X-Forwarded-For", "X-Real-IP"} { // default header keys
			if !strings.EqualFold(header, httpHeader) {
				if val := strings.TrimSpace(req.Header.Get(header)); val != "" {
					srcIP = val
					break
				}
			}
		}
	}

	if srcIP != "" {
		l.Debugf("header remote addr: %s", srcIP)

		parts := strings.Split(srcIP, ",")
		if len(parts) > 0 {
			srcIP = parts[0] // 注意：此处只取第一个 IP 作为源 IP
			netIP := net.ParseIP(srcIP)

			if netIP == nil {
				ipType = IPTypeIllegal
			} else {
				if isPrivateIP(netIP) {
					ipType = IPTypePrivate
				} else {
					ipType = IPTypePublic
				}
			}
			return
		}
	} else { // default use gin client IP
		l.Debugf("gin remote addr: %s", req.RemoteAddr)

		ipType = IPTypeRemoteAddr
		host, _, err := net.SplitHostPort(req.RemoteAddr)
		if err == nil {
			srcIP = host
			return
		} else {
			l.Warnf("net.SplitHostPort(%s): %s, ignored", req.RemoteAddr, err)
		}
	}

	return srcIP, ipType
}

func resetSpecificIPInfo(ipInfo *ipdb.IPdbRecord) *ipdb.IPdbRecord {
	switch ipInfo.Country { // #issue 354
	case "TW":
		ipInfo.Country = "CN"
		ipInfo.Region = "Taiwan"
	case "MO":
		ipInfo.Country = "CN"
		ipInfo.Region = "Macao"
	case "HK":
		ipInfo.Country = "CN"
		ipInfo.Region = "Hong Kong"
	}

	return ipInfo
}

func geoTags(srcip string) (*ipdb.IPdbRecord, error) {
	ipInfo, err := Geo(srcip)
	if err != nil {
		return nil, err
	}

	if ipInfo == nil {
		return nil, nil
	}

	ipInfo = resetSpecificIPInfo(ipInfo)

	if isp := SearchISP(srcip); len(isp) > 0 {
		ipInfo.Isp = isp
	}

	return ipInfo, nil
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	if dotIP := ip.To4(); dotIP != nil {
		switch {
		case dotIP[0] == 10:
			return true
		case dotIP[0] == 172 && dotIP[1] >= 16 && dotIP[1] <= 31:
			return true
		case dotIP[0] == 192 && dotIP[1] == 168:
			return true
		}
	}

	return false
}
