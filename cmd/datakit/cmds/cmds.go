package cmds

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline"
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
	if ipdbInstance, err := pipeline.InitIPdb(config.Cfg.Pipeline); err != nil {
		return nil, err
	} else {
		x, err := ipdbInstance.Geo(ip)
		if err != nil {
			return nil, err
		}

		return map[string]string{
			"city":     x.City,
			"province": x.Region,
			"country":  x.Country,
			"isp":      ipdbInstance.SearchIsp(ip),
			"ip":       ip,
		}, nil
	}
}

func setCmdRootLog(rl string) {
	lopt := &logger.Option{
		Path:  rl,
		Flags: logger.OPT_DEFAULT,
		Level: logger.DEBUG,
	}

	if rl == "stdout" {
		lopt.Path = ""
		lopt.Flags = logger.OPT_DEFAULT | logger.OPT_STDOUT
	}

	if err := logger.InitRoot(lopt); err != nil {
		l.Error(err)
		return
	}

	// setup config module logger, redirect to @rl
	config.SetLog()

	l = logger.SLogger("cmds")
}

func infof(fmtstr string, args ...interface{}) {
	if FlagJSON { // under json mode, there should no color message(aka, error message)
		return
	}

	color.Set(color.FgGreen)
	output(fmtstr, args...)
	color.Unset()
}

func warnf(fmtstr string, args ...interface{}) {
	if FlagJSON { // under json mode, there should no color message(aka, error message)
		return
	}

	color.Set(color.FgYellow)
	output(fmtstr, args...)
	color.Unset()
}

func errorf(fmtstr string, args ...interface{}) {
	if FlagJSON { // under json mode, there should no color message(aka, error message)
		return
	}

	color.Set(color.FgRed)
	output(fmtstr, args...)
	color.Unset()
}

func output(fmtstr string, args ...interface{}) {
	fmt.Printf(fmtstr, args...)
}
