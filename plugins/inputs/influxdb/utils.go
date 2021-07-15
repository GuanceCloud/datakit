package influxdb

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type NoMoreDataError struct {
	err string
}

func (e NoMoreDataError) Error() string {
	return e.err
}

type Point struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Values map[string]interface{} `json:"values"`
}

type Memstats struct { // influxdb v2 （若接口/debug/vars未移除）包含该数据
	Alloc         float64      `json:"Alloc"`
	TotalAlloc    float64      `json:"TotalAlloc"`
	Sys           float64      `json:"Sys"`
	Lookups       float64      `json:"Lookups"`
	Mallocs       float64      `json:"Mallocs"`
	Frees         float64      `json:"Frees"`
	HeapAlloc     float64      `json:"HeapAlloc"`
	HeapSys       float64      `json:"HeapSys"`
	HeapIdle      float64      `json:"HeapIdle"`
	HeapInuse     float64      `json:"HeapInuse"`
	HeapReleased  float64      `json:"HeapReleased"`
	HeapObjects   float64      `json:"HeapObjects"`
	StackInuse    float64      `json:"StackInuse"`
	StackSys      float64      `json:"StackSys"`
	MSpanInuse    float64      `json:"MSpanInuse"`
	MSpanSys      float64      `json:"MSpanSys"`
	MCacheInuse   float64      `json:"MCacheInuse"`
	MCacheSys     float64      `json:"MCacheSys"`
	BuckHashSys   float64      `json:"BuckHashSys"`
	GCSys         float64      `json:"GCSys"`
	OtherSys      float64      `json:"OtherSys"`
	NextGC        float64      `json:"NextGC"`
	LastGC        float64      `json:"LastGC"`
	PauseTotalNs  float64      `json:"PauseTotalNs"`
	PauseNs       [256]float64 `json:"PauseNs"`
	NumGC         float64      `json:"NumGC"`
	NumForcedGC   float64      `json:"NumForcedGC"`
	GCCPUFraction float64      `json:"GCCPUFraction"`
}

func DebugVarsDataParse2Point(respBody []byte,
	metricMap map[string]map[string]string,
) (func() (*Point, error), error) {
	var err error
	dataMap := map[string]json.RawMessage{}
	if err = json.Unmarshal(respBody, &dataMap); err != nil {
		return nil, err
	}
	kList := []string{}
	for k := range dataMap {
		kList = append(kList, k)
	}
	index := 0
	return func() (*Point, error) {
		point := Point{}
		var err error
		if index >= len(dataMap) {
			return nil, NoMoreDataError{"no more data"}
		}
		keyStr := kList[index]
		index++
		if keyStr == "memstats" {
			p := Memstats{}
			if err := json.Unmarshal(dataMap[keyStr], &p); err != nil {
				return nil, fmt.Errorf("parse memstats failed")
			} else {
				numGC, _ := strconv.Atoi(fmt.Sprintf("%.0f", p.NumGC))
				point = Point{
					Name: "memstats",
					Tags: map[string]string{},
					Values: map[string]interface{}{
						"Alloc":         p.Alloc,
						"TotalAlloc":    p.TotalAlloc,
						"Sys":           p.Sys,
						"Lookups":       p.Lookups,
						"Mallocs":       p.Mallocs,
						"Frees":         p.Frees,
						"HeapAlloc":     p.HeapAlloc,
						"HeapSys":       p.HeapSys,
						"HeapIdle":      p.HeapIdle,
						"HeapInuse":     p.HeapInuse,
						"HeapReleased":  p.HeapReleased,
						"HeapObjects":   p.HeapObjects,
						"StackInuse":    p.StackInuse,
						"StackSys":      p.StackSys,
						"MSpanInuse":    p.MSpanInuse,
						"MSpanSys":      p.MSpanSys,
						"MCacheInuse":   p.MCacheInuse,
						"MCacheSys":     p.MCacheSys,
						"BuckHashSys":   p.BuckHashSys,
						"GCSys":         p.GCSys,
						"OtherSys":      p.OtherSys,
						"NextGC":        p.NextGC,
						"LastGC":        p.LastGC,
						"PauseTotalNs":  p.PauseTotalNs,
						"PauseNs":       p.PauseNs[(numGC+255)%256],
						"NumGC":         p.NumGC,
						"NumForcedGC":   p.NumForcedGC,
						"GCCPUFraction": p.GCCPUFraction,
					},
				}
			}
		} else if keyStr == "system" || keyStr == "cmdline" {
		} else {
			if err := json.Unmarshal(dataMap[keyStr], &point); err != nil {
				return nil, err //
			}
		}

		// 映射名
		if metricMap != nil {
			values := map[string]interface{}{}
			if nameMap, ok := metricMap[point.Name]; ok {
				for oldMName, newMName := range nameMap {
					if pV, ok := point.Values[oldMName]; ok {
						values[newMName] = pV
					}
				}
			} else { // map 中不含的则返回空
				return nil, nil
			}
			point.Values = values
		}

		if point.Name == "" || len(point.Values) == 0 {
			return nil, nil
		}

		return &point, err
	}, err
}
