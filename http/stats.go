package http

import (
	"net/http"
)

type enabledInput struct {
	Input     string   `json:"input"`
	Instances int      `json:"instances"`
	Cfgs      []string `json:"configs"`
	Panics    int      `json:"panic"`
}

type datakitStats struct {
	InputsStats     map[string]*io.InputsStat `json:"inputs_status"`
	EnabledInputs   []*enabledInput           `json:"enabled_inputs"`
	AvailableInputs []string                  `json:"available_inputs"`

	Version      string    `json:"version"`
	BuildAt      string    `json:"build_at"`
	Branch       string    `json:"branch"`
	Uptime       string    `json:"uptime"`
	OSArch       string    `json:"os_arch"`
	Reload       time.Time `json:"reload,omitempty"`
	ReloadCnt    int       `json:"reload_cnt"`
	WithinDocker bool      `json:"docker"`
	IOChanStat   string    `json:"io_chan_stats"`
}

func (datakitStats *x) InputsStatsTable() string {
	/*
		"category": "metric",
		"frequency": "20.49/min",
		"avg_size": 1,
		"total": 29,
		"count": 28,
		"first": "2021-05-07T19:41:22.943257+08:00",
		"last": "2021-05-07T19:42:44.252574+08:00",
		"last_error": "mocked error from demo input",
		"last_error_ts": "2021-05-07T19:42:43.923569+08:00",
		"max_collect_cost": 1001554420,
		"avg_collect_cost": 1000723212
	*/

	const tblHeader = `
| 采集器 | 数据类型 | 频率 | 平均 IO 大小 | 总次/点数 | 首/末采集时间 | 当前错误/时间 | 平均/最大采集消耗 |
	`

	if len(x.EnabledInputs) == 0 {
		return "没有开启任何采集器"
	}

	rows := []string{}
	for _, v := range x.EnabledInputs {

		rows = append(rows, fmt.Sprintf("|%s|%s|%s|%d|%d/%d|%s/%s|%s/%s|%s/%s|", v.Input))
	}
}

const (
	monitorTmpl = `
# DataKit 运行展示

## 基本信息

版本         : {{.Version}}
运行时间     : {{.Uptime}}
发布日期     : {{.BuildAt}}
分支         : {{.Branch}}
系统类型     : {{.OSArch}}
是否容器运行 : {{.WithinDocker}}
Reload 情况  : {{.ReloadCnt}} 次/{{.SinceReload}} 以前
IO 消耗统计  : {{.IOChanStat}}

## 采集器运行情况

{{.InputsStatsTable}}
`
)

func getStats() (*datakitStats, error) {
	_ = r

	stats := &datakitStats{
		Version:      git.Version,
		BuildAt:      git.BuildAt,
		Branch:       git.Branch,
		Uptime:       fmt.Sprintf("%v", time.Since(uptime)),
		OSArch:       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		ReloadCnt:    reloadCnt,
		WithinDocker: datakit.Docker,
		IOChanStat:   io.ChanStat(),
	}

	if reloadCnt > 0 {
		stats.Reload = reload
	}

	var err error

	stats.InputsStats, err = io.GetStats(time.Second * 5) // get all inputs stats
	if err != nil {
		return nil, err
	}

	for k := range inputs.Inputs {
		if !datakit.Enabled(k) {
			continue
		}

		n, cfgs := inputs.InputEnabled(k)
		npanic := inputs.GetPanicCnt(k)
		if n > 0 {
			stats.EnabledInputs = append(stats.EnabledInputs, &enabledInput{Input: k, Instances: n, Cfgs: cfgs, Panics: npanic})
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

func apiGetDatakitMonitor(w http.ResponseWriter, r *http.Request) {
}

func apiGetDatakitStats(w http.ResponseWriter, r *http.Request) {

	s, err := getStats()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	body, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
