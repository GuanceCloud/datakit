//go:build linux
// +build linux

package sensors

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cmd"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	defCommand  = "sensors"
	defPath     = "/usr/bin/sensors"
	defInterval = datakit.Duration{Duration: 10 * time.Second}
	defTimeout  = datakit.Duration{Duration: 3 * time.Second}
)

type Input struct {
	Path     string            `toml:"path"`
	Interval datakit.Duration  `toml:"interval"`
	Timeout  datakit.Duration  `toml:"timeout"`
	Tags     map[string]string `toml:"tags"`

	semStop          *cliutils.Sem // start stop signal
	semStopCompleted *cliutils.Sem // stop completed signal
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLinux}
}

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{&sensorsMeasurement{}}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Info("sensors input started")

	var err error
	if ipt.Path == "" || !path.IsFileExists(ipt.Path) {
		if ipt.Path, err = exec.LookPath(defCommand); err != nil {
			l.Errorf("Can not find executable sensor command, install 'lm-sensors' first.")

			return
		}
		l.Info("Command fallback to %q due to invalide path provided in 'sensors' input", ipt.Path)
	}

	tick := time.NewTicker(ipt.Interval.Duration)
	for {
		select {
		case <-tick.C:
			if err = ipt.gather(); err != nil {
				l.Errorf("gather: %s", err.Error())
				io.FeedLastError(inputName, err.Error())
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("sensors input exit")

			return

		case <-ipt.semStop.Wait():
			l.Info("sensors input return")

			if ipt.semStopCompleted != nil {
				ipt.semStopCompleted.Close()
			}
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()

		// wait stop completed
		if ipt.semStopCompleted != nil {
			for range ipt.semStopCompleted.Wait() {
				return
			}
		}
	}
}

func (ipt *Input) gather() error {
	start := time.Now()
	output, err := cmd.RunWithTimeout(ipt.Timeout.Duration, false, ipt.Path, "-u")
	if err != nil {
		l.Errorf("Command process failed: %q", output)

		return err
	}

	if cache, err := ipt.parse(string(output)); err != nil {
		return err
	} else {
		return inputs.FeedMeasurement(inputName,
			datakit.Metric,
			cache,
			&io.Option{CollectCost: time.Since(start)})
	}
}

func (ipt *Input) getCustomerTags() map[string]string {
	tags := make(map[string]string)
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	return tags
}

func (ipt *Input) parse(output string) ([]inputs.Measurement, error) {
	var (
		lines  = strings.Split(strings.TrimSpace(output), "\n")
		tags   = ipt.getCustomerTags()
		fields = make(map[string]interface{})
		cache  []inputs.Measurement
	)

	for _, line := range lines {
		if line == "" {
			cache = append(cache, &sensorsMeasurement{
				name:   inputName,
				tags:   tags,
				fields: fields,
				ts:     time.Now(),
			})
			tags = ipt.getCustomerTags()
			fields = make(map[string]interface{})
			continue
		}

		if !strings.Contains(line, ":") {
			tags["chip"] = line
			continue
		}

		parts := strings.Split(line, ":")
		switch {
		case strings.HasSuffix(line, ":"):
			if len(fields) != 0 {
				cache = append(cache, &sensorsMeasurement{
					name:   inputName,
					tags:   tags,
					fields: fields,
					ts:     time.Now(),
				})

				tmp := make(map[string]string)
				for k, v := range tags {
					tmp[k] = v
				}
				tags = tmp
				fields = make(map[string]interface{})
			}
			tags["feature"] = strings.ToLower(strings.ReplaceAll(strings.TrimSpace(parts[0]), " ", "_"))
		case strings.HasPrefix(parts[0], " "):
			if value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err != nil {
				log.Println(err.Error())

				return nil, err
			} else {
				fields[strings.ToLower(strings.TrimSpace(parts[0]))] = value
			}
		default:
			tags[strings.ToLower(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if len(fields) != 0 {
		cache = append(cache, &sensorsMeasurement{name: inputName, tags: tags, fields: fields, ts: time.Now()})
	}

	return cache, nil
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Path:     defPath,
			Interval: defInterval,
			Timeout:  defTimeout,

			semStop:          cliutils.NewSem(),
			semStopCompleted: cliutils.NewSem(),
		}
	})
}
