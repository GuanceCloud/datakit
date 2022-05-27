// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package hostobject

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

// go test -v -timeout 30s -run ^TestCollect$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/hostobject
func TestCollect(t *testing.T) {
	ipt := DefaultHostObject()

	// ipt.OnlyPhysicalDevice = true

	ipt.isTestMode = true
	if err := ipt.doCollect(); err != nil {
		t.Error(err)
	}

	var pts []*io.Point
	if pt, err := ipt.collectData.LineProto(); err != nil {
		t.Error(err)
	} else {
		pts = append(pts, pt)
	}

	mpts := make(map[string][]*io.Point)
	mpts[datakit.Object] = pts

	if len(mpts) == 0 {
		t.Error("collect empty!")
		return
	}

	for category := range mpts {
		if category != "/v1/write/object" {
			t.Errorf("category not object: %s", category)
			return
		}
	}

	t.Log("TestCollect succeeded!")
}
