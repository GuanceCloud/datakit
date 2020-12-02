package process

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/geo"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type Procedure struct {
	*influxdb.Point
}

var (
	Log *logger.Logger
)

func (p *Procedure) GetPoint() *influxdb.Point {
	return p.Point
}

func (p *Procedure) GetByte() []byte {
	return []byte(p.Point.String())
}

func (p *Procedure) GetString() string {
	return p.Point.String()
}

func NewProcedure(pt *influxdb.Point) *Procedure {
	return &Procedure{
		pt,
	}
}

func (p *Procedure) Geo(ip string) *Procedure {
	ipLocInfo, err := geo.Geo(ip)
	if err != nil {
		Log.Errorf("ProcessGeo err: %v", err)
	} else {
		p.Point.Tags()["country"]  = ipLocInfo.Country_short
		p.Point.Tags()["province"] = ipLocInfo.Region
		p.Point.Tags()["city"]     = ipLocInfo.City
		f, _ := p.Point.Fields()
		f["ip"] = ip
	}
	return p
}

func init() {
	Log = logger.SLogger("process")
}


