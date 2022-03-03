package sender

import (
	"fmt"
	"io/ioutil"
	"testing"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/stretchr/testify/assert"
	lp "gitlab.jiagouyun.com/cloudcare-tools/cliutils/lineproto"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/cache"
)

type MockWriter struct {
	isFailed bool
}

func (w *MockWriter) Write(category string, pts []*influxdb.Point) error {
	fmt.Println("write points", category, pts)

	if w.isFailed {
		return fmt.Errorf("failed")
	}
	return nil
}

func TestSender(t *testing.T) {
	testDir, err := ioutil.TempDir("./", "__test")
	if err != nil {
		t.Fatal(err)
	}
	writer := &MockWriter{}
	sinkerInstance := &Sinker{map[string]Writer{"metric": writer}}
	sender := NewSender(sinkerInstance, nil)
	p, _ := lp.MakeLineProtoPoint("metric_name", map[string]string{"t1": "t1"}, map[string]interface{}{"f": "f"}, nil)

	err = sender.Write("metric", []*influxdb.Point{p})
	assert.NoError(t, err)

	err = sender.Wait()
	assert.NoError(t, err)

	assert.True(t, sender.Stat.successCount > 0)

	t.Run("cache data when failed", func(t *testing.T) {
		sender := NewSender(sinkerInstance, &Option{Cache: true, CacheDir: testDir})
		isCalled := false
		writer.isFailed = true
		err := sender.Write("metric", []*influxdb.Point{p})

		assert.NoError(t, err)

		sender.Wait()

		fmt.Println(cache.Info())

		assert.True(t, isCalled)
	})
}

func TestPoints(t *testing.T) {
	p, _ := lp.MakeLineProtoPoint("metric_name", map[string]string{"t1": "t1"}, map[string]interface{}{"f": "f"}, nil)
	fmt.Println(p.String())
	pts, err := lp.ParsePoints([]byte(p.String()+"\n"+p.String()), nil)

	fmt.Println(pts, len(pts), err)
}
