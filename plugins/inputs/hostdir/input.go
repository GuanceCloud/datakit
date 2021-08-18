package hostdir

import (
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

func (i *Input) SampleConfig() string {
	return sample
}

func (i *Input) appendMeasurement(name string, tags map[string]string, fields map[string]interface{}, ts time.Time) {
	tmp := &Measurement{name: name, tags: tags, fields: fields, ts: ts}
	i.collectCache = append(i.collectCache, tmp)
}

func (i *Input) Catalog() string {
	return "host"
}

func (i *Input) Collect() error {
	timeNow := time.Now()
	var tags map[string]string
	path := i.Dir
	if i.platform == "windows" {
		filesystem, err := GetFileSystemType(path)
		if err != nil {
			return err
		}
		dirMode, err := Getdirmode(path)
		if err != nil {
			return err
		}
		tags = map[string]string{
			"file_mode":      dirMode,
			"file_system":    filesystem,
			"host_directory": i.Dir,
		}
	} else {
		filesystem, err := GetFileSystemType(path)
		if err != nil {
			return err
		}
		dirMode, err := Getdirmode(path)
		if err != nil {
			return err
		}
		fileownership, err := GetFileOwnership(path, i.platform)
		if err != nil {
			return err
		}
		tags = map[string]string{
			"file_mode":      dirMode,
			"file_system":    filesystem,
			"file_ownership": fileownership,
			"host_directory": i.Dir,
		}
	}

	for k, v := range i.Tags {
		tags[k] = v
	}
	filesize, filecount, dircount := Startcollect(i.Dir, i.ExcludePatterns)
	fields := map[string]interface{}{
		"file_size":  filesize,
		"file_count": filecount,
		"dir_count":  dircount,
	}
	i.appendMeasurement(inputName, tags, fields, timeNow)
	return nil
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("hostdir input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	tick := time.NewTicker(i.Interval.Duration)
	isfirstRun := true
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if errFeed := inputs.FeedMeasurement(metricName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since((start))}); errFeed != nil {
					if !isfirstRun {
						io.FeedLastError(inputName, errFeed.Error())
					} else {
						isfirstRun = false
					}
				}
			} else {
				io.FeedLastError(inputName, err.Error())
				l.Error(err)
			}
			i.collectCache = make([]inputs.Measurement, 0)
		case <-datakit.Exit.Wait():
			l.Infof("hostdir input exit")
			return
		}
	}
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&Measurement{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		s := &Input{
			Interval: datakit.Duration{Duration: time.Second * 5},
			platform: runtime.GOOS,
		}
		return s
	})
}
