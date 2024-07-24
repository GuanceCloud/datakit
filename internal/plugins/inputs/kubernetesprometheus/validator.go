// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"k8s.io/apimachinery/pkg/labels"
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

func buildSelector(m map[string]string) string {
	return labels.Set(m).String()
}
