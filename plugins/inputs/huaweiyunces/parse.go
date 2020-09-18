package huaweiyunces

import (
	"encoding/json"
	"fmt"
	"time"
)

type metricItem struct {
	Namespace  string       `json:"namespace"`
	MetricName string       `json:"metric_name"`
	Dimensions []*Dimension `json:"dimensions"`
}

type batchReq struct {
	Period  string        `json:"period"`
	Filter  string        `json:"filter"`
	From    int64         `json:"from"`
	To      int64         `json:"to"`
	Metrics []*metricItem `json:"metrics"`
}

type hwClient struct {
	ak        string
	sk        string
	endpoint  string
	projectid string
}

type dataPoint struct {
	value float64
	tm    int64
	unit  string
}

type metricResult struct {
	metricName string
	datapoints []*dataPoint
}

func (m *metricResult) String() string {

	fs := ""
	for _, dp := range m.datapoints {
		s := fmt.Sprintf("%s %s %v%s", m.metricName, time.Unix(dp.tm/1000, 0), dp.value, dp.unit)
		fs += s + "\n"
	}
	return fs
}

type batchMetricResultItem struct {
	namespace  string
	metricName string
	unit       string
	datapoints []*dataPoint
}

type batchMetricResult struct {
	results []*batchMetricResultItem
}

func (m *batchMetricResultItem) String() string {

	fs := ""
	for _, dp := range m.datapoints {
		s := fmt.Sprintf("%s(%s) %s %v%s", m.metricName, m.namespace, time.Unix(dp.tm/1000, 0), dp.value, m.unit)
		fs += s + "\n"
	}
	return fs
}

func parseMetricResponse(resp []byte, filter string) *metricResult {
	var resJSON map[string]interface{}
	err := json.Unmarshal(resp, &resJSON)
	if err != nil {
		moduleLogger.Errorf("fail to unmarshal, %s", err)
		return nil
	}
	var result metricResult
	if s, ok := resJSON["metric_name"].(string); ok {
		result.metricName = s
	}
	dps := resJSON["datapoints"]
	if dpArr, ok := dps.([]interface{}); ok {
		for _, dp := range dpArr {
			var datapoint dataPoint
			if dpMap, ok := dp.(map[string]interface{}); ok {
				if v, ok := dpMap[filter].(float64); ok {
					datapoint.value = v
				}
				if t, ok := dpMap["timestamp"].(float64); ok {
					datapoint.tm = int64(t)
				}
				if u, ok := dpMap["unit"].(string); ok {
					datapoint.unit = u
				}
			}
			result.datapoints = append(result.datapoints, &datapoint)
		}

	}

	return &result
}

func parseBatchResponse(resp []byte, filter string) *batchMetricResult {
	var results map[string]interface{}
	err := json.Unmarshal(resp, &results)
	if err != nil {
		moduleLogger.Errorf("fail to unmarshal, %s", err)
		return nil
	}
	metrics := results["metrics"]
	//log.Printf("%v", reflect.TypeOf(metrics))
	var batchResult batchMetricResult
	if metricArr, ok := metrics.([]interface{}); ok {
		for _, item := range metricArr {
			var resItem batchMetricResultItem
			if mitem, ok := item.(map[string]interface{}); ok {
				if s, ok := mitem["namespace"].(string); ok {
					resItem.namespace = s
				}
				if s, ok := mitem["metric_name"].(string); ok {
					resItem.metricName = s
				}
				if s, ok := mitem["unit"].(string); ok {
					resItem.unit = s
				}
				if dpArr, ok := mitem["datapoints"].([]interface{}); ok {
					for _, dp := range dpArr {
						var datapoint dataPoint
						if dpMap, ok := dp.(map[string]interface{}); ok {
							if v, ok := dpMap[filter].(float64); ok {
								datapoint.value = v
							}
							if t, ok := dpMap["timestamp"].(float64); ok {
								datapoint.tm = int64(t)
							}
						}
						resItem.datapoints = append(resItem.datapoints, &datapoint)
					}
				}
			}
			batchResult.results = append(batchResult.results, &resItem)
		}

	}

	return &batchResult
}
