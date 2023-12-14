// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package promremote handle promremote remote write data.
package promremote

import (
	"compress/gzip"
	"crypto/subtle"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/golang/snappy"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

var (
	l                = logger.DefaultSLogger(inputName)
	_ inputs.InputV2 = (*Input)(nil)
)

const (
	body                   = "body"
	query                  = "query"
	defaultRemoteWritePath = "/prom_remote_write"
)

type Input struct {
	Path            string            `toml:"path"`
	Methods         []string          `toml:"methods"`
	DataSource      string            `toml:"data_source"`
	MaxBodySize     int64             `toml:"max_body_size"`
	BasicUsername   string            `toml:"basic_username"`
	BasicPassword   string            `toml:"basic_password"`
	HTTPHeaderTags  map[string]string `toml:"http_header_tags"`
	Tags            map[string]string `toml:"tags"`
	TagsIgnore      []string          `toml:"tags_ignore"`
	TagsIgnoreRegex []string          `toml:"tags_ignore_regex"`
	TagsOnly        []string          `toml:"tags_only"`
	TagsOnlyRegex   []string          `toml:"tags_only_regex"`
	TagsRename      map[string]string `toml:"tags_rename"`
	Overwrite       bool              `toml:"overwrite"`
	Output          string            `toml:"output"`

	Election bool // forever false

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger

	Parser
}

func (ipt *Input) RegHTTPHandler() {
	l = logger.SLogger(inputName)
	if ipt.Path == "" {
		ipt.Path = defaultRemoteWritePath
	}
	for _, m := range ipt.Methods {
		httpapi.RegHTTPHandler(m, ipt.Path, ipt.ServeHTTP)
	}
}

func (*Input) Catalog() string {
	return catalog
}

func (*Input) Terminate() {
	// do nothing
}

func (ipt *Input) Run() {
	l.Infof("%s input started...", inputName)
	for i, m := range ipt.Methods {
		ipt.Methods[i] = strings.ToUpper(m)
	}
}

// ServeHTTP accepts prometheus remote writing, then parses received
// metrics, and sends them to datakit io or local disk file.
func (ipt *Input) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler := ipt.serveWrite

	ipt.authenticateIfSet(handler, res, req)
}

func (ipt *Input) authenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if ipt.BasicUsername != "" && ipt.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(ipt.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(ipt.BasicPassword)) != 1 {
			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
	}
	handler(res, req)
}

func (ipt *Input) serveWrite(res http.ResponseWriter, req *http.Request) {
	t := time.Now()
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > ipt.MaxBodySize {
		if err := tooLarge(res); err != nil {
			l.Debugf("error in too-large: %v", err)
		}
		return
	}

	// Check if the requested HTTP method was specified in config.
	if !ipt.isAcceptedMethod(req.Method) {
		if err := methodNotAllowed(res); err != nil {
			l.Debugf("error in method-not-allowed: %v", err)
		}
		return
	}

	var bytes []byte
	var ok bool
	switch strings.ToLower(ipt.DataSource) {
	case query:
		bytes, ok = ipt.collectQuery(res, req)
	default:
		bytes, ok = ipt.collectBody(res, req)
	}
	if !ok {
		return
	}

	// If h.Output is configured, data is written to disk file path specified by h.Output.
	// Data will no more be written to datakit io.
	if ipt.Output != "" {
		err := ipt.writeFile(bytes)
		if err != nil {
			l.Warnf("fail to write data to file: %v", err)
		}
		res.WriteHeader(http.StatusNoContent)
		return
	}

	metrics, err := ipt.Parse(bytes, ipt)
	if err != nil {
		l.Debugf("parse error: %s", err.Error())
		if err := badRequest(res); err != nil {
			l.Debugf("error in bad-request: %v", err)
		}
		return
	}

	queryTags := map[string]string{}
	for k, v := range req.URL.Query() {
		if len(v) > 0 {
			queryTags[k] = v[0]
		}
	}

	// Add HTTP header tags and custom tags.
	for idx := range metrics {
		for headerName, tagName := range ipt.HTTPHeaderTags {
			headerValues := req.Header.Get(headerName)
			if len(headerValues) > 0 {
				metrics[idx].AddTag(tagName, headerValues)
			}
		}

		for k, v := range queryTags {
			metrics[idx].AddTag(k, v)
		}

		ipt.SetTags(metrics[idx])
	}
	if len(metrics) > 0 {
		if err := ipt.feeder.Feed(inputName,
			point.Metric,
			metrics,
			&dkio.Option{CollectCost: time.Since(t)}); err != nil {
			l.Warnf("Feed failed: %s, ignored", err)
		}
	}
	res.WriteHeader(http.StatusNoContent)
}

func (ipt *Input) isAcceptedMethod(method string) bool {
	for _, m := range ipt.Methods {
		if method == m {
			return true
		}
	}
	return false
}

func (ipt *Input) SetTags(pt *point.Point) {
	ipt.addTags(pt)
	ipt.renameTags(pt)

	if (len(ipt.TagsIgnoreRegex) > 0 || len(ipt.TagsIgnore) > 0) && (len(ipt.TagsOnlyRegex) > 0 || len(ipt.TagsOnly) > 0) {
		return // If both white list and black list, all list cancel.
	} else {
		ipt.ignoreTags(pt)
		ipt.ignoreTagsRegex(pt)
		ipt.onlyTags(pt)
	}
}

func (ipt *Input) addTags(pt *point.Point) {
	for k, v := range ipt.Tags {
		pt.AddTag(k, v)
	}
}

func (ipt *Input) ignoreTags(pt *point.Point) {
	for _, t := range ipt.TagsIgnore {
		pt.Del(t)
	}
}

func (ipt *Input) ignoreTagsRegex(pt *point.Point) {
	if len(ipt.TagsIgnoreRegex) == 0 {
		return
	}

	for _, kv := range pt.KVs() {
		if !kv.IsTag {
			continue
		}

		for _, r := range ipt.TagsIgnoreRegex {
			match, err := regexp.MatchString(r, kv.Key)
			if err != nil {
				continue
			}
			if match {
				pt.Del(kv.Key)
				break
			}
		}
	}
}

func (ipt *Input) onlyTags(pt *point.Point) {
	if len(ipt.TagsOnly) == 0 && len(ipt.TagsOnlyRegex) == 0 {
		return
	}

	keys := map[string]bool{}
	for _, kv := range pt.Tags() {
		key := kv.Key
		keys[key] = false

		for _, r := range ipt.TagsOnlyRegex {
			match, err := regexp.MatchString(r, key)
			if err != nil {
				continue
			}
			if match {
				keys[key] = true
				break
			}
		}

		if keys[key] {
			continue
		}

		for _, t := range ipt.TagsOnly {
			if key == t {
				keys[key] = true
				break
			}
		}
	}

	for k, v := range keys {
		if !v {
			pt.Del(k)
		}
	}
}

func (ipt *Input) renameTags(pt *point.Point) {
	for oldKey, newKey := range ipt.TagsRename {
		valOld := pt.GetTag(oldKey)
		if len(valOld) == 0 {
			continue
		}
		has := true
		if val := pt.GetTag(newKey); len(val) == 0 {
			has = false
		}
		if has && ipt.Overwrite || !has {
			pt.Del(oldKey)
			pt.Del(newKey)
			pt.AddTag(newKey, valOld)
		}
	}
}

// writeFile writes data to path specified by h.Output.
// If file already exists, simply truncate it.
func (ipt *Input) writeFile(data []byte) error {
	fp := ipt.Output
	if !path.IsAbs(fp) {
		dir := datakit.InstallDir
		fp = filepath.Join(dir, fp)
	}

	f, err := os.Create(filepath.Clean(fp))
	if err != nil {
		return err
	}

	defer f.Close() //nolint:errcheck,gosec
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func (ipt *Input) collectBody(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
	encoding := req.Header.Get("Content-Encoding")

	switch encoding {
	case "gzip":
		r, err := gzip.NewReader(req.Body)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		defer r.Close() //nolint:errcheck
		maxReader := http.MaxBytesReader(res, r, ipt.MaxBodySize)
		bytes, err := io.ReadAll(maxReader)
		if err != nil {
			if err := tooLarge(res); err != nil {
				l.Debugf("error in too-large: %v", err)
			}
			return nil, false
		}
		return bytes, true
	case "snappy":
		defer req.Body.Close() //nolint:errcheck
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		// snappy block format is only supported by decode/encode not snappy reader/writer
		bytes, err = snappy.Decode(nil, bytes)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		return bytes, true
	default:
		defer req.Body.Close() //nolint:errcheck
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			l.Debug(err.Error())
			if err := badRequest(res); err != nil {
				l.Debugf("error in bad-request: %v", err)
			}
			return nil, false
		}
		return bytes, true
	}
}

func (ipt *Input) collectQuery(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
	rawQuery := req.URL.RawQuery

	query, err := url.QueryUnescape(rawQuery)
	if err != nil {
		l.Debugf("Error parsing query: %s", err.Error())
		if err := badRequest(res); err != nil {
			l.Debugf("error in bad-request: %v", err)
		}
		return nil, false
	}

	return []byte(query), true
}

func tooLarge(res http.ResponseWriter) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusRequestEntityTooLarge)
	_, err := res.Write([]byte(`{"error":"http: request body too large"}`))
	return err
}

func methodNotAllowed(res http.ResponseWriter) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusMethodNotAllowed)
	_, err := res.Write([]byte(`{"error":"http: method not allowed"}`))
	return err
}

func badRequest(res http.ResponseWriter) error {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusBadRequest)
	_, err := res.Write([]byte(`{"error":"http: bad request"}`))
	return err
}

func (ipt *Input) SampleConfig() string {
	return sample
}

func (ipt *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

func (ipt *Input) AvailableArchs() []string {
	return datakit.AllOS
}

func defaultInput() *Input {
	i := Input{
		Methods:        []string{"POST", "PUT"},
		DataSource:     body,
		Tags:           map[string]string{},
		TagsRename:     map[string]string{},
		HTTPHeaderTags: map[string]string{},
		TagsIgnore:     []string{},
		MaxBodySize:    defaultMaxBodySize,
		semStop:        cliutils.NewSem(),
		feeder:         dkio.DefaultFeeder(),
		Tagger:         datakit.DefaultGlobalTagger(),
	}
	return &i
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return defaultInput()
	})
}
