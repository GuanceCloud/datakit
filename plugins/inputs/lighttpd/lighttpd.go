package lighttpd

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "lighttpd"

	defaultMeasurement = "lighttpd"

	sampleCfg = `
[[inputs.lighttpd]]
    # lighttpd status url
    # required
    url = "http://127.0.0.1:8080/server-status"
    
    # lighttpd version is "v1" or "v2"
    # required
    version = "v1"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # [inputs.lighttpd.tags]
    # from = "127.0.0.1:8080"
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Lighttpd{}
	})
}

type Lighttpd struct {
	URL      string            `toml:"url"`
	Version  string            `toml:"version"`
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`

	statusURL     string
	statusVersion Version

	duration time.Duration
}

func (*Lighttpd) SampleConfig() string {
	return sampleCfg
}

func (*Lighttpd) Catalog() string {
	return inputName
}

func (h *Lighttpd) Run() {
	l = logger.SLogger(inputName)

	if h.loadcfg() {
		return
	}
	ticker := time.NewTicker(h.duration)
	defer ticker.Stop()

	l.Infof("lighttpd input started.")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := h.getMetrics()
			if err != nil {
				l.Error(err)
				continue
			}
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (h *Lighttpd) loadcfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		d, err := time.ParseDuration(h.Interval)
		if err != nil || d <= 0 {
			l.Errorf("invalid interval, %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		h.duration = d

		if h.Version == "v1" {
			h.statusURL = fmt.Sprintf("%s?json", h.URL)
			h.statusVersion = v1
			break
		} else if h.Version == "v2" {
			h.statusURL = fmt.Sprintf("%s?format=plain", h.URL)
			h.statusVersion = v2
			break
		} else {
			l.Error("invalid lighttpd version")
			time.Sleep(time.Second)
		}
	}

	return false
}
