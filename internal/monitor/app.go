// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package monitor implements datakit tool monitor
package monitor

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/GuanceCloud/cliutils/metrics"
	"github.com/dustin/go-humanize"
	dto "github.com/prometheus/client_model/go"
	"github.com/rivo/tview"
)

var (
	l = logger.DefaultSLogger("monitor")

	inputsFeedCols   = strings.Split(`Input|Cat|Feeds|P90Lat|P90Pts|Filtered|LastFeed|AvgCost|Errors`, "|")
	plStatsCols      = strings.Split("Script|Cat|Namespace|TotalPts|DropPts|ErrPts|PLUpdate|AvgCost", "|")
	walStatsCols     = strings.Split("Cat|Points(mem/disk/drop/total)", "|")
	enabledInputCols = strings.Split(`Input|Count|Crashed`, "|")
	goroutineCols    = strings.Split(`Name|Running|Done|TotalCost`, "|")
	httpAPIStatCols  = strings.Split(`API|Status|Total|Latency|BodySize(P90/Total)`, "|")
	filterRuleCols   = strings.Split("Cat|Total|Filtered(%)|Cost", "|")
	dwptsStatCols    = strings.Split(`Cat|Points(ok/total)|Bytes(ok/total/gz)`, "|")
	dwCols           = strings.Split(`API|Status|Count|Latency|Retry`, "|")

	moduleGoroutine = []string{"G", "goroutine"}
	moduleBasic     = []string{"B", "basic"}
	moduleRuntime   = []string{"R", "runtime"}
	moduleFilter    = []string{"F", "filter"}
	moduleHTTP      = []string{"H", "http"}
	moduleInputs    = []string{"In", "inputs"}
	modulePipeline  = []string{"P", "pipeline"}
	moduleIO        = []string{"IO", "io_stats"}
	moduleDataway   = []string{"W", "dataway"}
	moduleWAL       = []string{"WAL", "wal"}

	labelCategory = "category"
	labelName     = "name"
)

type monitorAPP struct {
	app *tview.Application

	// UI elements
	basicInfoTable        *tview.Table
	golangRuntime         *tview.Table
	inputsStatTable       *tview.Table
	plStatTable           *tview.Table
	walStatTable          *tview.Table
	enabledInputTable     *tview.Table
	goroutineStatTable    *tview.Table
	httpServerStatTable   *tview.Table
	dwTable               *tview.Table
	dwptsTable            *tview.Table
	filterStatsTable      *tview.Table
	filterRulesStatsTable *tview.Table

	exitPrompt     *tview.TextView
	anyErrorPrompt *tview.TextView

	flex *tview.Flex

	mfs map[string]*dto.MetricFamily

	src monitorSource

	inputsStats map[string]string

	anyError error

	start time.Time

	// options
	verbose       bool
	dumpMetrics   bool
	maxTableWidth int
	maxRun        int
	refresh       time.Duration

	url string

	// If now not set, we refresh now each loop.
	// For replay exist metrics, now will be specified by bug report time.
	now          time.Time
	specifiedNow bool

	proxy                   string
	onlyInputs, onlyModules []string
}

func defaultApp() *monitorAPP {
	return &monitorAPP{
		app:     tview.NewApplication(),
		start:   time.Now(),
		refresh: time.Second * 5,
		url:     "localhost:9529",
	}
}

func Start(opts ...APPOption) {
	l = logger.SLogger("monitor")

	app := defaultApp()

	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}

	go app.refreshData()
	app.setup()
	if err := app.run(); err != nil {
		l.Errorf("app.run: %s", err.Error())
	}
}

func (app *monitorAPP) refreshData() {
	go func() {
		tick := time.NewTicker(app.refresh)
		defer tick.Stop()
		var err error

		n := 0

		for {
			app.anyError = nil
			app.mfs, err = app.src.FetchData()
			if err != nil {
				app.anyError = fmt.Errorf("request stats failed: %w", err)
			}

			if app.dumpMetrics && len(app.mfs) > 0 {
				var arr []*dto.MetricFamily
				for _, v := range app.mfs {
					arr = append(arr, v)
				}

				if err := ioutil.WriteFile(".monitor-metrics", []byte(metrics.MetricFamily2Text(arr)), os.ModePerm); err != nil {
					l.Warnf("dumpMetrics: %s, ignored", err.Error())
				} else {
					l.Debug("dump to .monitor-metrics ok")
				}
			}

			app.render()
			app.app = app.app.Draw() // NOTE: cause DATA RACE

			<-tick.C // wait
			n++

			if app.maxRun > 0 && n >= app.maxRun {
				app.app.Stop()
				return
			}
		}
	}()
}

func (app *monitorAPP) run() error {
	return app.app.Run() // NOTE: cause DATA RACE
}

func (app *monitorAPP) inputClicked(input string) func() bool {
	if app.anyError != nil { // some error occurred, we just gone
		return func() bool {
			return true
		}
	}

	return func() bool {
		conf := "没有此采集器的配置"
		if app.inputsStats != nil && len(app.inputsStats[input]) > 0 {
			conf = app.inputsStats[input]
		}
		app.renderInputConfigView(conf)

		return true
	}
}

func number(i interface{}) string {
	switch x := i.(type) {
	case int:
		return humanize.SIWithDigits(float64(x), 3, "")
	case uint:
		return humanize.SIWithDigits(float64(x), 3, "")
	case int64:
		return humanize.SIWithDigits(float64(x), 3, "")
	case uint64:
		return humanize.SIWithDigits(float64(x), 3, "")
	case float32:
		return humanize.SIWithDigits(float64(x), 3, "")
	case float64:
		return humanize.SIWithDigits(x, 3, "")
	default:
		return ""
	}
}

func metricWithLabel(mf *dto.MetricFamily, vals ...string) *dto.Metric {
	if mf == nil {
		return nil
	}

	labelMatch := func(lps []*dto.LabelPair) bool {
		if len(lps) < len(vals) {
			return false
		}

		for i, v := range vals {
			if lps[i].GetValue() != v {
				return false
			}
		}
		return true
	}

	for _, m := range mf.Metric {
		if labelMatch(m.GetLabel()) {
			return m
		}
	}
	return nil
}
