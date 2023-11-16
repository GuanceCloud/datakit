// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package client wrap kubernetes client functions
package client

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	prometheusclientv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	prometheusmonitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	guancev1beta1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/typed/guance/v1beta1"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	extensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

type Client interface {
	GetNamespaces() corev1.NamespaceInterface
	GetNodes() corev1.NodeInterface

	// resources
	GetDeployments(ns string) appsv1.DeploymentInterface
	GetDaemonSets(ns string) appsv1.DaemonSetInterface
	GetReplicaSets(ns string) appsv1.ReplicaSetInterface
	GetStatefulSets(ns string) appsv1.StatefulSetInterface
	GetJobs(ns string) batchv1.JobInterface
	GetCronJobs(ns string) batchv1.CronJobInterface
	GetEndpoints(ns string) corev1.EndpointsInterface
	GetServices(ns string) corev1.ServiceInterface
	GetPods(ns string) corev1.PodInterface
	GetIngress(ns string) extensionsv1beta1.IngressInterface
	GetEvents(ns string) corev1.EventInterface

	// CRDs
	GetDatakits(ns string) guancev1beta1.DatakitInterface
	GetPrmetheusPodMonitors(ns string) prometheusmonitoringv1.PodMonitorInterface
	GetPrmetheusServiceMonitors(ns string) prometheusmonitoringv1.ServiceMonitorInterface

	// plugins
	// metrics-server
	GetPodMetricses(ns string) metricsv1beta1.PodMetricsInterface
	GetNodeMetricses(ns string) metricsv1beta1.NodeMetricsInterface

	// kubelet
	GetMetricsFromKubelet() (*statsv1alpha1.Summary, error)
}

const (
	NAMESPACE = "" // Use all namespace

	LimiteQPS  = float32(1000)
	LimitBurst = 1000
)

var ErrInvalidBaseURL = errors.New("invalid baseURL, cannot be empty")

func NewKubernetesClientFromBearerToken(baseURL, path string) (Client, error) {
	if baseURL == "" {
		return nil, ErrInvalidBaseURL
	}

	token, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return NewKubernetesClientFromBearerTokenString(baseURL, strings.TrimSpace(string(token)))
}

func NewKubernetesClientFromBearerTokenString(baseURL, token string) (Client, error) {
	if baseURL == "" {
		return nil, ErrInvalidBaseURL
	}

	restConfig := &rest.Config{
		Host:        baseURL,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
		RateLimiter: flowcontrol.NewTokenBucketRateLimiter(LimiteQPS, LimitBurst), // setting default limit
	}
	return newKubernetesClient(restConfig)
}

func NewKubernetesClientFromTLS(baseURL string, tlsconfig *net.TLSClientConfig) (Client, error) {
	if baseURL == "" {
		return nil, ErrInvalidBaseURL
	}

	if tlsconfig == nil {
		return nil, fmt.Errorf("tlsconfig is empty pointer")
	}

	if len(tlsconfig.CaCerts) == 0 {
		return nil, fmt.Errorf("tlsconfig cacerts is empty")
	}

	if _, err := tlsconfig.TLSConfig(); err != nil {
		return nil, err
	}

	restConfig := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: tlsconfig.InsecureSkipVerify,
			CAFile:   tlsconfig.CaCerts[0],
			CertFile: tlsconfig.Cert,
			KeyFile:  tlsconfig.CertKey,
		},
		Host:        baseURL,
		RateLimiter: flowcontrol.NewTokenBucketRateLimiter(LimiteQPS, LimitBurst), // setting default limit
	}

	return newKubernetesClient(restConfig)
}

type client struct {
	clientset              *kubernetes.Clientset
	kubeletClient          *kubeletClient
	metricsV1beta1         *metricsv1beta1.MetricsV1beta1Client
	guanceV1beta1          *guancev1beta1.GuanceV1Client
	prometheusMonitoringV1 *prometheusclientv1.Clientset
}

func newKubernetesClient(restConfig *rest.Config) (*client, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	guanceClient, err := guancev1beta1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	if err := guancev1beta1.AddToScheme(clientsetscheme.Scheme); err != nil {
		return nil, err
	}

	prometheusClient, err := prometheusclientv1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsv1beta1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	kubeletConfig := KubeletClientConfig{
		Client:      restConfig,
		Scheme:      "https",
		DefaultPort: 10250,
	}
	kubeletClient, err := NewKubeletClientForConfig(&kubeletConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		clientset:              clientset,
		kubeletClient:          kubeletClient,
		metricsV1beta1:         metricsClient,
		guanceV1beta1:          guanceClient,
		prometheusMonitoringV1: prometheusClient,
	}, nil
}

func (c *client) GetNamespaces() corev1.NamespaceInterface {
	return c.clientset.CoreV1().Namespaces()
}

func (c *client) GetNodes() corev1.NodeInterface {
	return c.clientset.CoreV1().Nodes()
}

func (c *client) GetDeployments(ns string) appsv1.DeploymentInterface {
	return c.clientset.AppsV1().Deployments(ns)
}

func (c *client) GetDaemonSets(ns string) appsv1.DaemonSetInterface {
	return c.clientset.AppsV1().DaemonSets(ns)
}

func (c *client) GetReplicaSets(ns string) appsv1.ReplicaSetInterface {
	return c.clientset.AppsV1().ReplicaSets(ns)
}

func (c *client) GetStatefulSets(ns string) appsv1.StatefulSetInterface {
	return c.clientset.AppsV1().StatefulSets(ns)
}

func (c *client) GetJobs(ns string) batchv1.JobInterface {
	return c.clientset.BatchV1().Jobs(ns)
}

func (c *client) GetCronJobs(ns string) batchv1.CronJobInterface {
	return c.clientset.BatchV1().CronJobs(ns)
}

func (c *client) GetEndpoints(ns string) corev1.EndpointsInterface {
	return c.clientset.CoreV1().Endpoints(ns)
}

func (c *client) GetServices(ns string) corev1.ServiceInterface {
	return c.clientset.CoreV1().Services(ns)
}

func (c *client) GetPods(ns string) corev1.PodInterface {
	return c.clientset.CoreV1().Pods(ns)
}

func (c *client) GetIngress(ns string) extensionsv1beta1.IngressInterface {
	return c.clientset.ExtensionsV1beta1().Ingresses(ns)
}

func (c *client) GetEvents(ns string) corev1.EventInterface {
	return c.clientset.CoreV1().Events(ns)
}

/// CRDs

func (c *client) GetDatakits(ns string) guancev1beta1.DatakitInterface {
	return c.guanceV1beta1.Datakits(ns)
}

func (c *client) GetPrmetheusPodMonitors(ns string) prometheusmonitoringv1.PodMonitorInterface {
	return c.prometheusMonitoringV1.MonitoringV1().PodMonitors(ns)
}

func (c *client) GetPrmetheusServiceMonitors(ns string) prometheusmonitoringv1.ServiceMonitorInterface {
	return c.prometheusMonitoringV1.MonitoringV1().ServiceMonitors(ns)
}

/// plugins

func (c *client) GetPodMetricses(ns string) metricsv1beta1.PodMetricsInterface {
	return c.metricsV1beta1.PodMetricses(ns)
}

func (c *client) GetNodeMetricses(ns string) metricsv1beta1.NodeMetricsInterface {
	return c.metricsV1beta1.NodeMetricses()
}

func (c *client) GetMetricsFromKubelet() (*statsv1alpha1.Summary, error) {
	return c.kubeletClient.GetMetrics()
}
