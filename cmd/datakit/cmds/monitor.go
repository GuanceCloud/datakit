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

var categoryMap = map[string]string{
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

func inputMonitors(app *tview.Application, ds *dkhttp.DatakitStats) (*tview.Table, error) {
	table := tview.NewTable().SetBorders(true)

	colArr := strings.Split(`Input,Category,Freqency,Avg Feed Pts,Total Feed,Total Points,First Feed,Last Feed,Avg Cost,Max Cost,Error(date),`, ",")

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

	return table, nil
}

func runMonitorFlags() error {
	addr := fmt.Sprintf("http://%s/stats", config.Cfg.HTTPAPI.Listen)
	if *flagMonitorRefreshInterval < time.Second {
		*flagMonitorRefreshInterval = time.Second
	}

	app := tview.NewApplication()

	monitor := func() {
		ds, err := requestStats(addr)
		if err != nil {
			return // TODO: handle this error
		}

		// refer to this pg example: https://gist.github.com/rivo/2893c6740a6c651f685b9766d1898084

		table, err := inputMonitors(app, ds)
		//box := tview.NewBox().SetBorder(true).SetBorderAttributes(true).SetTitle("Inputs Running Status")
		//if err != nil {
		//}

		if err := app.SetRoot(table, true).EnableMouse(false).Run(); err != nil {
			panic(err)
		}
	}

	tick := time.NewTicker(*flagMonitorRefreshInterval)
	defer tick.Stop()
	for range tick.C {
		monitor()
	}

	return nil
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
