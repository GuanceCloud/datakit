// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package cmds used to define various tools used daily
package cmds

import (
	"github.com/GuanceCloud/cliutils/logger"
	prompt "github.com/c-bata/go-prompt"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/export"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/pipeline/plval"
)

var (
	StaticCDN   = "static.guance.com"
	suggestions = []prompt.Suggest{
		{Text: "exit", Description: "exit cmd"},
		{Text: "Q", Description: "exit cmd"},
		{Text: "flushall", Description: "k8s interactive command to generate deploy file"},
	}

	l = logger.DefaultSLogger("cmds")
	g = datakit.G("cmds")
)

type completer struct{}

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
	if ipdbInstance, err := plval.InitIPdb(datakit.DataDir, config.Cfg.Pipeline); err != nil {
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
		lopt.Flags = logger.OPT_DEFAULT | logger.OPT_STDOUT | logger.OPT_COLOR
	}

	if err := logger.InitRoot(lopt); err != nil {
		panic(err)
	}

	// setup module log, redirect to @rl
	config.SetLog()
	export.SetLog()

	l = logger.SLogger("cmds")

	l.Infof("set root log file to %q ok", rl)
}
