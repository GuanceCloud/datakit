package httpProb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
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
    drop_body = false
    # pipeline = "all_route.p" # datakit/pipeline/all_route.p

	[[inputs.httpProb.url]]
    # uri = "/user/info"
    # uri_regex = "/user/info/*"
    drop_body = false
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
	DropBody     bool               `toml:"drop_body"`
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

	h.InitPipeline()

	listen := fmt.Sprintf("%s:%v", h.Bind, h.Port)
	l.Info("HttpProb server start...", h.Port)

	http.ListenAndServe(listen, h)
}

// init pipeline
func (h *HttpProb) InitPipeline() {
	for _, item := range h.Url {
		if item.PipelinePath != "" {
			var err error
			item.Pipeline, err = pipeline.NewPipelineFromFile(item.PipelinePath)
			if err != nil {
				l.Errorf("pipline init fail %v", err)
			}
		}
	}
}

func (h *HttpProb) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "ok")

	go h.handle(w, req)
}

func (h *HttpProb) handle(w http.ResponseWriter, req *http.Request) {
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

		if item.Uri == "/" || req.URL.Path == item.Uri {
			filter = true
		}

		// match path
		if filter {
			if item.Pipeline != nil {
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
			if !item.DropBody && req.Body != nil {
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

			tags["version"] = req.Proto
			tags["method"] = req.Method
			tags["uri"] = req.URL.Path

			var resData = make(map[string]interface{})
			resData["version"] = req.Proto
			resData["method"] = req.Method
			resData["uri"] = req.URL.Path

			resData["queryParams"] = query

			if len(header) != 0 {
				resData["header"] = header
			}

			if !item.DropBody && buf != nil {
				contentType := header["Content-Type"]
				if strings.Contains(contentType, "application/json") {
					var body interface{}
					err := json.Unmarshal(buf.Bytes(), &body)
					if err != nil {
						l.Errorf("body json parse error %v", err)
					} else {
						resData["body"] = body
					}
				}
			}

			data, err := json.Marshal(resData)
			if item.Pipeline != nil {
				fields, err = item.Pipeline.Run(string(data)).Result()
				if err != nil {
					l.Errorf("run pipeline error, %s", err)
				}
			} else {
				fields = resData
				delete(fields, "header")
				delete(fields, "queryParams")
				delete(fields, "body")

				for k, v := range query {
					key := "query." + k
					fields[key] = v
				}

				for k, v := range header {
					key := "header." + k
					fields[key] = v
				}
			}

			pt, err := io.MakeMetric(h.Source, tags, fields, time.Now())
			if err != nil {
				l.Errorf("make metric point error %v", err)
			}

			l.Info("point ======>", string(pt))

			err = io.HighFreqFeed([]byte(pt), io.Logging, inputName)
			if err != nil {
				l.Errorf("push metric point error %v", err)
			}
		}
	}
}
