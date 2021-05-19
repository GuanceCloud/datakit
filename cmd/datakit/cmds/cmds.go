package cmds

import (
	"fmt"
	nhttp "net/http"

	"github.com/c-bata/go-prompt"
	"github.com/kardianos/service"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
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

func StopDatakit() error {
	svc, err := datakit.NewService()
	if err != nil {
		return err
	}

	status, err := svc.Status()
	if err != nil {
		return err
	}

	if status == service.StatusStopped {
		return nil
	}

	l.Info("stoping datakit...")
	if err := service.Control(svc, "stop"); err != nil {
		return err
	}
	return nil
}

func StartDatakit() error {
	svc, err := datakit.NewService()
	if err != nil {
		return err
	}

	status, err := svc.Status()
	if err != nil {
		return err
	}

	if status == service.StatusRunning {
		l.Info("datakit service is already running")
		return nil
	}

	if err := service.Control(svc, "install"); err != nil {
		l.Warnf("install service failed: %s, ignored", err)
	}

	if err := service.Control(svc, "start"); err != nil {
		return err
	}

	return nil
}

func RestartDatakit() error {
	if err := StopDatakit(); err != nil {
		return err
	}

	if err := StartDatakit(); err != nil {
		return err
	}

	return nil
}

func ReloadDatakit(port int) error {
	// FIXME: 如果没有绑定在 localhost 怎么办? 此处需解析 datakit 所用的 conf
	resp, err := nhttp.Get(fmt.Sprintf("http://127.0.0.1:%d/reload", port))
	if err != nil {
		return err
	}

	if resp.StatusCode == 200 {
		l.Info("datakit reload successful")
		return nil
	} else {
		return fmt.Errorf("datakit reload failed: %d", resp.StatusCode)
	}
}
