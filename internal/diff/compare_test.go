// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestCompare(t *testing.T) {
	var int1 int64 = 123
	var int2 int64 = 456

	cases := []struct {
		inOldVal interface{}
		inNewVal interface{}
		outEqual bool
	}{
		{
			inOldVal: apicorev1.ResourceRequirements{
				Limits: apicorev1.ResourceList{
					"cpu": resource.MustParse("100Mi"),
				},
			},
			inNewVal: apicorev1.ResourceRequirements{
				Limits: apicorev1.ResourceList{
					"cpu": resource.MustParse("200Mi"),
				},
			},
			outEqual: false,
		},
		{
			inOldVal: &int1,
			inNewVal: &int2,
			outEqual: false,
		},
		{
			inOldVal: nil,
			inNewVal: nil,
			outEqual: true,
		},
	}

	for _, tc := range cases {
		equal, difftext := Compare(tc.inOldVal, tc.inNewVal)
		assert.Equal(t, tc.outEqual, equal)

		t.Log(difftext)
	}
}
