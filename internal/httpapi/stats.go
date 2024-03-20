// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/metrics"
	plstats "github.com/GuanceCloud/cliutils/pipeline/stats"
	dto "github.com/prometheus/client_model/go"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	dkm "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/resourcelimit"
)

type DatakitStats struct {
	GoroutineStats *goroutineSummary `json:"goroutine_stats"`

	EnabledInputsDeprecated []*enabledInput          `json:"enabled_inputs,omitempty"`
	EnabledInputs           map[string]*enabledInput `json:"enabled_input_list"`

	GolangRuntime *RuntimeInfo `json:"golang_runtime"`

	AvailableInputs []string `json:"available_inputs"`

	HostName      string `json:"hostname"`
	Version       string `json:"version"`
	BuildAt       string `json:"build_at"`
	Branch        string `json:"branch"`
	Uptime        string `json:"uptime"`
	OSArch        string `json:"os_arch"`
	IOChanStat    string `json:"io_chan_stats"`
	Elected       string `json:"elected"`
	ResourceLimit string `json:"resource_limit"`
	CSS           string `json:"-"`

	OpenFiles int `json:"open_files"`

	HTTPMetrics map[string]*APIStat        `json:"http_metrics"`
	InputsStats map[string]inputsStat      `json:"inputs_status"`
	IOStats     *Stats                     `json:"io_stats"`
	PLStats     []plstats.ScriptStatsROnly `json:"pl_stats"`
	FilterStats *FilterStats               `json:"filter_stats"`

	WithinDocker bool `json:"docker"`
	AutoUpdate   bool `json:"auto_update"`

	// markdown options
	DisableMonofont bool `json:"-"`
}

type goroutineSummary struct {
	Total        int64 `json:"finished_goroutines"`
	RunningTotal int64 `json:"running_goroutines"`
	CostTime     int64 `json:"total_cost_time"`
	AvgCostTime  int64 `json:"avg_cost_time"`

	Items map[string]goroutine.RunningStatInfo
}
type ScriptStatsROnly struct {
	Pt, PtDrop, PtError uint64

	RunLastErrs []string

	TotalCost int64 // ns
	MetaTS    time.Time

	Script            string
	FirstTS           time.Time
	ScriptTS          time.Time
	ScriptUpdateTimes uint64

	Category, NS, Name string

	Enable       bool
	Deleted      bool
	CompileError string
}

type Stats struct {
	ChanUsage map[string][2]int `json:"chan_usage"`

	COSendPts uint64 `json:"CO_send_pts"`
	ESendPts  uint64 `json:"E_send_pts"`
	LSendPts  uint64 `json:"L_send_pts"`
	MSendPts  uint64 `json:"M_send_pts"`
	NSendPts  uint64 `json:"N_chan_pts"`
	OSendPts  uint64 `json:"O_send_pts"`
	PSendPts  uint64 `json:"P_chan_pts"`
	RSendPts  uint64 `json:"R_send_pts"`
	SSendPts  uint64 `json:"S_send_pts"`
	TSendPts  uint64 `json:"T_send_pts"`

	COFailPts uint64 `json:"CO_fail_pts"`
	EFailPts  uint64 `json:"E_fail_pts"`
	LFailPts  uint64 `json:"L_fail_pts"`
	MFailPts  uint64 `json:"M_fail_pts"`
	NFailPts  uint64 `json:"N_fail_pts"`
	OFailPts  uint64 `json:"O_fail_pts"`
	PFailPts  uint64 `json:"P_fail_pts"`
	RFailPts  uint64 `json:"R_fail_pts"`
	SFailPts  uint64 `json:"S_fail_pts"`
	TFailPts  uint64 `json:"T_fail_pts"`

	FeedDropPts uint64 `json:"drop_pts"`

	// 0: ok, others: beyond usage
	BeyondUsage uint64 `json:"beyond_usage"`

	// TODO: add disk cache stats
}
type FilterStats struct {
	RuleStats map[string]*ruleStat `json:"rule_stats"`

	PullCount    int           `json:"pull_count"`
	PullInterval time.Duration `json:"pull_interval"`
	PullFailed   int           `json:"pull_failed"`

	RuleSource  string        `json:"rule_source"`
	PullCost    time.Duration `json:"pull_cost"`
	PullCostAvg time.Duration `json:"pull_cost_avg"`
	PullCostMax time.Duration `json:"pull_cost_max"`

	LastUpdate  time.Time `json:"last_update"`
	LastErr     string    `json:"last_err"`
	LastErrTime time.Time `json:"last_err_time"`
}

type ruleStat struct {
	Total        int64         `json:"total"`
	Filtered     int64         `json:"filtered"`
	Cost         time.Duration `json:"cost"`
	CostPerPoint time.Duration `json:"cost_per_point"`
	Conditions   int           `json:"conditions"`
}
type inputsStat struct {
	// Name      string    `json:"name"`
	Category       string        `json:"category"`
	Frequency      string        `json:"frequency,omitempty"`
	AvgSize        int64         `json:"avg_size"`
	FeedTotal      int64         `json:"feed_total"`
	PtsTotal       int64         `json:"pts_total"`
	Count          int64         `json:"count"`
	Filtered       int64         `json:"filtered"`
	Errors         int64         `json:"errors"`
	First          time.Time     `json:"first"`
	Last           time.Time     `json:"last"`
	LastErr        string        `json:"last_error,omitempty"`
	LastErrTS      time.Time     `json:"last_error_ts,omitempty"`
	Version        string        `json:"version,omitempty"`
	MaxCollectCost time.Duration `json:"max_collect_cost"`
	AvgCollectCost time.Duration `json:"avg_collect_cost"`
}

type RuntimeInfo struct {
	Goroutines int     `json:"goroutines"`
	HeapAlloc  uint64  `json:"heap_alloc"`
	Sys        uint64  `json:"total_sys"`
	CPUUsage   float64 `json:"cpu_usage"`

	GCPauseTotal uint64        `json:"gc_pause_total"`
	GCNum        uint32        `json:"gc_num"`
	GCAvgCost    time.Duration `json:"gc_avg_bytes"`
}

type enabledInput struct {
	Input     string `json:"input"`
	Instances int    `json:"instances"`
	Panics    int    `json:"panic"`
}

type APIStat struct {
	TotalCount     int     `json:"total_count"`
	Limited        int     `json:"limited"`
	LimitedPercent float64 `json:"limited_percent"`

	Status2xx int `json:"2xx"`
	Status3xx int `json:"3xx"`
	Status4xx int `json:"4xx"`
	Status5xx int `json:"5xx"`

	MaxLatency   time.Duration `json:"max_latency"`
	AvgLatency   time.Duration `json:"avg_latency"`
	totalLatency time.Duration
}

const categoryKey = "category"

func getRuntimeInfo() *RuntimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var usage float64
	if u, err := resourcelimit.MyCPUPercent(0); err != nil {
		l.Warnf("get CPU usage failed: %s, ignored", err.Error())
	} else {
		usage = u
	}

	return &RuntimeInfo{
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  m.HeapAlloc,
		Sys:        m.Sys,
		CPUUsage:   usage,

		GCPauseTotal: m.PauseTotalNs,
		GCNum:        m.NumGC,
	}
}

func GetStats() (*DatakitStats, error) {
	var err error
	family, err := metrics.Gather()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	elecMetric := election.GetElectionInfo(family)

	stats := &DatakitStats{
		Version:       datakit.Version,
		BuildAt:       git.BuildAt,
		Branch:        git.Branch,
		Uptime:        fmt.Sprintf("%v", now.Sub(dkm.Uptime)),
		OSArch:        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		WithinDocker:  datakit.Docker,
		Elected:       fmt.Sprintf("%s::%s|%s", elecMetric.Status, elecMetric.Namespace, elecMetric.WhoElected),
		AutoUpdate:    datakit.AutoUpdate,
		HostName:      datakit.DatakitHostName,
		ResourceLimit: resourcelimit.Info(),
		OpenFiles:     datakit.OpenFiles(),
		GolangRuntime: getRuntimeInfo(),

		IOStats:         &Stats{ChanUsage: make(map[string][2]int, 5)},
		AvailableInputs: []string{},
		FilterStats:     &FilterStats{},
		GoroutineStats:  &goroutineSummary{Items: make(map[string]goroutine.RunningStatInfo, 10)},
		PLStats:         []plstats.ScriptStatsROnly{},
		HTTPMetrics:     make(map[string]*APIStat),
		InputsStats:     make(map[string]inputsStat),
		EnabledInputs:   make(map[string]*enabledInput),
	}

	prefix := "datakit_"

	for _, mfamily := range family {
		name := mfamily.GetName()
		pts := mfamily.GetMetric()

		if strings.HasPrefix(name, prefix+"io_dataway_") || strings.HasPrefix(name, prefix+"io_chan_") {
			l.Debugf("get io stats...")
			switch name {
			case "datakit_io_dataway_point_total":
				getDatawayPointTotal(name, stats, pts)

			case "datakit_io_dataway_api_latency":
			case "datakit_io_dataway_api_request_total":
			case "datakit_io_dataway_point_bytes_total":
			case "datakit_io_chan_usage":
				getIOChanUsage(name, stats, pts)

			case "datakit_io_chan_capacity":
				getIOChanCap(name, stats, pts)
			}
			continue
		}

		if strings.HasPrefix(name, prefix+"goroutine_") {
			l.Debugf("get stat...")
			getGoroutineStats(name, stats, pts)
			continue
		}

		if strings.HasPrefix(name, prefix+"http_api") {
			l.Debugf("get http stats...")
			getHTTPStats(name, stats, pts)
			continue
		}

		if strings.HasPrefix(name, prefix+"filter_") {
			l.Debugf("get filter stats...")
			getFilterStats(name, stats, pts)
			continue
		}

		if strings.HasPrefix(name, prefix+"io_") || strings.HasPrefix(name, prefix+"input_collect_latency") || strings.HasPrefix(name, "last_err") {
			l.Debugf("get inputs stats...")
			getInputsStats(name, stats, pts)
			continue
		}

		if strings.HasPrefix(name, prefix+"inputs_instance") {
			l.Debugf("get avaliabel inputs...")
			getInputsList(name, stats, pts)
			continue
		}

		if strings.HasPrefix(name, prefix+"election_status") {
			if ei := election.MetricElectionInfo(mfamily); ei != nil {
				if ei.ElectedTime > 0 {
					stats.Elected = fmt.Sprintf("%s::%s|%s(elected: %s)",
						ei.Namespace, ei.Status, ei.WhoElected, ei.ElectedTime.String())
				} else {
					stats.Elected = fmt.Sprintf("%s::%s|%s", ei.Namespace, ei.Status, ei.WhoElected)
				}
			}
			continue
		}
	}

	// stats.PLStats = plstats.ReadStats()

	return stats, nil
}

func getIOChanUsage(name string, stats *DatakitStats, pts []*dto.Metric) {
	if stats.IOStats == nil {
		stats.IOStats = &Stats{ChanUsage: make(map[string][2]int)}
	}

	for _, pt := range pts {
		label := ""
		for _, la := range pt.GetLabel() {
			if la.GetName() == categoryKey {
				label = la.GetValue()
			}
		}
		if len(label) == 0 {
			continue
		}

		if _, ok := chanUsage[label]; !ok {
			stats.IOStats.ChanUsage[label] = [2]int{}
		}

		s := stats.IOStats.ChanUsage[label]
		s[0] = int(pt.GetCounter().GetValue())
		stats.IOStats.ChanUsage[label] = s
	}
}

func getIOChanCap(name string, stats *DatakitStats, pts []*dto.Metric) {
	if stats.IOStats == nil {
		stats.IOStats = &Stats{ChanUsage: make(map[string][2]int)}
	}

	for _, pt := range pts {
		label := ""
		for _, la := range pt.GetLabel() {
			if la.GetName() == categoryKey {
				label = la.GetValue()
			}
		}
		if len(label) == 0 {
			continue
		}

		if _, ok := chanUsage[label]; !ok {
			stats.IOStats.ChanUsage[label] = [2]int{}
		}

		s := stats.IOStats.ChanUsage[label]
		s[1] = int(pt.GetCounter().GetValue())
		stats.IOStats.ChanUsage[label] = s
	}
}

func getInputsList(name string, stats *DatakitStats, pts []*dto.Metric) {
	if stats.AvailableInputs == nil {
		stats.AvailableInputs = []string{}
	}
	if stats.EnabledInputs == nil {
		stats.EnabledInputs = make(map[string]*enabledInput)
	}
	input := map[string]struct{}{}

	for _, pt := range pts {
		inputName := pt.GetLabel()[0].GetValue()
		in, ok := stats.EnabledInputs[inputName]
		if !ok {
			in = &enabledInput{Input: inputName}
		}

		if name == "datakit_inputs_instance" {
			in.Instances = int(pt.GetGauge().GetValue())
		} else if name == "datakit_inputs_crash_total" {
			in.Panics = int(pt.GetCounter().GetValue())
		}

		stats.EnabledInputs[inputName] = in
		input[inputName] = struct{}{}
	}

	if len(input) > 0 {
		for name := range input {
			stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("[D] %s", name))
		}
		sort.Strings(stats.AvailableInputs)
		stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("total %d, datakit %d",
			len(stats.AvailableInputs), len(inputs.Inputs)))
	}
}

func getInputsStats(name string, stats *DatakitStats, pts []*dto.Metric) {
	if stats.InputsStats == nil {
		stats.InputsStats = make(map[string]inputsStat, 0)
	}

	feedCost := map[string]float64{}

	for _, pt := range pts {
		labels := pt.GetLabel()
		if len(labels) == 0 {
			break
		}

		inputName, cat, err := "", "", ""
		for _, la := range labels {
			switch la.GetName() {
			case "name":
				inputName = la.GetValue()
			case "source":
				inputName = la.GetValue()
			case "category":
				cat = la.GetValue()
			case "error":
				err = la.GetValue()
			}
		}

		if len(inputName) == 0 {
			continue
		}

		item, ok := stats.InputsStats[inputName]
		if !ok {
			item = inputsStat{Category: cat}
		}

		switch name {
		case "datakit_io_feed_total":
			item.FeedTotal = int64(pt.GetCounter().GetValue())
		case "datakit_io_feed_point_total":
			item.PtsTotal = int64(pt.GetCounter().GetValue())
		case "datakit_io_input_filter_point_total":
			item.Filtered = int64(pt.GetCounter().GetValue())
		case "datakit_io_last_feed":
			item.Last = time.Unix(int64(pt.GetGauge().GetValue()), 0)
		case "datakit_last_err":
			item.LastErr = err
			item.LastErrTS = time.Unix(int64(pt.GetGauge().GetValue()), 0)
		case "datakit_input_collect_latency":
			cost := pt.GetCounter().GetValue()
			if item.MaxCollectCost > time.Duration(cost) {
				item.MaxCollectCost = time.Duration(cost)
			}
			feedCost[inputName] += cost
		case "datakit_input_collect_latency_seconds":
			item.AvgCollectCost = time.Duration(
				float64(time.Second) * pt.GetSummary().GetSampleSum() /
					float64(pt.GetSummary().GetSampleCount()))
		case "datakit_io_last_feed_timestamp_seconds":
			item.Last = time.Unix(int64(pt.GetGauge().GetValue()), 0)
		case "datakit_error_total":
			item.Errors = int64(pt.GetCounter().GetValue())
		}

		stats.InputsStats[inputName] = item
	}
}

func getFilterStats(name string, stats *DatakitStats, pts []*dto.Metric) {
	for _, pt := range pts {
		switch name {
		case "datakit_filter_point_total":
			stats.FilterStats.PullCount = int(pt.GetCounter().GetValue())

		case "datakit_filter_point_dropped_total":
			stats.FilterStats.PullFailed = int(pt.GetCounter().GetValue())

		case "datakit_filter_avg_cost":
		case "datakit_filter_max_cost":
		case "datakit_filter_last_update":
			stats.FilterStats.LastUpdate = time.Unix(int64(pt.GetGauge().GetValue()), 0)
		}
	}
}

func getHTTPStats(name string, stats *DatakitStats, pts []*dto.Metric) {
	if stats.HTTPMetrics == nil {
		stats.HTTPMetrics = make(map[string]*APIStat)
	}
	for _, pt := range pts {
		labels := pt.GetLabel()
		if len(labels) > 0 {
			api := ""
			status := 0

			for _, label := range labels {
				switch label.GetName() {
				case "api":
					api = label.GetValue()
				case "status":
					status = statusContextToStatusCode[label.GetValue()]
				}
			}

			if _, ok := stats.HTTPMetrics[api]; !ok {
				stats.HTTPMetrics[api] = &APIStat{}
			}

			hm := stats.HTTPMetrics[api]

			switch status / 100 {
			case 2:
				hm.Status2xx++
			case 3:
				hm.Status3xx++
			case 4:
				hm.Status4xx++
			case 5:
				hm.Status5xx++
			}

			switch name {
			case "datakit_http_api_total":
				hm.TotalCount += int(pt.GetCounter().GetValue())
			case "datakit_http_api_elapsed":
				// 单位 ms
				la := time.Duration(pt.GetSummary().GetSampleCount())
				if la > hm.MaxLatency {
					hm.MaxLatency = la
				}
				hm.totalLatency += la
			}
		}
	}
}

func getGoroutineStats(name string, stats *DatakitStats, pts []*dto.Metric) {
	if stats.GoroutineStats == nil {
		stats.GoroutineStats = &goroutineSummary{Items: make(map[string]goroutine.RunningStatInfo, 10)}
	}

	for _, pt := range pts {
		labels := pt.GetLabel()
		if len(labels) > 0 {
			item, ok := stats.GoroutineStats.Items[labels[0].GetValue()]
			if !ok {
				item = goroutine.RunningStatInfo{}
			}

			switch name {
			case "datakit_goroutine_stopped_total":
				item.Total = int64(pt.GetCounter().GetValue())
				stats.GoroutineStats.Total += item.Total

			case "datakit_goroutine_alive":
				item.RunningTotal = int64(pt.GetGauge().GetValue())
				stats.GoroutineStats.RunningTotal += item.RunningTotal

			case "datakit_goroutine_total_cost_time":
				item.CostTime = fmt.Sprintf("%v", (pt.GetCounter().GetValue()))
				stats.GoroutineStats.CostTime += int64(pt.GetCounter().GetValue())

			case "datakit_goroutine_min_cost_time":
				item.MinCostTime = fmt.Sprintf("%v", (pt.GetCounter().GetValue()))

			case "datakit_goroutine_max_cost_time":
				item.MaxCostTime = fmt.Sprintf("%v", (pt.GetCounter().GetValue()))

			case "datakit_goroutine_err_count":
				item.ErrCount = int64(pt.GetCounter().GetValue())

			case "datakit_goroutine_cost":
				item.CostTimeDuration = time.Duration(pt.GetSummary().GetSampleCount())

			case "datakit_goroutine_min_cost_time_ns":
				item.MinCostTimeDuration = time.Duration(pt.GetCounter().GetValue())

			case "datakit_goroutine_max_cost_time_ns":
				item.MaxCostTimeDuration = time.Duration(pt.GetCounter().GetValue())

				// case "datakit_goroutine_cost":
				// item.CostTime = fmt.Sprintf("%v", pt.GetSummary().GetSampleCount())
			}

			stats.GoroutineStats.Items[labels[0].GetValue()] = item
		}
	}

	if stats.GoroutineStats.Total != 0 {
		stats.GoroutineStats.AvgCostTime = stats.GoroutineStats.CostTime / stats.GoroutineStats.Total
	}
}

func getDatawayPointTotal(name string, stats *DatakitStats, pts []*dto.Metric) {
	if stats.IOStats == nil {
		stats.IOStats = &Stats{ChanUsage: make(map[string][2]int, 5)}
	}

	for _, pt := range pts {
		cat, status := "", ""

		labels := pt.GetLabel()
		for _, label := range labels {
			switch label.GetName() {
			case "category":
				cat = label.GetValue()
			case "status":
				status = label.GetValue()
			}
		}

		v := uint64(pt.GetCounter().GetValue())

		switch cat {
		case "/v1/write/custom_object":
			assignUint64Pts(v, &stats.IOStats.COSendPts, &stats.IOStats.COFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/metric":
			assignUint64Pts(v, &stats.IOStats.MSendPts, &stats.IOStats.MFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/network":
			assignUint64Pts(v, &stats.IOStats.NSendPts, &stats.IOStats.NFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/keyevent":
			assignUint64Pts(v, &stats.IOStats.ESendPts, &stats.IOStats.EFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/object":
			assignUint64Pts(v, &stats.IOStats.OSendPts, &stats.IOStats.OFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/logging":
			assignUint64Pts(v, &stats.IOStats.LSendPts, &stats.IOStats.LFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/tracing":
			assignUint64Pts(v, &stats.IOStats.TSendPts, &stats.IOStats.TFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/rum":
			assignUint64Pts(v, &stats.IOStats.RSendPts, &stats.IOStats.RFailPts, &stats.IOStats.FeedDropPts, status)

		case "/v1/write/security":
			assignUint64Pts(v, &stats.IOStats.SSendPts, &stats.IOStats.SFailPts, &stats.IOStats.FeedDropPts, status)

			// case "/v1/upload/profiling":
		case "/v1/write/profiling":
			assignUint64Pts(v, &stats.IOStats.PSendPts, &stats.IOStats.PFailPts, &stats.IOStats.FeedDropPts, status)
		}
	}
}

func assignUint64Pts(count uint64, send, fail, drop *uint64, status string) {
	switch status {
	case "ok":
		*send = count
	case "failed":
		*fail = count
	case "drroped":
		*drop += count
	}
}

var statusContextToStatusCode = map[string]int{
	"Continue":                        100,
	"Switching Protocols":             101,
	"Processing":                      102,
	"Early Hints":                     103,
	"OK":                              200,
	"Created":                         201,
	"Accepted":                        202,
	"Non-Authoritative Information":   203,
	"No Content":                      204,
	"Reset Content":                   205,
	"Partial Content":                 206,
	"Multi-Status":                    207,
	"Already Reported":                208,
	"IM Used":                         226,
	"Multiple Choices":                300,
	"Moved Permanently":               301,
	"Found":                           302,
	"See Other":                       303,
	"Not Modified":                    304,
	"Use Proxy":                       305,
	"Temporary Redirect":              307,
	"Permanent Redirect":              308,
	"Bad Request":                     400,
	"Unauthorized":                    401,
	"Payment Required":                402,
	"Forbidden":                       403,
	"Not Found":                       404,
	"Method Not Allowed":              405,
	"Not Acceptable":                  406,
	"Proxy Authentication Required":   407,
	"Request Timeout":                 408,
	"Conflict":                        409,
	"Gone":                            410,
	"Length Required":                 411,
	"Precondition Failed":             412,
	"Request Entity Too Large":        413,
	"Request URI Too Long":            414,
	"Unsupported Media Type":          415,
	"Requested Range Not Satisfiable": 416,
	"Expectation Failed":              417,
	"I'm a teapot":                    418,
	"Misdirected Request":             421,
	"Unprocessable Entity":            422,
	"Locked":                          423,
	"Failed Dependency":               424,
	"Too Early":                       425,
	"Upgrade Required":                426,
	"Precondition Required":           428,
	"Too Many Requests":               429,
	"Request Header Fields Too Large": 431,
	"Unavailable For Legal Reasons":   451,
	"Internal Server Error":           500,
	"Not Implemented":                 501,
	"Bad Gateway":                     502,
	"Service Unavailable":             503,
	"Gateway Timeout":                 504,
	"HTTP Version Not Supported":      505,
	"Variant Also Negotiates":         506,
	"Insufficient Storage":            507,
	"Loop Detected":                   508,
	"Not Extended":                    510,
	"Network Authentication Required": 511,
}

var chanUsage = map[string]string{
	"metric":        "/v1/write/metric",
	"metrics":       "/v1/write/metrics",
	"network":       "/v1/write/network",
	"event":         "/v1/write/event",
	"object":        "/v1/write/object",
	"custom_object": "/v1/write/custom_object",
	"logging":       "/v1/write/logging",
	"tracing":       "/v1/write/tracing",
	"rum":           "/v1/write/rum",
	"security":      "/v1/write/security",
	"profiling":     "/v1/write/profiling",
}
