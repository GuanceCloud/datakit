package system

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	conntrackutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/conntrack"
	filefdutil "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/hostutil/filefd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	minInterval = time.Second
	maxInterval = time.Minute
)

const (
	inputName           = "system"
	metricNameSystem    = "system"
	metricNameConntrack = "conntrack"
	metricNameFilefd    = "filefd"
	sampleCfg           = `
[[inputs.system]]
  ##(optional) collect interval, default is 10 seconds
  interval = '10s'

  [inputs.system.tags]
    # some_tag = "some_value"
    # more_tag = "some_other_value"
`
)

var l = logger.DefaultSLogger(inputName)

type Input struct {
	Interval  datakit.Duration
	Fielddrop []string // Deprecated
	Tags      map[string]string

	collectCache []inputs.Measurement
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) SampleConfig() string {
	return sampleCfg
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&systemMeasurement{},
		&conntrackMeasurement{},
		&filefdMeasurement{},
	}
}

func (i *Input) Collect() error {
	// clear collectCache
	i.collectCache = make([]inputs.Measurement, 0)

	ts := time.Now()

	loadAvg, err := load.Avg()

	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}
	numCPUs, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	tags := map[string]string{}
	for k, v := range i.Tags {
		tags[k] = v
	}

	if runtime.GOOS == "linux" {
		conntrackStat := conntrackutil.GetConntrackInfo()

		conntrackM := conntrackMeasurement{
			name: metricNameConntrack,
			fields: map[string]interface{}{
				"entries":             conntrackStat.Current,
				"entries_limit":       conntrackStat.Limit,
				"stat_found":          conntrackStat.Found,
				"stat_invalid":        conntrackStat.Invalid,
				"stat_ignore":         conntrackStat.Ignore,
				"stat_insert":         conntrackStat.Insert,
				"stat_insert_failed":  conntrackStat.InsertFailed,
				"stat_drop":           conntrackStat.Drop,
				"stat_early_drop":     conntrackStat.EarlyDrop,
				"stat_search_restart": conntrackStat.SearchRestart,
			},
			tags: tags,
			ts:   ts,
		}

		i.collectCache = append(i.collectCache, &conntrackM)

		filefdStat, err := filefdutil.GetFileFdInfo()
		if err != nil {
			l.Warnf("filefdutil.GetFileFdInfo(): %s, ignored", err.Error())
		} else {
			filefdM := filefdMeasurement{
				name: metricNameFilefd,
				fields: map[string]interface{}{
					"allocated":    filefdStat.Allocated,
					"maximum_mega": filefdStat.MaximumMega,
				},
				tags: tags,
				ts:   ts,
			}

			i.collectCache = append(i.collectCache, &filefdM)
		}
	}

	sysM := systemMeasurement{
		name: metricNameSystem,
		fields: map[string]interface{}{
			"load1":           loadAvg.Load1,
			"load5":           loadAvg.Load5,
			"load15":          loadAvg.Load15,
			"load1_per_core":  loadAvg.Load1 / float64(numCPUs),
			"load5_per_core":  loadAvg.Load5 / float64(numCPUs),
			"load15_per_core": loadAvg.Load15 / float64(numCPUs),
			"n_cpus":          numCPUs,
		},
		tags: tags,
		ts:   ts,
	}

	users, err := host.Users()
	if err == nil {
		sysM.fields["n_users"] = len(users)
	} else if os.IsNotExist(err) {
		l.Debugf("Reading users: %s", err.Error())
	} else if os.IsPermission(err) {
		l.Debug(err.Error())
	}
	uptime, err := host.Uptime()
	if err == nil {
		sysM.fields["uptime"] = uptime
	}

	i.collectCache = append(i.collectCache, &sysM)

	return err
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("system input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)}); errFeed != nil {
					io.FeedLastError(inputName, errFeed.Error())
					l.Error(errFeed)
				}
				// i.collectCache = make([]inputs.Measurement, 0)
			} else {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
		case <-datakit.Exit.Wait():
			l.Infof("system input exit")
			return
		}
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval: datakit.Duration{Duration: time.Second * 10},
		}
	})
}
