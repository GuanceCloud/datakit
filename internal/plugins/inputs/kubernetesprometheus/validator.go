// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/kubernetes/pkg/labels"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/selection"
)

type resourceValidator struct {
	namespaces []string
	selector   labels.Selector
}

func newResourceValidator(namespaces []string, selector string) (*resourceValidator, error) {
	validator := &resourceValidator{
		namespaces: namespaces,
		selector:   nil,
	}
	if selector != "" {
		s, err := labels.Parse(selector)
		if err != nil {
			return nil, err
		}
		validator.selector = s
	}
	return validator, nil
}

func (v *resourceValidator) Matches(namespace string, targetLables map[string]string) (pass bool) {
	if len(v.namespaces) != 0 {
		matched := false
		for _, ns := range v.namespaces {
			if ns == namespace {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if v.selector != nil {
		return v.selector.Matches(labels.Set(targetLables))
	}
	return true
}

func selectorToString(m map[string]string) string {
	return labels.Set(m).String()
}

func labelSelectorToString(selector *metav1.LabelSelector) string {
	s := labels.Set(selector.MatchLabels).AsSelector()

	for _, expr := range selector.MatchExpressions {
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
			// unreachable
		}

		requirement, err := labels.NewRequirement(expr.Key, op, expr.Values)
		if err != nil {
			continue
		}

		s = s.Add(*requirement)
	}

	return s.String()
}
