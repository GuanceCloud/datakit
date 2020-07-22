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
# [inputs.containerd]
# 	# containerd sock file, use default
# 	host_path = "/run/containerd/containerd.sock"
#
# 	# containerd namespace
# 	# 'ps -ef | grep containerd | grep containerd-shim' print detail
# 	namespace = "moby"
#
# 	# containerd ID list，ID is string and length 64.
# 	# if value is "*", collect all ID
# 	ID_list = ["*"]
#
# 	# valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
# 	collect_cycle = "60s"
#
# 	# [inputs.containerd.tags]
# 	# tags1 = "tags1"
`
)

var (
	l *logger.Logger

	testAssert = false
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Containerd{}
	})
}

type Containerd struct {
	HostPath     string            `toml:"host_path"`
	Namespace    string            `toml:"namespace"`
	IDList       []string          `toml:"ID_list"`
	CollectCycle string            `toml:"collect_cycle"`
	Tags         map[string]string `toml:"tags"`
	// get all ids metrics
	isAll bool
	// id cache
	ids map[string]interface{}
}

func (_ *Containerd) Catalog() string {
	return inputName
}

func (_ *Containerd) SampleConfig() string {
	return sampleCfg
}

func (c *Containerd) Run() {
	l = logger.SLogger(inputName)

	d, err := time.ParseDuration(c.CollectCycle)
	if err != nil || d <= 0 {
		l.Errorf("invalid duration of collect_cycle")
		return
	}
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	c.initcfg()

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
			if testAssert {
				fmt.Printf("containerd data: %s", string(data))
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

func (c *Containerd) initcfg() {
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

}
