// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

type IApiWrite interface {
	sendToIO(string, string, []*point.Point, *io.Option) error
	geoInfo(string) map[string]string
}

type apiWriteImpl struct{}

type jsonPoint struct {
	Measurement string                 `json:"measurement"`
	Tags        map[string]string      `json:"tags,omitempty"`
	Fields      map[string]interface{} `json:"fields"`
	Time        int64                  `json:"time,omitempty"`
}

// convert json point to lineproto point.
func (jp *jsonPoint) point(opt *lp.Option) (*point.Point, error) {
	p, err := lp.MakeLineProtoPoint(jp.Measurement, jp.Tags, jp.Fields, opt)
	if err != nil {
		return nil, err
	}

	return &point.Point{Point: p}, nil
}

func (x *apiWriteImpl) sendToIO(input, category string, pts []*point.Point, opt *io.Option) error {
	return io.Feed(input, category, pts, opt)
}

func (x *apiWriteImpl) geoInfo(ip string) map[string]string {
	return geoTags(ip)
}

//nolint:stylecheck
const (
	ArgPrecision            = "precision"
	ArgInput                = "input"
	ArgIgnoreGlobalTags     = "ignore_global_tags"      // deprecated, use IGNORE_GLOBAL_HOST_TAGS
	ArgIgnoreGlobalHostTags = "ignore_global_host_tags" // default enabled
	ArgGlobalElectionTags   = "global_election_tags"    // default disabled
	ArgEchoLineProto        = "echo_line_proto"
	ArgVersion              = "version"
	ArgPipelineSource       = "source"
	ArgLoose                = "loose"

	defaultInput     = "datakit-http" // 当 API 调用方未亮明自己身份时，默认它作为数据源名称
	DefaultPrecision = "n"
)

func apiWrite(w http.ResponseWriter, req *http.Request, x ...interface{}) (interface{}, error) {
	var body []byte
	var err error

	if x == nil || len(x) != 1 {
		l.Errorf("invalid handler")
		return nil, ErrInvalidAPIHandler
	}

	h, ok := x[0].(IApiWrite)
	if !ok {
		l.Errorf("not IApiWrite, got %s", reflect.TypeOf(x).String())
		return nil, ErrInvalidAPIHandler
	}

	input := defaultInput
	category := req.URL.Path
	opt := lp.NewDefaultOption()

	// extra tags comes from global-host-tag or global-env-tags
	opt.ExtraTags = map[string]string{}
	opt.Precision = DefaultPrecision
	opt.Time = time.Now()

	switch category {
	case datakit.Metric,
		datakit.Network,
		datakit.Logging,
		datakit.Object,
		datakit.Tracing,
		datakit.KeyEvent:

		opt.MaxFieldValueLen = point.MaxFieldValueLen
		opt.MaxTagKeyLen = point.MaxTagKeyLen
		opt.MaxFieldKeyLen = point.MaxFieldKeyLen
		opt.MaxTagValueLen = point.MaxTagValueLen

		if category == datakit.Metric {
			opt.EnablePointInKey = true
			opt.DisableStringField = true
		}

	case datakit.CustomObject:
		input = "custom_object"

	case datakit.Security:
		input = "scheck"
	default:
		l.Debugf("invalid category: %s", category)
		return nil, ErrInvalidCategory
	}

	q := req.URL.Query()

	opt.Strict = (q.Get(ArgLoose) == "")

	if x := q.Get(ArgInput); x != "" {
		input = x
	}

	if x := q.Get(ArgPrecision); x != "" {
		opt.Precision = x
	}

	for _, arg := range []string{
		ArgIgnoreGlobalHostTags,
		ArgIgnoreGlobalTags, // deprecated
	} {
		if x := q.Get(arg); x != "" {
			opt.ExtraTags = map[string]string{}
			break
		} else {
			for k, v := range point.GlobalHostTags() {
				l.Debugf("arg=%s, add host tag %s: %s", arg, k, v)
				opt.ExtraTags[k] = v
			}
		}
	}

	if x := q.Get(ArgGlobalElectionTags); x != "" {
		for k, v := range point.GlobalElectionTags() {
			l.Debugf("add env tag %s: %s", k, v)
			opt.ExtraTags[k] = v
		}
	}

	var version string
	if x := q.Get(ArgVersion); x != "" {
		version = x
	}

	var pipelineSource string
	if x := q.Get(ArgPipelineSource); x != "" {
		pipelineSource = x
	}

	switch opt.Precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		l.Warnf("invalid precision %s", opt.Precision)
		return nil, ErrInvalidPrecision
	}

	body, err = uhttp.ReadBody(req)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, ErrEmptyBody
	}

	isjson := (strings.Contains(req.Header.Get("Content-Type"), "application/json"))

	var pts []*point.Point

	pts, err = handleWriteBody(body, isjson, opt)
	if err != nil {
		return nil, err
	}

	// check if object is ok
	if category == datakit.Object {
		for _, pt := range pts {
			if err := checkObjectPoint(pt); err != nil {
				return nil, err
			}
		}
	}

	if len(pts) == 0 {
		return nil, ErrNoPoints
	}

	l.Debugf("received %d(%s) points from %s, pipeline source: %v", len(pts), category, input, pipelineSource)

	feedOpt := &io.Option{Version: version}
	if pipelineSource != "" {
		feedOpt.PlScript = map[string]string{pipelineSource: pipelineSource + ".p"}
	}

	if err := h.sendToIO(input, category, pts, feedOpt); err != nil {
		return nil, err
	}

	if q.Get(ArgEchoLineProto) != "" {
		res := []*point.JSONPoint{}
		for _, pt := range pts {
			x, err := pt.ToJSON()
			if err != nil {
				l.Warnf("ToJSON: %s, ignored", err)
				continue
			}
			res = append(res, x)
		}

		return res, nil
	}

	return nil, nil
}

func handleWriteBody(body []byte, isJSON bool, opt *lp.Option) ([]*point.Point, error) {
	switch isJSON {
	case true:
		return jsonPoints(body, opt)

	default:
		pts, err := lp.ParsePoints(body, opt)
		if err != nil {
			return nil, uhttp.Error(ErrInvalidLinePoint, err.Error())
		}

		return point.WrapPoint(pts), nil
	}
}

func jsonPoints(body []byte, opt *lp.Option) ([]*point.Point, error) {
	var jps []jsonPoint
	err := json.Unmarshal(body, &jps)
	if err != nil {
		l.Error(err)
		return nil, ErrInvalidJSONPoint
	}

	if opt == nil {
		opt = lp.DefaultOption
	}

	var pts []*point.Point
	for _, jp := range jps {
		if jp.Time != 0 { // use time from json point
			opt.Time = getTimeFromInt64(jp.Time, opt)
		}

		if p, err := jp.point(opt); err != nil {
			l.Error(err)
			return nil, uhttp.Error(ErrInvalidJSONPoint, err.Error())
		} else {
			pts = append(pts, p)
		}
	}
	return pts, nil
}

func getTimeFromInt64(n int64, opt *lp.Option) time.Time {
	if opt != nil {
		switch opt.Precision {
		case "h":
			return time.Unix(n*3600, 0).UTC()
		case "m":
			return time.Unix(n*60, 0).UTC()
		case "s":
			return time.Unix(n, 0).UTC()
		case "ms":
			return time.Unix(0, n*1000).UTC()
		case "u":
			return time.Unix(0, n*1000000).UTC()
		default:
		}
	}

	// nanoseconds
	return time.Unix(0, n).UTC()
}

func checkObjectPoint(p *point.Point) error {
	tags := p.Point.Tags()
	if _, ok := tags["name"]; !ok {
		return ErrInvalidObjectPoint
	}
	return nil
}
