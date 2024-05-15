// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
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
				Source:          "kube-coredns",
				MeasurementName: "kube-coredns",
				URL:             fmt.Sprintf("http://%s/metrics", net.JoinHostPort(address.IP, strconv.Itoa(targetPort))),
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
	pm     *promRunner
	client k8sClient
}

func newKubeletCadvisorCollection(client k8sClient, feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	config := promConfig{
		Source:          "kubelet-cadvisor",
		MeasurementName: "kubelet-cadvisor",
		Tags:            map[string]string{"node_name": nodeName},
	}
	pm, err := newPromRunner(&config, feeder)
	if err != nil {
		return nil, err
	}
	return &kubeletCadvisorCollection{pm, client}, nil
}
func (k *kubeletCadvisorCollection) Election() bool { return false }
func (k *kubeletCadvisorCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()

	body, err := k.client.GetMetricsCadvisor()
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer body.Close()
	k.pm.parseReader(body)
	return nil
}

type kubeletMetricsResourceCollection struct {
	pm     *promRunner
	client k8sClient
}

func newKubeletMetricsResourceCollection(client k8sClient, feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	config := promConfig{
		Source:          "kubelet-resource",
		MeasurementName: "kubelet-resource",
		Tags:            map[string]string{"node_name": nodeName},
	}
	pm, err := newPromRunner(&config, feeder)
	if err != nil {
		return nil, err
	}
	return &kubeletMetricsResourceCollection{pm, client}, nil
}
func (k *kubeletMetricsResourceCollection) Election() bool { return false }
func (k *kubeletMetricsResourceCollection) Collect() error {
	if k.pm == nil {
		return fmt.Errorf("unexpected")
	}
	k.pm.lastTime = time.Now()

	body, err := k.client.GetMetricsResource()
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer body.Close()
	k.pm.parseReader(body)
	return nil
}

type kubeletProxyCollection struct {
	pm *promRunner
}

func newKubeletProxyCollection(feeder dkio.Feeder, nodeName string) (kubeMetricsWithProm, error) {
	config := promConfig{
		Source:          "kube-proxy",
		MeasurementName: "kube-proxy",
		Tags:            map[string]string{"node_name": nodeName},
		URL:             "http://127.0.0.1:10249/metrics",
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
	pm         *promRunner
	httpclient *k8sclient.BaseClient
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

	httpclient, err := k8sclient.NewBaseClient(client.RestConfig())
	if err != nil {
		return nil, err
	}
	collection := kubeControllerManagerCollection{
		httpclient: httpclient,
	}

	for _, pod := range list.Items {
		config := promConfig{
			Source:          "kube-controller-manager",
			MeasurementName: "kube-controller-manager",
			Tags:            map[string]string{"node_name": pod.Spec.NodeName},
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

	resp, err := k.httpclient.Get("https://127.0.0.1:10257/metrics")
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer resp.Body.Close()
	k.pm.parseReader(resp.Body)
	return nil
}

type kubeSchedulerCollection struct {
	pm         *promRunner
	httpclient *k8sclient.BaseClient
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

	httpclient, err := k8sclient.NewBaseClient(client.RestConfig())
	if err != nil {
		return nil, err
	}
	collection := kubeSchedulerCollection{
		httpclient: httpclient,
	}

	for _, pod := range list.Items {
		config := promConfig{
			Source:          "kube-scheduler",
			MeasurementName: "kube-scheduler",
			Tags:            map[string]string{"node_name": pod.Spec.NodeName},
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

	resp, err := k.httpclient.Get("https://127.0.0.1:10259/metrics")
	if err != nil {
		return err
	}
	// nolint:errcheck
	defer resp.Body.Close()
	k.pm.parseReader(resp.Body)
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

	collection := kubeEtcdCollection{}

	for _, pod := range list.Items {
		config := promConfig{
			Source:          "kube-etcd",
			MeasurementName: "kube-etcd",
			URL:             "http://127.0.0.1:2381/metrics",
			Tags:            map[string]string{"node_name": pod.Spec.NodeName},
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
	Source          string
	URL             string
	MeasurementName string
	BearerToken     string
	CaFile          string
	Election        bool
	Tags            map[string]string
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
			dkio.WithBlocking(true))
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
		iprom.WithKeepAlive(defaultPrometheusioConnectKeepAlive),
		iprom.WithTags(c.Tags),
		iprom.WithMetricNameFilterIgnore([]string{".*bucket$"}),
		iprom.WithMaxBatchCallback(1 /*streamSize*/, callbackFunc),
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
