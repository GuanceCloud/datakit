package httpProb

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

func (h *HttpProb) Run() {
	l = logger.SLogger(inputName)
	l.Infof("HttpProb input started...")

	h.InitPipeline()

	listen := fmt.Sprintf("%s:%v", h.Bind, h.Port)
	l.Info("HttpProb server start...", h.Port)

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		h.handle(req)
		fmt.Fprintf(w, "ok")
	})

	// server
	srv := http.Server{
		Addr:    listen,
		Handler: handler,
	}

	go func() {
		<-datakit.Exit.Wait()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); nil != err {
			l.Fatalf("server shutdown failed, err: %v\n", err)
		}
		l.Info("server gracefully shutdown")
	}()

	err := srv.ListenAndServe()
	if http.ErrServerClosed != err {
		l.Fatalf("server not gracefully shutdown, err :%v\n", err)
	}
}

// init pipeline
func (h *HttpProb) InitPipeline() {
	for _, item := range h.Url {
		if item.PipelinePath != "" {
			var err error
			item.Pipeline, err = pipeline.NewPipelineByScriptPath(item.PipelinePath)
			if err != nil {
				l.Errorf("pipline init fail %v", err)
			}
		}
	}
}

func (h *HttpProb) handle(req *http.Request) {
	var err error

	var pts []*io.Point

	for _, item := range h.Url {
		var filter bool

		if item.Uri == "" && item.UriRegex != "" {
			re := regexp.MustCompile(item.UriRegex)
			filter = re.MatchString(req.URL.Path)
		}

		if item.Uri == "/" || req.URL.Path == item.Uri {
			filter = true
		}

		if !filter {
			continue
		}

		// match path
		if item.Pipeline != nil {
		}

		var body []byte
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
			body, err = ioutil.ReadAll(req.Body)
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

		message := req.Method + " " + req.URL.Path + " " + req.Proto

		if !item.DropBody && body != nil {
			var bodyS interface{}
			contentType := header["Content-Type"]
			if strings.Contains(contentType, "application/json") {
				err := json.Unmarshal(body, &bodyS)
				if err != nil {
					l.Errorf("body json parse error %v", err)
				} else {
					resData["body"] = bodyS
				}
			} else if strings.Contains(contentType, "text/plain") {
				resData["body"] = bodyS
			}
		}

		data, err := json.Marshal(resData)
		if item.Pipeline != nil {
			l.Info("pipeline input data ======>", string(data))
			fields, err = item.Pipeline.Run(string(data)).Result()
			l.Info("pipeline output data ======>", fields)
			if err != nil {
				l.Errorf("run pipeline error, %s", err)
			}
		} else {
			for k, v := range query {
				key := "query." + k
				fields[key] = v
			}

			for k, v := range header {
				key := "header." + k
				fields[key] = v
			}
		}

		fields["message"] = message

		pt, err := io.MakePoint(h.Source, tags, fields, time.Now())
		if err != nil {
			l.Errorf("make metric point error %v", err)
		}
		pts = append(pts, pt)
	}

	if err = io.Feed(inputName, datakit.Logging, pts, &io.Option{HighFreq: true}); err != nil {
		l.Errorf("push metric point error %v", err)
	}
}
