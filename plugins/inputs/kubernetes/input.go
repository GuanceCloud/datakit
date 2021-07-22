package kubernetes

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/influxdata/telegraf/filter"
	"github.com/influxdata/telegraf/plugins/common/tls"
	"k8s.io/client-go/rest"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

const (
	maxInterval = 30 * time.Minute
	minInterval = 15 * time.Second
)

var (
	inputName   = "kubernetes"
	catalogName = "kubernetes"
	l           = logger.DefaultSLogger("kubernetes")
)

type Input struct {
	Service                string `toml:"service"`
	Interval               datakit.Duration
	ObjectInterval         string                          `toml:"object_Interval"`
	ObjectIntervalDuration time.Duration                   `toml:"_"`
	Tags                   map[string]string               `toml:"tags"`
	collectCache           map[string][]inputs.Measurement `toml:"-"`
	collectObjectCache     []inputs.Measurement            `toml:"-"`
	lastErr                error

	KubeConfigPath    string `toml:"kube_config_path"`
	URL               string `toml:"url"`
	BearerToken       string `toml:"bearer_token"`
	BearerTokenString string `toml:"bearer_token_string"`
	Namespace         string `toml:"namespace"`
	Timeout           string `toml:"timeout"`
	TimeoutDuration   time.Duration
	ResourceExclude   []string `toml:"resource_exclude"`
	ResourceInclude   []string `toml:"resource_include"`

	SelectorInclude []string `toml:"selector_include"`
	SelectorExclude []string `toml:"selector_exclude"`
	DiscoveryDir    string   `toml:"discovery_dir"`

	tls.ClientConfig
	start          time.Time
	client         *client
	mu             sync.Mutex
	selectorFilter filter.Filter
	collectors     map[string]func(collector string) error
	chPause        chan bool
	pause          bool
}

func (i *Input) initCfg() error {
	TimeoutDuration, err := time.ParseDuration(i.Timeout)
	if err != nil {
		TimeoutDuration = 5 * time.Second
	}

	var (
		config *rest.Config
	)

	i.TimeoutDuration = TimeoutDuration

	if i.KubeConfigPath != "" {
		config, err = createConfigByKubePath(i.KubeConfigPath)
		if err != nil {
			return err
		}
	} else if i.URL != "" {
		token, err := ioutil.ReadFile(i.BearerToken)
		if err != nil {
			return err
		}

		i.BearerTokenString = string(token)

		config, err = createConfigByToken(i.URL, i.BearerTokenString, i.TLSCA, i.InsecureSkipVerify)
		if err != nil {
			return err
		}
	}

	// create discovery dir
	if i.DiscoveryDir == "" {
		i.DiscoveryDir = "/usr/local/datakit/data/exporter_urls"
	}

	if !datakit.FileExist(i.DiscoveryDir) {
		if err := os.MkdirAll(i.DiscoveryDir, os.ModePerm); err != nil {
			l.Errorf("create %s failed: %s", i.DiscoveryDir, err)
			return err
		}
	}

	i.client, err = newClient(config, 5*time.Second)
	if err != nil {
		return err
	}

	if i.client.Clientset == nil {
		return fmt.Errorf("config error init client fail")
	}

	// 通过 ServerVersion 方法来获取版本号
	_, err = i.client.ServerVersion()
	if err != nil {
		l.Errorf("get k8s version error %v", err)
		return err
	}

	i.globalTag()

	return nil
}

func (i *Input) globalTag() {
	if i.URL != "" {
		i.Tags["url"] = i.URL
	}

	if i.Service != "" {
		i.Tags["service_name"] = i.Service
	}
}

func (i *Input) Collect() error {
	i.collectors = map[string]func(collector string) error{
		"kubernetes": i.collectKubernetes,
	}

	i.collectCache = make(map[string][]inputs.Measurement)

	resourceFilter, err := filter.NewIncludeExcludeFilter(i.ResourceInclude, i.ResourceExclude)
	if err != nil {
		return err
	}

	i.selectorFilter, err = filter.NewIncludeExcludeFilter(i.SelectorInclude, i.SelectorExclude)
	if err != nil {
		return err
	}

	// collect kubernetes
	for collector, f := range i.collectors {
		if resourceFilter.Match(collector) {
			i.collectCache[collector] = make([]inputs.Measurement, 0)
			err := f(collector)
			if err != nil {
				l.Errorf("%s exec %v", collector, err)
				io.FeedLastError(inputName, err.Error())
			} else {
				resData := i.collectCache[collector]
				if len(resData) > 0 {
					if err := inputs.FeedMeasurement(inputName,
						datakit.Metric,
						resData,
						&io.Option{CollectCost: time.Since(i.start)}); err != nil {
						l.Error(err)
					}
				}
			}
		}
	}

	return nil
}

func (i *Input) Run() {
	l = logger.SLogger("kubernetes")
	i.Interval.Duration = config.ProtectedInterval(minInterval, maxInterval, i.Interval.Duration)

	i.ObjectIntervalDuration = 5 * time.Minute
	i.ObjectIntervalDuration, _ = time.ParseDuration(i.ObjectInterval)

	err := i.initCfg()
	if err != nil {
		l.Errorf("init config error %v", err)
		io.FeedLastError(inputName, err.Error())
		return
	}

	tick := time.NewTicker(i.Interval.Duration)
	defer tick.Stop()

	disTick := time.NewTicker(i.Interval.Duration)
	defer disTick.Stop()

	for {
		select {
		case <-tick.C:
			if i.pause {
				l.Debugf("not leader, skipped")
				continue
			}

			l.Debugf("kubernetes metric input gathering...")
			i.start = time.Now()
			i.Collect()

			// clear cache
			i.clear()
		case <-disTick.C:
			if i.pause {
				l.Debugf("not leader, skipped")
				continue
			}

			l.Debugf("exec discovery server...")
			if err := i.collectPodsExporter(); err != nil {
				l.Errorf("%s discovery exec error %v", err)
				io.FeedLastError(inputName, err.Error())
			}
		case <-datakit.Exit.Wait():
			l.Info("kubernetes exit")
			return

		case i.pause = <-i.chPause:
		}
	}
}

func (i *Input) Pause() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chPause <- true:
		return nil
	case <-tick.C:
		return fmt.Errorf("pause %s failed", inputName)
	}
}

func (i *Input) Resume() error {
	tick := time.NewTicker(time.Second * 5)
	select {
	case i.chPause <- false:
		return nil
	case <-tick.C:
		return fmt.Errorf("resume %s failed", inputName)
	}
}

func (i *Input) clear() {
	for cate, _ := range i.collectors {
		i.collectCache[cate] = i.collectCache[cate][:0]
	}
}

func (i *Input) Catalog() string { return catalogName }

func (i *Input) SampleConfig() string {
	if runtime.GOOS == "windows" {
		return configSampleWin
	}

	return configSampleLinux
}

func (i *Input) AvailableArchs() []string {
	return datakit.AllArch
}

func (i *Input) SampleMeasurement() []inputs.Measurement {
	return []inputs.Measurement{
		&kubernetesMetric{},
	}
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Input{
			chPause: make(chan bool, 1),
		}
	})
}
