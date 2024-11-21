// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package discovery

import (
	"fmt"
	"os"
	"strings"

	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type labelSelector struct {
	s labels.Selector
}

func newLabelSelector(matchLabels map[string]string, matchExpressions []metav1.LabelSelectorRequirement) *labelSelector {
	s := labels.Set(matchLabels).AsSelector()

	for _, expr := range matchExpressions {
		var op selection.Operator

		switch expr.Operator {
		case metav1.LabelSelectorOpIn:
			op = selection.In
		case metav1.LabelSelectorOpNotIn:
			op = selection.NotIn
		case metav1.LabelSelectorOpExists:
			op = selection.Exists
		case metav1.LabelSelectorOpDoesNotExist:
			op = selection.DoesNotExist
		default:
		}

		requirement, err := labels.NewRequirement(expr.Key, op, expr.Values)
		if err != nil {
			continue
		}

		s = s.Add(*requirement)
	}

	return &labelSelector{s}
}

func (ls *labelSelector) String() string {
	return ls.s.String()
}

func replaceLabelKey(s string) string {
	return strings.ReplaceAll(s, ".", "_")
}

func completePromConfig(config string, item *apicorev1.Pod) string {
	config = strings.ReplaceAll(config, "$IP", item.Status.PodIP)
	config = strings.ReplaceAll(config, "$NAMESPACE", item.Namespace)
	config = strings.ReplaceAll(config, "$PODNAME", item.Name)
	config = strings.ReplaceAll(config, "$NODENAME", item.Spec.NodeName)

	return config
}

func getLocalNodeName() (string, error) {
	var e string
	if os.Getenv("NODE_NAME") != "" {
		e = os.Getenv("NODE_NAME")
	}
	if os.Getenv("ENV_K8S_NODE_NAME") != "" {
		e = os.Getenv("ENV_K8S_NODE_NAME")
	}
	if e != "" {
		return e, nil
	}
	return "", fmt.Errorf("invalid ENV_K8S_NODE_NAME environment, cannot be empty")
}
