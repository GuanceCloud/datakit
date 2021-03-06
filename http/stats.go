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
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/man"
	plstats "gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/stats"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
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
	HTTPMetrics map[string]*apiStat        `json:"http_metrics"`

	WithinDocker bool            `json:"docker"`
	AutoUpdate   bool            `json:"auto_update"`
	FilterStats  *io.FilterStats `json:"filter_stats"`

	// markdown options
	DisableMonofont bool `json:"-"`
}

var (
	part1 = `
## ????????????

- ?????????     ???{{.HostName}}
- ??????       : {{.Version}}
- ????????????   : {{.Uptime}}
- ????????????   : {{.BuildAt}}
- ??????       : {{.Branch}}
- ????????????   : {{.OSArch}}
- ????????????   : {{.WithinDocker}}
- IO ????????????: {{.IOChanStat}}
- ????????????   ???{{.AutoUpdate}}
- ????????????   ???{{.Elected}}
	`

	part2 = `
## ?????????????????????

{{.InputsStatsTable}}
`

	part3 = `
## ?????????????????????

{{.InputsConfTable}}
`

	part4 = `
## Goroutine????????????

{{.GoroutineStatTable}}
`

	verboseMonitorTmpl = `
{{.CSS}}

# DataKit ????????????
` + part1 + part2 + part3 + part4

	monitorTmpl = `
{{.CSS}}

# DataKit ????????????
` + part1 + part2
)

func (x *DatakitStats) InputsConfTable() string {
	const (
		tblHeader = `
| ????????? | ???????????? | ???????????? |
| ----   | :----:   |  :----:  |
`
	)

	rowFmt := "|`%s`|%d|%d|"
	if x.DisableMonofont {
		rowFmt = "|%s|%d|%d|"
	}

	if len(x.EnabledInputs) == 0 {
		return "???????????????????????????"
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
| ????????? | ???????????? | ??????   | ?????? IO ?????? | ????????? | ??????  | ???????????? | ???????????? | ?????????????????? | ?????????????????? | ????????????(??????) |
| ----   | :----:   | :----: | :----:       | :----: | :---: | :----:   | :---:    | :----:       | :---:        | :----:         |
`
	)

	rowFmt := "|`%s`|`%s`|%s|%d|%d|%d|%s|%s|%s|%s|`%s`(%s)|"
	if x.DisableMonofont {
		rowFmt = "|%s|%s|%s|%d|%d|%d|%s|%s|%s|%s|%s(%s)|"
	}

	if len(x.InputsStats) == 0 {
		return "???????????????????????????"
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
- ?????????: %d
- ?????????: %d
- ?????????: %s
- ????????????: %s
`
		tblHeader = `
| ?????? | ????????? | ????????? | ????????? | ???????????? | ???????????? | ???????????? |
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
		Version:        datakit.Version,
		BuildAt:        git.BuildAt,
		Branch:         git.Branch,
		Uptime:         fmt.Sprintf("%v", now.Sub(uptime)),
		OSArch:         fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		WithinDocker:   datakit.Docker,
		IOStats:        io.GetIOStats(),
		PLStats:        plstats.ReadStats(),
		Elected:        fmt.Sprintf("%s::%s|%s", ns, elected, who),
		Cgroup:         cgroup.Info(),
		AutoUpdate:     datakit.AutoUpdate,
		GoroutineStats: goroutine.GetStat(),
		HostName:       datakit.DatakitHostName,
		EnabledInputs:  map[string]*enabledInput{},
		HTTPMetrics:    getMetrics(),
		GolangRuntime:  getRuntimeInfo(),
		FilterStats:    io.GetFilterStats(),
		OpenFiles:      datakit.OpenFiles(),
	}

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
