// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package proxy

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/testutils"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestProxyServer(t *testing.T) {
	var pts []*io.Point
	for i := 0; i < 100; i++ {
		pts = append(pts, &io.Point{Point: testutils.RandPoint("test_point", 10, 30)})
	}

	for _, pt := range pts {
		log.Info(pt.String())
	}
}
