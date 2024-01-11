// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildLabelsOption(t *testing.T) {
	t.Run("build-labels-option-for-non-metric", func(t *testing.T) {
		{
			enableLabelAsTags := true
			asTagKeys := []string{"app", "name"}
			customerKeys := []string{"sink-project", "name"}

			out := labelsOption{all: true, keys: nil}

			res := buildLabelsOptionForNonMetric(enableLabelAsTags, asTagKeys, customerKeys)
			assert.Equal(t, out, res)
		}
		{
			enableLabelAsTags := false
			asTagKeys := []string{"app", "name"}
			customerKeys := []string{"sink-project", "name"}

			out := labelsOption{all: false, keys: []string{"app", "name", "sink-project"}}

			res := buildLabelsOptionForNonMetric(enableLabelAsTags, asTagKeys, customerKeys)
			assert.Equal(t, out, res)
		}
		{
			enableLabelAsTags := false
			var asTagKeys []string
			var customerKeys []string

			out := labelsOption{all: false, keys: []string{}}

			res := buildLabelsOptionForNonMetric(enableLabelAsTags, asTagKeys, customerKeys)
			assert.Equal(t, out, res)
		}
	})

	t.Run("build-labels-option-for-metric", func(t *testing.T) {
		{
			asTagKeys := []string{"app", "name"}
			customerKeys := []string{"sink-project", "name"}

			out := labelsOption{all: false, keys: []string{"app", "name", "sink-project"}}

			res := buildLabelsOptionForMetric(asTagKeys, customerKeys)
			assert.Equal(t, out, res)
		}
		{
			asTagKeys := []string{""}
			customerKeys := []string{"sink-project", "name"}
			out := labelsOption{all: true}

			res := buildLabelsOptionForMetric(asTagKeys, customerKeys)
			assert.Equal(t, out, res)
		}
		{
			var asTagKeys []string
			customerKeys := []string{"sink-project"}
			out := labelsOption{all: false, keys: []string{"sink-project"}}

			res := buildLabelsOptionForMetric(asTagKeys, customerKeys)
			assert.Equal(t, out, res)
		}
	})
}
