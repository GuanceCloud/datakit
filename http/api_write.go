package http

import (
	"encoding/json"
	"net/http"
	"reflect"
	"time"

	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	plw "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/worker"
)

type IApiWrite interface {
	sendToPipLine(*plw.Task) error
	sendToIO(string, string, []*io.Point, *io.Option) error
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
func (jp *jsonPoint) point(opt *lp.Option) (*io.Point, error) {
	p, err := lp.MakeLineProtoPoint(jp.Measurement, jp.Tags, jp.Fields, opt)
	if err != nil {
		return nil, err
	}

	return &io.Point{Point: p}, nil
}

func (x *apiWriteImpl) sendToIO(input, category string, pts []*io.Point, opt *io.Option) error {
	return io.Feed(input, category, pts, opt)
}

func (x *apiWriteImpl) geoInfo(ip string) map[string]string {
	return geoTags(ip)
}

// sendToPipLine will send each point from @pts to pipeline module
func (x *apiWriteImpl) sendToPipLine(t *plw.Task) error {
	return plw.FeedPipelineTaskBlock(t)
}

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

	input := DEFAULT_INPUT

	category := req.URL.Path

	switch category {
	case datakit.Metric,
		datakit.Network,
		datakit.Logging,
		datakit.Object,
		datakit.Tracing,
		datakit.KeyEvent:

	case datakit.CustomObject:
		input = "custom_object"

	case datakit.RUM:
		input = "rum"
	case datakit.Security:
		input = "scheck"
	default:
		l.Debugf("invalid category: %s", category)
		return nil, ErrInvalidCategory
	}

	q := req.URL.Query()

	if x := q.Get(INPUT); x != "" {
		input = x
	}

	precision := DEFAULT_PRECISION
	if x := q.Get(PRECISION); x != "" {
		precision = x
	}

	extags := extraTags
	if x := q.Get(IGNORE_GLOBAL_TAGS); x != "" {
		extags = nil
	}

	var version string
	if x := q.Get(VERSION); x != "" {
		version = x
	}

	var pipelineSource string
	if x := q.Get(PIPELINE_SOURCE); x != "" {
		pipelineSource = x
	}

	switch precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		l.Warnf("invalid precision %s", precision)
		return nil, ErrInvalidPrecision
	}

	body, err = uhttp.ReadBody(req)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, ErrEmptyBody
	}

	isjson := req.Header.Get("Content-Type") == "application/json"

	var pts []*io.Point

	switch category {
	case datakit.RUM:
		pts, err = handleRUMBody(body,
			precision, isjson,
			h.geoInfo(getSrcIP(apiConfig, req)),
			apiConfig.RUMAppIDWhiteList)

		if err != nil {
			return nil, err
		}

	default:
		pts, err = handleWriteBody(body, isjson, &lp.Option{
			Precision: precision,
			Time:      time.Now(),
			ExtraTags: extags,
			Strict:    true,
		})
		if err != nil {
			return nil, err
		}

		switch category {
		case datakit.Object:
			for _, pt := range pts {
				if err := checkObjectPoint(pt); err != nil {
					return nil, err
				}
			}
		}
	}

	if len(pts) == 0 {
		return nil, ErrNoPoints
	}

	l.Debugf("received %d(%s) points from %s, pipeline source: %v", len(pts), category, input, pipelineSource)

	if category == datakit.Logging && pipelineSource != "" {
		// Currently on logging support pipeline.
		// We try to find some @input.p to split logging, for example, if @input is nginx
		// the default pipeline is nginx.p.
		// If nginx.p missing, pipeline do nothing on incomming logging data.

		// for logging upload, we redirect them to pipeline
		l.Debugf("send pts to pipeline")
		err = h.sendToPipLine(buildLogPLTask(input, pipelineSource, version, category, pts))
	} else {
		err = h.sendToIO(input, category, pts, &io.Option{HighFreq: true, Version: version})
	}

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func handleWriteBody(body []byte, isJSON bool, opt *lp.Option) ([]*io.Point, error) {
	switch isJSON {
	case true:
		return jsonPoints(body, opt)

	default:
		pts, err := lp.ParsePoints(body, opt)
		if err != nil {
			return nil, uhttp.Error(ErrInvalidLinePoint, err.Error())
		}

		return io.WrapPoint(pts), nil
	}
}

func jsonPoints(body []byte, opt *lp.Option) ([]*io.Point, error) {
	var jps []jsonPoint
	err := json.Unmarshal(body, &jps)
	if err != nil {
		l.Error(err)
		return nil, ErrInvalidJSONPoint
	}

	if opt == nil {
		opt = lp.DefaultOption
	}

	var pts []*io.Point
	for _, jp := range jps {
		if p, err := jp.point(opt); err != nil {
			l.Error(err)
			return nil, uhttp.Error(ErrInvalidJSONPoint, err.Error())
		} else {
			pts = append(pts, p)
		}
	}
	return pts, nil
}

func checkObjectPoint(p *io.Point) error {
	tags := p.Point.Tags()
	if _, ok := tags["name"]; !ok {
		return ErrInvalidObjectPoint
	}
	return nil
}
