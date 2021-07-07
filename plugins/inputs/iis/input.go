// +build windows

package iis

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/win_utils/pdh"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"golang.org/x/sys/windows"
)

const (
	minInterval = time.Second * 5
	maxInterval = time.Minute * 10
)

var (
	inputName            = "iis"
	metricNameWebService = "iis_web_service"
	metricNameAppPoolWas = "iis_app_pool_was"
	l                    = logger.DefaultSLogger("iis")
)

type Input struct {
	Interval datakit.Duration

	Log  *inputs.TailerOption `toml:"log"`
	Tags map[string]string

	collectCache []inputs.Measurement
}

func (i *Input) Catalog() string {
	return "iis"
}

func (i *Input) SampleConfig() string {
	return sampleConfig
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&IISAppPoolWas{},
		&IISWebService{},
	}
}

// TODO
func (*Input) RunPipeline() {}

func (i *Input) PipelineConfig() map[string]string {
	pipelineConfig := map[string]string{
		inputName: pipelineCfg,
	}
	return pipelineConfig
}

func (i *Input) AvailableArchs() []string {
	return []string{
		// datakit.OSArchWin386,
		datakit.OSArchWinAmd64,
	}
}

func (i *Input) Run() {
	l = logger.SLogger(inputName)
	l.Infof("iis input started")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)
	if i.Log != nil && len(i.Log.Files) > 0 {
		go i.gatherLog()
	}
	tick := time.NewTicker(i.Interval.Duration)

	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			start := time.Now()
			if err := i.Collect(); err == nil {
				if feedErr := inputs.FeedMeasurement(inputName, datakit.Metric, i.collectCache,
					&io.Option{CollectCost: time.Since(start)}); feedErr != nil {
					l.Error(feedErr)
					io.FeedLastError(inputName, feedErr.Error())
				}
			} else {
				l.Error(err)
				io.FeedLastError(inputName, err.Error())
			}
		case <-datakit.Exit.Wait():
			l.Infof("iis input exit")
			return
		}
	}
}

func (i *Input) Collect() error {
	for mName, metricCounterMap := range PerfObjMetricMap {
		ts := time.Now()
		for objName := range metricCounterMap {
			// measurement name -> instance name -> metric name -> counter query handle list index
			indexMap := map[string]map[string]map[string]int{
				mName: {}}
			// counter name 被本地化，无法使用
			instanceList, _, ret := pdh.PdhEnumObjectItems(objName)
			if ret != uint32(windows.ERROR_SUCCESS) {
				return fmt.Errorf("failed to enumerate the instance and counter of object %s", objName)
			}
			var pathList = make([]string, 0)
			pathListIndex := 0
			// instance
			for i := 0; i < len(instanceList); i++ {
				indexMap[mName][instanceList[i]] = map[string]int{}
				// counter
				for key_counter := range metricCounterMap[objName] {
					if metricName, ok := metricCounterMap[objName][key_counter]; ok {
						// make full counter path
						tmpCounterFullPath := pdh.MakeFullCounterPath(objName, instanceList[i], key_counter)
						// if r := pdh.PdhValidatePath(tmpCounterFullPath); r != uint32(windows.ERROR_SUCCESS) {
						// 	// Check full counter path
						// 	l.Errorf("path %s invalid", tmpCounterFullPath)
						// } else {
						// 	// append to path list; save index
						// 	pathList = append(pathList, tmpCounterFullPath)
						// 	indexMap[mName][instanceList[i]][metricName] = pathListIndex
						// 	pathListIndex += 1
						// }
						pathList = append(pathList, tmpCounterFullPath)

						indexMap[mName][instanceList[i]][metricName] = pathListIndex
						pathListIndex += 1
					}
				}
			}
			if len(pathList) < 1 {
				return fmt.Errorf("obj %s no vaild counter ", objName)
			}
			var handle pdh.PDH_HQUERY
			var counterHandle pdh.PDH_HCOUNTER
			if ret = pdh.PdhOpenQuery(0, 0, &handle); ret != uint32(windows.ERROR_SUCCESS) {
				return fmt.Errorf("object: %s, PdhOpenQuery return: %x", objName, ret)
			}

			var counterHandleList = make([]pdh.PDH_HCOUNTER, len(pathList))
			var valueList = make([]interface{}, len(pathList))
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
				for k, v := range i.Tags {
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
				i.collectCache = append(i.collectCache, &measurement{
					name:   mName,
					tags:   tags,
					fields: fields,
					ts:     ts,
				})
			}
		}
	}
	return nil
}

func (i *Input) gatherLog() {
	inputs.JoinPipelinePath(i.Log, inputName+".p")
	i.Log.Source = inputName
	for k, v := range i.Tags {
		i.Log.Tags[k] = v
	}
	if tailer, err := inputs.NewTailer(i.Log); err != nil {
		return
	} else {
		defer tailer.Close()
		tailer.Run()
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			Interval: datakit.Duration{Duration: time.Second * 15},
		}
	})
}
