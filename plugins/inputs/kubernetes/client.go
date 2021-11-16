package kubernetes

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

type client struct {
	namespace string
	*kubernetes.Clientset
}

func newClientFromBearerToken(baseURL, path string) (*client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("invalid baseURL, cannot be empty")
	}

	token, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return newClientFromBearerTokenString(baseURL, strings.TrimSpace(string(token)))
}

func newClientFromBearerTokenString(baseURL, token string) (*client, error) {
	restConfig := &rest.Config{
		Host:        baseURL,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	return newClient(restConfig)
}

func newClientFromTLS(baseURL string, tlsconfig *net.TLSClientConfig) (*client, error) {
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

	return newClient(restConfig)
}

func newClient(restConfig *rest.Config) (*client, error) {
	config, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return &client{
		Clientset: config,
	}, nil
}

var metav1ListOption = metav1.ListOptions{}

func (c *client) getDeployments() kubev1apps.DeploymentInterface {
	return c.AppsV1().Deployments(c.namespace)
}

func (c *client) getDaemonSets() kubev1apps.DaemonSetInterface {
	return c.AppsV1().DaemonSets(c.namespace)
}

func (c *client) getReplicaSets() kubev1apps.ReplicaSetInterface {
	return c.AppsV1().ReplicaSets(c.namespace)
}

func (c *client) getStatefulSets() kubev1apps.StatefulSetInterface {
	return c.AppsV1().StatefulSets(c.namespace)
}

func (c *client) getJobs() kubev1batch.JobInterface {
	return c.BatchV1().Jobs(c.namespace)
}

func (c *client) getCronJobs() kubev1batchbeta1.CronJobInterface {
	return c.BatchV1beta1().CronJobs(c.namespace)
}

func (c *client) getEndpoints() kubev1core.EndpointsInterface {
	return c.CoreV1().Endpoints(c.namespace)
}

func (c *client) getServices() kubev1core.ServiceInterface {
	return c.CoreV1().Services(c.namespace)
}

func (c *client) getNodes() kubev1core.NodeInterface {
	return c.CoreV1().Nodes()
}

func (c *client) getNamespaces() kubev1core.NamespaceInterface {
	return c.CoreV1().Namespaces()
}

func (c *client) getPods(namespace string) kubev1core.PodInterface {
	return c.CoreV1().Pods(namespace)
}

func (c *client) getClusters() kubev1rbac.ClusterRoleInterface {
	return c.RbacV1().ClusterRoles()
}

func (c *client) getIngress() kubev1extensionsbeta1.IngressInterface {
	return c.ExtensionsV1beta1().Ingresses(c.namespace)
}
