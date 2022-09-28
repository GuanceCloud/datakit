// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package point

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	mrand "math/rand"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

func doRandomPoints(count int) ([]*Point, error) {
	if count <= 0 {
		return nil, nil
	}

	buf := make([]byte, 30)

	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}

	var pts []*Point

	option := defaultPointOption()
	option.Category = datakit.Logging

	for i := 0; i < count; i++ {
		if pt, err := NewPoint("mock_random_point",
			map[string]string{
				base64.StdEncoding.EncodeToString(buf): base64.StdEncoding.EncodeToString(buf[1:]),
			},
			map[string]interface{}{
				base64.StdEncoding.EncodeToString(buf[2:]): base64.StdEncoding.EncodeToString(buf[3:]),
				base64.StdEncoding.EncodeToString(buf[3:]): mrand.Int(),         //nolint:gosec
				base64.StdEncoding.EncodeToString(buf[4:]): mrand.NormFloat64(), //nolint:gosec
			},
			option,
		); err != nil {
			log.Fatal(err.Error())
		} else {
			pts = append(pts, pt)
		}
	}

	return pts, nil
}

func RandPoints(count int) []*Point {
	if pts, err := doRandomPoints(count); err != nil {
		panic(fmt.Sprintf("doRandomPoints failed: %s", err))
	} else {
		return pts
	}
}
