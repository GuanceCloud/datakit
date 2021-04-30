package cmds

import (
	"fmt"
	"github.com/c-bata/go-prompt"
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	nhttp "net/http"
	"runtime"
)

var (
	suggestions = []prompt.Suggest{
		{Text: "exit", Description: "exit cmd"},
		{Text: "Q", Description: "exit cmd"},
	}
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

func StopDatakit() {
	svc, err := datakit.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	status, err := svc.Status()
	if err != nil {
		l.Warnf("get datakit service status: %s, ignored", err.Error())
	}
	if status == service.StatusStopped {
		return
	}

	l.Info("stoping datakit...")
	if err := service.Control(svc, "stop"); err != nil {
		l.Warnf("stop service: %s, ignored", err.Error())
	}
	l.Info("stop datakit successful")
}

func StartDatakit() {
	svc, err := datakit.NewService()
	if err != nil {
		l.Errorf("new %s service failed: %s", runtime.GOOS, err.Error())
		return
	}

	status, err := svc.Status()
	if err != nil {
		l.Warnf("get datakit service status: %s, ignored", err.Error())
	}
	if status == service.StatusRunning {
		l.Info("datakit service is already running")
		return
	}

	if err := service.Control(svc, "start"); err != nil {
		l.Warnf("start datakit service: %s, ignored", err.Error())
	}

	l.Info("start datakit successful")
}

func RestartDatakit() {
	StopDatakit()
	StartDatakit()
}

func ReloadDatakit(port int) {
	resp, err := nhttp.Get(fmt.Sprintf("http://127.0.0.1:%d/reload", port))
	if err != nil {
		l.Warn(err)
		return
	}

	if resp.StatusCode == 200 {
		l.Info("datakit reload successful")
	} else {
		l.Warn("datakit reload failed: %d", resp.StatusCode)
	}
}
