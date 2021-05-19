package http

import (
	"fmt"
	"strings"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"

	"github.com/gin-gonic/gin"
	influxm "github.com/influxdata/influxdb1-client/models"
	influxdb "github.com/influxdata/influxdb1-client/v2"
)

var (
	RUMOriginIPHeader = "X-Forward-For"

	metricNames = map[string]bool{
		`rum_web_page_performance`:          true,
		`rum_web_resource_performance`:      true,
		`rum_app_startup`:                   true,
		`rum_app_system_performance`:        true,
		`rum_app_view`:                      true,
		`rum_app_freeze`:                    true,
		`rum_app_resource_performance`:      true,
		"rum_mini_app_startup":              true,
		"rum_mini_app_page_performance":     true,
		"rum_mini_app_resource_performance": true,
	}

	esNames = map[string]bool{
		`js_error`: true,
		`page`:     true,
		`resource`: true,
		`view`:     true,
		`crash`:    true,
		`freeze`:   true,
	}
)

func isMetricData(name string) bool {
	_, ok := metricNames[name]
	return ok
}

func isRUMData(name string) bool {
	_, ok := esNames[name]
	return ok
}

func geoTags(srcip string) (tags map[string]string) {
	tags = map[string]string{}

	ipInfo, err := geo.Geo(srcip)

	l.Debugf("ipinfo: %+#v", ipInfo)

	if err != nil {
		l.Errorf("geo failed: %s, ignored", err)
		return
	} else {
		// 无脑填充 geo 数据
		tags = map[string]string{
			"city":     ipInfo.City,
			"province": ipInfo.Region,
			"country":  ipInfo.Country_short,
			"isp":      ip2isp.SearchIsp(srcip),
		}
	}

	return
}

func handleBody(body []byte, precision, srcip string) (mpts, rumpts []*influxdb.Point, err error) {
	extraTags := geoTags(srcip)

	mpts, err = lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: extraTags,
		Strict:    true,

		// 由于 RUM 数据需要分别处理，故用回调函数来区分
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())
			if isRUMData(name) { // ignore RUM data
				return nil, nil
			}

			if !isMetricData(name) {
				return nil, fmt.Errorf("unknow metric name %s", name)
			}

			return p, nil
		},
	})

	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	// 只将 IP 注入到 RUM 数据中（RUM 类型的数据存储在 ES 中）
	// 注意：不要将 IP 加到时序数据 tag 中，这会造成时间线暴涨
	extraTags["ip"] = srcip

	rumpts, err = lp.ParsePoints(body, &lp.Option{
		Time:      time.Now(),
		Precision: precision,
		ExtraTags: extraTags,
		Strict:    true,
		Callback: func(p influxm.Point) (influxm.Point, error) {
			name := string(p.Name())
			if isMetricData(name) { // ignore Metric data
				return nil, nil
			}

			if !isRUMData(name) {
				return nil, fmt.Errorf("unknow metric name %s", name)
			}

			// add extra `message' tag
			p.AddTag("message", p.String())
			return p, nil
		},
	})
	if err != nil {
		l.Error(err)
		return nil, nil, err
	}

	return mpts, rumpts, nil
}

func handleRUMBody(c *gin.Context, precision, input string, body []byte) {

	srcip := c.Request.Header.Get(RUMOriginIPHeader)
	if srcip != "" {
		parts := strings.Split(srcip, ",")
		if len(parts) > 0 {
			srcip = parts[0] // 注意：此处只取第一个 IP 作为源 IP
		}
	} else { // 默认取 gin 框架带进来的 IP
		parts := strings.Split(c.Request.RemoteAddr, ":")
		if len(parts) > 0 {
			srcip = parts[0]
		}
	}

	metricpts, rumpts, err := handleBody(body, precision, srcip)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
		return
	}

	if input == DEFAULT_INPUT { // RUM 默认源不好直接用 datakit，故单独以 `rum' 标记之
		input = "rum"
	}

	if len(metricpts) > 0 {
		if err = io.Feed(input, datakit.Metric, io.WrapPoint(metricpts), &io.Option{HighFreq: true}); err != nil {
			uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
			return
		}
	}

	if len(rumpts) > 0 {
		if err = io.Feed(input, datakit.Rum, io.WrapPoint(rumpts), &io.Option{HighFreq: true}); err != nil {
			uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
			return
		}
	}

	ErrOK.HttpBody(c, nil)
}
