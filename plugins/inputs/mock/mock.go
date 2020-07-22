package mock

import (
	"os"
	"time"

	"github.com/Pallinder/go-randomdata"
	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/influxdata/telegraf"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l *logger.Logger

	inputName = "mock"

	sampleCfg = `
# [inputs.mock]
# interval = '3s'
# metric = 'mock-testing'
	`
)

type Mock struct {
	Interval string `toml:"interval"`
	Metric   string `toml:"metric"`
}

func (m *Mock) SampleConfig() string {
	return sampleCfg
}

func (m *Mock) Description() string {
	return "mock testing data"
}

func (m *Mock) Catalog() string {
	return "mock"
}

func (m *Mock) Gather(acc telegraf.Accumulator) error {
	return nil
}

func (m *Mock) Run() {

	l = logger.SLogger("mock")
	host, err := os.Hostname()
	if err != nil {
		host = randomdata.SillyName()
		l.Warnf("get hostname failed: %s, use random silly name(%s) instead", err, host)
	}

	l.Info("mock input started...")

	interval, err := time.ParseDuration(m.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			pt, err := influxdb.NewPoint(m.Metric,
				map[string]string{
					"from": host,
				},
				map[string]interface{}{
					"f1": randomdata.Number(0, 100),
					"f2": randomdata.Decimal(0, 100),
					"f3": randomdata.SillyName(),
					"f4": randomdata.Boolean()},
				time.Now())
			if err != nil {
				l.Error(err)
				return
			}

			data := []byte(pt.String())
			if err := io.NamedFeed(data, io.Metric, inputName); err != nil {
				l.Error(err)
			} else {
				l.Debugf("feed %d bytes to io ok", len(data))
			}

		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func init() {
	inputs.Add("mock", func() inputs.Input {
		return &Mock{}
	})
}
