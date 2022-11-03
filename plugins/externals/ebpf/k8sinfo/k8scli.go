// Package k8sinfo used to create k8s client and get some k8s info
package k8sinfo

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var metaV1ListOption = metav1.ListOptions{}

type K8sClient struct {
	*kubernetes.Clientset
}

func NewK8sClientFromBearerToken(baseURL, path string) (*K8sClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("invalid baseURL, cannot be empty")
	}

	token, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return NewK8sClientFromBearerTokenString(baseURL, strings.TrimSpace(string(token)))
}

func NewK8sClientFromBearerTokenString(baseURL, token string) (*K8sClient, error) {
	restConfig := &rest.Config{
		Host:        baseURL,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	return newK8sClient(restConfig)
}

func newK8sClient(restConfig *rest.Config) (*K8sClient, error) {
	config, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	k := &K8sClient{Clientset: config}

	return k, nil
}

func (c *K8sClient) GetNamespaces() ([]string, error) {
	result := []string{}
	nss := c.CoreV1().Namespaces()
	list, err := nss.List(context.Background(), metaV1ListOption)
	if err != nil {
		return result, err
	}
	for _, v := range list.Items {
		result = append(result, v.GetName())
	}
	return result, nil
}

type Port struct {
	Port     uint32
	Protocol string
}

func (p Port) String() string {
	return fmt.Sprintf("%d/%s", p.Port, p.Protocol)
}

type K8sEndpointsNet struct {
	Namespace string
	Name      string // == service name
	IPPort    map[string][]Port
}

func (c *K8sClient) GetEndpointNet(ns string) (map[string]*K8sEndpointsNet, error) {
	result := map[string]*K8sEndpointsNet{}
	ep := c.CoreV1().Endpoints(ns)
	list, err := ep.List(context.Background(), metaV1ListOption)
	if err != nil {
		return result, err
	}

	for _, v := range list.Items {
		ep := &K8sEndpointsNet{
			Namespace: v.GetNamespace(),
			Name:      v.GetName(),
			IPPort:    map[string][]Port{},
		}
		for _, subset := range v.Subsets {
			for _, v := range subset.Addresses {
				ep.IPPort[v.IP] = []Port{}
				for _, p := range subset.Ports {
					ep.IPPort[v.IP] = append(ep.IPPort[v.IP],
						Port{
							Port:     uint32(p.Port),
							Protocol: string(p.Protocol),
						},
					)
				}
			}
		}
		result[ep.Name] = ep
	}

	return result, nil
}

type K8sServicesNet struct {
	Namespace      string
	Name           string
	DeploymentName string

	ClusterIPs  []string
	ExternalIPs []string
	// format: <port>/protocol 9529/TCP, 80/TCP, 53/UDP, ...
	Port       []Port
	NodePort   []Port
	TargetPort []Port

	// NodePort, ...
	Type string

	Selector map[string]string
}

func (c *K8sClient) GetServiceNet(ns string) (map[string]*K8sServicesNet, error) {
	result := map[string]*K8sServicesNet{}

	services := c.CoreV1().Services(ns)

	list, err := services.List(context.Background(), metaV1ListOption)
	if err != nil {
		return result, err
	}
	for _, v := range list.Items {
		svc := &K8sServicesNet{
			Namespace:      v.GetNamespace(),
			Name:           v.GetName(),
			DeploymentName: "N/A",
			ClusterIPs:     []string{},
			ExternalIPs:    []string{},
			Port:           []Port{},
			NodePort:       []Port{},
			TargetPort:     []Port{},
			Selector:       map[string]string{},
		}

		svc.Type = string(v.Spec.Type)
		for k, v := range v.Spec.Selector {
			svc.Selector[k] = v
		}

		svc.ClusterIPs = append(svc.ClusterIPs, v.Spec.ClusterIPs...)
		svc.ExternalIPs = append(svc.ExternalIPs, v.Spec.ExternalIPs...)
		for _, v := range v.Spec.Ports {
			if v.Port != 0 {
				svc.Port = append(svc.Port,
					Port{
						Port:     uint32(v.Port),
						Protocol: string(v.Protocol),
					},
				)
			}
			if v.NodePort != 0 {
				svc.NodePort = append(svc.NodePort,
					Port{
						Port:     uint32(v.NodePort),
						Protocol: string(v.Protocol),
					},
				)
			}
			if v.TargetPort.IntVal != 0 {
				svc.TargetPort = append(svc.TargetPort,
					Port{
						Port:     uint32(v.TargetPort.IntVal),
						Protocol: string(v.Protocol),
					},
				)
			}
		}
		result[svc.Name] = svc
	}

	return result, nil
}

type K8sPodNet struct {
	Namespace   string
	ClusterName string
	Name        string

	DeploymentName string
	ServiceName    string
	NodeName       string

	Labels map[string]string
	PodIPs []string

	HostIPC     bool
	HostNetwork bool
	HostPID     bool
}

func (c *K8sClient) GetPodNet(ns string) (map[string][]*K8sPodNet, error) {
	result := map[string][]*K8sPodNet{}
	pods := c.CoreV1().Pods(ns)
	list, err := pods.List(context.Background(), metaV1ListOption)
	if err != nil {
		return result, err
	}
	for _, v := range list.Items {
		pod := &K8sPodNet{
			Namespace: v.GetNamespace(),
			Name:      v.GetName(),

			DeploymentName: "N/A",
			ServiceName:    "N/A",
			NodeName:       v.Spec.NodeName,

			Labels: map[string]string{},
			PodIPs: []string{},

			HostIPC:     v.Spec.HostIPC,
			HostNetwork: v.Spec.HostNetwork,
			HostPID:     v.Spec.HostPID,
		}

		for _, v := range v.Status.PodIPs {
			pod.PodIPs = append(pod.PodIPs, v.IP)
		}
		for k, v := range v.GetLabels() {
			pod.Labels[k] = v
		}

		for _, v := range pod.PodIPs {
			if _, ok := result[v]; !ok {
				result[v] = []*K8sPodNet{}
			}
			result[v] = append(result[v], pod)
		}
	}

	return result, nil
}

type K8sDeployment struct {
	Namespace string
	Name      string

	MatchLabels map[string]string
}

func (c *K8sClient) GetDeployment(ns string) (map[string]*K8sDeployment, error) {
	result := map[string]*K8sDeployment{}
	deploymentIface := c.AppsV1().Deployments(ns)
	list, err := deploymentIface.List(context.Background(), metaV1ListOption)
	if err != nil {
		return result, err
	}

	for _, v := range list.Items {
		result[v.Name] = &K8sDeployment{
			Namespace: v.GetNamespace(),
			Name:      v.GetName(),

			MatchLabels: func() map[string]string {
				r := map[string]string{}
				for k, v := range v.Spec.Selector.MatchLabels {
					r[k] = v
				}
				return r
			}(),
		}
	}
	return result, nil
}

func MatchLabel(selector, labels map[string]string) bool {
	if len(selector) > len(labels) {
		return false
	}

	for k, v := range selector {
		if v2, ok := labels[k]; !ok || v2 != v {
			return false
		}
	}

	return true
}
