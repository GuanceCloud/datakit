package kubernetes

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchbetav1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	kubev1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/net"
)

type client struct {
	namespace string
	*kubernetes.Clientset
}

func newClientFromBearerToken(baseURL, path string) (*client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("invalid baseURL, cannot be empty")
	}

	token, err := ioutil.ReadFile(path)
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

	_, err := tlsconfig.TLSConfig()
	if err != nil {
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

func (c *client) getClusters() (*rbacv1.ClusterRoleList, error) {
	return c.RbacV1().ClusterRoles().List(context.Background(), metav1.ListOptions{})
}

func (c *client) getPods() (*corev1.PodList, error) {
	return c.CoreV1().Pods(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getDeployments() (*appsv1.DeploymentList, error) {
	return c.AppsV1().Deployments(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getReplicaSets() (*appsv1.ReplicaSetList, error) {
	return c.AppsV1().ReplicaSets(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getServices() (*corev1.ServiceList, error) {
	return c.CoreV1().Services(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getNodes() (*corev1.NodeList, error) {
	return c.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
}

func (c *client) getJobs() (*batchv1.JobList, error) {
	return c.BatchV1().Jobs(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getCronJobs() (*batchbetav1.CronJobList, error) {
	return c.BatchV1beta1().CronJobs(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getEndpoints() (*corev1.EndpointsList, error) {
	return c.CoreV1().Endpoints(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getNamespaces() (*corev1.NamespaceList, error) {
	return c.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
}

func (c *client) getDaemonSets() (*appsv1.DaemonSetList, error) {
	return c.AppsV1().DaemonSets(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getStatefulSets() (*appsv1.StatefulSetList, error) {
	return c.AppsV1().StatefulSets(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getIngress() (*v1beta1.IngressList, error) {
	return c.ExtensionsV1beta1().Ingresses(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getPersistentVolumes() (*corev1.PersistentVolumeList, error) {
	return c.CoreV1().PersistentVolumes().List(context.Background(), metav1.ListOptions{})
}

func (c *client) getPersistentVolumeClaims() (*corev1.PersistentVolumeClaimList, error) {
	return c.CoreV1().PersistentVolumeClaims(c.namespace).List(context.Background(), metav1.ListOptions{})
}

func (c *client) getEvent() (kubev1core.EventInterface, error) {
	return c.CoreV1().Events(c.namespace), nil
}
