package process

import (
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/ip2isp"
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
		Log.Errorf("Geo err: %v", err)
	} else {
		tags := p.Tags()
		tags["isp"]      = ip2isp.SearchIsp(ip)
		tags["country"]  = ipLocInfo.Country_short
		tags["province"] = ipLocInfo.Region
		tags["city"]     = ipLocInfo.City

		f, _ := p.Point.Fields()
		f["ip"] = ip

		newPoint, err := influxdb.NewPoint(p.Point.Name(), tags, f, p.Time())
		if err != nil {
			Log.Errorf("New influxdb Point err: %v", err)
		} else {
			p.Point = newPoint
		}

		Log.Debugf("%v %v", ip, p.String())
	}
	return p
}

func init() {
	Log = logger.SLogger("process")
}


