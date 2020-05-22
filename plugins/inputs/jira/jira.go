package jira

import (
	"context"
	"log"
	"strings"

	"github.com/influxdata/telegraf"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type JiraTarget struct {
	Interval int
	Active   bool
	Host     string
	Username string
	Password string
	Project  string
	Issue    string
}

type Jira struct {
	MetricName string `toml:"metric_name"`
	Targets    []JiraTarget
}

type JiraInput struct {
	JiraTarget
	MetricName string
}

type JiraOutput struct {
	acc telegraf.Accumulator
}

type JiraParam struct {
	input  JiraInput
	output JiraOutput
}

const (
	jiraConfigSample = `### metric_name: the name of metric, default is "jira"
### You need to configure an [[targets]] for each jira to be monitored.
### active     : whether to gather jira data.
### host       : jira service url.
### project    : project id. If no configuration, get all projects.
### issue      : issue id.  If no configuration, get all issues.
### username   : the username to access jira.
### password   : the password to access jira.
### interval   : batch interval, unit is second. The default value is 60.

#metric_name="jira"
#[[targets]]
#	active      = true
#	host        = "https://jira.jiagouyun.com/"
#	project     = "11902"
#	issue       = "52922"
#	username    = "user"
#	password    = "password"
#	interval    = 60


#[[targets]]
#	active      = true
#	host        = "https://jira.jiagouyun.com/"
#	project     = "11902"
#	issue       = "52922"
#	username    = "user"
#	password    = "password"
#	interval    = 60
`
	defaultInterval  = 60
	maxIssuesPerQueue = 1000
)

var (
	ctx        context.Context
	cfun       context.CancelFunc
	acc        telegraf.Accumulator
	metricName = "jira"
)

func (g *Jira) Catalog() string {
	return "jira"
}

func (j *Jira) SampleConfig() string {
	return jiraConfigSample
}

func (j *Jira) Description() string {
	return "Collect Jira Data"
}

func (j *Jira) Gather(telegraf.Accumulator) error {
	return nil
}

func (j *Jira) Start(ac telegraf.Accumulator) error {
	log.Printf("I! [jira] start")
	ctx, cfun = context.WithCancel(context.Background())

	acc = ac
	if j.MetricName != "" {
		metricName = j.MetricName
	}

	for _, target := range j.Targets {
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

func (j *Jira) Stop() {
	cfun()
}

func init() {
	inputs.Add("jira", func() inputs.Input {
		jira := &Jira{}
		return jira
	})
}