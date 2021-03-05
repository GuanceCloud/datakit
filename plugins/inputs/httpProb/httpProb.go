package httpProb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"net/http"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "httpProb"

	defaultMeasurement = "httpProb"

	sampleCfg = `
[[inputs.httpProb]]
    bind = "0.0.0.0"
	port = 9530
	drop_body = false
	# log source(required)
	source = "xxx-app"

    # global tags
    [inputs.httpProb.tags]
    # tag1 = val1
    # tag2 = val2

    [[inputs.httpProb.url]]
    # uri or uri_regex
    # uri = "/"         # regist all routes
    # uri_regex = "/*"
    # pipeline = "all_route.p" # datakit/pipeline/all_route.p

	[[inputs.httpProb.url]]
    # uri = "/user/info"
    # uri_regex = "/user/info/*"
    # pipeline = "user_info.p" # datakit/pipeline/user_info.p
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &HttpProb{}
	})
}

type Url struct {
	Uri          string             `toml:"uri"`
	UriRegex     string             `toml:"uri_regex"`
	Pipeline     *pipeline.Pipeline `toml:"-"`
	PipelinePath string             `toml:"pipeline"`
}

type HttpProb struct {
	Bind   string            `toml:"bind"`
	Port   int               `toml:"port"`
	Source string            `toml:"source"`
	Tags   map[string]string `toml:"tags"`
	Url    []*Url            `toml:"url"`
}

func (*HttpProb) SampleConfig() string {
	return sampleCfg
}

func (*HttpProb) Catalog() string {
	return "network"
}

func (HttpProb) Test() (result *inputs.TestResult, err error) {
	return
}

func (h *HttpProb) Run() {
	l = logger.SLogger(inputName)
	l.Infof("HttpProb input started...")

	listen := fmt.Sprintf("%s:%v", h.Bind, h.Port)
	l.Info("server start...", h.Port)

	http.ListenAndServe(listen, h)
}

func (h *HttpProb) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var err error
	for _, item := range h.Url {
		var filter bool

		if item.Uri == "" && item.UriRegex != "" {
			filter, err = regexp.MatchString(item.UriRegex, req.URL.Path)
			if err != nil {
				l.Errorf("config uri %s regex error %v", item.UriRegex, err)
				continue
			}
		}

		if item.Uri == "/" ||  req.URL.Path == item.Uri {
			filter = true
		}

		// match path
		if filter  {
			if item.PipelinePath != "" {
				item.Pipeline, err = pipeline.NewPipelineFromFile(item.PipelinePath)
				if err != nil {
					l.Errorf("pipline init fail %v", err)
				}
			}

			var buf = bytes.NewBuffer([]byte{})
			var header = make(map[string]string)

			// 写入header
			for key, item := range req.Header {
				header[key] = strings.Join(item, ",")
			}

			query := make(map[string]interface{})
			for k, v := range req.URL.Query() {
				if len(v) == 1 && len(v[0]) != 0 {
					query[k] = v[0]
				} else {
					break
				}
			}

			// 写入body
			if req.Body != nil {
				_, err := buf.ReadFrom(req.Body)
				if err != nil {
					l.Errorf("read body error", err)
				}
			}

			tags := make(map[string]string)
			fields := make(map[string]interface{})

			for tag, tagV := range h.Tags {
				tags[tag] = tagV
			}

			var resData = make(map[string]interface{})
			resData["method"] = req.Method
			resData["url"] = req.URL.Path

			if len(query) != 0 {
				resData["query"] = query
			}

			if len(header) != 0 {
				resData["header"] = header
			}

			resData["body"] = buf.String()

			data, err := json.Marshal(resData)

			if item.Pipeline != nil {
				fields, err = item.Pipeline.Run(string(data)).Result()
				if err != nil {
					l.Errorf("run pipeline error, %s", err)
				}
			} else {
				fields["message"] = string(data)
			}

			pt, err := io.MakeMetric(h.Source, tags, fields, time.Now())
			if err != nil {
				l.Errorf("make metric point error %v", err)
			}

			l.Info("point ======>", string(pt))

			err = io.NamedFeed([]byte(pt), io.Metric, inputName)
			if err != nil {
				l.Errorf("push metric point error %v", err)
			}
		}
	}

	fmt.Fprintf(w, "ok")
}
