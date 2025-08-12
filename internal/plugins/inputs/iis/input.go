// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build windows && amd64
// +build windows,amd64

package iis

import (
	"context"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils"
	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/point"
	"golang.org/x/sys/windows"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/tailer"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/win_utils/pdh"
)

const (
	minInterval = time.Second * 5
	maxInterval = time.Minute * 10
)

var (
	inputName                           = "iis"
	metricNameWebService                = "iis_web_service"
	metricNameAppPoolWas                = "iis_app_pool_was"
	l                                   = logger.DefaultSLogger("iis")
	_                    inputs.InputV2 = (*Input)(nil)
)

type Input struct {
	Interval datakit.Duration

	Tags map[string]string

	Log  *iisLog `toml:"log"`
	tail *tailer.Tailer

	collectCache []*point.Point

	semStop *cliutils.Sem // start stop signal
	feeder  dkio.Feeder
	Tagger  datakit.GlobalTagger
}

type iisLog struct {
	Files    []string `toml:"files"`
	Pipeline string   `toml:"pipeline"`
}

func (*Input) Catalog() string { return "iis" }

func (*Input) SampleConfig() string { return sampleConfig }

func (*Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&IISAppPoolWas{},
		&IISWebService{},
	}
}

// RunPipeline TODO.
func (ipt *Input) RunPipeline() {
	if ipt.Log == nil || len(ipt.Log.Files) == 0 {
		return
	}

	opts := []tailer.Option{
		tailer.WithSource("iis"),
		tailer.WithService("iis"),
		tailer.WithMaxMultilineLength(int64(float64(config.Cfg.Dataway.MaxRawBodySize) * 0.8)),
		tailer.WithPipeline(ipt.Log.Pipeline),
		tailer.WithGlobalTags(inputs.MergeTags(ipt.Tagger.HostTags(), ipt.Tags, "")),
		tailer.EnableDebugFields(config.Cfg.EnableDebugFields),
	}

	var err error
	if ipt.tail, err = tailer.NewTailer(ipt.Log.Files, opts...); err != nil {
		l.Error(err)
		metrics.FeedLastError(inputName, err.Error())
		return
	}

	g := goroutine.NewGroup(goroutine.Option{Name: "inputs_iis"})
	g.Go(func(ctx context.Context) error {
		ipt.tail.Start()
		return nil
	})
}

func (*Input) PipelineConfig() map[string]string {
	pipelineConfig := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineConfig
}

func (ipt *Input) GetPipeline() []tailer.Option {
	opts := []tailer.Option{
		tailer.WithSource(inputName),
		tailer.WithService(inputName),
	}
	if ipt.Log != nil {
		opts = append(opts, tailer.WithPipeline(ipt.Log.Pipeline))
	}
	return opts
}

func (ipt *Input) AvailableArchs() []string {
	return []string{datakit.OSLabelWindows}
}

func (ipt *Input) Run() {
	l = logger.SLogger(inputName)

	l.Infof("iis input started")

	ipt.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, ipt.Interval.Duration)
	tick := time.NewTicker(ipt.Interval.Duration)
	start := ntp.Now()

	defer tick.Stop()
	for {
		collectStart := time.Now()
		if err := ipt.Collect(start.UnixNano()); err == nil {
			if feedErr := ipt.feeder.Feed(point.Metric, ipt.collectCache,
				dkio.WithCollectCost(time.Since(collectStart)),
				dkio.WithSource(inputName),
			); feedErr != nil {
				l.Error(feedErr)
				metrics.FeedLastError(inputName, feedErr.Error())
			}
		} else {
			l.Error(err)
			metrics.FeedLastError(inputName, err.Error())
		}
		ipt.collectCache = ipt.collectCache[:0]

		select {
		case tt := <-tick.C:
			start = inputs.AlignTime(tt, start, ipt.Interval.Duration)

		case <-datakit.Exit.Wait():
			ipt.exit()
			l.Infof("iis input exit")
			return

		case <-ipt.semStop.Wait():
			ipt.exit()
			l.Infof("iis input return")
			return
		}
	}
}

func (ipt *Input) exit() {
	if ipt.tail != nil {
		ipt.tail.Close()
		l.Infof("iis logging exit")
	}
}

func (ipt *Input) Terminate() {
	if ipt.semStop != nil {
		ipt.semStop.Close()
	}
}

func (ipt *Input) Collect(ptTS int64) error {
	for mName, metricCounterMap := range PerfObjMetricMap {
		for objName := range metricCounterMap {
			// measurement name -> instance name -> metric name -> counter query handle list index
			indexMap := map[string]map[string]map[string]int{mName: {}}

			// counter name is localized and cannot be used
			instanceList, _, ret := pdh.PdhEnumObjectItems(objName)
			if ret != uint32(windows.ERROR_SUCCESS) {
				return fmt.Errorf("failed to enumerate the instance and counter of object %s", objName)
			}

			pathList := make([]string, 0)
			pathListIndex := 0
			// instance
			for i := 0; i < len(instanceList); i++ {
				indexMap[mName][instanceList[i]] = map[string]int{}
				for keyCounter := range metricCounterMap[objName] {
					if metricName, ok := metricCounterMap[objName][keyCounter]; ok {
						// make full counter path
						tmpCounterFullPath := pdh.MakeFullCounterPath(objName, instanceList[i], keyCounter)
						pathList = append(pathList, tmpCounterFullPath)

						indexMap[mName][instanceList[i]][metricName] = pathListIndex
						pathListIndex += 1
					}
				}
			}
			if len(pathList) < 1 {
				return fmt.Errorf("obj %s no valid counter ", objName)
			}
			var handle pdh.PDH_HQUERY
			var counterHandle pdh.PDH_HCOUNTER
			if ret = pdh.PdhOpenQuery(0, 0, &handle); ret != uint32(windows.ERROR_SUCCESS) {
				return fmt.Errorf("object: %s, PdhOpenQuery return: %x", objName, ret)
			}

			counterHandleList := make([]pdh.PDH_HCOUNTER, len(pathList))
			valueList := make([]interface{}, len(pathList))
			for i := range pathList {
				ret = pdh.PdhAddEnglishCounter(handle, pathList[i], 0, &counterHandle)
				counterHandleList[i] = counterHandle
				if ret != uint32(windows.ERROR_SUCCESS) {
					return fmt.Errorf("add query counter %s failed", pathList[i])
				}
			}
			// Call PDH query function,
			// for some counter, it need to call twice
			if ret = pdh.PdhCollectQueryData(handle); ret != uint32(windows.ERROR_SUCCESS) {
				return fmt.Errorf("object: %s, PdhCollectQueryData return: %x", objName, ret)
			}

			// If object name is `Web Service` and only call the func once,
			// will cause func such as PdhGetFormattedCounterValueDouble
			// return PDH_INVALID_DATA
			if ret = pdh.PdhCollectQueryData(handle); ret != uint32(windows.ERROR_SUCCESS) {
				return fmt.Errorf("object: %s, PdhCollectQueryData return: %x", objName, ret)
			}

			// Get value
			var counterValue pdh.PDH_FMT_COUNTERVALUE_DOUBLE
			for i := 0; i < len(counterHandleList); i++ {
				ret = pdh.PdhGetFormattedCounterValueDouble(counterHandleList[i], nil, &counterValue)
				if ret != uint32(windows.ERROR_SUCCESS) {
					return fmt.Errorf("PdhGetFormattedCounterValueDouble return: %x\n\t"+
						"CounterFullPath: %s", ret, pathList[i])
				}
				valueList[i] = counterValue.DoubleValue
			}

			// Close query
			ret = pdh.PdhCloseQuery(handle)
			if ret != uint32(windows.ERROR_SUCCESS) {
				return fmt.Errorf("object: %s, PdhCloseQuery return: %x", objName, ret)
			}

			for instanceName := range indexMap[mName] {
				tags := map[string]string{}
				fields := map[string]interface{}{}
				for k, v := range ipt.Tags {
					tags[k] = v
				}
				for metricName := range indexMap[mName][instanceName] {
					fields[metricName] = valueList[indexMap[mName][instanceName][metricName]]
				}
				switch objName {
				case "Web Service":
					tags["website"] = instanceName
				case "APP_POOL_WAS":
					tags["app_pool"] = instanceName
				default:
					return fmt.Errorf("action not defined, obj name: %s  measurement name: %s", objName, mName)
				}

				opts := point.DefaultMetricOptions()
				opts = append(opts, point.WithTimestamp(ptTS))

				tags = inputs.MergeTags(ipt.Tagger.HostTags(), tags, "")

				pt := point.NewPoint(mName,
					append(point.NewTags(tags), point.NewKVs(fields)...),
					opts...)

				ipt.collectCache = append(ipt.collectCache, pt)
			}
		}
	}
	return nil
}

func init() { // nolint:gochecknoinits
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval: datakit.Duration{Duration: time.Second * 15},

			semStop: cliutils.NewSem(),
			feeder:  dkio.DefaultFeeder(),
			Tagger:  datakit.DefaultGlobalTagger(),
		}
	})
}
