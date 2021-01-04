package process

import (
	influxdb "github.com/influxdata/influxdb1-client/v2"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/ip2isp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/process/parser"
)

type Procedure struct {
	*influxdb.Point
	Content   []byte
	lastErr   error
}

var (
	l = logger.DefaultSLogger("process")
)

func (p *Procedure) LastError() error {
	return p.lastErr
}

func (p *Procedure) GetPoint() *influxdb.Point {
	return p.Point
}

func (p *Procedure) GetByte() []byte {
	if p.Point != nil {
		return []byte(p.Point.String())
	}
	return p.Content
}

func (p *Procedure) GetString() string {
	if p.Point != nil {
		p.Point.String()
	}

	return string(p.Content)
}

func NewProcedure(pt *influxdb.Point, nodes []parser.Node) *Procedure {
	return &Procedure{
		Point: pt,
	}
}

func (p *Procedure) Geo(ip string) *Procedure {
	if p.lastErr != nil {
		return p
	}

	ipLocInfo, err := geo.Geo(ip)
	if err != nil {
		l.Errorf("Geo err: %v", err)
	} else {
		tags := p.Tags()
		tags["isp"] = ip2isp.SearchIsp(ip)
		tags["country"] = ipLocInfo.Country_short
		tags["province"] = ipLocInfo.Region
		tags["city"] = ipLocInfo.City

		f, _ := p.Point.Fields()
		f["ip"] = ip

		newPoint, err := influxdb.NewPoint(p.Point.Name(), tags, f, p.Time())
		if err != nil {
			p.lastErr = err
		} else {
			p.Point = newPoint
		}

		l.Debugf("%v %v", ip, p.String())
	}
	return p
}
