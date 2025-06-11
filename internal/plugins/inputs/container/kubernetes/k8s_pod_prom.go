// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	apicorev1 "k8s.io/api/core/v1"
)

var (
	defaultPrometheusioConnectKeepAlive = time.Second * 20
	defaultPromElection                 = false /*collect self node, not election*/

	promRunnersChan = make(chan []*promRunner)
)

func startPromWorker() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	var runners []*promRunner

	for {
		select {
		case <-datakit.Exit.Wait():
			klog.Info("prom worker exit")
			return
		case rs := <-promRunnersChan:
			runners = mergePromRunners(runners, rs)
			podAnnotationPromVec.WithLabelValues("prom").Observe(float64(len(runners)))

		case <-ticker.C:
			for _, runner := range runners {
				runner.scrapOnce()
			}
		}
	}
}

func mergePromRunners(oldRunners, newRunners []*promRunner) []*promRunner {
	for _, oldRunner := range oldRunners {
		tickReused := false

		for _, newRunner := range newRunners {
			if newRunner.identifier != oldRunner.identifier {
				continue
			}
			newRunner.tick = oldRunner.tick
			newRunner.lastTime = oldRunner.lastTime
			tickReused = true
			break
		}

		if !tickReused {
			oldRunner.tick.Stop()
		}
	}
	return newRunners
}

type promRunner struct {
	identifier string
	conf       *promConfig
	pm         *iprom.Prom
	feeder     dkio.Feeder

	tick     *time.Ticker
	lastTime time.Time

	currentURL   string
	instanceTags map[string]string // map["urlstr"] = "url.Host"
}

func newPromRunnersForPod(pod *apicorev1.Pod, inputConfig string, cfg *Config) []*promRunner {
	cfgStr := completePromConfig(pod, inputConfig)

	runners, err := newPromRunnersWithTomlConfig(cfg.Feeder, cfgStr)
	if err != nil {
		klog.Warnf("failed to new prom runner of pod %s export-config, err: %s", pod.Name, err)
		return nil
	}

	for idx := range runners {
		if runners[idx].conf == nil {
			continue
		}

		if runners[idx].conf.Source == "" {
			runners[idx].conf.Source = pod.Namespace + "/" + pod.Name
		}

		for _, key := range cfg.LabelAsTagsForMetric.Keys {
			v, exist := pod.Labels[key]
			if !exist {
				continue
			}
			if runners[idx].conf.Tags == nil {
				runners[idx].conf.Tags = make(map[string]string)
			}

			newKey := pointutil.ReplaceLabelKey(key)
			if _, exist := runners[idx].conf.Tags[newKey]; !exist {
				runners[idx].conf.Tags[newKey] = v
			}
		}
	}

	return runners
}

func newPromRunnersWithTomlConfig(feeder dkio.Feeder, configStr string) ([]*promRunner, error) {
	cfgs, err := parsePromConfigs(configStr)
	if err != nil {
		return nil, fmt.Errorf("parse config error: %w", err)
	}

	var runners []*promRunner

	for _, c := range cfgs {
		p, err := newPromRunnerWithConfig(feeder, c)
		if err != nil {
			return nil, err
		}

		if p.conf.Interval > 0 {
			p.tick = time.NewTicker(p.conf.Interval)
		} else {
			klog.Warnf("ignore prom scrap due to invalid interval(%v), ignored", p.conf.Interval)
			continue
		}

		runners = append(runners, p)
	}

	return runners, nil
}

func newPromRunnerWithConfig(feeder dkio.Feeder, c *promConfig) (*promRunner, error) {
	var (
		p = &promRunner{
			identifier:   fmt.Sprintf("%s: %v", c.Source, c.URLs),
			conf:         c,
			feeder:       feeder,
			lastTime:     ntp.Now(),
			instanceTags: make(map[string]string),
		}

		start = time.Now()
	)

	hosts, err := parseURLHost(c)
	if err != nil {
		return nil, fmt.Errorf("parse urls error: %w", err)
	}
	p.instanceTags = hosts

	callbackFunc := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		// append instance tag to points
		if instance, ok := p.instanceTags[p.currentURL]; ok {
			for _, pt := range pts {
				pt.AddTag("instance", instance)
			}
		}

		if p.conf.AsLogging != nil && p.conf.AsLogging.Enable {
			for _, pt := range pts {
				err := p.feeder.FeedV2(point.Logging, []*point.Point{pt},
					dkio.WithCollectCost(time.Since(start)),
					dkio.WithElection(defaultPromElection),
					dkio.WithInputName(pt.Name()),
				)
				if err != nil {
					klog.Warnf("failed to feed prom logging: %s, ignored", err)
				}
			}
		} else {
			err := p.feeder.FeedV2(point.Metric, pts,
				dkio.WithCollectCost(time.Since(start)),
				dkio.WithElection(defaultPromElection),
				dkio.WithInputName(p.conf.Source),
			)
			if err != nil {
				klog.Warnf("failed to feed prom metrics: %s, ignored", err)
			}
		}
		return nil
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(klog), // WithLogger must in the first
		iprom.WithSource(c.Source),
		iprom.WithTimeout(c.Timeout),
		iprom.WithKeepAlive(defaultPrometheusioConnectKeepAlive),
		iprom.WithIgnoreReqErr(c.IgnoreReqErr),
		iprom.WithMetricTypes(c.MetricTypes),
		iprom.WithMetricNameFilter(c.MetricNameFilter),
		iprom.WithMetricNameFilterIgnore(c.MetricNameFilterIgnore),
		iprom.WithMeasurementPrefix(c.MeasurementPrefix),
		iprom.WithMeasurementName(c.MeasurementName),
		iprom.WithMeasurements(c.Measurements),
		iprom.WithTLSOpen(c.TLSOpen),
		iprom.WithUDSPath(c.UDSPath),
		iprom.WithCacertFiles([]string{c.CacertFile}),
		iprom.WithCertFile(c.CertFile),
		iprom.WithKeyFile(c.KeyFile),
		iprom.WithInsecureSkipVerify(c.InsecureSkipVerify),
		iprom.WithTagsIgnore(c.TagsIgnore),
		iprom.WithTagsRename(c.TagsRename),
		iprom.WithAsLogging(c.AsLogging),
		iprom.WithIgnoreTagKV(c.IgnoreTagKV),
		iprom.WithHTTPHeaders(c.HTTPHeaders),
		iprom.WithTags(c.Tags),
		iprom.WithDisableInfoTag(c.DisableInfoTag),
		iprom.WithAuth(c.Auth),
		iprom.WithMaxBatchCallback(1, callbackFunc),
	}

	if c.BearerTokenFile != "" {
		token, err := os.ReadFile(c.BearerTokenFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, iprom.WithBearerToken(string(token)))
	}

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create prom: %w", err)
	}

	p.pm = pm
	return p, nil
}

func (p *promRunner) scrapOnce() {
	if p.conf == nil {
		return
	}

	select {
	case tt := <-p.tick.C:
		p.lastTime = inputs.AlignTime(tt, p.lastTime, p.conf.Interval)

		klog.Debugf("running collect from source %s", p.conf.Source)

		for _, u := range p.conf.URLs {
			p.currentURL = u
			// use callback processor, not return pts
			_, err := p.pm.CollectFromHTTPV2(u, iprom.WithTimestamp(p.lastTime.UnixNano()))
			if err != nil {
				klog.Warnf("failed to collect prom: %s", err)
				return
			}
		}

	default: // pass: not on current scrap tick
		return
	}
}

const (
	annotationPromExport = "datakit/prom.instances"
	defaultInterval      = time.Second * 30
)

type promConfig struct {
	Source   string        `toml:"source" json:"source"`
	Interval time.Duration `toml:"interval"`
	Timeout  time.Duration `toml:"timeout"`

	URL  string   `toml:"url" json:"url"` // deprecated
	URLs []string `toml:"urls" json:"urls"`

	IgnoreReqErr           bool         `toml:"ignore_req_err" json:"ignore_req_err"`
	MetricTypes            []string     `toml:"metric_types" json:"metric_types"`
	MetricNameFilter       []string     `toml:"metric_name_filter" json:"metric_name_filter"`
	MetricNameFilterIgnore []string     `toml:"metric_name_filter_ignore" json:"metric_name_filter_ignore"`
	MeasurementPrefix      string       `toml:"measurement_prefix" json:"measurement_prefix"`
	MeasurementName        string       `toml:"measurement_name" json:"measurement_name"`
	Measurements           []iprom.Rule `toml:"measurements" json:"measurements"`

	TLSOpen            bool   `toml:"tls_open" json:"tls_open"`
	UDSPath            string `toml:"uds_path" json:"uds_path"`
	CacertFile         string `toml:"tls_ca"`
	CertFile           string `toml:"tls_cert"`
	KeyFile            string `toml:"tls_key"`
	InsecureSkipVerify bool   `toml:"insecure_skip_verify" json:"insecure_skip_verify"`
	BearerTokenFile    string `toml:"bearer_token_file" json:"bearer_token_file"`

	TagsIgnore  []string            `toml:"tags_ignore" json:"tags_ignore"`
	TagsRename  *iprom.RenameTags   `toml:"tags_rename" json:"tags_rename"`
	AsLogging   *iprom.AsLogging    `toml:"as_logging" json:"as_logging"`
	IgnoreTagKV map[string][]string `toml:"ignore_tag_kv_match" json:"ignore_tag_kv_match"`
	HTTPHeaders map[string]string   `toml:"http_headers" json:"http_headers"`

	Tags           map[string]string
	DisableInfoTag bool `toml:"disable_info_tag" json:"disable_info_tag"`

	Auth map[string]string `toml:"auth" json:"auth"`
}

type wrapPromConfig struct {
	Inputs struct {
		Prom []*promConfig `toml:"prom"`
	} `toml:"inputs"`
}

func parseURLHost(cfg *promConfig) (map[string]string, error) {
	res := make(map[string]string)
	for _, urlstr := range cfg.URLs {
		u, err := url.Parse(urlstr)
		if err != nil {
			return nil, fmt.Errorf("invalid url %s, err: %w", urlstr, err)
		}
		res[urlstr] = u.Host
	}
	return res, nil
}

func parsePromConfigs(str string) ([]*promConfig, error) {
	c := wrapPromConfig{}
	if err := bstoml.Unmarshal([]byte(str), &c); err != nil {
		return nil, fmt.Errorf("unable to parse toml: %w", err)
	}
	for _, cfg := range c.Inputs.Prom {
		if cfg.URL != "" {
			cfg.URLs = append(cfg.URLs, cfg.URL)
		}
		if cfg.Interval <= 0 {
			cfg.Interval = defaultInterval
		}
	}
	return c.Inputs.Prom, nil
}

func completePromConfig(item *apicorev1.Pod, config string) string {
	config = strings.ReplaceAll(config, "$IP", item.Status.PodIP)
	config = strings.ReplaceAll(config, "$NAMESPACE", item.Namespace)
	config = strings.ReplaceAll(config, "$PODNAME", item.Name)
	config = strings.ReplaceAll(config, "$NODENAME", item.Spec.NodeName)

	return config
}
