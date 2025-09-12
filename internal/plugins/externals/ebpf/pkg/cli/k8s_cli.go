// Package cli used to create k8s client and get some k8s info
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var log = logger.DefaultSLogger("k8s")

func SetLogger() {
	log = logger.SLogger("k8s")
}

type K8sConfig struct {
	URL             string `toml:"url"`
	BearerToken     string `toml:"bearer_token"`
	BearerTokenPath string `toml:"bearer_token_path"`

	WorkloadLabels      []string `toml:"workload_labels"`
	WorkloadLabelPrefix string   `toml:"workload_label_prefix"`
}

func NewK8sClientFromBearer(cfg K8sConfig, stopCh <-chan struct{}) (*K8sClient, error) {
	var (
		k8sURL          = cfg.URL
		bearerTokenPath = cfg.BearerTokenPath
		bearerToken     = cfg.BearerToken
	)
	if k8sURL == "" {
		k8sURL = "https://kubernetes.default:443"
	}

	// net.LookupHost()

	if bearerTokenPath == "" && bearerToken == "" {
		//nolint:gosec
		bearerTokenPath = "/run/secrets/kubernetes.io/serviceaccount/token"
	}

	lbs := cfg.WorkloadLabels
	lbPrefix := cfg.WorkloadLabelPrefix

	var cli *K8sClient
	var err error
	if bearerTokenPath != "" {
		cli, err = NewK8sClientFromBearerToken(
			stopCh, k8sURL,
			bearerTokenPath, lbs, lbPrefix)
		if err != nil {
			return nil, err
		}
	} else {
		cli, err = NewK8sClientFromBearerTokenString(
			stopCh, k8sURL,
			bearerToken, lbs, lbPrefix)
		if err != nil {
			return nil, err
		}
	}

	if cli == nil {
		return nil, fmt.Errorf("new k8s client")
	}

	return cli, nil
}

func NewK8sClientFromKubeConfig(stopCh <-chan struct{}, kubeconfig string, lbs []string, lbPrefix string) (*K8sClient, error) {
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("unable to find home directory")
		}
	}

	// use kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %w", err)
	}

	// create k8s client
	return newK8sClient(config, stopCh, lbs, lbPrefix)
}

func NewK8sClientFromBearerToken(stopCh <-chan struct{}, baseURL, path string, lbs []string, lbPrefix string) (*K8sClient, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("invalid baseURL, cannot be empty")
	}

	token, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	return NewK8sClientFromBearerTokenString(
		stopCh,
		baseURL, strings.TrimSpace(string(token)),
		lbs, lbPrefix)
}

func NewK8sClientFromBearerTokenString(stopCh <-chan struct{}, baseURL, token string, lbs []string, lbPrefix string) (*K8sClient, error) {
	restConfig := &rest.Config{
		Host:        baseURL,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	return newK8sClient(restConfig, stopCh, lbs, lbPrefix)
}

func newK8sClient(restConfig *rest.Config, stopCh <-chan struct{}, lbs []string, lbPrefix string) (*K8sClient, error) {
	config, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	k := &K8sClient{
		Clientset:           config,
		informer:            NewInformaers(config, stopCh),
		workloadLabels:      lbs,
		workloadLablePrefix: lbPrefix,
	}

	return k, nil
}
