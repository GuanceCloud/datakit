package http

import (
	"fmt"
	"time"

	influxm "github.com/influxdata/influxdb1-client/models"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
)

var (
	rumMetricNames = map[string]bool{
		`view`:      true,
		`resource`:  true,
		`error`:     true,
		`long_task`: true,
		`action`:    true,
	}

	rumMetricAppID = "app_id"
)

func geoTags(srcip string) map[string]string {
	ipInfo, err := pipeline.Geo(srcip)

	l.Debugf("ipinfo(%s): %+#v", srcip, ipInfo)

	if err != nil {
		l.Warnf("geo failed: %s, ignored", err)
		return nil
	}

	switch ipInfo.Country_short { // #issue 354
	case "TW":
		ipInfo.Country_short = "CN"
		ipInfo.Region = "Taiwan"
	case "MO":
		ipInfo.Country_short = "CN"
		ipInfo.Region = "Macao"
	case "HK":
		ipInfo.Country_short = "CN"
		ipInfo.Region = "Hong Kong"
	}

	// 无脑填充 geo 数据
	tags := map[string]string{
		"city":     ipInfo.City,
		"province": ipInfo.Region,
		"country":  ipInfo.Country_short,
		"isp":      ip2isp.SearchIsp(srcip),
		"ip":       srcip,
	}

	return tags
}

func doHandleRUMBody(body []byte,
	precision string,
	isjson bool,
	extraTags map[string]string,
	appIDWhiteList []string) ([]*io.Point, error) {
	if isjson {
		rumpts, err := jsonPoints(body, precision, extraTags)
		if err != nil {
			return nil, err
		}
		for _, p := range rumpts {
			tags := p.Tags()
			if tags != nil {
				if !contains(tags[rumMetricAppID], appIDWhiteList) {
					return nil, ErrRUMAppIDNotInWhiteList
				}
			}
		}
		return rumpts, nil
	}

	rumpts, err := lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: extraTags,
		Strict:    true,

		// 由于 RUM 数据需要分别处理，故用回调函数来区分
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())

			if !contains(p.Tags().GetString(rumMetricAppID), appIDWhiteList) {
				return nil, ErrRUMAppIDNotInWhiteList
			}

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

func contains(str string, list []string) bool {
	if len(list) == 0 {
		return true
	}
	for _, a := range list {
		if a == str {
			return true
		}
	}
	return false
}

func handleRUMBody(body []byte,
	precision,
	srcip string,
	isjson bool,
	list []string) ([]*io.Point, error) {
	return doHandleRUMBody(body, precision, isjson, geoTags(srcip), list)
}
