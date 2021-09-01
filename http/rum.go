package http

import (
	"fmt"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"

	influxm "github.com/influxdata/influxdb1-client/models"
)

var (
	rumMetricNames = map[string]bool{
		`view`:      true,
		`resource`:  true,
		`error`:     true,
		`long_task`: true,
		`action`:    true,
	}
)

func geoTags(srcip string) (tags map[string]string) {
	tags = map[string]string{}

	ipInfo, err := geo.Geo(srcip)

	l.Debugf("ipinfo(%s): %+#v", srcip, ipInfo)

	if err != nil {
		l.Warnf("geo failed: %s, ignored", err)
		return
	} else {
		// 无脑填充 geo 数据
		tags = map[string]string{
			"city":     ipInfo.City,
			"province": ipInfo.Region,
			"country":  ipInfo.Country_short,
			"isp":      ip2isp.SearchIsp(srcip),
			"ip":       srcip,
		}
	}

	return
}

func handleRUMBody(body []byte, precision, srcip string, isjson bool) ([]*io.Point, error) {
	extags := geoTags(srcip)

	if isjson {
		return jsonPoints(body, precision, extags)
	}

	rumpts, err := lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: extags,
		Strict:    true,

		// 由于 RUM 数据需要分别处理，故用回调函数来区分
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())

			if _, ok := rumMetricNames[name]; !ok {
				return nil, fmt.Errorf("unknow RUM data-type %s", name)
			}

			return p, nil
		},
	})

	if err != nil {
		l.Error(err)
		return nil, uhttp.Error(ErrInvalidLinePoint, err.Error())
	}

	return io.WrapPoint(rumpts), nil
}
