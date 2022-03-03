package container

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubev1apps "k8s.io/client-go/kubernetes/typed/apps/v1"
	kubev1batch "k8s.io/client-go/kubernetes/typed/batch/v1"
	kubev1batchbeta1 "k8s.io/client-go/kubernetes/typed/batch/v1beta1"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	kubev1extensionsbeta1 "k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	kubev1rbac "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/rest"
)

var (
	metaV1ListOption = metav1.ListOptions{}
	metaV1GetOption  = metav1.GetOptions{}
)

type k8sClientX interface {
	getDeployments() kubev1apps.DeploymentInterface
	getDaemonSets() kubev1apps.DaemonSetInterface
	getReplicaSets() kubev1apps.ReplicaSetInterface
	getStatefulSets() kubev1apps.StatefulSetInterface
	getJobs() kubev1batch.JobInterface
	getCronJobs() kubev1batchbeta1.CronJobInterface
	getEndpoints() kubev1core.EndpointsInterface
	getServices() kubev1core.ServiceInterface
	getNodes() kubev1core.NodeInterface
	getNamespaces() kubev1core.NamespaceInterface
	getPods() kubev1core.PodInterface
	getPodsForNamespace(namespace string) kubev1core.PodInterface
	getClusters() kubev1rbac.ClusterRoleInterface
	getIngress() kubev1extensionsbeta1.IngressInterface
	getEvents() kubev1core.EventInterface
}

type k8sClient struct {
	namespace string

	metricsClient k8sMetricsClientX

	*kubernetes.Clientset
}

func newK8sClientFromBearerToken(baseURL, path string) (*k8sClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("invalid baseURL, cannot be empty")
	}

	token, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return newK8sClientFromBearerTokenString(baseURL, strings.TrimSpace(string(token)))
}

func newK8sClientFromBearerTokenString(baseURL, token string) (*k8sClient, error) {
	restConfig := &rest.Config{
		Host:        baseURL,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	return newK8sClient(restConfig)
}

//nolint:deadcode,unused
func newK8sClientFromTLS(baseURL string, tlsconfig *net.TLSClientConfig) (*k8sClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("invalid baseURL, cannot be empty")
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
		Host: baseURL,
	}

	return newK8sClient(restConfig)
}

func newK8sClient(restConfig *rest.Config) (*k8sClient, error) {
	config, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	k := &k8sClient{Clientset: config}

	if c, err := newK8sMetricsClient(restConfig); err != nil {
		l.Errorf("failed to connect k8s metrics-server, error: %w", err)
	} else {
		k.metricsClient = c
	}

	return k, nil
}

//nolint:deadcode,unused
func (c *k8sClient) setNamespace(namespace string) {
	c.namespace = namespace
}

func (c *k8sClient) getDeployments() kubev1apps.DeploymentInterface {
	return c.AppsV1().Deployments(c.namespace)
}

func (c *k8sClient) getDaemonSets() kubev1apps.DaemonSetInterface {
	return c.AppsV1().DaemonSets(c.namespace)
}

func (c *k8sClient) getReplicaSets() kubev1apps.ReplicaSetInterface {
	return c.AppsV1().ReplicaSets(c.namespace)
}

func (c *k8sClient) getStatefulSets() kubev1apps.StatefulSetInterface {
	return c.AppsV1().StatefulSets(c.namespace)
}

func (c *k8sClient) getJobs() kubev1batch.JobInterface {
	return c.BatchV1().Jobs(c.namespace)
}

func (c *k8sClient) getCronJobs() kubev1batchbeta1.CronJobInterface {
	return c.BatchV1beta1().CronJobs(c.namespace)
}

func (c *k8sClient) getEndpoints() kubev1core.EndpointsInterface {
	return c.CoreV1().Endpoints(c.namespace)
}

func (c *k8sClient) getServices() kubev1core.ServiceInterface {
	return c.CoreV1().Services(c.namespace)
}

func (c *k8sClient) getNodes() kubev1core.NodeInterface {
	return c.CoreV1().Nodes()
}

func (c *k8sClient) getNamespaces() kubev1core.NamespaceInterface {
	return c.CoreV1().Namespaces()
}

func (c *k8sClient) getPods() kubev1core.PodInterface {
	return c.CoreV1().Pods(c.namespace)
}

func (c *k8sClient) getPodsForNamespace(namespace string) kubev1core.PodInterface {
	return c.CoreV1().Pods(namespace)
}

func (c *k8sClient) getClusters() kubev1rbac.ClusterRoleInterface {
	return c.RbacV1().ClusterRoles()
}

func (c *k8sClient) getIngress() kubev1extensionsbeta1.IngressInterface {
	return c.ExtensionsV1beta1().Ingresses(c.namespace)
}

func (c *k8sClient) getEvents() kubev1core.EventInterface {
	return c.CoreV1().Events(c.namespace)
}
