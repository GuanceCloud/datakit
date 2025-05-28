// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package phpfpm

import (
	"testing"
	"time"

	pfpm "github.com/hipages/php-fpm_exporter/phpfpm"
)

// testLogger 适配 testing.T 到 pfpm.logger 接口
type testLogger struct {
	t *testing.T
}

func (tl *testLogger) Info(args ...interface{}) {
	tl.t.Log(args...)
}

func (tl *testLogger) Infof(format string, args ...interface{}) {
	tl.t.Logf(format, args...)
}

func (tl *testLogger) Debug(args ...interface{}) {
	tl.t.Log(args...)
}

func (tl *testLogger) Debugf(format string, args ...interface{}) {
	tl.t.Logf(format, args...)
}

func (tl *testLogger) Error(args ...interface{}) {
	tl.t.Log(args...)
}

func (tl *testLogger) Errorf(format string, args ...interface{}) {
	tl.t.Logf(format, args...)
}

func TestCollect(t *testing.T) {
	pfpm.SetLogger(&testLogger{t: t})

	i := defaultInput()
	i.StatusURL = "http://localhost/status"
	i.UseFastCGI = false
	i.Tags = map[string]string{
		"some": "xxx",
	}

	if err := i.collect(); err != nil {
		t.Errorf("collect failed: %v", err)
	}
	time.Sleep(time.Second * 1)

	if len(i.collectCache) < 1 {
		t.Log("Failed to collect, no data returned")
	}

	if len(i.collectCache) > 0 {
		tmap := map[string]bool{}
		for _, pt := range i.collectCache {
			tmap[pt.Time().String()] = true
		}
		if len(tmap) != 1 {
			t.Error("Multiple timestamps in collectCache, expected one")
		}
	}

	// fmt.Fprintln(os.Stderr, i.collectCache)
}
