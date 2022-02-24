// Package iploc implement ipdb.
package iploc

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ip2location/ip2location-go"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ipdb"
)

var l = logger.DefaultSLogger("iploc")

type DB interface {
	Get_all(ipaddress string) (ip2location.IP2Locationrecord, error)
}

var openDB = func(f string) (DB, error) {
	db, err := ip2location.OpenDB(f)
	return db, err
}

const (
	FileSeparator = " "

	ErrInvalidIP = "Invalid IP address"
	ErrInvalidDB = "Invalid database file"

	geoDefaultVal = "unknown"
)

type IPloc struct {
	db    DB
	ispDB map[string]string
}

func (iploc *IPloc) Init(dataDir string, config map[string]string) {
	l = logger.SLogger("iploc")
	ipdbDir := filepath.Join(dataDir, "ipdb", "iploc")
	iplocFile := "iploc.bin"
	ispFile := "ip2isp.txt"

	if file, ok := config["iploc_file"]; ok {
		if len(file) > 0 {
			iplocFile = file
		}
	}

	if file, ok := config["isp_file"]; ok {
		if len(file) > 0 {
			ispFile = file
		}
	}

	if err := iploc.loadIPLib(filepath.Join(ipdbDir, iplocFile)); err != nil {
		l.Warnf("iploc load ip lib error: %s", err.Error())
	}

	iploc.ispDB = map[string]string{}

	if err := iploc.loadISP(filepath.Join(ipdbDir, ispFile)); err != nil {
		l.Warnf("isp file load error: %s", err.Error())
	}
}

func (iploc *IPloc) loadIPLib(f string) error {
	if !datakit.FileExist(f) {
		l.Warnf("%v not found", f)
		return nil
	}

	db, err := openDB(f)
	if err != nil {
		return err
	} else {
		iploc.db = db
	}

	return nil
}

func (iploc *IPloc) loadISP(f string) error {
	m := make(map[string]string)

	if !datakit.FileExist(f) {
		l.Warnf("%v not found", f)
		return nil
	}

	fd, err := os.Open(filepath.Clean(f))
	if err != nil {
		return err
	}

	defer fd.Close() //nolint:errcheck,gosec

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		contents := strings.Split(scanner.Text(), FileSeparator)
		if len(contents) != 2 {
			continue
		}

		ipBitStr, err := ipdb.ParseIPCIDR(contents[0])
		if err != nil {
			continue
		}
		m[ipBitStr] = contents[1]
	}

	if len(m) != 0 {
		iploc.ispDB = m
		l.Infof("found new %d rules", len(m))
	} else {
		l.Infof("no rules founded")
	}

	return nil
}

func (iploc *IPloc) SearchIsp(ip string) string {
	if len(iploc.ispDB) == 0 {
		return "unknown"
	}

	for i := 32; i > 0; i-- {
		ipCidr := fmt.Sprintf("%s/%v", ip, i)
		ipBitStr, _ := ipdb.ParseIPCIDR(ipCidr)
		if v, ok := iploc.ispDB[ipBitStr]; ok {
			return v
		}
	}
	return "unknown"
}

func (iploc *IPloc) Geo(ip string) (*ipdb.IPdbRecord, error) {
	record := &ipdb.IPdbRecord{}
	if iploc.db == nil {
		return record, nil
	}

	r, err := iploc.Get(ip)
	if err != nil {
		return nil, err
	}

	switch r.City {
	case ErrInvalidIP, ErrInvalidDB:
		r.City = geoDefaultVal
	}

	switch r.Region {
	case ErrInvalidIP, ErrInvalidDB:
		r.Region = geoDefaultVal
	}

	switch r.Country_short {
	case ErrInvalidIP, ErrInvalidDB:
		r.Country_short = geoDefaultVal
	}

	record.City = r.City
	record.Region = r.Region
	record.Country = r.Country_short

	return record, err
}

func (iploc *IPloc) Get(ip string) (*ip2location.IP2Locationrecord, error) {
	if iploc.db == nil {
		return nil, fmt.Errorf("GEO DB not set")
	}

	r, err := iploc.db.Get_all(ip)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
