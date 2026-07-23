// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	bstoml "github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/pointutil"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/podutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/ntp"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	apicorev1 "k8s.io/api/core/v1"
)

var (
	defaultPrometheusioConnectKeepAlive = time.Second * 20
	defaultPromElection                 = false /*collect self node, not election*/
)

func stopPromRunners(runners []*promRunner) {
	for _, runner := range runners {
		if runner != nil {
			runner.close()
		}
	}
}

type promRunner struct {
	conf       *promConfig
	pm         *iprom.Prom
	feeder     dkio.Feeder
	promSource string
	mu         sync.Mutex
	closed     bool

	tlsFileRevision [sha256.Size]byte
	reloadTLSFiles  bool

	collectStart,
	lastTime time.Time

	currentURL   string
	instanceTags map[string]string // map["urlstr"] = "url.Host"
}

func clonePromConfigs(configs []*promConfig) []*promConfig {
	cloned := make([]*promConfig, len(configs))
	for idx, config := range configs {
		if config == nil {
			continue
		}
		value := *config
		value.Tags = maps.Clone(config.Tags)
		cloned[idx] = &value
	}
	return cloned
}

func applyPromPodMetadata(pod *apicorev1.Pod, configs []*promConfig, labelKeys []string) {
	name := pod.Name
	if _, ownerName := podutil.PodOwner(pod); ownerName != "" {
		name = ownerName
	}

	for _, config := range configs {
		if config == nil {
			continue
		}
		if config.Source == "" {
			config.Source = pod.Namespace + "/" + name
		}
		for _, key := range labelKeys {
			value, found := pod.Labels[key]
			if !found {
				continue
			}
			if config.Tags == nil {
				config.Tags = make(map[string]string)
			}
			tagKey := pointutil.ReplaceLabelKey(key)
			if _, found := config.Tags[tagKey]; !found {
				config.Tags[tagKey] = value
			}
		}
	}
}

func newPromRunnersForPod(pod *apicorev1.Pod, configs []*promConfig, cfg *Config) ([]*promRunner, error) {
	runners := make([]*promRunner, 0, len(configs))
	for _, key := range cfg.LabelAsTagsForMetric.Keys {
		if _, found := pod.Labels[key]; !found {
			continue
		}
		for _, config := range configs {
			if config != nil && config.Tags == nil {
				config.Tags = make(map[string]string)
			}
		}
		break
	}
	for _, promConfig := range configs {
		runner, err := newPromRunnerWithConfig(cfg.Feeder, promConfig)
		if err != nil {
			stopPromRunners(runners)
			return nil, err
		}
		runners = append(runners, runner)
	}
	applyPromPodMetadata(pod, configs, cfg.LabelAsTagsForMetric.Keys)

	return runners, nil
}

func readPromBearerToken(path string) (string, error) {
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("read bearer token file %q: %w", path, err)
	}
	token := strings.TrimSpace(string(content))
	if token == "" {
		return "", fmt.Errorf("bearer token file %q is empty", path)
	}
	return token, nil
}

func newPromRunnerWithConfig(feeder dkio.Feeder, c *promConfig) (*promRunner, error) {
	p := &promRunner{
		conf:         c,
		feeder:       feeder,
		promSource:   c.Source,
		lastTime:     ntp.Now(),
		instanceTags: make(map[string]string),
	}

	hosts, err := parseURLHost(c)
	if err != nil {
		return nil, fmt.Errorf("parse urls error: %w", err)
	}
	p.instanceTags = hosts

	if c.BearerTokenFile != "" {
		if _, err := readPromBearerToken(c.BearerTokenFile); err != nil {
			return nil, err
		}
	}

	pm, tlsFileRevision, reloadTLSFiles, err := p.buildPromWithStableTLSFiles()
	if err != nil {
		return nil, err
	}
	p.pm = pm
	p.tlsFileRevision = tlsFileRevision
	p.reloadTLSFiles = reloadTLSFiles
	return p, nil
}

func (p *promRunner) buildProm() (*iprom.Prom, error) {
	c := p.conf
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
				err := p.feeder.Feed(point.Logging, []*point.Point{pt},
					dkio.WithCollectCost(time.Since(p.collectStart)),
					dkio.WithElection(defaultPromElection),
					dkio.WithSource(pt.Name()),
				)
				if err != nil {
					klog.Warnf("failed to feed prom logging: %s, ignored", err)
				}
			}
		} else {
			err := p.feeder.Feed(point.Metric, pts,
				dkio.WithCollectCost(time.Since(p.collectStart)),
				dkio.WithElection(defaultPromElection),
				dkio.WithSource(p.conf.Source),
				dkio.WithInput("container"),
			)
			if err != nil {
				klog.Warnf("failed to feed prom metrics: %s, ignored", err)
			}
		}
		return nil
	}

	opts := []iprom.PromOption{
		iprom.WithLogger(klog), // WithLogger must in the first
		iprom.WithSource(p.promSource),
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

	pm, err := iprom.NewProm(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create prom: %w", err)
	}
	return pm, nil
}

func promTLSFileRevision(c *promConfig) ([sha256.Size]byte, bool, error) {
	var zero [sha256.Size]byte
	if c == nil || !c.TLSOpen {
		return zero, false, nil
	}

	paths := [...]string{c.CacertFile, c.CertFile, c.KeyFile}
	material := make([]byte, 0, len(paths)*(sha256.Size+1))
	watched := false
	for _, path := range paths {
		if path == "" {
			material = append(material, 0)
			material = append(material, zero[:]...)
			continue
		}
		watched = true
		content, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return zero, true, fmt.Errorf("read TLS credential file %q: %w", path, err)
		}
		contentHash := sha256.Sum256(content)
		material = append(material, 1)
		material = append(material, contentHash[:]...)
	}
	if !watched {
		return zero, false, nil
	}
	return sha256.Sum256(material), true, nil
}

func (p *promRunner) buildPromWithStableTLSFiles() (*iprom.Prom, [sha256.Size]byte, bool, error) {
	var zero [sha256.Size]byte
	for range 2 {
		before, watched, err := promTLSFileRevision(p.conf)
		if err != nil {
			return nil, zero, watched, err
		}
		pm, err := p.buildProm()
		if err != nil {
			return nil, zero, watched, err
		}
		after, stillWatched, err := promTLSFileRevision(p.conf)
		if err != nil {
			pm.CloseIdleConnections()
			return nil, zero, watched, err
		}
		if watched == stillWatched && before == after {
			return pm, after, watched, nil
		}
		pm.CloseIdleConnections()
	}
	return nil, zero, true, fmt.Errorf("TLS credential files changed while reloading")
}

func (p *promRunner) promForRequest() (*iprom.Prom, error) {
	revision, watched, err := promTLSFileRevision(p.conf)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return nil, context.Canceled
	}
	if !watched || !p.reloadTLSFiles || revision == p.tlsFileRevision {
		return p.pm, nil
	}

	pm, stableRevision, reloadTLSFiles, err := p.buildPromWithStableTLSFiles()
	if err != nil {
		return nil, err
	}
	if stableRevision == p.tlsFileRevision {
		pm.CloseIdleConnections()
		return p.pm, nil
	}

	oldProm := p.pm
	p.pm = pm
	p.tlsFileRevision = stableRevision
	p.reloadTLSFiles = reloadTLSFiles
	oldProm.CloseIdleConnections()
	return p.pm, nil
}

func (p *promRunner) scrape(ctx context.Context, scheduledAt time.Time, requestTimeout time.Duration) error {
	if p.conf == nil {
		return nil
	}
	requestTimeout = maxPromRequestTimeout(requestTimeout, p.conf.Timeout)

	p.lastTime = inputs.AlignTime(scheduledAt, p.lastTime, p.conf.Interval)
	klog.Debugf("running collect from source %s", p.conf.Source)

	for _, u := range p.conf.URLs {
		if err := ctx.Err(); err != nil {
			return err
		}

		p.currentURL = u
		p.collectStart = time.Now()
		pm, err := p.promForRequest()
		if err != nil {
			podAnnotationPromScrapesTotal.WithLabelValues(promScrapeResult(err)).Inc()
			return err
		}
		collectOptions := []iprom.PromOption{iprom.WithTimestamp(p.lastTime.UnixNano())}
		if p.conf.BearerTokenFile != "" {
			token, err := readPromBearerToken(p.conf.BearerTokenFile)
			if err != nil {
				podAnnotationPromScrapesTotal.WithLabelValues(promScrapeResult(err)).Inc()
				return err
			}
			collectOptions = append(collectOptions, iprom.WithBearerToken(token))
		}
		err = p.collectFromURL(ctx, pm, u, requestTimeout, collectOptions...)
		podAnnotationPromScrapesTotal.WithLabelValues(promScrapeResult(err)).Inc()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *promRunner) collectFromURL(ctx context.Context, pm *iprom.Prom, u string,
	requestTimeout time.Duration, opts ...iprom.PromOption,
) (err error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	podAnnotationPromInflightScrapes.Inc()
	defer podAnnotationPromInflightScrapes.Dec()

	// use callback processor, not return pts
	_, err = pm.CollectFromHTTPV2Context(requestCtx, u, opts...)
	if requestErr := requestCtx.Err(); requestErr != nil {
		err = requestErr
	}
	return err
}

func maxPromRequestTimeout(managerTimeout, configuredTimeout time.Duration) time.Duration {
	if configuredTimeout > managerTimeout {
		return configuredTimeout
	}
	return managerTimeout
}

func promScrapeResult(err error) string {
	if err == nil {
		return "success"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "timeout"
	}
	return "error"
}

func (p *promRunner) close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	pm := p.pm
	p.mu.Unlock()
	if pm != nil {
		pm.CloseIdleConnections()
	}
}

const (
	annotationPromExport = "datakit/prom.instances"
	defaultInterval      = time.Second * 30
	minPromInterval      = 10 * time.Second
	maxPromInterval      = 5 * time.Minute
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
		cfg.Interval = config.ProtectedInterval(minPromInterval, maxPromInterval, cfg.Interval)
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
