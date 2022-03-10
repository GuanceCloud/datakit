package sinkinfluxdb

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	// this is important because of the bug in go mod
	_ "github.com/influxdata/influxdb1-client"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

//------------------------------------------------------------------------------

// how to use influxdb v2 SDK:
// https://github.com/influxdata/influxdb1-client/blob/master/v2/example_test.go

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
				"id":        "influxdb_1",
				"addr":      "http://10.200.7.21:8086",
				"precision": "ns",
				"database":  "db0",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			si := &SinkInfluxDB{}
			err := si.LoadConfig(tc.in)
			assert.Equal(t, tc.expectLoadConfigError, err)

			pts := getTestPoints(t, 1000, 42)
			var newPts []sinkcommon.ISinkPoint
			for _, v := range pts {
				newPts = append(newPts, sinkcommon.ISinkPoint(v))
			}
			err = si.Write(newPts)
			assert.Equal(t, tc.expectWriteError, err)
		})
	}
}

//------------------------------------------------------------------------------

type testPoint struct {
	*client.Point
}

var _ sinkcommon.ISinkPoint = new(testPoint)

func (p *testPoint) ToPoint() *client.Point {
	return p.Point
}

func getTestPoints(t *testing.T, sampleSize int, seed int64) []*testPoint {
	t.Helper()

	rand.Seed(seed)

	var pts []*testPoint
	for i := 0; i < sampleSize; i++ {
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

		pt, err := client.NewPoint(
			"cpu_usage",
			tags,
			fields,
			time.Now(),
		)
		assert.NoError(t, err, fmt.Sprintf("client.NewPoint failed: %v", err))
		pts = append(pts, &testPoint{pt})
	}
	return pts
}
