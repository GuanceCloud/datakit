// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
	iprom "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/prom"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type kubeMetricsWithProm interface {
	Election() bool
	Collect() error
}

type kubeApiserverCollection struct {
	pm     *promRunner
	client k8sClient
}

func newKubeApiserverCollection(client k8sClient, feeder dkio.Feeder) (kubeMetricsWithProm, error) {
	config := promConfig{
		Source:          "kube-apiserver",
		MeasurementName: "kube-apiserver",
		Election:        true,
	}
	pm, err := newPromRunner(&config, feeder)
	if err != nil {
		return nil, err
	}
	return &kubeApiserverCollection{pm, client}, nil
}
func (k *kubeApiserverCollection) Election() bool { return true }
func (k *kubeApiserverCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()

	body, err := k.client.GetAbsPath("/metrics").Stream(context.Background())
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer body.Close()
	k.pm.parseReader(body)
	return nil
}

type kubeCorednsCollection struct {
	pm []*promRunner
}

func newKubeCorednsCollection(client k8sClient, feeder dkio.Feeder) (kubeMetricsWithProm, error) {
	ep, err := client.GetEndpoints("kube-system").Get(context.Background(), "kube-dns", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	collection := kubeCorednsCollection{}

	for _, subset := range ep.Subsets {
		targetPort := 9153
		for _, port := range subset.Ports {
			if port.Name == "metrics" {
				targetPort = int(port.Port)
				break
			}
		}

		for _, address := range subset.Addresses {
			config := promConfig{
				URL:             fmt.Sprintf("http://%s/metrics", net.JoinHostPort(address.IP, strconv.Itoa(targetPort))),
				Source:          "kube-coredns",
				MeasurementName: "kube-coredns",
				Election:        true,
			}
			if address.NodeName != nil {
				config.Tags = map[string]string{"node_name": *address.NodeName}
			}

			pm, err := newPromRunner(&config, feeder)
			if err != nil {
				return nil, err
			}

			collection.pm = append(collection.pm, pm)
		}
	}

	return &collection, nil
}
func (k *kubeCorednsCollection) Election() bool { return true }
func (k *kubeCorednsCollection) Collect() error {
	for _, pm := range k.pm {
		pm.lastTime = time.Now()
		pm.collectOnce()
	}
	return nil
}

type kubeletCadvisorCollection struct {
	pm *promRunner
}

func newKubeletCadvisorCollection(client k8sClient, feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	config := promConfig{
		URL:             fmt.Sprintf("https://%s/metrics/cadvisor", k8sclient.DefaultKubeletHostInCluster()),
		Source:          "kubelet-cadvisor",
		MeasurementName: "kubelet-cadvisor",
		TLSOpen:         true,
		BearerTokenFile: k8sclient.TokenFile,
		Tags:            map[string]string{"node_name": nodeName},
	}

	pm, err := newPromRunner(&config, feeder)
	if err != nil {
		return nil, err
	}
	return &kubeletCadvisorCollection{pm}, nil
}
func (k *kubeletCadvisorCollection) Election() bool { return false }
func (k *kubeletCadvisorCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()
	k.pm.collectOnce()
	return nil
}

type kubeletMetricsResourceCollection struct {
	pm *promRunner
}

func newKubeletMetricsResourceCollection(client k8sClient, feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	config := promConfig{
		URL:             fmt.Sprintf("https://%s/metrics/resource", k8sclient.DefaultKubeletHostInCluster()),
		Source:          "kubelet-resource",
		MeasurementName: "kubelet-resource",
		Tags:            map[string]string{"node_name": nodeName},
		TLSOpen:         true,
		BearerTokenFile: k8sclient.TokenFile,
	}

	pm, err := newPromRunner(&config, feeder)
	if err != nil {
		return nil, err
	}
	return &kubeletMetricsResourceCollection{pm}, nil
}
func (k *kubeletMetricsResourceCollection) Election() bool { return false }
func (k *kubeletMetricsResourceCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()
	k.pm.collectOnce()
	return nil
}

type kubeletProxyCollection struct {
	pm *promRunner
}

func newKubeletProxyCollection(feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	config := promConfig{
		URL:             "http://127.0.0.1:10249/metrics",
		Source:          "kube-proxy",
		MeasurementName: "kube-proxy",
		Tags:            map[string]string{"node_name": nodeName},
	}

	pm, err := newPromRunner(&config, feeder)
	if err != nil {
		return nil, err
	}
	return &kubeletProxyCollection{pm}, nil
}
func (k *kubeletProxyCollection) Election() bool { return false }
func (k *kubeletProxyCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()
	k.pm.collectOnce()
	return nil
}

type kubeControllerManagerCollection struct {
	pm *promRunner
}

func newKubeControllerManagerCollection(client k8sClient, feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	listopts := metav1.ListOptions{
		LabelSelector: "tier=control-plane,component=kube-controller-manager",
		FieldSelector: "spec.nodeName=" + nodeName,
	}
	list, err := client.GetPods("kube-system").List(context.Background(), listopts)
	if err != nil {
		return nil, err
	}

	collection := kubeControllerManagerCollection{}

	for _, pod := range list.Items {
		config := promConfig{
			URL:             "https://127.0.0.1:10257/metrics",
			Source:          "kube-controller-manager",
			MeasurementName: "kube-controller-manager",
			Tags:            map[string]string{"node_name": pod.Spec.NodeName},
			TLSOpen:         true,
			BearerTokenFile: k8sclient.TokenFile,
		}

		pm, err := newPromRunner(&config, feeder)
		if err != nil {
			return nil, err
		} else {
			collection.pm = pm
			break
		}
	}

	return &collection, nil
}
func (k *kubeControllerManagerCollection) Election() bool { return false }
func (k *kubeControllerManagerCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()
	k.pm.collectOnce()
	return nil
}

type kubeSchedulerCollection struct {
	pm *promRunner
}

func newKubeSchedulerCollection(client k8sClient, feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	listopts := metav1.ListOptions{
		LabelSelector: "tier=control-plane,component=kube-scheduler",
		FieldSelector: "spec.nodeName=" + nodeName,
	}

	list, err := client.GetPods("kube-system").List(context.Background(), listopts)
	if err != nil {
		return nil, err
	}

	collection := kubeSchedulerCollection{}

	for _, pod := range list.Items {
		config := promConfig{
			URL:             "https://127.0.0.1:10259/metrics",
			Source:          "kube-scheduler",
			MeasurementName: "kube-scheduler",
			Tags:            map[string]string{"node_name": pod.Spec.NodeName},
			TLSOpen:         true,
			BearerTokenFile: k8sclient.TokenFile,
		}

		pm, err := newPromRunner(&config, feeder)
		if err != nil {
			return nil, err
		} else {
			collection.pm = pm
			break
		}
	}

	return &collection, nil
}
func (k *kubeSchedulerCollection) Election() bool { return false }
func (k *kubeSchedulerCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()
	k.pm.collectOnce()
	return nil
}

type kubeEtcdCollection struct {
	pm *promRunner
}

func newKubeEtcdCollection(client k8sClient, feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	listopts := metav1.ListOptions{
		LabelSelector: "tier=control-plane,component=etcd",
		FieldSelector: "spec.nodeName=" + nodeName,
	}
	list, err := client.GetPods("kube-system").List(context.Background(), listopts)
	if err != nil {
		return nil, err
	}

	cfg := getSelfMetricConfig()
	collection := kubeEtcdCollection{}

	for _, pod := range list.Items {
		config := promConfig{
			URL:             "https://127.0.0.1:2379/metrics",
			Source:          "kube-etcd",
			MeasurementName: "kube-etcd",
			Tags:            map[string]string{"node_name": pod.Spec.NodeName},
			TLSOpen:         true,
			CaFile:          cfg.Etcd.CaFile,
			CertFile:        cfg.Etcd.CertFile,
			KeyFile:         cfg.Etcd.KeyFile,
		}

		pm, err := newPromRunner(&config, feeder)
		if err != nil {
			return nil, err
		} else {
			collection.pm = pm
			break
		}
	}

	return &collection, nil
}
func (k *kubeEtcdCollection) Election() bool { return false }
func (k *kubeEtcdCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()
	k.pm.collectOnce()
	return nil
}

type promConfig struct {
	Source                    string
	URL                       string
	MeasurementName           string
	TLSOpen                   bool
	CaFile, CertFile, KeyFile string
	BearerTokenFile           string
	Election                  bool
	Tags                      map[string]string
}

type promRunner struct {
	conf       *promConfig
	feeder     dkio.Feeder
	pm         *iprom.Prom
	currentURL string
	lastTime   time.Time
}

func newPromRunner(c *promConfig, feeder dkio.Feeder) (*promRunner, error) {
	p := &promRunner{
		conf:     c,
		feeder:   feeder,
		lastTime: time.Now(),
	}

	host, err := parseURLHost(c.URL)
	if err != nil {
		klog.Warnf("failed to parse url %s", err)
		return nil, fmt.Errorf("parse url error: %w", err)
	}
	p.currentURL = host

	callbackFunc := func(pts []*point.Point) error {
		if len(pts) == 0 {
			return nil
		}

		// append instance tag to points
		if p.currentURL != "" {
			for _, pt := range pts {
				pt.AddTag("instance", p.currentURL)
			}
		}

		err := p.feeder.FeedV2(
			point.Metric,
			pts,
			dkio.WithCollectCost(time.Since(p.lastTime)),
			dkio.WithElection(p.conf.Election),
			dkio.WithInputName(p.conf.Source),
		)
		if err != nil {
			klog.Warnf("failed to feed prom metrics: %s, ignored", err)
		}
		return nil
	}

	defaultPrometheusioConnectKeepAlive := time.Second * 20
	opts := []iprom.PromOption{
		iprom.WithLogger(klog), // WithLogger must in the first
		iprom.WithSource(c.Source),
		iprom.WithMeasurementName(c.MeasurementName),
		iprom.WithTLSOpen(c.TLSOpen),
		iprom.WithKeepAlive(defaultPrometheusioConnectKeepAlive),
		iprom.WithTags(c.Tags),
		iprom.WithMetricNameFilterIgnore([]string{".*bucket$"}),
		iprom.WithMaxBatchCallback(1 /*streamSize*/, callbackFunc),
	}

	if c.CaFile != "" && c.CertFile != "" {
		opts = append(opts,
			iprom.WithCacertFiles([]string{c.CaFile}),
			iprom.WithCertFile(c.CertFile),
			iprom.WithKeyFile(c.KeyFile),
		)
		klog.Warn("ETCD with tls")
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
		klog.Warnf("failed to create prom %s", err)
		return nil, fmt.Errorf("failed to create prom: %w", err)
	}

	p.pm = pm
	return p, nil
}

func (p *promRunner) parseReader(body io.Reader) {
	_, err := p.pm.ProcessMetrics(body, "")
	if err != nil {
		klog.Warnf("failed to parse prom: %s", err)
	}
}

func (p *promRunner) collectOnce() {
	_, err := p.pm.CollectFromHTTPV2(p.conf.URL)
	if err != nil {
		klog.Warnf("failed to collect prom: %s", err)
		return
	}
}

func parseURLHost(urlstr string) (string, error) {
	u, err := url.Parse(urlstr)
	if err != nil {
		return "", fmt.Errorf("invalid url %s, err: %w", urlstr, err)
	}
	return u.Host, nil
}

type selfMetricConfig struct {
	Etcd struct {
		CaFile   string `json:"ca_file"`
		CertFile string `json:"cert_file"`
		KeyFile  string `json:"key_file"`
	} `json:"etcd"`
}

func getSelfMetricConfig() *selfMetricConfig {
	s := os.Getenv("ENV_INPUT_CONTAINER_K8S_SELF_METRIC_CONFIG")

	var cfg selfMetricConfig
	if err := json.Unmarshal([]byte(s), &cfg); err != nil {
		klog.Warnf("failed to parse config: %s", err)
	}

	return &cfg
}
