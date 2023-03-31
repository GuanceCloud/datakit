// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"time"

	plstats "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
)

type DatakitStats struct {
	// EnabledInputs           map[string]*EnabledInput `json:"enabled_input_list"`

	// IOStats     *io.Stats                  `json:"io_stats"`
	PLStats []plstats.ScriptStatsROnly `json:"pl_stats"`

	// markdown options
	DisableMonofont bool `json:"-"`
}

type APIStat struct {
	TotalCount     int     `json:"total_count"`
	Limited        int     `json:"limited"`
	LimitedPercent float64 `json:"limited_percent"`

	Status2xx int `json:"2xx"`
	Status3xx int `json:"3xx"`
	Status4xx int `json:"4xx"`
	Status5xx int `json:"5xx"`

	MaxLatency time.Duration `json:"max_latency"`
	AvgLatency time.Duration `json:"avg_latency"`
}

func GetStats() (*DatakitStats, error) {
	// now := time.Now()
	// elected, ns, who := election.Elected()
	// if ns == "" {
	//	ns = "<default>"
	//}

	// stats := &DatakitStats{
	//	Version:       datakit.Version,
	//	BuildAt:       git.BuildAt,
	//	Branch:        git.Branch,
	//	Uptime:        fmt.Sprintf("%v", now.Sub(datakit.Uptime)),
	//	OSArch:        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	//	WithinDocker:  datakit.Docker,
	//	Elected:       fmt.Sprintf("%s::%s|%s", ns, elected, who),
	//	AutoUpdate:    datakit.AutoUpdate,
	//	HostName:      datakit.DatakitHostName,
	//	EnabledInputs: map[string]*EnabledInput{},
	//}

	// stats.InputsStats, err = io.GetInputsStats() // get all inputs stats
	// if err != nil {
	//	return nil, err
	//}

	// for k := range inputs.InputsInfo {
	//	n := inputs.InputEnabled(k)
	//	npanic := inputs.GetPanicCnt(k)
	//	if n > 0 {
	//		stats.EnabledInputs[k] = &EnabledInput{Input: k, Instances: n, Panics: npanic}
	//	}
	//}

	// for k := range inputs.Inputs {
	//	stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("[D] %s", k))
	//}

	// add available inputs(datakit) stats

	//	len(stats.AvailableInputs), len(inputs.Inputs)))

	return nil, nil
}
