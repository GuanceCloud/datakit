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

func MockWrite(category string, pts []sinkcommon.ISinkPoint) error {
	fmt.Println("write points", category, pts)

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

	assert.True(t, sender.Stat.SuccessCount > 0)

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
	})

	t.Run("exit when receive global exit", func(t *testing.T) {
		sender, err := NewSender(&Option{Cache: true, CacheDir: testDir, Write: MockWrite})
		assert.NoError(t, err)
		go datakit.Exit.Close()
		sender.Wait()
	})
}
