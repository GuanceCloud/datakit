package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	//"github.com/gdamore/tcell/v2"
	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkhttp "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"golang.org/x/term"
)

var (
	categoryMap = map[string]string{
		datakit.MetricDeprecated: "M",
		datakit.Metric:           "M",
		datakit.Network:          "N",
		datakit.KeyEvent:         "E",
		datakit.Object:           "O",
		datakit.Logging:          "L",
		datakit.Tracing:          "T",
		datakit.RUM:              "R",
		datakit.Security:         "S",
	}

	inputsStatsCols = strings.Split(`Input,Category,Freqency,Avg Feed Pts,Total Feed,Total Points,First Feed,Last Feed,Avg Cost,Max Cost,Error(date),`, ",")
)

func renderBasicInfoTable(table *tview.Table, ds *dkhttp.DatakitStats) {
	row := 0
	table.SetCell(row, 0,
		tview.NewTableCell("Hostname").
			SetAlign(tview.AlignCenter))

	table.SetCell(row, 1,
		tview.NewTableCell(ds.HostName).
			SetAlign(tview.AlignCenter))
	row++

	table.SetCell(row, 0,
		tview.NewTableCell("Version").
			SetAlign(tview.AlignCenter))

	table.SetCell(row, 1,
		tview.NewTableCell(ds.Version).
			SetAlign(tview.AlignCenter))
	row++

	table.SetCell(row, 0,
		tview.NewTableCell("Uptime").
			SetAlign(tview.AlignCenter))

	table.SetCell(row, 1,
		tview.NewTableCell(ds.Uptime).
			SetAlign(tview.AlignCenter))
	row++

	// TODO
}

func renderInputsStatTable(table *tview.Table, ds *dkhttp.DatakitStats, colArr []string) {
	// set table header
	for idx := range colArr {
		table.SetCell(0, idx,
			tview.NewTableCell(colArr[idx]).
				SetAlign(tview.AlignCenter))
	}

	row := 1
	now := time.Now()
	for k, v := range ds.InputsStats {
		table.SetCell(row, 0,
			tview.NewTableCell(k).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 1,
			tview.NewTableCell(func() string {
				return categoryMap[v.Category]
			}()).SetAlign(tview.AlignCenter))

		table.SetCell(row, 2,
			tview.NewTableCell(func() string {
				if v.Frequency == "" {
					return "-"
				}
				return v.Frequency
			}()).SetAlign(tview.AlignCenter))

		table.SetCell(row, 3,
			tview.NewTableCell(fmt.Sprintf("%d", v.AvgSize)).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 4,
			tview.NewTableCell(fmt.Sprintf("%d", v.Count)).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 5,
			tview.NewTableCell(fmt.Sprintf("%d", v.Total)).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 6,
			tview.NewTableCell(func() string {
				return humanize.RelTime(v.First, now, "ago", "")
			}()).SetAlign(tview.AlignCenter))

		table.SetCell(row, 7,
			tview.NewTableCell(func() string {
				return humanize.RelTime(v.Last, now, "ago", "")
			}()).SetAlign(tview.AlignCenter))

		table.SetCell(row, 8,
			tview.NewTableCell(v.AvgCollectCost.String()).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 9,
			tview.NewTableCell(v.MaxCollectCost.String()).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 9,
			tview.NewTableCell(func() string {
				if v.LastErr == "" {
					return "-"
				}
				return fmt.Sprintf("%s(%s)", v.LastErr, humanize.RelTime(v.LastErrTS, now, "ago", ""))
			}()).SetAlign(tview.AlignCenter))

		row++
	}
}

type monitorAPP struct {
	app            *tview.Application
	basicInfoBox   *tview.Box
	basicInfoTable *tview.Table

	inputsStatBox   *tview.Box
	inputsStatTable *tview.Table

	refresh time.Duration
	url     string
}

func (m *monitorAPP) setup() {
	m.basicInfoBox = tview.NewTable().SetBorder(true).SetTitle("Basic Info")
	m.basicInfoTable = tview.NewTable().SetBorders(true)
	m.basicInfoTable.SetBorder(true)

	m.inputsStatBox = tview.NewTable().SetBorder(true).SetTitle("Inputs Stats")
	m.inputsStatTable = tview.NewTable().SetBorders(true)
	m.inputsStatTable.SetBorders(true)

	flex := tview.NewFlex().AddItem(m.basicInfoBox, 0, 1, false).
		AddItem(m.basicInfoTable, 0, 2, false).
		AddItem(m.inputsStatBox, 0, 1, false).
		AddItem(m.inputsStatTable, 0, 2, false)

	go func() {
		tick := time.NewTicker(m.refresh)
		defer tick.Stop()

		for {

			l.Debugf("try get stats...")

			ds, err := requestStats(m.url)
			if err != nil {
				return // TODO: handle this error
			}

			m.render(ds)

			select { // wait
			case <-tick.C:
			}
		}
	}()

	pages := tview.NewPages().AddPage("", flex, true, true)
	m.app.SetRoot(pages, true)
}

func (m *monitorAPP) run() error {
	return m.app.Run()
}

func (m *monitorAPP) render(ds *dkhttp.DatakitStats) {
	// inputMonitors(m.app, ds)

	m.basicInfoTable.Clear()
	m.inputsStatTable.Clear()

	renderBasicInfoTable(m.inputsStatTable, ds)
	renderInputsStatTable(m.inputsStatTable, ds, inputsStatsCols)
}

func runMonitorFlags() error {
	if *flagMonitorRefreshInterval < time.Second {
		*flagMonitorRefreshInterval = time.Second
	}

	m := monitorAPP{
		app:     tview.NewApplication(),
		refresh: *flagMonitorRefreshInterval,
		url:     fmt.Sprintf("http://%s/stats", config.Cfg.HTTPAPI.Listen),
	}

	m.setup()

	return m.run()
}

func cmdMonitor(interval time.Duration, verbose bool) {
	addr := fmt.Sprintf("http://%s/stats", config.Cfg.HTTPAPI.Listen)

	if interval < time.Second {
		interval = time.Second
	}

	run := func() {
		fmt.Print("\033[H\033[2J") // clean screen

		x, err := doCMDMonitor(addr, verbose)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(string(x))
			fmt.Printf("(Refresh at %s)Press ctrl+c to exit.\n", interval)
		}
	}

	run() // run before sleep

	tick := time.NewTicker(interval)
	defer tick.Stop()
	for range tick.C {
		run()
	}
}

func requestStats(url string) (*dkhttp.DatakitStats, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s", string(body))
	}

	ds := dkhttp.DatakitStats{
		DisableMonofont: true,
	}
	if err = json.Unmarshal(body, &ds); err != nil {
		return nil, err
	}

	return &ds, nil
}

func doCMDMonitor(url string, verbose bool) ([]byte, error) {
	ds, err := requestStats(url)
	if err != nil {
		return nil, err
	}

	mdtxt, err := ds.Markdown("", verbose)
	if err != nil {
		return nil, err
	}

	width := 100
	if term.IsTerminal(0) {
		if width, _, err = term.GetSize(0); err != nil {
			width = 100
		}
	}

	leftPad := 2
	if err != nil {
		return nil, err
	} else {
		if len(mdtxt) == 0 {
			return nil, fmt.Errorf("no monitor info available")
		} else {
			result := markdown.Render(string(mdtxt), width, leftPad)
			return result, nil
		}
	}
}
