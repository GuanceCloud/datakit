package fluentdlog

import (
	"bufio"
	iowrite "io"
	"net/http"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "fluentdlog"

	defaultMeasurement = "fluentd"

	sampleCfg = `
[[inputs.fluentdlog]]
    # http server route path
    # input url(required)
	path = "/fluentd"
	# log source(required)
	source = ""
    # [inputs.fluentdlog.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Fluentd{}
	})
}

type Fluentd struct {
	Path         string             `toml:"path"`
	Pipeline     *pipeline.Pipeline `toml:"-"`
	PipelinePath string             `toml:"pipeline_path"`
	Metric       string             `toml:"source"`
	Tags         map[string]string  `toml:"tags"`
}

func (*Fluentd) SampleConfig() string {
	return sampleCfg
}

func (*Fluentd) Catalog() string {
	return "fluentd"
}

func (f *Fluentd) Run() {
	l = logger.SLogger(inputName)
	l.Infof("Fluentd input started...")
	var err error
	f.Pipeline, err = pipeline.NewPipelineFromFile(f.PipelinePath)
	if err != nil {
		l.Errorf("new pipeline error, %v", err)
	}
}

func (f *Fluentd) RegHttpHandler() {
	httpd.RegHttpHandler("POST", f.Path, f.Handle)
}

func (f *Fluentd) Handle(w http.ResponseWriter, r *http.Request) {
	if err := extract(f.Pipeline, r.Body, f.Metric, f.Tags); err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func extract(p *pipeline.Pipeline, body iowrite.Reader, metric string, tags map[string]string) error {
	r := bufio.NewReader(body)

	var pts []*io.Point

	for {
		bytes, err := r.ReadBytes(byte('\n'))
		if err == iowrite.EOF || string(bytes) == "" {
			break
		}

		_tags := make(map[string]string)
		_fields := make(map[string]interface{})

		for key, val := range tags {
			_tags[key] = val
		}

		tm := time.Now()

		n := len(bytes)

		if p != nil {
			_fields, err = p.Run(string(bytes[0 : n-1])).Result()
			if err != nil {
				l.Errorf("run pipeline error, %v", err)
				continue
			}
		} else {
			_fields["content"] = string(bytes[0 : n-1])
		}

		pt, err := io.MakePoint(metric, _tags, _fields, tm)
		if err != nil {
			l.Errorf("make metric point error %v", err)
			continue
		}
		pts = append(pts, pt)

	}

	if err := io.Feed(inputName, datakit.Logging, pts, &io.Option{HighFreq: true}); err != nil {
		l.Errorf("push metric point error %v", err)
		return err
	}

	return nil
}
