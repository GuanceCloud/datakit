// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/container/typed"
	apicorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestComposePodMetric(t *testing.T) {
	t.Run("compose pod metric", func(t *testing.T) {
		in := &apicorev1.PodList{
			Items: []apicorev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-name-testing",
						Namespace: "pod-namespace-testing",
						UID:       "pod-uid-testing",
					},
					Spec: apicorev1.PodSpec{
						NodeName: "node-name-testing",
					},
				},
			},
		}

		out := typed.NewPointKV(podMetricMeasurement)
		out.SetTag("uid", "pod-uid-testing")
		out.SetTag("pod", "pod-name-testing")
		out.SetTag("pod_name", "pod-name-testing")
		out.SetTag("node_name", "node-name-testing")
		out.SetTag("namespace", "pod-namespace-testing")
		out.SetField("ready", 0)
		out.SetField("restarts", 0)

		outPts := pointKVs{out}

		p := &podMetadata{
			parent: &pod{client: nil, counter: make(map[string]map[string]int)},
			list:   in,
		}

		conf := Config{}
		res := p.newMetric(&conf)

		assert.Equal(t, outPts, res)
	})
}
