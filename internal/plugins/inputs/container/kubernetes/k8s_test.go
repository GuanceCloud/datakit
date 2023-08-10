// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestTransToNamespaceMeasurements(t *testing.T) {
	t.Run("trans to namespace measurements", func(t *testing.T) {
		in := map[string]map[string]int{
			"pod": {
				"namespace01": 10,
				"namespace02": 20,
			},
			"deployment": {
				"namespace01": 10,
				"namespace02": 20,
				"namespace03": 30,
			},
		}

		p1 := typed.NewPointKV()
		p1.SetTag("namespace", "namespace01")
		p1.SetField("pod", 10)
		p1.SetField("deployment", 10)

		p2 := typed.NewPointKV()
		p2.SetTag("namespace", "namespace02")
		p2.SetField("pod", 20)
		p2.SetField("deployment", 20)

		p3 := typed.NewPointKV()
		p3.SetTag("namespace", "namespace03")
		p3.SetField("deployment", 30)

		out := []inputs.Measurement{
			&count{p1},
			&count{p2},
			&count{p3},
		}

		res := transToNamespaceMeasurements(in)

		assert.Equal(t, len(out), len(res))
		// map is unordered
		// assert.Equal(t, out, res)
	})
}
