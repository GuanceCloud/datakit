package jira

import (
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IoFeed func(data []byte, category, name string) error

type Jira struct {
	Interval    interface{}
	Active      bool
	Host        string
	Username    string
	Password    string
	Project     string
	Issue       string
	MetricsName string
	Tags        map[string]string
}

type JiraInput struct {
	Jira
}

type JiraOutput struct {
	IoFeed
}

type JiraParam struct {
	input  JiraInput
	output JiraOutput
	log    *logger.Logger
}

const (
	jiraConfigSample = `### You need to configure an [[inputs.jira]] for each jira to be monitored.
### active     : whether to gather jira data.
### host       : jira service url.
### project    : project id. If no configuration, get all projects.
### issue      : issue id.  If no configuration, get all issues.
### username   : the username to access jira.
### password   : the password to access jira.
### interval   : monitor interval, the default value is "60s".
### metricsName: the name of metric, default is "jira"

#[[inputs.jira]]
#	active      = true
#	host        = "https://jira.jiagouyun.com/"
#	project     = "11902"
#	issue       = "52922"
#	username    = "user"
#	password    = "password"
#	interval    = "60s"
#	metricsName = "jira"
#	[inputs.jira.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"

#[[inputs.jira]]
#	active      = true
#	host        = "https://jira.jiagouyun.com/"
#	project     = "11902"
#	issue       = "52922"
#	username    = "user"
#	password    = "password"
#	interval    = "60s"
#	metricsName = "jira"
#	[inputs.jira.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	inputName         = "jira"
	defaultInterval   = "60s"
	defaultMetricName = inputName
	maxIssuesPerQueue = 1000
)

func (g *Jira) Catalog() string {
	return "jira"
}

func (j *Jira) SampleConfig() string {
	return jiraConfigSample
}

func (j *Jira) Run() {
	if !j.Active || j.Host == "" {
		return
	}
	if j.MetricsName == "" {
		j.MetricsName = defaultMetricName
	}
	if j.Interval == nil {
		j.Interval = defaultInterval
	}
	j.Host = strings.Trim(j.Host, " ")

	input := JiraInput{*j}
	output := JiraOutput{io.NamedFeed}
	log := logger.SLogger("jira")

	p := JiraParam{input, output, log}
	p.log.Info("jira input started...")
	p.active()
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		jira := &Jira{}
		return jira
	})
}
