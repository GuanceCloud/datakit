// Package cli used to create k8s client and get some k8s info
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/GuanceCloud/cliutils/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	k8sclient "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/client"
)

var log = logger.DefaultSLogger("k8s")

func SetLogger() {
	log = logger.SLogger("k8s")
}

type K8sConfig struct {
	URL             string `toml:"url"`
	BearerToken     string `toml:"bearer_token"`
	BearerTokenPath string `toml:"bearer_token_path"`
	KubeConfig      string `toml:"kubeconfig"`

	// OperatorURL 非空时在已有 informer 基础上叠加使用 operator 已实现的接口（如 Pod）；未实现的接口仍走 informer
	OperatorURL string `toml:"operator_url"`

	WorkloadLabels      []string `toml:"workload_labels"`
	WorkloadLabelPrefix string   `toml:"workload_label_prefix"`
}

func NewK8sClientFromBearer(cfg K8sConfig, stopCh <-chan struct{}) (*K8sClient, error) {
	if cfg.URL == "" {
		cfg.URL = "https://kubernetes.default:443"
	}

	if cfg.BearerTokenPath == "" && cfg.BearerToken == "" {
		//nolint:gosec
		cfg.BearerTokenPath = "/run/secrets/kubernetes.io/serviceaccount/token"
	}

	var cli *K8sClient
	var err error
	if cfg.BearerTokenPath != "" {
		cli, err = NewK8sClientFromBearerToken(stopCh, cfg)
		if err != nil {
			return nil, err
		}
	} else {
		cli, err = NewK8sClientFromBearerTokenString(stopCh, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cli == nil {
		return nil, fmt.Errorf("new k8s client")
	}

	return cli, nil
}

// AttachOperator 在已有 K8sClient（informer）上挂载 operator 客户端；已由 operator 实现的接口（如 Pod）将走 operator，其余仍走 informer。
func AttachOperator(k *K8sClient, operatorURL string) error {
	if k == nil {
		return fmt.Errorf("K8sClient is nil")
	}
	opClient, err := k8sclient.NewKubernetesClientForOperator(operatorURL)
	if err != nil {
		return fmt.Errorf("create operator client: %w", err)
	}
	k.operatorClient = opClient
	return nil
}

// NewK8sClientFromOperator 仅通过 operator 创建客户端（无 informer），用于仅能访问 operator 的场景；当前仅支持 Pod。
func NewK8sClientFromOperator(cfg K8sConfig) (*K8sClient, error) {
	opClient, err := k8sclient.NewKubernetesClientForOperator(cfg.OperatorURL)
	if err != nil {
		return nil, fmt.Errorf("create operator client: %w", err)
	}

	k := &K8sClient{
		informer:            nil,
		operatorClient:      opClient,
		workloadLabels:      cfg.WorkloadLabels,
		workloadLabelPrefix: cfg.WorkloadLabelPrefix,
		Pods:                make(map[string][]*corev1.Pod),
	}
	return k, nil
}

func NewK8sClientFromKubeConfig(stopCh <-chan struct{}, kubeconfig K8sConfig) (*K8sClient, error) {
	if kubeconfig.KubeConfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig.KubeConfig = filepath.Join(home, ".kube", "config")
		} else {
			return nil, fmt.Errorf("unable to find home directory")
		}
	}

	// use kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %w", err)
	}

	if config, err := kubernetes.NewForConfig(config); err != nil {
		return nil, err
	} else {
		return &K8sClient{
			informer:            NewInformers(config, kubeconfig.OperatorURL == "", stopCh),
			workloadLabels:      kubeconfig.WorkloadLabels,
			workloadLabelPrefix: kubeconfig.WorkloadLabelPrefix,
		}, nil
	}
}

func NewK8sClientFromBearerToken(stopCh <-chan struct{}, cfg K8sConfig) (*K8sClient, error) {
	path := cfg.BearerTokenPath

	token, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	cfg.BearerToken = strings.TrimSpace(string(token))

	return NewK8sClientFromBearerTokenString(stopCh, cfg)
}

func NewK8sClientFromBearerTokenString(stopCh <-chan struct{}, cfg K8sConfig) (*K8sClient, error) {
	restConfig := &rest.Config{
		Host:        cfg.URL,
		BearerToken: cfg.BearerToken,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	if restConfig, err := kubernetes.NewForConfig(restConfig); err != nil {
		return nil, err
	} else {
		return &K8sClient{
			informer:            NewInformers(restConfig, cfg.OperatorURL == "", stopCh),
			workloadLabels:      cfg.WorkloadLabels,
			workloadLabelPrefix: cfg.WorkloadLabelPrefix,
		}, nil
	}
}
