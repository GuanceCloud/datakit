// +build linux

package nfsstat

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"

	"github.com/prometheus/procfs/nfs"
)

const (
	inputName = "nfsstat"

	defaultMeasurement = "nfsstat"

	sampleCfg = `
[inputs.nfsstat]
    # nfsstat file location. default "/proc/net/rpc/nfsd"
    # required
    location = "/proc/net/rpc/nfsd"

    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    # required
    interval = "10s"

    # [inputs.nfsstat.tags]
    # tags1 = "value1"
`
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &NFSstat{
			Interval: config.Cfg.IntervalDeprecated,
			Tags:     make(map[string]string),
		}
	})
}

type NFSstat struct {
	Location string            `toml:"location"`
	Interval string            `toml:"interval"`
	Tags     map[string]string `toml:"tags"`
	duration time.Duration
}

func (*NFSstat) SampleConfig() string {
	return sampleCfg
}

func (*NFSstat) Catalog() string {
	return inputName
}

func (n *NFSstat) Run() {
	l = logger.SLogger(inputName)

	if n.initCfg() {
		return
	}

	ticker := time.NewTicker(n.duration)
	defer ticker.Stop()

	l.Infof("nfsstat input started...")

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return

		case <-ticker.C:
			data, err := buildPoint(n.Location, n.Tags)
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

func (n *NFSstat) initCfg() bool {
	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		if err := n.loadCfg(); err != nil {
			l.Error(err)
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	return false
}

func (n *NFSstat) loadCfg() (err error) {
	if n.Location == "" {
		err = fmt.Errorf("location cannot be empty")
		return
	}

	n.duration, err = time.ParseDuration(n.Interval)
	if err != nil {
		err = fmt.Errorf("invalid interval, %s", err.Error())
		return
	} else if n.duration <= 0 {
		err = fmt.Errorf("invalid interval, cannot be less than zero")
		return
	}

	if _, ok := n.Tags["location"]; !ok {
		n.Tags["location"] = n.Location
	}

	return
}

func buildPoint(fn string, tags map[string]string) ([]byte, error) {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, fmt.Errorf("could not open %s", fn)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("file is empty")
	}

	stats, err := nfs.ParseServerRPCStats(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}

	var fields = make(map[string]interface{})
	deepHit(*stats, "nfs_", fields)

	return io.MakeMetric(defaultMeasurement, tags, fields)
}

func deepHit(data interface{}, prefix string, m map[string]interface{}) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanInterface() {
			key := strings.ToLower(t.Field(i).Name)
			switch v.Field(i).Kind() {
			case reflect.Struct:
				deepHit(v.Field(i).Interface(), prefix+key+"_", m)

			case reflect.Uint64:
				m[prefix+key] = int64(v.Field(i).Uint())

			default:
				// nil
			}
		}
	}
}
