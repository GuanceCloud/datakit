// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package client wrap kubernetes client functions
package client

import (
	"errors"
	"fmt"
	"io/ioutil"
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
	rbacv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"

	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

type Client interface {
	// resources
	GetDeployments() appsv1.DeploymentInterface
	GetDaemonSets() appsv1.DaemonSetInterface
	GetReplicaSets() appsv1.ReplicaSetInterface
	GetStatefulSets() appsv1.StatefulSetInterface
	GetJobs() batchv1.JobInterface
	GetCronJobs() batchv1.CronJobInterface
	GetEndpoints() corev1.EndpointsInterface
	GetServices() corev1.ServiceInterface
	GetNodes() corev1.NodeInterface
	GetNamespaces() corev1.NamespaceInterface
	GetPods() corev1.PodInterface
	GetClusterRoles() rbacv1.ClusterRoleInterface
	GetIngress() extensionsv1beta1.IngressInterface
	GetEvents() corev1.EventInterface

	// CRDs
	GetDatakits() guancev1beta1.DatakitInterface
	GetPrmetheusPodMonitors() prometheusmonitoringv1.PodMonitorInterface
	GetPrmetheusServiceMonitors() prometheusmonitoringv1.ServiceMonitorInterface

	// for namespace
	GetPodsForNamespace(string) corev1.PodInterface
	GetReplicaSetsForNamespace(string) appsv1.ReplicaSetInterface
	GetDaemonSetsForNamespace(string) appsv1.DaemonSetInterface
	GetDeploymentsForNamespace(string) appsv1.DeploymentInterface
	GetServicesForNamespace(string) corev1.ServiceInterface

	// plugins
	// metrics-server
	GetPodMetricses() metricsv1beta1.PodMetricsInterface
	GetPodMetricsesForNamespace(namespace string) metricsv1beta1.PodMetricsInterface
	GetNodeMetricses() metricsv1beta1.NodeMetricsInterface
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

	token, err := ioutil.ReadFile(filepath.Clean(path))
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

	return &client{
		clientset:              clientset,
		metricsV1beta1:         metricsClient,
		guanceV1beta1:          guanceClient,
		prometheusMonitoringV1: prometheusClient,
	}, nil
}

func (c *client) GetDeployments() appsv1.DeploymentInterface {
	return c.clientset.AppsV1().Deployments(NAMESPACE)
}

func (c *client) GetDeploymentsForNamespace(namespace string) appsv1.DeploymentInterface {
	return c.clientset.AppsV1().Deployments(namespace)
}

func (c *client) GetDaemonSets() appsv1.DaemonSetInterface {
	return c.clientset.AppsV1().DaemonSets(NAMESPACE)
}

func (c *client) GetDaemonSetsForNamespace(namespace string) appsv1.DaemonSetInterface {
	return c.clientset.AppsV1().DaemonSets(namespace)
}

func (c *client) GetReplicaSets() appsv1.ReplicaSetInterface {
	return c.clientset.AppsV1().ReplicaSets(NAMESPACE)
}

func (c *client) GetStatefulSets() appsv1.StatefulSetInterface {
	return c.clientset.AppsV1().StatefulSets(NAMESPACE)
}

func (c *client) GetJobs() batchv1.JobInterface {
	return c.clientset.BatchV1().Jobs(NAMESPACE)
}

func (c *client) GetCronJobs() batchv1.CronJobInterface {
	return c.clientset.BatchV1().CronJobs(NAMESPACE)
}

func (c *client) GetEndpoints() corev1.EndpointsInterface {
	return c.clientset.CoreV1().Endpoints(NAMESPACE)
}

func (c *client) GetServices() corev1.ServiceInterface {
	return c.clientset.CoreV1().Services(NAMESPACE)
}

func (c *client) GetServicesForNamespace(namespace string) corev1.ServiceInterface {
	return c.clientset.CoreV1().Services(namespace)
}

func (c *client) GetNodes() corev1.NodeInterface {
	return c.clientset.CoreV1().Nodes()
}

func (c *client) GetNamespaces() corev1.NamespaceInterface {
	return c.clientset.CoreV1().Namespaces()
}

func (c *client) GetPods() corev1.PodInterface {
	return c.clientset.CoreV1().Pods(NAMESPACE)
}

func (c *client) GetPodsForNamespace(namespace string) corev1.PodInterface {
	return c.clientset.CoreV1().Pods(namespace)
}

func (c *client) GetReplicaSetsForNamespace(namespace string) appsv1.ReplicaSetInterface {
	return c.clientset.AppsV1().ReplicaSets(namespace)
}

func (c *client) GetClusterRoles() rbacv1.ClusterRoleInterface {
	return c.clientset.RbacV1().ClusterRoles()
}

func (c *client) GetIngress() extensionsv1beta1.IngressInterface {
	return c.clientset.ExtensionsV1beta1().Ingresses(NAMESPACE)
}

func (c *client) GetEvents() corev1.EventInterface {
	return c.clientset.CoreV1().Events(NAMESPACE)
}

/// CRDs

func (c *client) GetDatakits() guancev1beta1.DatakitInterface {
	return c.guanceV1beta1.Datakits(NAMESPACE)
}

func (c *client) GetPrmetheusPodMonitors() prometheusmonitoringv1.PodMonitorInterface {
	return c.prometheusMonitoringV1.MonitoringV1().PodMonitors(NAMESPACE)
}

func (c *client) GetPrmetheusServiceMonitors() prometheusmonitoringv1.ServiceMonitorInterface {
	return c.prometheusMonitoringV1.MonitoringV1().ServiceMonitors(NAMESPACE)
}

/// plugins

func (c *client) GetPodMetricses() metricsv1beta1.PodMetricsInterface {
	return c.metricsV1beta1.PodMetricses(NAMESPACE)
}

func (c *client) GetPodMetricsesForNamespace(namespace string) metricsv1beta1.PodMetricsInterface {
	return c.metricsV1beta1.PodMetricses(namespace)
}

func (c *client) GetNodeMetricses() metricsv1beta1.NodeMetricsInterface {
	return c.metricsV1beta1.NodeMetricses()
}
