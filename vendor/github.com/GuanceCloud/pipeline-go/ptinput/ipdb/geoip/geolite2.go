// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package geoip implement ipdb.
package geoip

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/pipeline-go/ptinput/ipdb"
	"github.com/GuanceCloud/pipeline-go/ptinput/utils"
	"github.com/oschwald/geoip2-golang"
)

const (
	INVALIDIP    = "Invalid IP address"
	CfgGeoIPFile = "geoip_file"
	CfgISPFile   = "isp_file"
)

var _ ipdb.IPdb = (*Geoip)(nil)

var (
	l = logger.DefaultSLogger("geoip")

	openDB = func(f string) (*geoip2.Reader, error) {
		db, err := geoip2.Open(f)
		return db, err
	}
)

func InitLog() {
	l = logger.SLogger("geoip")
}

type Geoip struct {
	geo *geoip2.Reader
	isp *geoip2.Reader
}

func (g *Geoip) loadIPLib(geo string, isp string) error {
	if !utils.FileExist(geo) {
		l.Warnf("%v not found", geo)
		return nil
	}

	db, err := openDB(geo)
	if err != nil {
		return err
	} else {
		g.geo = db
	}

	if utils.FileExist(isp) {
		if ispDB, err := openDB(isp); err == nil {
			g.isp = ispDB
		}
	}

	return nil
}

func NewGeoip(dir string, config map[string]string) *Geoip {
	g := Geoip{}

	ipdbFile := "GeoLite2-City.mmdb"

	if file, ok := config[CfgGeoIPFile]; ok {
		if len(file) > 0 {
			ipdbFile = file
		}
	}

	ispFile := "GeoIP2-ISP.mmdb"
	if file, ok := config[CfgISPFile]; ok {
		if len(file) > 0 {
			ispFile = file
		}
	}

	if err := g.loadIPLib(
		filepath.Join(dir, ipdbFile),
		filepath.Join(dir, ispFile)); err != nil {
		l.Warnf("geolite2 load ip lib error: %s", err.Error())
	}

	return &g
}

// Init deprecated
func (g *Geoip) Init(dataDir string, config map[string]string) {
	l.Debug("use geolite2 db")
	dir := filepath.Join(dataDir, "ipdb", "geolite2", "GeoLite2-City_20220617")
	ipdbFile := "GeoLite2-City.mmdb"

	if file, ok := config[CfgGeoIPFile]; ok {
		if len(file) > 0 {
			ipdbFile = file
		}
	}

	var ispFile string

	if file, ok := config[CfgISPFile]; ok {
		if len(file) > 0 {
			ispFile = file
		}
	}

	if err := g.loadIPLib(
		filepath.Join(dir, ipdbFile),
		filepath.Join(dir, ispFile)); err != nil {
		l.Warnf("geolite2 load ip lib error: %s", err.Error())
	}
}

func (g *Geoip) Geo(ip string) (*ipdb.IPdbRecord, error) {
	record := &ipdb.IPdbRecord{}
	if g.geo == nil {
		return record, nil
	}

	ipParse := net.ParseIP(ip)

	r, err := g.get(ipParse)
	if err != nil {
		return record, err
	}
	// ip invalid
	if r == nil {
		record.City = INVALIDIP
		record.Timezone = INVALIDIP
		record.Region = INVALIDIP
		record.Country = INVALIDIP
		record.Isp = INVALIDIP
		return record, nil
	} else {
		record.City = r.City.Names["en"]
		record.Timezone = r.Location.TimeZone
		if len(r.Subdivisions) != 0 {
			record.Region = r.Subdivisions[0].Names["en"]
		}
		record.Country = r.Country.IsoCode
		record.Isp = g.searchISP(ipParse)
		return record.CheckData(), err
	}
}

func (g *Geoip) get(ip net.IP) (*geoip2.City, error) {
	if ip == nil {
		return nil, fmt.Errorf("empty ip")
	}
	if g.geo == nil {
		return nil, fmt.Errorf("GEO DB not set")
	}

	r, err := g.geo.City(ip)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (g *Geoip) searchISP(ip net.IP) string {
	if g.isp != nil && ip != nil {
		if s, err := g.isp.ISP(ip); err == nil {
			return s.ISP
		}
	}
	return "unknown"
}

func (g *Geoip) SearchIsp(ip string) string {
	ipParse := net.ParseIP(ip)
	// ip invalid
	if ipParse == nil {
		return "unknown"
	}

	return g.searchISP(ipParse)
}
