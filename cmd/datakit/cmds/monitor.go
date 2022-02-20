//nolint:lll
package cmds

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"time"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
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
	MaxTableWidth = 128

	inputsStatsCols  = strings.Split(`Input,Category,Freq,Avg Pts,Total Feed,Total Pts,1st Feed,Last Feed,Avg Cost,Max Cost,Error(date)`, ",")
	enabledInputCols = strings.Split(`Input,Instaces,Crashed`, ",")
	goroutineCols    = strings.Split(`Name,Done,Running,Total Cost,Min Cost,Max Cost,Failed`, ",")
)

func (m *monitorAPP) renderGolangRuntime(ds *dkhttp.DatakitStats) {
	table := m.golangRuntime
	row := 0

	if m.anyError != nil { // some error occurred, we just gone
		return
	}

	table.SetCell(row, 0,
		tview.NewTableCell("Goroutines").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%d", ds.GolangRuntime.Goroutines)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("Memory").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(humanize.IBytes(ds.GolangRuntime.HeapAlloc)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("Stack").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(humanize.IBytes(ds.GolangRuntime.StackInuse)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("GC Pasued").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(time.Duration(ds.GolangRuntime.GCPauseTotal).String()).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0,
		tview.NewTableCell("GC Count").SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1,
		tview.NewTableCell(fmt.Sprintf("%d", ds.GolangRuntime.GCNum)).
			SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignLeft))
}

func (m *monitorAPP) renderBasicInfoTable(ds *dkhttp.DatakitStats) {
	table := m.basicInfoTable
	row := 0

	if m.anyError != nil { // some error occurred, we just gone
		table.SetCell(row, 0, tview.NewTableCell("Error").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(m.anyError.Error()).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft).SetTextColor(tcell.ColorRed))
		return
	}

	table.SetCell(row, 0, tview.NewTableCell("Hostname").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.HostName).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Version").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Version).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Build").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.BuildAt).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Branch").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Branch).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Uptime").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Uptime).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("OS/Arch").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.OSArch).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("IO").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.IOChanStat).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Pipeline").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.PLWorkerStat).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Elected").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(ds.Elected).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("From").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(m.url).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))

	row++
	table.SetCell(row, 0, tview.NewTableCell("Monitor Time").SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
	table.SetCell(row, 1, tview.NewTableCell(time.Since(m.start).String()).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignLeft))
}

func (m *monitorAPP) renderEnabledInputTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.enabledInputTable

	if m.anyError != nil {
		return
	}

	if len(ds.EnabledInputs) == 0 {
		m.enabledInputTable.SetTitle("Enabled Inputs(no inputs enabled)")
		return
	} else {
		m.enabledInputTable.SetTitle(fmt.Sprintf("Enabled Inputs(%d inputs)", len(ds.EnabledInputs)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).SetMaxWidth(*flagMonitorMaxTableWidth).SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	// sort enabled inputs(by name)
	names := []string{}
	for k := range ds.EnabledInputs {
		names = append(names, k)
	}
	sort.Strings(names)

	row := 1

	// sort inputs(by name)
	for _, k := range names {
		ei := ds.EnabledInputs[k]
		table.SetCell(row, 0, tview.NewTableCell(ei.Input).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", ei.Instances)).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", ei.Panics)).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		row++
	}
}

func (m *monitorAPP) renderGoroutineTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.goroutineStatTable

	if m.anyError != nil {
		return
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).SetMaxWidth(*flagMonitorMaxTableWidth).SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1

	// sort goroutines(by name)
	names := []string{}
	for k := range ds.GoroutineStats.Items {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		v := ds.GoroutineStats.Items[name]

		table.SetCell(row, 0, tview.NewTableCell(name).SetMaxWidth(MaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(fmt.Sprintf("%d", v.Total)).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 2, tview.NewTableCell(fmt.Sprintf("%d", v.RunningTotal)).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 3, tview.NewTableCell(v.CostTime).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 4, tview.NewTableCell(v.MinCostTime).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 5, tview.NewTableCell(v.MaxCostTime).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 6, tview.NewTableCell(fmt.Sprintf("%d", v.ErrCount)).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		row++
	}
}

func (m *monitorAPP) renderExitPrompt() {
	fmt.Fprintf(m.exitPrompt, "[yellow]Refreshed at %s, Press ctrl+c to exit monitor", *flagMonitorRefreshInterval)
}

func (m *monitorAPP) renderInputsStatTable(ds *dkhttp.DatakitStats, colArr []string) {
	table := m.inputsStatTable

	if m.anyError != nil {
		return
	}

	if len(ds.InputsStats) == 0 {
		m.inputsStatTable.SetTitle("Inputs Info(no data collected)")
		return
	} else {
		m.inputsStatTable.SetTitle(fmt.Sprintf("Inputs Info(%d inputs)", len(ds.InputsStats)))
	}

	// set table header
	for idx := range colArr {
		table.SetCell(0, idx, tview.NewTableCell(colArr[idx]).SetMaxWidth(*flagMonitorMaxTableWidth).SetTextColor(tcell.ColorGreen).SetAlign(tview.AlignRight))
	}

	row := 1
	now := time.Now()

	isSpecifiedInputs := func(n string) bool {
		for _, x := range *flagMonitorOnlyInputs {
			if x == n {
				return true
			}
		}
		return false
	}

	// sort inputs(by name)
	inputsNames := []string{}
	for k := range ds.InputsStats {
		if len(*flagMonitorOnlyInputs) == 0 || isSpecifiedInputs(k) {
			inputsNames = append(inputsNames, k)
		}
	}
	sort.Strings(inputsNames)

	if len(*flagMonitorOnlyInputs) > 0 {
		m.inputsStatTable.SetTitle(fmt.Sprintf("Inputs Info(total %d, %d selected)",
			len(ds.InputsStats), len(*flagMonitorOnlyInputs)))
	}

	for _, name := range inputsNames {
		v := ds.InputsStats[name]
		table.SetCell(row, 0,
			tview.NewTableCell(name).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 1, tview.NewTableCell(func() string {
			if v, ok := categoryMap[v.Category]; ok {
				return v
			} else {
				return "-"
			}
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignCenter))
		table.SetCell(row, 2, tview.NewTableCell(func() string {
			if v.Frequency == "" {
				return "-"
			}
			return v.Frequency
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 3,
			tview.NewTableCell(fmt.Sprintf("%d", v.AvgSize)).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 4,
			tview.NewTableCell(humanize.SI(float64(v.Count), "")).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 5,
			tview.NewTableCell(humanize.SI(float64(v.Total), "")).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 6, tview.NewTableCell(func() string {
			return humanize.RelTime(v.First, now, "ago", "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 7, tview.NewTableCell(func() string {
			return humanize.RelTime(v.Last, now, "ago", "")
		}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 8,
			tview.NewTableCell(v.AvgCollectCost.String()).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 9,
			tview.NewTableCell(v.MaxCollectCost.String()).
				SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))
		table.SetCell(row, 10,
			tview.NewTableCell(func() string {
				if v.LastErr == "" {
					return "-"
				}
				return fmt.Sprintf("%s(%s)", v.LastErr, humanize.RelTime(v.LastErrTS, now, "ago", ""))
			}()).SetMaxWidth(*flagMonitorMaxTableWidth).SetAlign(tview.AlignRight))

		row++
	}
}

type monitorAPP struct {
	app *tview.Application

	// UI elements
	basicInfoTable     *tview.Table
	golangRuntime      *tview.Table
	inputsStatTable    *tview.Table
	enabledInputTable  *tview.Table
	goroutineStatTable *tview.Table
	exitPrompt         *tview.TextView

	anyError error

	refresh time.Duration
	start   time.Time
	url     string
}

func (m *monitorAPP) setup() {
	// basic info
	m.basicInfoTable = tview.NewTable().SetSelectable(true, false).SetBorders(false)
	m.basicInfoTable.SetBorder(true).SetTitle("Basic Info").SetTitleAlign(tview.AlignLeft)

	m.golangRuntime = tview.NewTable().SetSelectable(true, false).SetBorders(false)
	m.golangRuntime.SetBorder(true).SetTitle("Runtime Info").SetTitleAlign(tview.AlignLeft)

	// inputs running stats
	m.inputsStatTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.inputsStatTable.SetBorder(true).SetTitle("Inputs Info").SetTitleAlign(tview.AlignLeft)

	// enabled inputs
	m.enabledInputTable = tview.NewTable().SetSelectable(true, false).SetBorders(false)
	m.enabledInputTable.SetBorder(true).SetTitle("Enabled Inputs").SetTitleAlign(tview.AlignLeft)

	// goroutine stats
	m.goroutineStatTable = tview.NewTable().SetSelectable(true, false).SetBorders(false).SetSeparator(tview.Borders.Vertical)
	m.goroutineStatTable.SetBorder(true).SetTitle("Goroutine Groups").SetTitleAlign(tview.AlignLeft)

	// buttom prompt
	m.exitPrompt = tview.NewTextView().SetDynamicColors(true)

	flex := tview.NewFlex()

	if *flagMonitorVerbose {
		flex.SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(m.basicInfoTable, 0, 10, false).
				AddItem(m.golangRuntime, 0, 10, false), 0, 10, false).
			AddItem(m.inputsStatTable, 0, 14, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(m.enabledInputTable, 0, 10, false).
				AddItem(m.goroutineStatTable, 0, 10, false), 0, 10, false).
			AddItem(m.exitPrompt, 0, 1, false)
	} else {
		flex.SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
				AddItem(m.basicInfoTable, 0, 10, false).
				AddItem(m.golangRuntime, 0, 10, false), 0, 10, false).
			AddItem(m.inputsStatTable, 0, 14, false).
			AddItem(m.exitPrompt, 0, 1, false)
	}

	go func() {
		tick := time.NewTicker(m.refresh)
		defer tick.Stop()

		for {
			m.anyError = nil

			l.Debugf("try get stats...")

			ds, err := requestStats(m.url)
			if err != nil {
				m.anyError = fmt.Errorf("request stats failed: %w", err)
			}

			m.render(ds)
			m.app = m.app.Draw()

			<-tick.C // wait
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
	m.basicInfoTable.Clear()
	m.golangRuntime.Clear()

	m.inputsStatTable.Clear()

	if *flagMonitorVerbose {
		m.enabledInputTable.Clear()
		m.goroutineStatTable.Clear()
	}
	m.exitPrompt.Clear()

	m.renderBasicInfoTable(ds)
	m.renderGolangRuntime(ds)
	m.renderInputsStatTable(ds, inputsStatsCols)
	if *flagMonitorVerbose {
		m.renderEnabledInputTable(ds, enabledInputCols)
		m.renderGoroutineTable(ds, goroutineCols)
	}
	m.renderExitPrompt()
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

// cmdMonitor deprecated
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
