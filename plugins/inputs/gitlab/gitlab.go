package gitlab

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type IoFeed func(data []byte, category, name string) error

type Gitlab struct {
	Interval    interface{}
	Active      bool
	Host        string
	Token       string
	Project     interface{}
	Branch      string
	StartDate   string
	HoursBatch  int
	MetricsName string
	Tags        map[string]string
	jsFile      string
}

type GitlabInput struct {
	Gitlab
}

type GitlabOutput struct {
	IoFeed
}

type GitlabParam struct {
	input  GitlabInput
	output GitlabOutput
	log    *logger.Logger
}

const (
	inputName          = "gitlab"
	gitlabConfigSample = `### You need to configure an [[inputs.gitlab]] for each gitlab to be monitored.
### active     : whether to gather gitlab data.
### host       : gitlab service url.
### project    : project name or project id. If no configuration, get all projects.
### branch     : branch name.  If no configuration, get all branches.
### token      : the token of access gitlab web.
### interval   : monitor interval, the default value is "60s".
### startDate  : gather data from this start time.
### hoursBatch : time range for gather data per batch.
### metricsName: the name of metric, default is "gitlab"

#[[inputs.gitlab]]
#	active      = true
#	host        = "https://gitlab.jiagouyun.com/api/v4"
#	project     = "493"
#	branch      = "dev"
#	token       = "KovnP_TmLX_VTmPcSzYqPx8"
#	interval    = "60s"
#	startDate   = "2019-01-01T00:00:00"
#	hoursBatch  = 720
#	metricsName = "gitlab"
#	[inputs.gitlab.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"

#[[inputs.gitlab]]
#	active      = true
#	host        = "https://gitlab.jiagouyun.com/api/v4"
#	project     = "493"
#	branch      = "dev"
#	token       = "KovnP_TmLX_VTmPcSzYqPx8"
#	interval    = "60s"
#	startDate   = "2019-01-01T00:00:00"
#	hoursBatch  = 720
#	metricsName = "gitlab"
#	[inputs.gitlab.tags]
#		tag1 = "tag1"
#		tag2 = "tag2"
#		tag3 = "tag3"
`
	defaultInterval   = "60s"
	defaultStartDate  = "2005-12-15T00:00:00"
	defaultMetricName = "gitlab"
	defaultDataDir    = "data"
	defaultGitlabDir  = "gitlab"
)

var locker sync.Mutex

func (g *Gitlab) Catalog() string {
	return "gitlab"
}

func (g *Gitlab) SampleConfig() string {
	return gitlabConfigSample
}

func (g *Gitlab) Run() {
	if !g.Active || g.Host == "" {
		return
	}

	if g.Interval == nil {
		g.Interval = defaultInterval
	}

	if g.StartDate == "" {
		g.StartDate = defaultStartDate
	}

	g.Host = strings.Trim(g.Host, " ")

	s1 := strings.ReplaceAll(g.Host, ":", "")
	s2 := strings.ReplaceAll(s1, "/", "")
	jsonFileName := filepath.Join(datakit.InstallDir, defaultDataDir, defaultGitlabDir, s2)
	g.jsFile = jsonFileName

	input := GitlabInput{*g}
	output := GitlabOutput{io.NamedFeed}

	p := GitlabParam{input, output, logger.SLogger("gitlab")}

	p.log.Info("gitlab input started...")
	p.log.Debugf("%#v", p)
	p.mkGitlabDataDir()
	p.gather()
}

func (g *GitlabParam) mkGitlabDataDir() {
	dataDir := filepath.Join(datakit.InstallDir, defaultDataDir)
	gitlabDir := filepath.Join(dataDir, defaultGitlabDir)

	if !PathExists(dataDir) {
		return
	}
	if PathExists(gitlabDir) {
		return
	}

	locker.Lock()
	defer locker.Unlock()
	if PathExists(gitlabDir) {
		return
	}

	err := os.MkdirAll(gitlabDir, 0666)
	if err != nil {
		g.log.Error("Mkdir gitlab err: %s", err.Error())
	}
}
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		git := &Gitlab{}
		return git
	})
}
