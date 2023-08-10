// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransLabels(t *testing.T) {
	t.Run("trans-labels", func(t *testing.T) {
		in := map[string]string{
			"io.kubernetes.name":      "testing",
			"io.kubernetes.namespace": "testing-ns",
		}

		out := map[string]interface{}{
			"df_label":            `["io.kubernetes.name:testing","io.kubernetes.namespace:testing-ns"]`,
			"df_label_permission": "read_only",
			"df_label_source":     "datakit",
		}

		res := transLabels(in)

		assert.Equal(t, len(out), len(res))
		for k, v := range out {
			assert.Equal(t, v, out[k])
		}
	})

	t.Run("trans-labels-noe", func(t *testing.T) {
		in := map[string]string{}

		out := map[string]interface{}{
			"df_label":            `[]`,
			"df_label_permission": "read_only",
			"df_label_source":     "datakit",
		}

		res := transLabels(in)

		assert.Equal(t, len(out), len(res))
		for k, v := range out {
			assert.Equal(t, v, out[k])
		}
	})
}
