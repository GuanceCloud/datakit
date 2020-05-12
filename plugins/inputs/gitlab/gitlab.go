package gitlab

import (
	"context"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type GitlabTarget struct {
	Interval   int
	Active     bool
	Host       string
	Token      string
	Project    interface{}
	Branch     string
	StartDate  string `toml:"start_date"`
	HoursBatch int `toml:"hours_batch"`
}

type Gitlab struct {
	MetricName string `toml:"metric_name"`
	Targets    []GitlabTarget
}

type GitlabInput struct {
	GitlabTarget
	MetricName string
}

type GitlabOutput struct {
	acc telegraf.Accumulator
}

type GitlabParam struct {
	input  GitlabInput
	output GitlabOutput
}

const (
	gitlabConfigSample = `### metric_name: the name of metric, default is "gitlab"
### You need to configure an [[targets]] for each gitlab to be monitored.
### active     : whether to gather gitlab data.
### host       : gitlab service url.
### project    : project name or project id. If no configuration, get all projects.
### branch     : branch name.  If no configuration, get all branches.
### token      : the token of access gitlab web.
### interval   : batch interval, unit is second. The default value is 60.
### start_date : gather data from this start time
### hours_batch: time range for gather data per batch.

#metric_name="gitlab"
#[[targets]]
#	active      = true
#	host        = "https://gitlab.jiagouyun.com/api/v4"
#	project     = "493"
#	branch      = "dev"
#	token       = "KovnP_TmLX_VTmPcSzYqPx8"
#	interval    = 60
#	start_date  = "2019-01-01T00:00:00"
#	hours_batch = 720

#[[targets]]
#	active      = true
#	host        = "https://gitlab.jiagouyun.com/api/v4"
#	project     = "493"
#	branch      = "dev"
#	token       = "KovnP_TmLX_VTmPcSzYqPx8"
#	interval    = 60
#	start_date  = "2019-01-01T00:00:00"
#	hours_batch = 720
`
	defaultInterval = 60
)

var (
	ctx        context.Context
	cfun       context.CancelFunc
	acc        telegraf.Accumulator
	metricName = "gitlab"
)

func (g *Gitlab) SampleConfig() string {
	return gitlabConfigSample
}

func (g *Gitlab) Description() string {
	return "Collect Gitlab Data"
}

func (g *Gitlab) Gather(telegraf.Accumulator) error {
	return nil
}

func (g *Gitlab) Start(ac telegraf.Accumulator) error {
	log.Printf("I! [gitlab] start")
	ctx, cfun = context.WithCancel(context.Background())

	acc = ac
	if g.MetricName != "" {
		metricName = g.MetricName
	}

	for _, target := range g.Targets {
		if target.Active && target.Host != "" {
			if target.Interval == 0 {
				target.Interval = defaultInterval
			}
			target.Host = strings.Trim(target.Host, " ")
			go target.active()
		}
	}
	return nil
}

func (g *Gitlab) Stop() {
	cfun()
}

func init() {
	inputs.Add("gitlab", func() telegraf.Input {
		git := &Gitlab{}
		return git
	})
}
