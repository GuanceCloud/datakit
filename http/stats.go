// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/git"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cgroup"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/goroutine"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/election"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/filter"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	plstats "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	StatInfoType   = "info"
	StatMetricType = "metric"
)

type enabledInput struct {
	Input     string `json:"input"`
	Instances int    `json:"instances"`
	Panics    int    `json:"panic"`
}

type runtimeInfo struct {
	Goroutines int     `json:"goroutines"`
	HeapAlloc  uint64  `json:"heap_alloc"`
	Sys        uint64  `json:"total_sys"`
	CPUUsage   float64 `json:"cpu_usage"`

	GCPauseTotal uint64        `json:"gc_pause_total"`
	GCNum        uint32        `json:"gc_num"`
	GCAvgCost    time.Duration `json:"gc_avg_bytes"`
}

func getRuntimeInfo() *runtimeInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var usage float64
	if u, err := cgroup.GetCPUPercent(0); err != nil {
		l.Warnf("get CPU usage failed: %s, ignored", err.Error())
	} else {
		usage = u
	}

	return &runtimeInfo{
		Goroutines: runtime.NumGoroutine(),
		HeapAlloc:  m.HeapAlloc,
		Sys:        m.Sys,
		CPUUsage:   usage,

		GCPauseTotal: m.PauseTotalNs,
		GCNum:        m.NumGC,
	}
}

type DatakitStats struct {
	GoroutineStats *goroutine.Summary `json:"goroutine_stats"`

	EnabledInputsDeprecated []*enabledInput          `json:"enabled_inputs,omitempty"`
	EnabledInputs           map[string]*enabledInput `json:"enabled_input_list"`

	GolangRuntime *runtimeInfo `json:"golang_runtime"`

	AvailableInputs []string `json:"available_inputs"`

	HostName   string `json:"hostname"`
	Version    string `json:"version"`
	BuildAt    string `json:"build_at"`
	Branch     string `json:"branch"`
	Uptime     string `json:"uptime"`
	OSArch     string `json:"os_arch"`
	IOChanStat string `json:"io_chan_stats"`
	Elected    string `json:"elected"`
	Cgroup     string `json:"cgroup"`
	CSS        string `json:"-"`

	OpenFiles int `json:"open_files"`

	InputsStats map[string]*io.InputsStat  `json:"inputs_status"`
	IOStats     *io.Stats                  `json:"io_stats"`
	PLStats     []plstats.ScriptStatsROnly `json:"pl_stats"`
	HTTPMetrics map[string]*APIStat        `json:"http_metrics"`

	WithinDocker bool                `json:"docker"`
	AutoUpdate   bool                `json:"auto_update"`
	FilterStats  *filter.FilterStats `json:"filter_stats"`

	// markdown options
	DisableMonofont bool `json:"-"`
}

var (
	part1 = `
## 基本信息

- 主机名     ：{{.HostName}}
- 版本       : {{.Version}}
- 运行时间   : {{.Uptime}}
- 发布日期   : {{.BuildAt}}
- 分支       : {{.Branch}}
- 系统类型   : {{.OSArch}}
- 容器运行   : {{.WithinDocker}}
- IO 消耗统计: {{.IOChanStat}}
- 自动更新   ：{{.AutoUpdate}}
- 选举状态   ：{{.Elected}}
	`

	part2 = `
## 采集器运行情况

{{.InputsStatsTable}}
`

	part3 = `
## 采集器配置情况

{{.InputsConfTable}}
`

	part4 = `
## Goroutine运行情况

{{.GoroutineStatTable}}
`

	verboseMonitorTmpl = `
{{.CSS}}

# DataKit 运行展示
` + part1 + part2 + part3 + part4

	monitorTmpl = `
{{.CSS}}

# DataKit 运行展示
` + part1 + part2
)

func (x *DatakitStats) InputsConfTable() string {
	const (
		tblHeader = `
| 采集器 | 实例个数 | 奔溃次数 |
| ----   | :----:   |  :----:  |
`
	)

	rowFmt := "|`%s`|%d|%d|"
	if x.DisableMonofont {
		rowFmt = "|%s|%d|%d|"
	}

	if len(x.EnabledInputs) == 0 {
		return "没有开启任何采集器"
	}

	rows := []string{}
	for _, v := range x.EnabledInputs {
		rows = append(rows, fmt.Sprintf(rowFmt,
			v.Input,
			v.Instances,
			v.Panics,
		))
	}

	sort.Strings(rows)
	return tblHeader + strings.Join(rows, "\n")
}

func (x *DatakitStats) InputsStatsTable() string {
	const (
		//nolint:lll
		tblHeader = `
| 采集器 | 数据类型 | 频率   | 平均 IO 大小 | 总次数 | 点数  | 首次采集 | 最近采集 | 平均采集消耗 | 最大采集消耗 | 当前错误(时间) |
| ----   | :----:   | :----: | :----:       | :----: | :---: | :----:   | :---:    | :----:       | :---:        | :----:         |
`
	)

	rowFmt := "|`%s`|`%s`|%s|%d|%d|%d|%s|%s|%s|%s|`%s`(%s)|"
	if x.DisableMonofont {
		rowFmt = "|%s|%s|%s|%d|%d|%d|%s|%s|%s|%s|%s(%s)|"
	}

	if len(x.InputsStats) == 0 {
		return "暂无采集器统计数据"
	}

	now := time.Now()

	rows := []string{}

	for k, s := range x.InputsStats {
		firstIO := humanize.RelTime(s.First, now, "ago", "")
		lastIO := humanize.RelTime(s.Last, now, "ago", "")

		lastErr := "-"
		if s.LastErr != "" {
			lastErr = s.LastErr
		}

		lastErrTime := "-"
		if s.LastErr != "" {
			lastErrTime = humanize.RelTime(s.LastErrTS, now, "ago", "")
		}

		freq := "-"
		if s.Frequency != "" {
			freq = s.Frequency
		}

		category := "-"
		if s.Category != "" {
			category = datakit.CategoryMap[s.Category]
		}

		rows = append(rows,
			fmt.Sprintf(rowFmt,
				k,
				category,
				freq,
				s.AvgSize,
				s.Count,
				s.Total,
				firstIO,
				lastIO,
				s.AvgCollectCost,
				s.MaxCollectCost,
				lastErr,
				lastErrTime,
			))
	}

	sort.Strings(rows)
	return tblHeader + strings.Join(rows, "\n")
}

func (x *DatakitStats) GoroutineStatTable() string {
	const (
		summaryFmt = `
- 已完成: %d
- 运行中: %d
- 总消耗: %s
- 平均消耗: %s
`
		tblHeader = `
| 名称 | 已完成 | 运行中 | 总消耗 | 最小消耗 | 最大消耗 | 失败次数 |
| ----   | :----:   |  :----:  |  :----:  |  :----:  |  :----:  |  :----:  |
`
	)

	rowFmt := "|%s|%d|%d|%s|%s|%s|%d|"

	rows := []string{}

	s := x.GoroutineStats

	summary := fmt.Sprintf(summaryFmt, s.Total, s.RunningTotal, s.CostTime, s.AvgCostTime)
	if len(s.Items) == 0 {
		return summary
	}

	for name, item := range s.Items {
		rows = append(rows, fmt.Sprintf(rowFmt,
			name,
			item.Total,
			item.RunningTotal,
			item.CostTime,
			item.MaxCostTime,
			item.MaxCostTime,
			item.ErrCount,
		))
	}

	sort.Strings(rows)
	return summary + "\n" + tblHeader + strings.Join(rows, "\n")
}

func GetStats() (*DatakitStats, error) {
	now := time.Now()
	elected, ns, who := election.Elected()
	if ns == "" {
		ns = "<default>"
	}

	stats := &DatakitStats{
		Version:       datakit.Version,
		BuildAt:       git.BuildAt,
		Branch:        git.Branch,
		Uptime:        fmt.Sprintf("%v", now.Sub(uptime)),
		OSArch:        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		WithinDocker:  datakit.Docker,
		Elected:       fmt.Sprintf("%s::%s|%s", ns, elected, who),
		AutoUpdate:    datakit.AutoUpdate,
		HostName:      datakit.DatakitHostName,
		EnabledInputs: map[string]*enabledInput{},
	}

	l.Debugf("io.GetStats()...")
	stats.IOStats = io.GetIOStats()

	l.Debugf("plstats.ReadStats()...")
	stats.PLStats = plstats.ReadStats()

	l.Debugf("cgroup.Info()...")
	stats.Cgroup = cgroup.Info()

	l.Debugf("goroutine.GetStat()...")
	stats.GoroutineStats = goroutine.GetStat()

	l.Debugf("http.getMetrics()...")
	stats.HTTPMetrics = GetMetrics()

	l.Debugf("getRuntimeInfo()...")
	stats.GolangRuntime = getRuntimeInfo()

	l.Debugf("io.GetFilterStats()...")
	stats.FilterStats = filter.GetFilterStats()

	l.Debugf("OpenFiles()...")
	stats.OpenFiles = datakit.OpenFiles()

	var err error

	l.Debugf("io.GetStats()...")
	stats.InputsStats, err = io.GetInputsStats() // get all inputs stats
	if err != nil {
		return nil, err
	}

	l.Debugf("get inputs info...")
	for k := range inputs.InputsInfo {
		n := inputs.InputEnabled(k)
		npanic := inputs.GetPanicCnt(k)
		if n > 0 {
			stats.EnabledInputs[k] = &enabledInput{Input: k, Instances: n, Panics: npanic}
		}
	}

	for k := range inputs.Inputs {
		if !datakit.Enabled(k) {
			continue
		}
		stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("[D] %s", k))
	}

	// add available inputs(datakit) stats
	stats.AvailableInputs = append(stats.AvailableInputs, fmt.Sprintf("tatal %d, datakit %d",
		len(stats.AvailableInputs), len(inputs.Inputs)))

	sort.Strings(stats.AvailableInputs)
	return stats, nil
}

func (x *DatakitStats) Markdown(css string, verbose bool) ([]byte, error) {
	tmpl := monitorTmpl
	if verbose {
		tmpl = verboseMonitorTmpl
	}

	temp, err := template.New("").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse markdown template failed: %w", err)
	}

	if css != "" {
		x.CSS = css
	}

	var buf bytes.Buffer
	if err := temp.Execute(&buf, x); err != nil {
		return nil, fmt.Errorf("execute markdown template failed: %w", err)
	}

	return buf.Bytes(), nil
}

func apiGetDatakitMonitor(c *gin.Context) {
	s, err := GetStats()
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
		return
	}

	mdbytes, err := s.Markdown(man.MarkdownCSS, true)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html", []byte(err.Error()))
		return
	}

	mdext := parser.CommonExtensions
	psr := parser.NewWithExtensions(mdext)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.CompletePage
	opts := html.RendererOptions{Flags: htmlFlags}
	// opts := html.RendererOptions{Flags: htmlFlags, Head: headerScript}
	renderer := html.NewRenderer(opts)

	out := markdown.ToHTML(mdbytes, psr, renderer)

	c.Data(http.StatusOK, "text/html; charset=UTF-8", out)
}

func apiGetDatakitStats(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	s, err := GetStats()
	if err != nil {
		return nil, err
	}

	body, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return nil, err
	}

	return body, nil
}

type ResponseJSON struct {
	Code      int         `json:"code"`
	Content   interface{} `json:"content"`
	ErrorCode string      `json:"errorCode"`
	ErrorMsg  string      `json:"errorMsg"`
	Success   bool        `json:"success"`
}

// StatInfo contains datakit stat info which not changes over time.
type StatInfo struct {
	EnabledInputs   map[string]*enabledInput  `json:"enabled_input_list"`
	AvailableInputs []string                  `json:"available_inputs"`
	HostName        string                    `json:"hostname"`
	Version         string                    `json:"version"`
	BuildAt         string                    `json:"build_at"`
	Branch          string                    `json:"branch"`
	OSArch          string                    `json:"os_arch"`
	WithinDocker    bool                      `json:"docker"`
	AutoUpdate      bool                      `json:"auto_update"`
	Cgroup          string                    `json:"cgroup"`
	ConfigInfo      map[string]*inputs.Config `json:"config_info"`
}

// StatMetric contains datakit stat metric which changes over time.
type StatMetric struct {
	GoroutineStats *goroutine.Summary         `json:"goroutine_stats"`
	GolangRuntime  *runtimeInfo               `json:"golang_runtime"`
	Elected        string                     `json:"elected"`
	OpenFiles      int                        `json:"open_files"`
	InputsStats    map[string]*io.InputsStat  `json:"inputs_status"`
	IOStats        *io.Stats                  `json:"io_stats"`
	PLStats        []plstats.ScriptStatsROnly `json:"pl_stats"`
	HTTPMetrics    map[string]*APIStat        `json:"http_metrics"`
	FilterStats    *filter.FilterStats        `json:"filter_stats"`
}

// getStatInfo return stat info.
func getStatInfo() *StatInfo {
	infoStat := &StatInfo{}
	s, err := GetStats()
	if err != nil {
		l.Warnf("get stats error: %s", err.Error())
	} else {
		infoStat.EnabledInputs = s.EnabledInputs
		infoStat.AvailableInputs = s.AvailableInputs
		infoStat.HostName = s.HostName
		infoStat.Version = s.Version
		infoStat.BuildAt = s.BuildAt
		infoStat.Branch = s.Branch
		infoStat.OSArch = s.OSArch
		infoStat.WithinDocker = s.WithinDocker
		infoStat.AutoUpdate = s.AutoUpdate
		infoStat.Cgroup = s.Cgroup
		infoStat.ConfigInfo = inputs.ConfigInfo
	}

	return infoStat
}

// getStatMetric return stat metric.
func getStatMetric() *StatMetric {
	metricStat := &StatMetric{}
	s, err := GetStats()
	if err != nil {
		l.Warnf("get stats error: %s", err.Error())
	} else {
		metricStat.GoroutineStats = s.GoroutineStats
		metricStat.GolangRuntime = s.GolangRuntime
		metricStat.Elected = s.Elected
		metricStat.OpenFiles = s.OpenFiles
		metricStat.InputsStats = s.InputsStats
		metricStat.IOStats = s.IOStats
		metricStat.PLStats = s.PLStats
		metricStat.HTTPMetrics = s.HTTPMetrics
		metricStat.FilterStats = s.FilterStats
	}

	return metricStat
}

func apiGetDatakitStatsByType(w http.ResponseWriter, r *http.Request, x ...interface{}) (interface{}, error) {
	var stat interface{}

	statType := r.URL.Query().Get("type")

	switch statType {
	case StatMetricType:
		stat = getStatMetric()
	case StatInfoType:
		stat = getStatInfo()
	default:
		stat = getStatInfo()
	}

	if stat == nil {
		stat = ResponseJSON{
			Code:      400,
			ErrorCode: "param.invalid",
			ErrorMsg:  fmt.Sprintf("invalid type, which should be '%s' or '%s'", StatInfoType, StatMetricType),
			Success:   false,
		}
	}

	body, err := json.MarshalIndent(stat, "", "    ")
	if err != nil {
		return nil, err
	}

	return body, nil
}
