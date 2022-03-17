// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import "testing"

// go test -v -timeout 30s -run ^TestCollect$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject
func TestCollect(t *testing.T) {
	ho := DefaultHostObject()
	pts, err := ho.Collect()
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(pts)

	if len(pts) == 0 {
		t.Error("collect empty!")
		return
	}

	for category := range pts {
		if category != "/v1/write/object" {
			t.Errorf("category not object: %s", category)
			return
		}
	}

	t.Log("TestCollect succeeded!")
}
