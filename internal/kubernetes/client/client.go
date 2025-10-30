// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package client wrap kubernetes client functions
package client

import (
	"fmt"
	"net"
	"os"

	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	batchv1 "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	extensionsv1beta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	statsv1alpha1 "k8s.io/kubelet/pkg/apis/stats/v1alpha1"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	prometheusclientv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned"
	prometheusmonitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"

	loggingclientv1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/client/clientset/versioned"
	loggingv1alpha1 "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/client/clientset/versioned/typed/datakits/v1alpha1"
)

type Client interface {
	// clients
	KubernetesClientset() *kubernetes.Clientset
	PrometheusClient() *prometheusclientv1.Clientset
	LoggingClient() *loggingclientv1.Clientset

	// resources
	GetNamespaces() corev1.NamespaceInterface
	GetNodes() corev1.NodeInterface
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
	GetPersistentVolumes() corev1.PersistentVolumeInterface
	GetPersistentVolumeClaims(ns string) corev1.PersistentVolumeClaimInterface

	// CRDs
	GetLoggingConfigs() loggingv1alpha1.ClusterLoggingConfigInterface
	GetPrmetheusPodMonitors(ns string) prometheusmonitoringv1.PodMonitorInterface
	GetPrmetheusServiceMonitors(ns string) prometheusmonitoringv1.ServiceMonitorInterface

	// plugins metrics-server
	GetPodMetricses(ns string) metricsv1beta1.PodMetricsInterface
	GetNodeMetricses(ns string) metricsv1beta1.NodeMetricsInterface

	KubeletClient
}

const (
	NAMESPACE = "" // Use all namespace

	LimiteQPS  = float32(1000)
	LimitBurst = 1000

	//nolint:gosec
	TokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

func DefaultConfigInCluster() (*rest.Config, error) {
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, fmt.Errorf("unable to load in-cluster configuration")
	}

	token, err := os.ReadFile(TokenFile)
	if err != nil {
		return nil, err
	}

	config := rest.Config{
		Host:            "https://" + net.JoinHostPort(host, port),
		TLSClientConfig: rest.TLSClientConfig{Insecure: true},
		BearerToken:     string(token),
		BearerTokenFile: TokenFile,
		RateLimiter:     flowcontrol.NewTokenBucketRateLimiter(LimiteQPS, LimitBurst), // setting default limit
	}
	return &config, nil
}

func NewKubernetesClientInCluster() (Client, error) {
	config, err := DefaultConfigInCluster()
	if err != nil {
		return nil, err
	}
	return newKubernetesClient(config)
}

type client struct {
	clientset        *kubernetes.Clientset
	kubeletClient    KubeletClient
	metricsClient    *metricsv1beta1.MetricsV1beta1Client
	prometheusClient *prometheusclientv1.Clientset
	loggingClient    *loggingclientv1.Clientset
}

func newKubernetesClient(restConfig *rest.Config) (*client, error) {
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	kubeletClient, err := NewDefaultKubeletClient(restConfig)
	if err != nil {
		return nil, err
	}

	prometheusClient, err := prometheusclientv1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	loggingClient, err := loggingclientv1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsv1beta1.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		clientset:        clientset,
		kubeletClient:    kubeletClient,
		metricsClient:    metricsClient,
		prometheusClient: prometheusClient,
		loggingClient:    loggingClient,
	}, nil
}

func (c *client) KubernetesClientset() *kubernetes.Clientset      { return c.clientset }
func (c *client) PrometheusClient() *prometheusclientv1.Clientset { return c.prometheusClient }
func (c *client) LoggingClient() *loggingclientv1.Clientset       { return c.loggingClient }

func (c *client) GetNamespaces() corev1.NamespaceInterface { return c.clientset.CoreV1().Namespaces() }
func (c *client) GetNodes() corev1.NodeInterface           { return c.clientset.CoreV1().Nodes() }

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

func (c *client) GetJobs(ns string) batchv1.JobInterface { return c.clientset.BatchV1().Jobs(ns) }

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

func (c *client) GetPersistentVolumes() corev1.PersistentVolumeInterface {
	return c.clientset.CoreV1().PersistentVolumes()
}

func (c *client) GetPersistentVolumeClaims(ns string) corev1.PersistentVolumeClaimInterface {
	return c.clientset.CoreV1().PersistentVolumeClaims(ns)
}

/// CRDSs

func (c *client) GetLoggingConfigs() loggingv1alpha1.ClusterLoggingConfigInterface {
	return c.loggingClient.LoggingV1alpha1().ClusterLoggingConfigs()
}

func (c *client) GetPrmetheusPodMonitors(ns string) prometheusmonitoringv1.PodMonitorInterface {
	return c.prometheusClient.MonitoringV1().PodMonitors(ns)
}

func (c *client) GetPrmetheusServiceMonitors(ns string) prometheusmonitoringv1.ServiceMonitorInterface {
	return c.prometheusClient.MonitoringV1().ServiceMonitors(ns)
}

/// plugins

func (c *client) GetPodMetricses(ns string) metricsv1beta1.PodMetricsInterface {
	return c.metricsClient.PodMetricses(ns)
}

func (c *client) GetNodeMetricses(ns string) metricsv1beta1.NodeMetricsInterface {
	return c.metricsClient.NodeMetricses()
}

// kubelet.
func (c *client) GetStatsSummary() (*statsv1alpha1.Summary, error) {
	return c.kubeletClient.GetStatsSummary()
}
