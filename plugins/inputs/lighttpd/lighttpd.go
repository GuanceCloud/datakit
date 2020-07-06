package lighttpd

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "lighttpd"

	defaultMeasurement = "lighttpd"

	configSample = `
# [[inputs.lighttpd]]
#       ## lighttpd status url
#	url = "http://127.0.0.1:8080/server-status"
#
#       ## 指定 lighttpd 版本为 "v1" 或 "v2"
#       version = "v1"
#
#	## 采集周期，时间单位是秒
#	collect_cycle = 60

	mea
`
)

var l *zap.SugaredLogger

type (
	Lighttpd struct {
		C []Impl `toml:"lighttpd"`
	}

	Impl struct {
		URL           string        `toml:"url"`
		Version       string        `toml:"version"`
		Cycle         time.Duration `toml:"collect_cycle"`
		statusURL     string
		statusVersion Version
	}
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Impl{}
	})
}

func (_ *Lighttpd) SampleConfig() string {
	return configSample
}

func (_ *Lighttpd) Catalog() string {
	return inputName
}

func (h *Lighttpd) Run() {
	l = logger.SLogger(inputName)

	for _, c := range h.C {
		go c.start()
	}
}

func (i *Impl) start() {

	switch i.Version {
	case "v1":
		i.statusURL = fmt.Sprintf("%s?json", i.URL)
		i.statusVersion = v1
	case "v2":
		i.statusURL = fmt.Sprintf("%s?format=plain", i.URL)
		i.statusVersion = v2
	default:
		l.Error("invalid lighttpd version")
		return
	}

	ticker := time.NewTicker(time.Second * i.Cycle)
	defer ticker.Stop()

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			pt, err := LighttpdStatusParse(i.statusURL, i.statusVersion, inputName)
			if err != nil {
				l.Error(err)
				continue
			}

			io.Feed([]byte(pt.String()), io.Metric)
		}
	}
}
