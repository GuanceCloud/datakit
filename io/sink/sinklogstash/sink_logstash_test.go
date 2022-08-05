// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sinklogstash

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/point"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

//------------------------------------------------------------------------------
type Message struct {
	Data string `json:"data"`
}

type Log struct {
	Action  string    `json:"action"`
	Time    time.Time `json:"time"`
	Message Message   `json:"message"`
}

// go test -v -timeout 30s -run ^TestAll$ gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinklogstash
func TestAll(t *testing.T) {
	if !checkDevHost() {
		return
	}

	cases := []struct {
		name                  string
		in                    map[string]interface{}
		expectLoadConfigError error
		expectWriteError      error
	}{
		{
			name: "required",
			in: map[string]interface{}{
				"host":         "10.200.7.21:8080",
				"protocol":     "http",
				"request_path": "/twitter/tweet/1",
				"timeout":      "5s",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			si := &SinkLogstash{}
			err := si.LoadConfig(tc.in)
			assert.Equal(t, tc.expectLoadConfigError, err)

			pts := getTestPoints(t, 41)

			err = si.Write("", pts)
			assert.Equal(t, tc.expectWriteError, err)
		})
	}
}

//------------------------------------------------------------------------------

func getTestPoints(t *testing.T, seed int64) []*point.Point {
	t.Helper()

	rand.Seed(seed)

	mms := []string{"mm1", "mm2", "mm3", "mm4"}
	var pts []*point.Point
	for i := 0; i < 4; i++ {
		regions := []string{"us-west1", "us-west2", "us-west3", "us-east1"}
		tags := map[string]string{
			"cpu":    "cpu-total",
			"host":   fmt.Sprintf("host%d", rand.Intn(1000)),
			"region": regions[rand.Intn(len(regions))],
		}

		idle := rand.Float64() * 100.0
		fields := map[string]interface{}{
			"idle": idle,
			"busy": 100.0 - idle,
		}
		pt, err := point.NewPoint(mms[i], tags, fields, nil)
		if err != nil {
			return nil
		}

		pts = append(pts, pt)
	}

	return pts
}
