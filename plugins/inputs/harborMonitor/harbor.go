package harborMonitor

import (
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

var (
	l         *logger.Logger
	inputName = "harborMonitor"
)

func (_ *HarborMonitor) SampleConfig() string {
	return harborConfigSample
}

func (_ *HarborMonitor) Catalog() string {
	return "harbor"
}

func (_ *HarborMonitor) Description() string {
	return "harbor monitor"
}

func (_ *HarborMonitor) Gather() error {
	return nil
}

func (h *HarborMonitor) Run() {
	l = logger.SLogger("harborMonitor")

	l.Info("harborMonitor input started...")

	if h.MetricName == "" {
		h.MetricName = "harborMonitor"
	}

	interval, err := time.ParseDuration(h.Interval)
	if err != nil {
		l.Error(err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			// handle
			h.command()
		case <-datakit.Exit.Wait():
			l.Info("exit")
			return
		}
	}
}

func (h *HarborMonitor) command() {
	baseUrl := fmt.Sprintf("http://%s:%s@%s", h.Username, h.Password, h.Domain)

	if h.Https {
		baseUrl = fmt.Sprintf("https://%s:%s@%s", h.Username, h.Password, h.Domain)
	}

	resp1 := h.getVolumes(baseUrl)
	resp2 := h.getStatistics(baseUrl)
	resp3 := h.getHealth(baseUrl)

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["url"] = h.Domain
	tags["product"] = "harbor"

	fields["total"] = gjson.Get(resp1, "storage.total").Int()
	fields["free"] = gjson.Get(resp1, "storage.free").Int()
	fields["total_project_count"] = gjson.Get(resp2, "total_project_count").Int()
	fields["public_project_count"] = gjson.Get(resp2, "public_project_count").Int()
	fields["private_project_count"] = gjson.Get(resp2, "private_project_count").Int()
	fields["public_repo_count"] = gjson.Get(resp2, "public_repo_count").Int()
	fields["total_repo_count"] = gjson.Get(resp2, "total_repo_count").Int()
	fields["private_repo_count"] = gjson.Get(resp2, "private_repo_count").Int()

	for _, item := range gjson.Parse(resp3).Get("components").Array() {
		idx := ""
		for key, val := range item.Map() {
			if key == "name" {
				idx = val.String()
			} else {
				fields[idx] = val.String()
			}
		}
	}

	pt, err := io.MakeMetric(h.MetricName, tags, fields, time.Now())
	if err != nil {
		l.Errorf("make metric point error %v", err)
	}

	h.resData = pt

	err = io.NamedFeed([]byte(pt), datakit.Metric, inputName)
	if err != nil {
		l.Errorf("push metric point error %v", err)
	}
}

func (r *HarborMonitor) getVolumes(baseUrl string) string {
	path := fmt.Sprintf("%s/api/systeminfo/volumes", baseUrl)
	_, resp := Get(path)

	return resp
}

func (r *HarborMonitor) getStatistics(baseUrl string) string {
	path := fmt.Sprintf("%s/api/statistics", baseUrl)
	_, resp := Get(path)

	return resp
}

func (r *HarborMonitor) getHealth(baseUrl string) string {
	path := fmt.Sprintf("%s/api/health", baseUrl)

	_, resp := Get(path)

	return resp
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &HarborMonitor{}
	})
}
