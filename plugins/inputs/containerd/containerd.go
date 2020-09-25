// +build linux

package containerd

import (
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
    # containerd sock file, default "/run/containerd/containerd.sock"
    # required
    location = "/run/containerd/containerd.sock"
    
    # containerd namespace
    # 'ps -ef | grep containerd | grep containerd-shim' print detail
    # required
    namespace = "moby"
    
    # containerd ID list，ID is string and length 64.
    # if value is "*", collect all ID
    # required
    ID_list = ["*"]
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"
    
    # [inputs.containerd.tags]
    # tags1 = "value1"
`
	containerdSock = "/run/containerd/containerd.sock"
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Containerd{}
	})
}

type Containerd struct {
	Location  string            `toml:"location"`
	Namespace string            `toml:"namespace"`
	Interval  string            `toml:"interval"`
	IDList    []string          `toml:"ID_list"`
	Tags      map[string]string `toml:"tags"`

	// forward compatibility
	HostPath     string `toml:"host_path"`
	CollectCycle string `toml:"collect_cycle"`
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

	if c.loadcfg() {
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
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (c *Containerd) loadcfg() bool {
	if c.Location == "" && c.HostPath != "" {
		c.Location = c.HostPath
	}
	if c.Location == "" {
		c.Location = containerdSock
		l.Infof("location is empty, use default location %s", containerdSock)
	}
	if c.Interval == "" && c.CollectCycle != "" {
		c.Interval = c.CollectCycle
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		d, err := time.ParseDuration(c.Interval)
		if err != nil || d <= 0 {
			l.Errorf("invalid interval, %s", err.Error())
			time.Sleep(time.Second)
			continue
		}
		c.duration = d
		break
	}

	if c.Tags == nil {
		c.Tags = make(map[string]string)
	}

	c.isAll = len(c.IDList) == 1 && c.IDList[0] == "*"

	c.ids = func() map[string]interface{} {
		m := make(map[string]interface{})
		for _, v := range c.IDList {
			m[v] = nil
		}
		return m
	}()

	return false
}
