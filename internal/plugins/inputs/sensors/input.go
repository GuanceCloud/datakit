// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sensors

import (
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/command"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/path"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
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

	feeder  dkio.Feeder
	semStop *cliutils.Sem // start stop signal
}

func (*Input) Catalog() string {
	return inputName
}

func (*Input) SampleConfig() string {
	return sampleConfig
}

func (*Input) AvailableArchs() []string {
	return []string{datakit.OSLabelLinux}
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
				dkio.FeedLastError(inputName, err.Error())
				continue
			}
		case <-datakit.Exit.Wait():
			l.Info("sensors input exit")
			return

		case <-ipt.semStop.Wait():
			l.Info("sensors input return")
			return
		}
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) gather() error {
	start := time.Now()
	output, err := command.RunWithTimeout(ipt.Timeout.Duration, false, ipt.Path, "-u")
	if err != nil {
		l.Errorf("Command process failed: %q", output)

		return err
	}

	if cache, err := ipt.parse(string(output)); err != nil {
		return err
	} else {
		return ipt.feeder.FeedV2(point.Metric, cache,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithInputName(inputName),
		)
	}
}

func (ipt *Input) getCustomerTags() map[string]string {
	tags := make(map[string]string)
	for k, v := range ipt.Tags {
		tags[k] = v
	}

	return tags
}

func (ipt *Input) parse(output string) ([]*point.Point, error) {
	var (
		lines  = strings.Split(strings.TrimSpace(output), "\n")
		tags   = ipt.getCustomerTags()
		fields = make(map[string]interface{})
		cache  []*point.Point
	)

	for _, line := range lines {
		if line == "" {
			cache = append(cache,
				point.NewPointV2(inputName,
					append(point.NewTags(tags), point.NewKVs(fields)...),
					point.DefaultMetricOptions()...))

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
				cache = append(cache,
					point.NewPointV2(inputName,
						append(point.NewTags(tags), point.NewKVs(fields)...),
						point.DefaultMetricOptions()...))

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
				l.Errorf("strconv.ParseFloat: %s", err.Error())

				return nil, err
			} else {
				fields[strings.ToLower(strings.TrimSpace(parts[0]))] = value
			}
		default:
			tags[strings.ToLower(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	if len(fields) != 0 {
		cache = append(cache,
			point.NewPointV2(inputName,
				append(point.NewTags(tags), point.NewKVs(fields)...),
				point.DefaultMetricOptions()...))
	}

	return cache, nil
}

func init() { //nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Path:     defPath,
			Interval: defInterval,
			Timeout:  defTimeout,
			feeder:   dkio.DefaultFeeder(),

			semStop: cliutils.NewSem(),
		}
	})
}
