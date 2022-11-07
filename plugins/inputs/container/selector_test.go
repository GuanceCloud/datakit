// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabelSelector(t *testing.T) {
	t.Run("to string", func(t *testing.T) {
		cases := []struct {
			matchLabels      map[string]string
			matchExpressions []metav1.LabelSelectorRequirement
			output           string
		}{
			{
				matchLabels:      map[string]string{"app": "nginx"},
				matchExpressions: nil,
				output:           "app=nginx",
			},
			{
				matchLabels: nil,
				matchExpressions: []metav1.LabelSelectorRequirement{
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
				output: "environment in (production,testing),image notin (nginx1,nginx2)",
			},
			{
				matchLabels: nil,
				matchExpressions: []metav1.LabelSelectorRequirement{
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
				output: "exists,!no-exists",
			},
			{
				matchLabels: map[string]string{"app": "nginx"},
				matchExpressions: []metav1.LabelSelectorRequirement{
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
				output: "app=nginx,environment in (production,testing),exists,image notin (nginx1,nginx2),!no-exists",
			},
		}
		for _, tc := range cases {
			s := newLabelSelector(tc.matchLabels, tc.matchExpressions)
			assert.Equal(t, tc.output, s.String())
		}
	})

	t.Run("to matches", func(t *testing.T) {
		cases := []struct {
			matchLabels      map[string]string
			matchExpressions []metav1.LabelSelectorRequirement

			targetLabels map[string]string
			matched      bool
		}{
			{
				matchLabels:      map[string]string{"app": "nginx"},
				matchExpressions: nil,

				targetLabels: map[string]string{"app": "nginx"},
				matched:      true,
			},
			{
				matchLabels: nil,
				matchExpressions: []metav1.LabelSelectorRequirement{
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

				targetLabels: map[string]string{"exists": "dummy"},
				matched:      true,
			},

			{
				matchLabels: nil,
				matchExpressions: []metav1.LabelSelectorRequirement{
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

				targetLabels: map[string]string{"exists": "dummy", "no-exists": "dummy"},
				matched:      false,
			},
		}

		for idx, tc := range cases {
			t.Run(fmt.Sprintf("%d", idx), func(t *testing.T) {
				s := newLabelSelector(tc.matchLabels, tc.matchExpressions)
				assert.Equal(t, tc.matched, s.Matches(tc.targetLabels))
			})
		}
	})
}
