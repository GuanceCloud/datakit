// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetesprometheus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabelSelectorToString(t *testing.T) {
	testcases := []struct {
		in  *metav1.LabelSelector
		out string
	}{
		{
			in: &metav1.LabelSelector{
				MatchLabels:      map[string]string{"app": "nginx"},
				MatchExpressions: nil,
			},
			out: "app=nginx",
		},
		{
			in: &metav1.LabelSelector{
				MatchLabels: nil,
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "environment",
						Operator: metav1.LabelSelectorOperator("In"),
						Values:   []string{"production", "testing"},
					},
					{
						Key:      "image",
						Operator: metav1.LabelSelectorOperator("NotIn"),
						Values:   []string{"nginx1", "nginx2"},
					},
				},
			},
			out: "environment in (production,testing),image notin (nginx1,nginx2)",
		},
		{
			in: &metav1.LabelSelector{
				MatchLabels: nil,
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "exists",
						Operator: metav1.LabelSelectorOperator("Exists"),
						Values:   nil,
					},
					{
						Key:      "no-exists",
						Operator: metav1.LabelSelectorOperator("DoesNotExist"),
						Values:   nil,
					},
				},
			},
			out: "exists,!no-exists",
		},
		{
			in: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{
						Key:      "environment",
						Operator: metav1.LabelSelectorOperator("In"),
						Values:   []string{"production", "testing"},
					},
					{
						Key:      "image",
						Operator: metav1.LabelSelectorOperator("NotIn"),
						Values:   []string{"nginx1", "nginx2"},
					},
					{
						Key:      "exists",
						Operator: metav1.LabelSelectorOperator("Exists"),
						Values:   nil,
					},
					{
						Key:      "no-exists",
						Operator: metav1.LabelSelectorOperator("DoesNotExist"),
						Values:   nil,
					},
				},
			},
			out: "app=nginx,environment in (production,testing),exists,image notin (nginx1,nginx2),!no-exists",
		},
	}

	for _, tc := range testcases {
		res := labelSelectorToString(tc.in)
		assert.Equal(t, tc.out, res)
	}
}
