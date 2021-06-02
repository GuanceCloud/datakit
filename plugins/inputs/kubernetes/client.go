package kubernetes

import (
	"errors"
	"fmt"
	"io/ioutil"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"net"
	"net/http"
	"os"
	"time"
	// netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	certutil "k8s.io/client-go/util/cert"

	"github.com/influxdata/telegraf/plugins/common/tls"
)

var ErrNotInCluster = errors.New("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")

type client struct {
	namespace string
	timeout   time.Duration
	*kubernetes.Clientset
	restClient *http.Client
}

func createConfigByKubePath(kubePath string) (*rest.Config, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubePath)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func createConfigByToken(baseURL, bearerToken string, caFile string, insecureSkipVerify bool) (*rest.Config, error) {
	const (
		tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
		rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	)

	if baseURL == "" {
		host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
		if len(host) == 0 || len(port) == 0 {
			return nil, ErrNotInCluster
		}
		baseURL = "https://" + net.JoinHostPort(host, port)
	}

	if bearerToken == "" {
		token, err := ioutil.ReadFile(tokenFile)
		if err != nil {
			return nil, err
		}
		bearerToken = string(token)
	}

	tlsClientConfig := rest.TLSClientConfig{
		Insecure: insecureSkipVerify,
	}

	if caFile == "" {
		caFile = rootCAFile
	}
	if _, err := certutil.NewPool(caFile); err != nil {
		return nil, fmt.Errorf("Expected to load root CA config from %s, but got err: %v", caFile, err)
	} else {
		tlsClientConfig.CAFile = caFile
	}

	return &rest.Config{
		Host:            baseURL,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     bearerToken,
	}, nil
}

func createConfigByCert(baseURL string, tlsConfig *tls.ClientConfig) *rest.Config {
	config := &rest.Config{
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: tlsConfig.InsecureSkipVerify,
			CAFile:   tlsConfig.TLSCA,
			CertFile: tlsConfig.TLSCert,
			KeyFile:  tlsConfig.TLSKey,
		},
		Host: baseURL,
	}

	return config
}

func newClient(config *rest.Config, timeout time.Duration) (*client, error) {
	cli := &client{
		timeout: timeout,
	}

	if config != nil {
		c, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		cli.Clientset = c
	}

	cli.restClient = &http.Client{
		Timeout: 1 * time.Second,
	}

	return cli, nil
}

func (c *client) promMetrics(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.restClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *client) getDaemonSets() (*appsv1.DaemonSetList, error) {
	return c.AppsV1().DaemonSets(c.namespace).List(metav1.ListOptions{})
}

func (c *client) getDeployments() (*appsv1.DeploymentList, error) {
	return c.AppsV1().Deployments(c.namespace).List(metav1.ListOptions{})
}

func (c *client) getEndpoints() (*corev1.EndpointsList, error) {
	return c.CoreV1().Endpoints(c.namespace).List(metav1.ListOptions{})
}

func (c *client) getNodes() (*corev1.NodeList, error) {
	return c.CoreV1().Nodes().List(metav1.ListOptions{})
}

func (c *client) getPersistentVolumes() (*corev1.PersistentVolumeList, error) {
	return c.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
}

func (c *client) getPersistentVolumeClaims() (*corev1.PersistentVolumeClaimList, error) {
	return c.CoreV1().PersistentVolumeClaims(c.namespace).List(metav1.ListOptions{})
}

func (c *client) getPods() (*corev1.PodList, error) {
	return c.CoreV1().Pods(c.namespace).List(metav1.ListOptions{})
}

func (c *client) getServices() (*corev1.ServiceList, error) {
	return c.CoreV1().Services(c.namespace).List(metav1.ListOptions{})
}

func (c *client) getStatefulSets() (*appsv1.StatefulSetList, error) {
	return c.AppsV1().StatefulSets(c.namespace).List(metav1.ListOptions{})
}

func (c *client) getIngress() (*v1beta1.IngressList, error) {
	return c.ExtensionsV1beta1().Ingresses(c.namespace).List(metav1.ListOptions{})
}
