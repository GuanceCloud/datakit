package cmds

import (
	"fmt"
	nhttp "net/http"
	"path/filepath"

	"github.com/c-bata/go-prompt"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/service"
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
	svc, err := dkservice.NewService()
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
	svc, err := dkservice.NewService()
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
	client := &nhttp.Client{
		CheckRedirect: func(req *nhttp.Request, via []*nhttp.Request) error {
			return nhttp.ErrUseLastResponse
		},
	}
	_, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/reload", port))
	if err == nhttp.ErrUseLastResponse {
		return nil
	}

	return err
}

func DatakitStatus() (string, error) {

	svc, err := dkservice.NewService()
	if err != nil {
		return "", err
	}

	status, err := svc.Status()
	if err != nil {
		return "", err
	}
	switch status {
	case service.StatusUnknown:
		return "unknown", nil
	case service.StatusRunning:
		return "running", nil
	case service.StatusStopped:
		return "stopped", nil
	default:
		return "", fmt.Errorf("should not been here")
	}
}

func IPInfo(ip string) (map[string]string, error) {

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
