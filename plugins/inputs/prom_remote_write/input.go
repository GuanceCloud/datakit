package prom_remote_write

import (
	"compress/gzip"
	"crypto/subtle"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	iod "gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"github.com/golang/snappy"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var l = logger.DefaultSLogger(inputName)

const (
	body                   = "body"
	query                  = "query"
	defaultRemoteWritePath = "/prom_remote_write"
)

type Input struct {
	Path           string            `toml:"path"`
	Methods        []string          `toml:"methods"`
	DataSource     string            `toml:"data_source"`
	MaxBodySize    int64             `toml:"max_body_size"`
	BasicUsername  string            `toml:"basic_username"`
	BasicPassword  string            `toml:"basic_password"`
	HTTPHeaderTags map[string]string `toml:"http_header_tags"`
	Tags           map[string]string `toml:"tags"`
	TagsIgnore     []string          `toml:"tags_ignore"`
	Output         string            `toml:"output"`
	Parser
}

func (h *Input) RegHttpHandler() {
	if h.Path == "" {
		h.Path = defaultRemoteWritePath
	}
	for _, m := range h.Methods {
		dkhttp.RegHttpHandler(m, h.Path, h.ServeHTTP)
	}
}

func (h *Input) Catalog() string {
	return catalog
}

func (h *Input) Run() {
	l.Infof("%s input started...", inputName)
	if h.MaxBodySize == 0 {
		h.MaxBodySize = defaultMaxBodySize
	}
	for i, m := range h.Methods {
		h.Methods[i] = strings.ToUpper(m)
	}
}

// ServeHTTP accepts prometheus remote writing, then parses received
// metrics, and sends them to datakit io or local disk file.
func (h *Input) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	handler := h.serveWrite

	h.authenticateIfSet(handler, res, req)
}

func (h *Input) authenticateIfSet(handler http.HandlerFunc, res http.ResponseWriter, req *http.Request) {
	if h.BasicUsername != "" && h.BasicPassword != "" {
		reqUsername, reqPassword, ok := req.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(reqUsername), []byte(h.BasicUsername)) != 1 ||
			subtle.ConstantTimeCompare([]byte(reqPassword), []byte(h.BasicPassword)) != 1 {
			http.Error(res, "Unauthorized.", http.StatusUnauthorized)
			return
		}
	}
	handler(res, req)
}

func (h *Input) serveWrite(res http.ResponseWriter, req *http.Request) {
	t := time.Now()
	// Check that the content length is not too large for us to handle.
	if req.ContentLength > h.MaxBodySize {
		if err := tooLarge(res); err != nil {
			l.Debugf("error in too-large: %v", err)
		}
		return
	}

	// Check if the requested HTTP method was specified in config.
	isAcceptedMethod := false
	for _, method := range h.Methods {
		if req.Method == method {
			isAcceptedMethod = true
			break
		}
	}
	if !isAcceptedMethod {
		if err := methodNotAllowed(res); err != nil {
			l.Debugf("error in method-not-allowed: %v", err)
		}
		return
	}

	var bytes []byte
	var ok bool
	switch strings.ToLower(h.DataSource) {
	case query:
		bytes, ok = h.collectQuery(res, req)
	default:
		bytes, ok = h.collectBody(res, req)
	}
	if !ok {
		return
	}

	// If h.Output is configured, data is written to disk file path specified by h.Output.
	// Data will no more be written to datakit io.
	if h.Output != "" {
		err := h.writeFile(bytes)
		if err != nil {
			l.Warnf("fail to write data to file: %v", err)
		}
		res.WriteHeader(http.StatusNoContent)
		return
	}

	metrics, err := h.Parse(bytes)
	if err != nil {
		l.Debugf("parse error: %s", err.Error())
		if err := badRequest(res); err != nil {
			l.Debugf("error in bad-request: %v", err)
		}
		return
	}

	// Add HTTP header tags and custom tags.
	for i := range metrics {
		m := metrics[i].(*Measurement)
		for headerName, measurementName := range h.HTTPHeaderTags {
			headerValues := req.Header.Get(headerName)
			if len(headerValues) > 0 {
				m.tags[measurementName] = headerValues
			}
		}
		h.AddAndIgnoreTags(m)
	}
	if len(metrics) > 0 {
		if err := inputs.FeedMeasurement(inputName,
			datakit.Metric,
			metrics,
			&iod.Option{CollectCost: time.Since(t)}); err != nil {
			l.Warnf("inputs.FeedMeasurement: %s, ignored", err)
		}
	}
	res.WriteHeader(http.StatusNoContent)
}

func (h *Input) AddAndIgnoreTags(m *Measurement) {
	for k, v := range h.Tags {
		m.tags[k] = v
	}
	for _, t := range h.TagsIgnore {
		if _, has := m.tags[t]; has {
			delete(m.tags, t)
		}
	}
}

// writeFile writes data to path specified by h.Output.
// If file already exists, simply truncate it.
func (h *Input) writeFile(data []byte) error {
	fp := h.Output
	if !path.IsAbs(fp) {
		dir := datakit.InstallDir
		fp = filepath.Join(dir, fp)
	}
	f, err := os.Create(fp)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return nil
}

func (h *Input) collectBody(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
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
		defer r.Close()
		maxReader := http.MaxBytesReader(res, r, h.MaxBodySize)
		bytes, err := io.ReadAll(maxReader)
		if err != nil {
			if err := tooLarge(res); err != nil {
				l.Debugf("error in too-large: %v", err)
			}
			return nil, false
		}
		return bytes, true
	case "snappy":
		defer req.Body.Close()
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
		defer req.Body.Close()
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

func (h *Input) collectQuery(res http.ResponseWriter, req *http.Request) ([]byte, bool) {
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

func (h *Input) SampleConfig() string {
	return sample
}

func (h *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&Measurement{}}
}

func (h *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func NewInput() *Input {
	i := Input{
		Methods:    []string{"POST", "PUT"},
		DataSource: body,
	}
	return &i
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return NewInput()
	})
}
