package geo

import (
	"fmt"
	"path/filepath"

	ipL "github.com/ip2location/ip2location-go"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

var (
	Db           *ipL.DB
	Ip2LocDbPath = filepath.Join(datakit.InstallDir, "data", "iploc.bin")
	l            = logger.DefaultSLogger("geo")
)

func Geo(ip string) (ipL.IP2Locationrecord, error) {
	if Db == nil {
		return ipL.IP2Locationrecord{}, fmt.Errorf("ip2location db nil")
	}

	record, err := Db.Get_all(ip)

	if record.City == "Invalid IP address." {
		record.City = "unknown"
	}

	if record.Region == "Invalid IP address." {
		record.Region = "unknown"
	}

	if record.Country_short == "Invalid IP address." {
		record.Country_short = "unknown"
	}

	return record, err
}

func init() {
	var err error
	Db, err = ipL.OpenDB(Ip2LocDbPath)
	if err != nil {
		l.Warnf("Open %v db err %v", Ip2LocDbPath, err)
	}
}
