package cmds

import (
	"fmt"
	nhttp "net/http"
	"path/filepath"

	"github.com/c-bata/go-prompt"
	"github.com/kardianos/service"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	dkservice "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/service"
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

func ReloadDatakit(host string) error {
	// FIXME: 如果没有绑定在 localhost 怎么办? 此处需解析 datakit 所用的 conf
	client := &nhttp.Client{
		CheckRedirect: func(req *nhttp.Request, via []*nhttp.Request) error {
			return nhttp.ErrUseLastResponse
		},
	}
	_, err := client.Get(fmt.Sprintf("http://%s/reload", host))
	if err == nhttp.ErrUseLastResponse {
		return nil
	}

	return err
}

func UninstallDatakit() error {
	svc, err := dkservice.NewService()
	if err != nil {
		return err
	}

	l.Info("uninstall datakit...")
	return service.Control(svc, "uninstall")
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

func SetCmdRootLog(rl string) {
	if err := logger.SetGlobalRootLogger(rl, logger.DEBUG, logger.OPT_DEFAULT); err != nil {
		l.Error(err)
		return
	}

	config.SetLog()

	l = logger.SLogger("cmds")
	l.Infof("root log path set to %s", rl)
}
