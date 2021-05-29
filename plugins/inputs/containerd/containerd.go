// +build linux

package containerd

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "containerd"

	defaultMeasurement = "containerd"

	sampleCfg = `
[inputs.containerd]
    ## containerd sock file, default "/run/containerd/containerd.sock"
    ## required
    location = "/run/containerd/containerd.sock"

    ## containerd namespace
    ## 'ps -ef | grep containerd | grep containerd-shim' print detail
    ## required
    namespace = "moby"

    ## containerd ID list，ID is string and length 64.
    ## if value is "*", collect all ID
    ## required
    ID_list = ["*"]

    ## valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    ## required, cannot be less than zero
    interval = "10s"

    [inputs.containerd.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Containerd{
			Interval: datakit.Cfg.IntervalDeprecated,
			Tags:     make(map[string]string),
			ids:      make(map[string]interface{}),
		}
	})
}

type Containerd struct {
	Location  string            `toml:"location"`
	Namespace string            `toml:"namespace"`
	Interval  string            `toml:"interval"`
	IDList    []string          `toml:"ID_list"`
	Tags      map[string]string `toml:"tags"`
	// get all ids metrics
	isAll bool
	// id cache
	ids map[string]interface{}

	duration time.Duration
}

func (*Containerd) Catalog() string {
	return inputName
}

func (*Containerd) SampleConfig() string {
	return sampleCfg
}

func (c *Containerd) Run() {
	l = logger.SLogger(inputName)

	if c.initCfg() {
		return
	}
	ticker := time.NewTicker(c.duration)
	defer ticker.Stop()

	l.Infof("containerd input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := c.collectContainerd()
			if err != nil {
				l.Error(err)
				continue
			}
			if err := io.NamedFeed(data, datakit.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (c *Containerd) initCfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := c.loadCfg(); err != nil {
			l.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return false
}

func (c *Containerd) loadCfg() (err error) {
	if c.Location == "" {
		err = fmt.Errorf("location cannot be empty")
		return
	}

	c.duration, err = time.ParseDuration(c.Interval)
	if err != nil {
		err = fmt.Errorf("invalid interval, %s", err.Error())
		return
	} else if c.duration <= 0 {
		err = fmt.Errorf("invalid interval, cannot be less than zero")
		return
	}

	c.isAll = len(c.IDList) == 1 && c.IDList[0] == "*"

	for _, v := range c.IDList {
		c.ids[v] = nil
	}

	return
}
