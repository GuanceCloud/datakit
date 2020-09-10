package ansible

import (
	"io/ioutil"
	"net/http"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	httpd "gitlab.jiagouyun.com/cloudcare-tools/datakit/http"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	inputName = "ansible"

	ConfigSample = `
[[inputs.ansible]]
   paht = "/ansible"

`
)

type Ansible struct {
	Path string `toml:"path"`
}

func (a *Ansible) Catalog() string {
	return inputName
}

func (a *Ansible) SampleConfig() string {
	return ConfigSample
}
var l = logger.DefaultSLogger(inputName)


func (a *Ansible) Run() {
	l = logger.SLogger(inputName)
	l.Infof("%s input started...", inputName)

}

func (a *Ansible) RegHttpHandler() {
	httpd.RegHttpHandler("POST", a.Path, Handle)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	dataType := r.URL.Query().Get("type")
	body, err := ioutil.ReadAll(r.Body)
	l.Infof("ansible body {}", string(body))
	defer r.Body.Close()

	if err != nil {
		l.Errorf("failed of http parsen body in ansible err:%s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	switch dataType {
	case "keyevent":
		if err := io.NamedFeed(body, io.KeyEvent, "ansible"); err != nil {
			l.Errorf("failed to io Feed, err: %s", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)

	case "metric":
		if err := io.NamedFeed(body, io.Metric, "ansible"); err != nil {
			l.Errorf("failed to io Feed, err: %s", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func init() {
	inputs.Add(inputName, func() inputs.Input {

		return &Ansible{}
	})
}
