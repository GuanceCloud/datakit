// Package rum real user monitoring
package rum

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin/binding"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	uhttp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/trace"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "rum"

	sampleConfig = `
[[inputs.rum]]
## profile Agent endpoints register by version respectively.
## Endpoints can be skipped listen by remove them from the list.
## Default value set as below. DO NOT MODIFY THESE ENDPOINTS if not necessary.
endpoints = ["/v1/write/rum"]

# proguard HOME
proguard_home = "/usr/local/datakit/data/rum/tools/proguard"

# android-ndk HOME
ndk_home = "/usr/local/datakit/data/rum/tools/android-ndk"

# atos or atosl bin path
# for macOS datakit use the built-in tool atos default
# for Linux there are several tools that can be used to instead of macOS atos partially,
# such as https://github.com/everettjf/atosl-rs
atos_bin_path = "/usr/local/datakit/data/rum/tools/atosl"

`
)

var (
	log                                    = logger.DefaultSLogger(inputName)
	_                     inputs.HTTPInput = &Input{}
	_                     inputs.InputV2   = &Input{}
	sourceMapTokenBuckets                  = newExecCmdTokenBuckets(64)
)

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{}
	})
}

type execCmdTokenBuckets struct {
	buckets chan struct{}
}

func newExecCmdTokenBuckets(size int) *execCmdTokenBuckets {
	if size < 16 {
		size = 16
	}
	tb := &execCmdTokenBuckets{
		buckets: make(chan struct{}, size),
	}
	for i := 0; i < size; i++ {
		tb.buckets <- struct{}{}
	}
	return tb
}

func (e *execCmdTokenBuckets) getToken() struct{} {
	return <-e.buckets
}

func (e *execCmdTokenBuckets) sendBackToken(token struct{}) {
	e.buckets <- token
}

type Input struct {
	Endpoints    []string `toml:"endpoints"`
	ProguardHome string   `toml:"proguard_home"`
	NDKHome      string   `toml:"ndk_home"`
	AtosBinPath  string   `toml:"atos_bin_path"`
}

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

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&trace.TraceMeasurement{Name: inputName}}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func (i *Input) Terminate() {
	log.Info("rum plugin over...")
}

func jsonPoints(body []byte, opt *lp.Option) ([]*point.Point, error) {
	var jps []jsonPoint
	err := json.Unmarshal(body, &jps)
	if err != nil {
		log.Error(err)
		return nil, dkhttp.ErrInvalidJSONPoint
	}

	if opt == nil {
		opt = lp.DefaultOption
	}

	var pts []*point.Point
	for _, jp := range jps {
		if jp.Time != 0 { // use time from json point
			opt.Time = time.Unix(0, jp.Time)
		}

		if p, err := jp.point(opt); err != nil {
			log.Error(err)
			return nil, uhttp.Error(dkhttp.ErrInvalidJSONPoint, err.Error())
		} else {
			pts = append(pts, p)
		}
	}
	return pts, nil
}

func (i *Input) handleRUM(req *http.Request) ([]*point.JSONPoint, error) {
	var body []byte
	var err error

	category := req.URL.Path

	q := req.URL.Query()

	precision := dkhttp.DEFAULT_PRECISION
	if x := q.Get(dkhttp.PRECISION); x != "" {
		precision = x
	}

	// extraTags comes from global-host-tag or global-env-tags
	extraTags := map[string]string{}
	for _, arg := range []string{
		dkhttp.IGNORE_GLOBAL_HOST_TAGS,
		dkhttp.IGNORE_GLOBAL_TAGS, // deprecated
	} {
		if x := q.Get(arg); x != "" {
			extraTags = map[string]string{}
			break
		} else {
			for k, v := range point.GlobalHostTags() {
				log.Debugf("arg=%s, add host tag %s: %s", arg, k, v)
				extraTags[k] = v
			}
		}
	}

	if x := q.Get(dkhttp.GLOBAL_ENV_TAGS); x != "" {
		for k, v := range point.GlobalEnvTags() {
			log.Debugf("add env tag %s: %s", k, v)
			extraTags[k] = v
		}
	}

	var version string
	if x := q.Get(dkhttp.VERSION); x != "" {
		version = x
	}

	var pipelineSource string
	if x := q.Get(dkhttp.PIPELINE_SOURCE); x != "" {
		pipelineSource = x
	}

	switch precision {
	case "h", "m", "s", "ms", "u", "n":
	default:
		log.Warnf("invalid precision %s", precision)
		return nil, dkhttp.ErrInvalidPrecision
	}

	body, err = uhttp.ReadBody(req)
	if err != nil {
		return nil, err
	}

	if len(body) == 0 {
		return nil, dkhttp.ErrEmptyBody
	}

	isjson := (strings.Contains(req.Header.Get("Content-Type"), "application/json"))

	var pts []*point.Point

	apiConfig := config.Cfg.HTTPAPI
	pts, err = handleRUMBody(body,
		precision, isjson,
		geoInfo(getSrcIP(apiConfig, req)),
		apiConfig.RUMAppIDWhiteList,
		i,
	)

	if err != nil {
		return nil, err
	}

	if len(pts) == 0 {
		return nil, dkhttp.ErrNoPoints
	}

	log.Debugf("received %d(%s) points from %s, pipeline source: %v", len(pts), category, inputName, pipelineSource)

	err = sendToIO(inputName, category, pts, &io.Option{HighFreq: true, Version: version})
	if err != nil {
		return nil, err
	}

	if q.Get(dkhttp.ECHO_LINE_PROTO) != "" {
		var res []*point.JSONPoint
		for _, pt := range pts {
			x, err := pt.ToJSON()
			if err != nil {
				log.Warnf("ToJSON: %s, ignored", err)
				continue
			}
			res = append(res, x)
		}

		return res, nil
	}

	return nil, nil
}

func writeBody(w http.ResponseWriter, statusCode int, contentType string, body []byte) error {
	w.WriteHeader(statusCode)
	if body != nil {
		w.Header().Set("Content-Type", contentType)
		n, err := w.Write(body)
		if n != len(body) {
			return fmt.Errorf("send partial http response, full length(%d), send length(%d) ", len(body), n)
		}
		if err != nil {
			return fmt.Errorf("send http response popup err: %w", err)
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, statusCode int, body []byte) error {
	return writeBody(w, statusCode, binding.MIMEJSON, body)
}

func jsonReturnf(he *uhttp.HttpError, w http.ResponseWriter, format string, args ...interface{}) {
	resp := &uhttp.BodyResp{
		HttpError: he,
	}

	if args != nil {
		resp.Message = fmt.Sprintf(format, args...)
	}

	j, err := json.Marshal(resp)
	if err != nil {
		jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "%s: %+#v", "json.Marshal() failed", resp)
		return
	}

	if err := writeJSON(w, he.HttpCode, j); err != nil {
		log.Error(err)
	}
}

func httpErr(w http.ResponseWriter, err error) {
	switch e := err.(type) { // nolint:errorlint
	case *uhttp.HttpError:
		jsonReturnf(e, w, "")
	case *uhttp.MsgError:
		if e.Args != nil {
			jsonReturnf(e.HttpError, w, e.Fmt, e.Args)
		}
	default:
		jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "")
	}
}

func httpOK(w http.ResponseWriter, body interface{}) {
	if body == nil {
		if err := writeJSON(w, dkhttp.OK.HttpCode, nil); err != nil {
			log.Error(err)
		}
		return
	}

	var bodyBytes []byte
	var contentType string
	var err error

	switch x := body.(type) {
	case []byte:
		bodyBytes = x
	default:
		resp := &uhttp.BodyResp{
			HttpError: dkhttp.OK,
			Content:   body,
		}
		contentType = `application/json`

		bodyBytes, err = json.Marshal(resp)
		if err != nil {
			jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "%s: %+#v", "json.Marshal() failed", resp)
			return
		}
	}

	if err := writeBody(w, dkhttp.OK.HttpCode, contentType, bodyBytes); err != nil {
		log.Error(err)
	}
}

func (i *Input) RegHTTPHandler() {
	for _, endpoint := range i.Endpoints {
		dkhttp.RegHTTPHandler(http.MethodPost, endpoint, func(w http.ResponseWriter, req *http.Request) {
			res, err := i.handleRUM(req)
			if err != nil {
				httpErr(w, err)
				return
			}

			httpOK(w, res)
		})

		log.Infof("pattern: %s registered", endpoint)
	}
}

func (i *Input) Catalog() string {
	return inputName
}

func (i *Input) Run() {
	log = logger.SLogger(inputName)
	log.Infof("the input %s is running...", inputName)

	log.Infof("rum config: %+#v\n", i)

	loadSourcemapFile()
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}
