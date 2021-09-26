package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type jsonPoint struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Field       map[string]interface{} `json:"fields"`
	Time        int64                  `json:"time,omitempty"`
}

// convert json point to real point.
func (jp *jsonPoint) pt(prec string, extags map[string]string) (*io.Point, error) {
	if prec == "" || prec == "n" {
		prec = "ns"
	}

	var t time.Time
	switch prec {
	case "h", "m", "s", "ms", "u", "ns":
		if jp.Time == 0 {
			t = time.Now()
		} else {
			du, err := time.ParseDuration(fmt.Sprintf("%d%s", jp.Time, prec))
			if err != nil {
				return nil, err
			}
			t = time.Unix(0, int64(du))
		}

	default:
		return nil, fmt.Errorf("invalid precision")
	}

	if extags == nil {
		return io.MakePointWithoutGlobalTags(jp.Measurement, jp.Tags, jp.Field, t)
	} else {
		return io.MakePoint(jp.Measurement, jp.Tags, jp.Field, t)
	}
}

func apiWrite(c *gin.Context) {
	var body []byte
	var err error

	input := DEFAULT_INPUT

	category := c.Request.URL.Path

	switch category {
	case datakit.Metric,
		datakit.Network,
		datakit.Logging,
		datakit.Object,
		datakit.Tracing,
		datakit.KeyEvent:

	case datakit.CustomObject:
		input = "custom_object"

	case datakit.Rum:
		input = "rum"
	case datakit.Security:
		input = "sechecker"
	default:
		l.Debugf("invalid category: %s", category)
		uhttp.HttpErr(c, ErrInvalidCategory)
		return
	}

	if x := c.Query(INPUT); x != "" {
		input = x
	}

	precision := DEFAULT_PRECISION
	if x := c.Query(PRECISION); x != "" {
		precision = x
	}

	switch precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		l.Warnf("invalid precision %s", precision)
		uhttp.HttpErr(c, ErrInvalidPrecision)
		return
	}

	body, err = uhttp.GinRead(c)
	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrHttpReadErr, err.Error()))
		return
	}

	isjson := (c.Request.Header.Get("Content-Type") == "application/json")

	var pts []*io.Point
	if category == datakit.Rum { // RUM 数据单独处理
		srcip := ""
		if apiConfig != nil {
			srcip = c.Request.Header.Get(apiConfig.RUMOriginIPHeader)
			l.Debugf("get ip from %s: %s", apiConfig.RUMOriginIPHeader, srcip)
			if srcip == "" {
				for k, v := range c.Request.Header {
					l.Debugf("%s: %s", k, strings.Join(v, ","))
				}
			}
		} else {
			l.Debugf("apiConfig not set")
		}

		if srcip != "" {
			l.Debugf("header remote addr: %s", srcip)
			parts := strings.Split(srcip, ",")
			if len(parts) > 0 {
				srcip = parts[0] // 注意：此处只取第一个 IP 作为源 IP
			}
		} else { // 默认取 gin 框架带进来的 IP
			l.Debugf("gin remote addr: %s", c.Request.RemoteAddr)
			host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
			if err == nil {
				srcip = host
			}
		}

		pts, err = handleRUMBody(body, precision, srcip, isjson, apiConfig.RUMAppIDWhiteList)
		// appid不在白名单中，当前 http 请求直接返回
		if errors.As(err, &ErrRUMAppIdNotInWhiteList) {
			uhttp.HttpErr(c, err)
			return
		}
	} else {
		extags := extraTags
		if x := c.Query(IGNORE_GLOBAL_TAGS); x != "" {
			extags = nil
		}

		pts, err = handleWriteBody(body, precision,
			extags, isjson, nil)
		if err != nil {
			uhttp.HttpErr(c, err)
			return
		}
	}

	l.Debugf("received %d(%s) points from %s", len(pts), category, input)

	err = io.Feed(input, category, pts, &io.Option{HighFreq: true})

	if err != nil {
		uhttp.HttpErr(c, uhttp.Error(ErrBadReq, err.Error()))
	} else {
		ErrOK.HttpBody(c, nil)
	}
}

func handleWriteBody(body []byte,
	precision string,
	extags map[string]string,
	isJson bool,
	appIdWhiteList []string,
) ([]*io.Point, error) {
	switch isJson {
	case true:
		return jsonPoints(body, precision, extags, appIdWhiteList)

	default:
		pts, err := lp.ParsePoints(body, &lp.Option{
			Time:      time.Now(),
			ExtraTags: extags,
			Strict:    true,
			Precision: precision,
		})
		if err != nil {
			return nil, uhttp.Error(ErrInvalidLinePoint, err.Error())
		}

		return io.WrapPoint(pts), nil
	}
}

func jsonPoints(body []byte,
	prec string,
	extags map[string]string,
	appIdWhiteList []string) ([]*io.Point, error) {
	var jps []jsonPoint
	err := json.Unmarshal(body, &jps)
	if err != nil {
		l.Error(err)
		return nil, ErrInvalidJsonPoint
	}

	var pts []*io.Point
	for _, jp := range jps {
		if p, err := jp.pt(prec, extags); err != nil {
			l.Error(err)
			return nil, uhttp.Error(ErrInvalidJsonPoint, err.Error())
		} else {
			tags := p.Tags()
			if len(tags) == 0 {
				return nil, fmt.Errorf("empty tags")
			}
			if !contains(tags[rumMetricAppID], appIdWhiteList) {
				return nil, ErrRUMAppIdNotInWhiteList
			}

			pts = append(pts, p)
		}
	}
	return pts, nil
}
