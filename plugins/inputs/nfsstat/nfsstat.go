// +build linux

package nfsstat

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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
    location = "/proc/net/rpc/nfsd"
    
    # valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"
    interval = "10s"
    
    # [inputs.nfsstat.tags]
    # tags1 = "value1"
`
	nfsStatFileLocation = "/proc/net/rpc/nfsd"
)

var l = logger.DefaultSLogger(inputName)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &NFSstat{}
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

	if n.loadcfg() {
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
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
				continue
			}
			l.Debugf("feed %d bytes to io ok", len(data))
		}
	}
}

func (n *NFSstat) loadcfg() bool {
	if n.Location == "" {
		n.Location = nfsStatFileLocation
		l.Infof("location is empty, use default location %s", nfsStatFileLocation)
	}

	for {
		select {
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return true
		default:
			// nil
		}

		d, err := time.ParseDuration(n.Interval)
		if err != nil || d <= 0 {
			l.Errorf("invalid interval")
			time.Sleep(time.Second)
			continue
		}
		n.duration = d

		if _, err := os.Stat(n.Location); err != nil {
			l.Error(err)
			time.Sleep(time.Second)
			continue
		}

		break
	}

	if n.Tags == nil {
		n.Tags = make(map[string]string)
	}
	if _, ok := n.Tags["location"]; !ok {
		n.Tags["location"] = n.Location
	}

	return false
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
