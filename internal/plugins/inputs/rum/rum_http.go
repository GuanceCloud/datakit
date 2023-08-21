// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package rum

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/gin-gonic/gin/binding"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
)

func httpStatusRespFunc(resp http.ResponseWriter, req *http.Request, err error) {
	if err != nil {
		httpErr(resp, err)
	} else {
		httpOK(resp, nil)
	}
}

func (ipt *Input) parseCallback(p *point.Point) (*point.Point, error) {
	name := string(p.Name())
	tags := p.InfluxTags()
	if !contains(tags[rumMetricAppID], config.Cfg.HTTPAPI.RUMAppIDWhiteList) {
		return nil, httpapi.ErrRUMAppIDNotInWhiteList
	}

	if _, ok := rumMetricNames[name]; !ok {
		return nil, uhttp.Errorf(httpapi.ErrUnknownRUMMeasurement, "unknow RUM measurement: %s", name)
	}

	if name == Error {
		// handle sourcemap
		sdkName := tags["sdk_name"]
		status := &sourceMapStatus{
			appid:   tags["app_id"],
			sdkName: sdkName,
			status:  StatusUnknown,
			remark:  "",
		}
		px, err := ipt.parseSourcemap(p, sdkName, status)
		if err != nil {
			log.Errorf("handle source map failed: %s", err.Error())
			if status.status != StatusZipNotFound && status.status != StatusToolNotFound {
				status.status = StatusError
			}
			status.remark = err.Error()
			// Do nothing, return original point.
			sourceMapCount.WithLabelValues(status.appid, status.sdkName, status.status, utf8SubStr(status.remark, 8)).Inc()
			return p, nil
		}
		return px, nil
	} else if name == Resource {
		// handle resource provider
		px, err := ipt.handleProvider(p)
		if err != nil {
			log.Errorf("unable to new point: %s", err)
			// If err prompt, we return the original point.
			return p, nil
		}
		return px, nil
	}
	return p, nil
}

func (ipt *Input) handleRUM(resp http.ResponseWriter, req *http.Request) {
	log.Debugf("### RUM request from %s", req.URL.String())

	opts := []point.Option{
		point.WithPrecision(point.NS),
		point.WithTime(time.Now()),
	}

	var (
		query                   = req.URL.Query()
		version, pipelineSource string
	)
	if x := query.Get(httpapi.ArgVersion); x != "" {
		version = x
	}
	if x := query.Get(httpapi.ArgPipelineSource); x != "" {
		pipelineSource = x
	}
	if x := query.Get(httpapi.ArgPrecision); x != "" {
		opts = append(opts, point.WithPrecision(point.PrecStr(x)))
	}

	body, err := uhttp.ReadBody(req)
	if err != nil {
		log.Error(err.Error())
		httpErr(resp, err)

		return
	}
	if len(body) == 0 {
		log.Debug(httpapi.ErrEmptyBody.Err.Error())
		httpErr(resp, httpapi.ErrEmptyBody)

		return
	}

	var (
		pts       []*point.Point
		apiConfig = config.Cfg.HTTPAPI
		isJSON    = strings.Contains(req.Header.Get("Content-Type"), "application/json")
	)

	ipStatus := &ipLocationStatus{}
	defer func() {
		ClientRealIPCounter.WithLabelValues(ipStatus.appid, ipStatus.ipStatus, ipStatus.locateStatus).Inc()
	}()

	opts = append(opts, point.WithExtraTags(geoTags(getSrcIP(apiConfig, req, ipStatus), ipStatus)), point.WithCallback(ipt.parseCallback))

	if pts, err = httpapi.HandleWriteBody(body, isJSON, opts...); err != nil {
		log.Error(err.Error())
		httpErr(resp, httpapi.ErrInvalidLinePoint)

		return
	}

	if len(pts) == 0 {
		log.Debug(httpapi.ErrNoPoints.Err.Error())
		httpErr(resp, httpapi.ErrNoPoints)

		return
	}

	tags := pts[0].InfluxTags()
	ipStatus.appid = tags["app_id"]

	log.Debugf("### received %d(%s) points from %s, pipeline source: %v", len(pts), req.URL.Path, inputName, pipelineSource)

	feedOpt := &dkio.Option{Version: version}
	if pipelineSource != "" {
		feedOpt.PlScript = map[string]string{pipelineSource: pipelineSource + ".p"}
	}
	if err = ipt.feeder.Feed(inputName, point.RUM, pts, feedOpt); err != nil {
		log.Error(err.Error())
		httpErr(resp, err)

		return
	}

	if query.Get(httpapi.ArgEchoLineProto) != "" {
		var arr []string
		for _, x := range pts {
			arr = append(arr, x.LineProto())
		}
		httpOK(resp, strings.Join(arr, "\n"))
		return
	}

	httpOK(resp, nil)
}

func httpOK(w http.ResponseWriter, body interface{}) {
	if body == nil {
		if err := writeBody(w, httpapi.OK.HttpCode, binding.MIMEJSON, nil); err != nil {
			log.Error(err.Error())
		}

		return
	}

	var (
		bodyBytes   []byte
		contentType string
		err         error
	)
	switch x := body.(type) {
	case []byte:
		bodyBytes = x
	default:
		resp := &uhttp.BodyResp{
			HttpError: httpapi.OK,
			Content:   body,
		}
		contentType = `application/json`

		if bodyBytes, err = json.Marshal(resp); err != nil {
			log.Error(err.Error())
			jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "%s: %+#v", "json.Marshal() failed", resp)

			return
		}
	}

	if err := writeBody(w, httpapi.OK.HttpCode, contentType, bodyBytes); err != nil {
		log.Error(err.Error())
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

func jsonReturnf(httpErr *uhttp.HttpError, w http.ResponseWriter, format string, args ...interface{}) {
	resp := &uhttp.BodyResp{
		HttpError: httpErr,
	}

	if args != nil {
		resp.Message = fmt.Sprintf(format, args...)
	}

	buf, err := json.Marshal(resp)
	if err != nil {
		jsonReturnf(uhttp.NewErr(err, http.StatusInternalServerError), w, "%s: %+#v", "json.Marshal() failed", resp)

		return
	}

	if err := writeBody(w, httpErr.HttpCode, binding.MIMEJSON, buf); err != nil {
		log.Error(err.Error())
	}
}
