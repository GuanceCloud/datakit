// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sender

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cache"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

var sinkName = "dataway"

func MockWrite(category string, pts []sinkcommon.ISinkPoint) error {
	fmt.Println("write points", category, pts)

	return nil
}

func MockMetricWrite(category string, pts []sinkcommon.ISinkPoint) error {
	FeedMetric(&SinkMetric{
		Name:      sinkName,
		IsSuccess: true,
	})
	return nil
}

func MockWriteFailed(category string, pts []sinkcommon.ISinkPoint) error {
	fmt.Println("write points", category, pts)

	return fmt.Errorf("failed")
}

func TestSender(t *testing.T) {
	testDir, err := ioutil.TempDir("./", "__test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)
	sender, err := NewSender(&Option{Write: MockWrite})
	assert.NoError(t, err)
	p, _ := lp.MakeLineProtoPoint("metric_name", map[string]string{"t1": "t1"}, map[string]interface{}{"f": "f"}, nil)

	err = sender.Write("metric", WrapPoint([]*influxdb.Point{p}))
	assert.NoError(t, err)

	err = sender.Wait()
	assert.NoError(t, err)

	assert.True(t, len(sender.Stat) > 0)

	t.Run("cache data when failed", func(t *testing.T) {
		sender, err := NewSender(&Option{Cache: true, CacheDir: testDir, FlushCacheInterval: time.Second, Write: MockWriteFailed})
		assert.NoError(t, err)
		isCached := false
		err = sender.Write("metric", WrapPoint([]*influxdb.Point{p}))

		assert.NoError(t, err)

		time.Sleep(2 * time.Second)
		cache.ForEach(cacheBucket, func(key, value []byte) error {
			isCached = true
			return nil
		}, false)
		sender.Stop()
		sender.Wait()

		assert.True(t, isCached)

		t.Run("should flush cache when sender run", func(t *testing.T) {
			sender, err := NewSender(&Option{Cache: true, CacheDir: testDir, FlushCacheInterval: time.Second, Write: MockWrite})
			assert.NoError(t, err)
			time.Sleep(2 * time.Second)
			cacheInfo, err := cache.GetInfo()
			assert.NoError(t, err)
			assert.EqualValues(t, 1, cacheInfo.CacheCount)
			assert.EqualValues(t, 1, cacheInfo.FlushedCount)

			sender.Stop()
			sender.Wait()
		})
	})

	t.Run("update sink metric", func(t *testing.T) {
		sender, err := NewSender(&Option{Write: MockMetricWrite})
		assert.NoError(t, err)

		err = sender.Write("metric", WrapPoint([]*influxdb.Point{p}))
		sender.Wait()
		assert.NoError(t, err)

		time.Sleep(1 * time.Second)
		stat := GetStat()
		fmt.Println(stat)
		sinkMetric, ok := stat[sinkName]
		if !ok {
			t.Fatalf("should feed %s ", sinkName)
		}
		assert.True(t, ok)
		assert.EqualValues(t, 1, sinkMetric.Count)
	})

	t.Run("exit when receive global exit", func(t *testing.T) {
		sender, err := NewSender(&Option{Cache: true, CacheDir: testDir, Write: MockWrite})
		assert.NoError(t, err)
		go datakit.Exit.Close()
		sender.Wait()
	})
}
