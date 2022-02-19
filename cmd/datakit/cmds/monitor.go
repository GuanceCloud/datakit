package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	//"github.com/gdamore/tcell/v2"
	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
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
	ErrMsgTrimLenght = 8
	MaxTableWidth    = 16
	inputsStatsCols  = strings.Split(`Input,Category,Freq,Avg Pts,Total Feed,Total Pts,1st Feed,Last Feed,Avg Cost,Max Cost,Error(date)`, ",")
)

func (m *monitorAPP) renderBasicInfoForm(ds *dkhttp.DatakitStats) {
	m.basicInfoForm.AddInputField("Hostname", ds.HostName, 32, nil, nil).
		AddInputField("Version", ds.Version, 32, nil, nil).
		AddInputField("Build", ds.BuildAt, 32, nil, nil).
		AddInputField("Branch", ds.Branch, 32, nil, nil).
		AddInputField("Uptime", ds.Uptime, 32, nil, nil).
		AddInputField("OS/Arch", ds.OSArch, 32, nil, nil).
		AddInputField("IO Chan", ds.IOChanStat, 32, nil, nil).
		AddInputField("PL Chan", ds.PLWorkerStat, 32, nil, nil).
		AddInputField("Elected", ds.Elected, 32, nil, nil).
		AddInputField("From", m.url, 32, nil, nil).
		AddInputField("Monitor Time", fmt.Sprintf("%s", time.Since(m.start)), 32, nil, nil)
}

func (m *monitorAPP) renderInputsStatTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.inputsStatTable

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx,
			tview.NewTableCell(colArr[idx]).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetTextColor(tcell.ColorGreen).
				SetAlign(tview.AlignCenter))
	}

	row := 1
	now := time.Now()

	// sort inputs(by name)
	inputsNames := []string{}
	for k := range ds.InputsStats {
		inputsNames = append(inputsNames, k)
	}
	sort.Strings(inputsNames)

	for _, name := range inputsNames {
		v := ds.InputsStats[name]

		table.SetCell(row, 0,
			tview.NewTableCell(name).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignRight))

		table.SetCell(row, 1,
			tview.NewTableCell(func() string {
				if v, ok := categoryMap[v.Category]; ok {
					return v
				} else {
					return "-"
				}
			}()).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 2,
			tview.NewTableCell(func() string {
				if v.Frequency == "" {
					return "-"
				}
				return v.Frequency
			}()).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 3,
			tview.NewTableCell(fmt.Sprintf("%d", v.AvgSize)).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 4,
			tview.NewTableCell(fmt.Sprintf("%s", humanize.Bytes(uint64(v.Count)))).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 5,
			tview.NewTableCell(fmt.Sprintf("%s", humanize.Bytes(uint64(v.Total)))).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 6,
			tview.NewTableCell(func() string {
				return humanize.RelTime(v.First, now, "ago", "")
			}()).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 7,
			tview.NewTableCell(func() string {
				return humanize.RelTime(v.Last, now, "ago", "")
			}()).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 8,
			tview.NewTableCell(v.AvgCollectCost.String()).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 9,
			tview.NewTableCell(v.MaxCollectCost.String()).
				SetMaxWidth(*flagMonitorMaxTableWidth).
				SetAlign(tview.AlignCenter))

		table.SetCell(row, 10,
			tview.NewTableCell(func() string {
				if v.LastErr == "" {
					return "-"
				}
				return fmt.Sprintf("%s(%s)", cliutils.StringTrim(v.LastErr, ErrMsgTrimLenght), humanize.RelTime(v.LastErrTS, now, "ago", ""))
			}()).
				SetMaxWidth(*flagMonitorMaxTableWidth+ErrMsgTrimLenght).
				SetAlign(tview.AlignCenter))

		row++
	}
}

type monitorAPP struct {
	app            *tview.Application
	basicInfoBox   *tview.Box
	basicInfoForm  *tview.Form
	basicInfoTable *tview.Table

	inputsStatTable *tview.Table
	exitPrompt      *tview.TextView

	refresh time.Duration
	start   time.Time
	url     string
}

func (m *monitorAPP) setup() {
	m.inputsStatTable = tview.NewTable().SetBorders(true)
	m.inputsStatTable.SetBorder(true).SetTitle("Inputs Info").SetTitleAlign(tview.AlignLeft)

	m.basicInfoForm = tview.NewForm()
	m.basicInfoForm.SetBorder(true).SetTitle("Basic Info").SetTitleAlign(tview.AlignLeft)
	m.basicInfoForm.SetFieldBackgroundColor(tcell.ColorDefault)
	m.basicInfoForm.SetItemPadding(0)

	m.exitPrompt = tview.NewTextView().SetDynamicColors(true)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(m.basicInfoForm, 0, 7, false).
		AddItem(m.inputsStatTable, 0, 14, false).
		AddItem(m.exitPrompt, 0, 1, false)

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
			m.app = m.app.Draw()

			select { // wait
			case <-tick.C:
			}
		}
	}()

	if err := m.app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func (m *monitorAPP) run() error {
	return m.app.Run()
}

func (m *monitorAPP) render(ds *dkhttp.DatakitStats) {
	m.basicInfoForm.Clear(true)
	m.inputsStatTable.Clear()
	m.exitPrompt.Clear()

	if ds == nil {
		return
	}

	m.renderBasicInfoForm(ds)
	m.renderInputsStatTable(ds, inputsStatsCols)
	fmt.Fprintf(m.exitPrompt, "[yellow]Refreshed at %s, Press ctrl+c to exit monitor", *flagMonitorRefreshInterval)
}

func runMonitorFlags() error {
	if *flagMonitorRefreshInterval < time.Second {
		*flagMonitorRefreshInterval = time.Second
	}

	to := config.Cfg.HTTPAPI.Listen
	if *flagMonitorTo != "" {
		to = *flagMonitorTo
	}

	m := monitorAPP{
		app:     tview.NewApplication(),
		refresh: *flagMonitorRefreshInterval,
		url:     fmt.Sprintf("http://%s/stats", to),
		start:   time.Now(),
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
