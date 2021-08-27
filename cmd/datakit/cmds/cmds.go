package cmds

import (
	"path/filepath"

	"github.com/c-bata/go-prompt"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
)

var (
	suggestions = []prompt.Suggest{
		{Text: "exit", Description: "exit cmd"},
		{Text: "Q", Description: "exit cmd"},
		{Text: "flushall", Description: "k8s interactive command to generate deploy file"},
	}

	l = logger.DefaultSLogger("cmds")
)

type completer struct{}

func SetLog() {
	l = logger.SLogger("cmds")
}

func newCompleter() (*completer, error) {
	return &completer{}, nil
}

func (c *completer) Complete(d prompt.Document) []prompt.Suggest {
	w := d.GetWordBeforeCursor()
	switch w {
	case "":
		return []prompt.Suggest{}
	default:
		return prompt.FilterFuzzy(suggestions, w, true)
	}
}

func ipInfo(ip string) (map[string]string, error) {

	datadir := datakit.DataDir

	if err := geo.LoadIPLib(filepath.Join(datadir, "iploc.bin")); err != nil {
		return nil, err
	}

	if err := ip2isp.Init(filepath.Join(datadir, "ip2isp.txt")); err != nil {
		return nil, err
	}

	x, err := geo.Geo(ip)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"city":     x.City,
		"province": x.Region,
		"country":  x.Country_short,
		"isp":      ip2isp.SearchIsp(ip),
		"ip":       ip,
	}, nil
}

func setCmdRootLog(rl string) {

	if err := logger.InitRoot(&logger.Option{Path: rl, Flags: logger.OPT_DEFAULT, Level: logger.DEBUG}); err != nil {
		l.Error(err)
		return
	}

	// setup config module logger, redirect to @rl
	config.SetLog()

	l = logger.SLogger("cmds")
	l.Infof("root log path set to %s", rl)
}
