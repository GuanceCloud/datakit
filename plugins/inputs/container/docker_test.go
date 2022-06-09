// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package container

import (
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetPodAnnotationState(t *testing.T) {
	cases := []struct {
		name string
		in1  *types.Container
		in2  *podMeta
		out  podAnnotationStateType
	}{
		{
			in1: &types.Container{
				Labels: map[string]string{
					"io.kubernetes.pod.name":       "t_name",
					"io.kubernetes.container.name": "t_container_name",
				},
			},
			in2: &podMeta{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"datakit/logs": `
						    [
						        {
						            "disable": false,
						    	    "only_images": ["image:testing_image"]
						        }
						    ]
						`,
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "t_container_name",
								Image: "testing_image",
							},
						},
					},
				},
			},
			out: podAnnotationEnable,
		},
		{
			in1: &types.Container{
				Labels: map[string]string{
					"io.kubernetes.pod.name":       "t_name",
					"io.kubernetes.container.name": "t_container_name",
				},
			},
			in2: &podMeta{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"datakit/logs": `
						    [
						        {
						            "disable": false,
						    	    "only_images": ["image:invalid"]
						        }
						    ]
						`,
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "t_container_name",
								Image: "testing_image",
							},
						},
					},
				},
			},
			out: podAnnotationDisable,
		},
		{
			in1: &types.Container{},
			in2: &podMeta{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"datakit/logs": `
						    [
						        {
						            "disable": true,
						    	    "only_images": ["image:invalid"]
						        }
						    ]
						`,
						},
					},
				},
			},
			out: podAnnotationDisable,
		},
		{
			in1: &types.Container{},
			in2: &podMeta{
				Pod: &v1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"datakit/logs": `
						    [
						        {
						            "disable": false,
						    	    "only_images": []
						        }
						    ]
						`,
						},
					},
				},
			},
			out: podAnnotationEnable,
		},
		{
			in1: &types.Container{},
			in2: &podMeta{
				Pod: &v1.Pod{},
			},
			out: podAnnotationNil,
		},
		{
			in1: &types.Container{},
			in2: nil,
			out: podAnnotationNil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := getPodAnnotationState(tc.in1, tc.in2)
			assert.Equal(t, tc.out, res)
		})
	}
}
