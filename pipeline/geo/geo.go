package geo

import (
	"fmt"

	ipL "github.com/ip2location/ip2location-go"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	Db = &ipL.DB{}
	l  = logger.DefaultSLogger("geo")
)

func Geo(ip string) (*ipL.IP2Locationrecord, error) {
	if Db == nil {
		return nil, fmt.Errorf("ip2location db nil")
	}

	record, err := Db.Get_all(ip)
	if err != nil {
		return nil, err
	}

	if record.City == "Invalid IP address." || record.City == "Invalid database file." {
		record.City = "unknown"
	}

	if record.Region == "Invalid IP address." || record.Region == "Invalid database file." {
		record.Region = "unknown"
	}

	if record.Country_short == "Invalid IP address." || record.Region == "Invalid database file." {
		record.Country_short = "unknown"
	}

	return &record, nil
}

func LoadIPLib(f string) error {
	var err error
	if !datakit.FileExist(f) {
		l.Warnf("%v not found", f)
		return nil
	}

	Db, err = ipL.OpenDB(f)
	return err
}
