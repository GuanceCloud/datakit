package tidb

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func TestMain(t *testing.T) {
	io.TestOutput()
	var ti = TiDB{
		PDServerURL: []string{"http://127.0.0.1:2379/pd/api/v1/stores"},
		Interval:    "5s",
	}
	ti.Run()
}

func TestKiB(t *testing.T) {
	t.Log(toKiB("66.66GiB"))
	t.Log(toKiB("55.55MiB"))
	t.Log(toKiB("10.93KiB"))
}
