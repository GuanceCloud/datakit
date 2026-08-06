// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

const (
	kubernetesNodeLabelPlaceholderPrefix = "__k8s_node_label:"
	kubernetesNodeLabelResolveTimeout    = 5 * time.Second
)

type kubernetesNodeLabelGetter func(context.Context, string) (map[string]string, error)

func (c *Config) setupKubernetesNodeLabelHostTags() {
	ctx, cancel := context.WithTimeout(context.Background(), kubernetesNodeLabelResolveTimeout)
	defer cancel()

	c.resolveKubernetesNodeLabelHostTags(ctx, getKubernetesNodeLabels)
}

func getKubernetesNodeLabels(ctx context.Context, nodeName string) (map[string]string, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("load in-cluster configuration: %w", err)
	}

	client, err := corev1.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes core client: %w", err)
	}

	node, err := client.Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get node %q: %w", nodeName, err)
	}
	return node.Labels, nil
}

func (c *Config) resolveKubernetesNodeLabelHostTags(ctx context.Context, getLabels kubernetesNodeLabelGetter) {
	placeholders := make(map[string]string)
	for tagKey, tagValue := range c.GlobalHostTags {
		if strings.HasPrefix(tagValue, kubernetesNodeLabelPlaceholderPrefix) {
			placeholders[tagKey] = strings.TrimPrefix(tagValue, kubernetesNodeLabelPlaceholderPrefix)
			delete(c.GlobalHostTags, tagKey)
		}
	}

	if len(placeholders) == 0 {
		return
	}

	if !datakit.Docker || !IsKubernetes() {
		l.Warnf("Kubernetes node-label host tags require Kubernetes container mode, tags omitted")
		return
	}

	nodeName, err := GetLocalNodeName()
	if err != nil {
		l.Warnf("unable to resolve Kubernetes node-label host tags: local node name is unavailable, tags omitted")
		return
	}

	labels, err := getLabels(ctx, nodeName)
	if err != nil {
		l.Warnf("failed to resolve Kubernetes node-label host tags on node %q: %s, tags omitted", nodeName, err)
		return
	}

	for tagKey, labelKey := range placeholders {
		labelValue, ok := labels[labelKey]
		if !ok || labelValue == "" {
			l.Warnf("Kubernetes node label %q is missing or empty on node %q, global host tag %q omitted",
				labelKey, nodeName, tagKey)
			continue
		}
		c.GlobalHostTags[tagKey] = labelValue
	}
}
