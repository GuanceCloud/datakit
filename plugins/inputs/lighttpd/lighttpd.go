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
# [[inputs.lighttpd]]
# 	# lighttpd status url
# 	url = "http://127.0.0.1:8080/server-status"
#
# 	# lighttpd version is "v1" or "v2"
# 	version = "v1"
#
# 	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
# 	collect_cycle = "60s"
#
# 	# [inputs.lighttpd.tags]
# 	# tags1 = "tags1"
`
)

var l *logger.Logger

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Lighttpd{}
	})
}

type Lighttpd struct {
	URL          string            `toml:"url"`
	Version      string            `toml:"version"`
	CollectCycle string            `toml:"collect_cycle"`
	Tags         map[string]string `toml:"tags"`

	statusURL     string
	statusVersion Version
}

func (_ *Lighttpd) SampleConfig() string {
	return sampleCfg
}

func (_ *Lighttpd) Catalog() string {
	return inputName
}

func (h *Lighttpd) Run() {
	l = logger.SLogger(inputName)

	d, err := time.ParseDuration(h.CollectCycle)
	if err != nil || d <= 0 {
		l.Errorf("invalid duration of collect_cycle")
		return
	}
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	if h.initcfg() {
		return
	}

	l.Infof("lighttpd input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := LighttpdStatusParse(h.statusURL, h.statusVersion, h.Tags)
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

func (h *Lighttpd) initcfg() bool {
	if h.Tags == nil {
		h.Tags = make(map[string]string)
	}
	if _, ok := h.Tags["url"]; !ok {
		h.Tags["url"] = h.URL
	}
	if _, ok := h.Tags["version"]; !ok {
		h.Tags["version"] = h.Version
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

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
