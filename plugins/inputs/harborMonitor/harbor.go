package harborMonitor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/selfstat"
	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/models"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type HarborMonitor struct {
	Harbor           []*HarborCfg
	runningInstances []*runningInstance
	ctx              context.Context
	cancelFun        context.CancelFunc
	accumulator      telegraf.Accumulator
	logger           *models.Logger
}

type runningInstance struct {
	cfg        *HarborCfg
	agent      *HarborMonitor
	logger     *models.Logger
	metricName string
}

func (_ *HarborMonitor) SampleConfig() string {
	return baiduIndexConfigSample
}

func (_ *HarborMonitor) Catalog() string {
	return "Harbor"
}

func (_ *HarborMonitor) Description() string {
	return ""
}

func (_ *HarborMonitor) Gather(telegraf.Accumulator) error {
	return nil
}

func (h *HarborMonitor) Start(acc telegraf.Accumulator) error {
	if len(h.Harbor) == 0 {
		log.Printf("W! [HarborMonitor] no configuration found")
		return nil
	}

	h.logger = &models.Logger{
		Errs: selfstat.Register("gather", "errors", nil),
		Name: `HarborMonitor`,
	}

	log.Printf("HarborMonitor cdn start")

	h.accumulator = acc

	for _, instCfg := range h.Harbor {
		r := &runningInstance{
			cfg:    instCfg,
			agent:  h,
			logger: h.logger,
		}

		r.metricName = instCfg.MetricName
		if r.metricName == "" {
			r.metricName = "baiduIndex"
		}

		if r.cfg.Interval.Duration == 0 {
			r.cfg.Interval.Duration = time.Minute * 10
		}

		h.runningInstances = append(h.runningInstances, r)

		go r.run(h.ctx)
	}

	return nil
}

func (h *HarborMonitor) Stop() {
	h.cancelFun()
}

func (r *runningInstance) run(ctx context.Context) error {
	defer func() {
		if e := recover(); e != nil {

		}
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		internal.SleepContext(ctx, r.cfg.Interval.Duration)
	}

	return nil
}

func (r *runningInstance) command() {
	baseUrl := fmt.Sprintf("http://%s:%s@%s", r.cfg.Username, r.cfg.Password, r.cfg.Domain)
	resp1 := r.getVolumes(baseUrl)
	resp2 := r.getStatistics(baseUrl)
	resp3 := r.getHealth(baseUrl)

	tags := map[string]string{}
	fields := map[string]interface{}{}

	tags["url"] = r.cfg.Domain
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
		for key, val := range item.Map() {
			fields[key] = val.String()
		}
	}

	r.agent.accumulator.AddFields(r.metricName, fields, tags)
}

func (r *runningInstance) getVolumes(baseUrl string) string {
	path := fmt.Sprintf("%s/systeminfo/volumes", baseUrl)
	_, resp := Get(path)

	fmt.Println("data ========>", resp)

	return resp
}

func (r *runningInstance) getStatistics(baseUrl string) string {
	path := fmt.Sprintf("%s/statistics", baseUrl)
	_, resp := Get(path)
	fmt.Println("data ========>", resp)

	return resp
}

func (r *runningInstance) getHealth(baseUrl string) string {
	path := fmt.Sprintf("%s/health", baseUrl)
	_, resp := Get(path)
	fmt.Println("data ========>", resp)

	return resp
}

func init() {
	inputs.Add("harborMonitor", func() inputs.Input {
		ac := &HarborMonitor{}
		ac.ctx, ac.cancelFun = context.WithCancel(context.Background())
		return ac
	})
}
