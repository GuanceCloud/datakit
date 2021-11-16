// Package tomcat collect Tomcat metrics.
package tomcat

import (
	"os"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName   = "tomcat"
	minInterval = time.Second * 10
	maxInterval = time.Minute * 20
)

var l = logger.DefaultSLogger(inputName)

type tomcatlog struct {
	Files             []string `toml:"files"`
	Pipeline          string   `toml:"pipeline"`
	IgnoreStatus      []string `toml:"ignore"`
	CharacterEncoding string   `toml:"character_encoding"`
	MultilineMatch    string   `toml:"multiline_match"`
}

type Input struct {
	inputs.JolokiaAgent
	Log  *tomcatlog        `toml:"log"`
	Tags map[string]string `toml:"tags"`

	tail *tailer.Tailer
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return tomcatSampleCfg
}

func (*Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&TomcatGlobalRequestProcessorM{},
		&TomcatJspMonitorM{},
		&TomcatThreadPoolM{},
		&TomcatServletM{},
		&TomcatCacheM{},
	}
}

func (*Input) PipelineConfig() map[string]string {
	pipelineMap := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineMap
}

func (i *Input) GetPipeline() []*tailer.Option {
	return []*tailer.Option{
		{
			Source:   inputName,
			Service:  inputName,
			Pipeline: i.Log.Pipeline,
		},
	}
}

func (i *Input) RunPipeline() {
	if i.Log == nil || len(i.Log.Files) == 0 {
		return
	}

	if i.Log.Pipeline == "" {
		i.Log.Pipeline = inputName + ".p" // use default
	}

	opt := &tailer.Option{
		Source:            inputName,
		Service:           inputName,
		GlobalTags:        i.Tags,
		IgnoreStatus:      i.Log.IgnoreStatus,
		CharacterEncoding: i.Log.CharacterEncoding,
		MultilineMatch:    i.Log.MultilineMatch,
	}

	pl, err := config.GetPipelinePath(i.Log.Pipeline)
	if err != nil {
		l.Error(err)
		io.FeedLastError(inputName, err.Error())
		return
	}
	if _, err := os.Stat(pl); err != nil {
		l.Warn("%s missing: %s", pl, err.Error())
	} else {
		opt.Pipeline = pl
	}

	i.tail, err = tailer.NewTailer(i.Log.Files, opt)
	if err != nil {
		l.Errorf("NewTailer: %s", err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	go i.tail.Start()
}

func (i *Input) Run() {
	go func() {
		for {
			<-datakit.Exit.Wait()
			if i.tail != nil {
				i.tail.Close() //nolint:errcheck
			}
		}
	}()
	if d, err := time.ParseDuration(i.Interval); err != nil {
		i.Interval = (time.Second * 10).String()
	} else {
		i.Interval = config.ProtectedInterval(minInterval, maxInterval, d).String()
	}
	i.PluginName = inputName
	i.JolokiaAgent.Tags = i.Tags
	i.JolokiaAgent.Types = TomcatMetricType
	i.JolokiaAgent.Collect()
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		i := &Input{}
		return i
	})
}
