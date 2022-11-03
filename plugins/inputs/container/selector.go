// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
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

func (ls *labelSelector) Matches(targetLables map[string]string) (exist bool) {
	return ls.s.Matches(labels.Set(targetLables))
}
