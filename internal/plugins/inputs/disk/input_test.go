// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package disk

import (
	"os"
	"path/filepath"
	"runtime"
	T "testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

func TestCollect(t *T.T) {
	i := defaultInput()
	intervalMillSec := i.Interval.Milliseconds()
	var lastAlignTime int64

	for x := 0; x < 1; x++ {
		tn := time.Now()
		lastAlignTime = inputs.AlignTimeMillSec(tn, lastAlignTime, intervalMillSec)
		if err := i.collect(lastAlignTime * 1e6); err != nil {
			t.Error(err)
		}
		time.Sleep(time.Second * 1)
	}
	if len(i.collectCache) < 1 {
		t.Error("Failed to collect, no data returned")
	}
	tmap := map[string]bool{}
	for _, pt := range i.collectCache {
		tmap[pt.Time().String()] = true
	}
	if len(tmap) != 1 {
		t.Error("Need to clear collectCache.")
	}
}

func TestFilterUsage(t *T.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipped on ", runtime.GOOS)
	}

	t.Run(`all-ignored`, func(t *T.T) {
		ipt := defaultInput()

		tmpdir := filepath.Join(t.TempDir(), "proc")
		os.Setenv("HOST_PROC", tmpdir)
		t.Cleanup(func() {
			os.Unsetenv("HOST_PROC")
		})

		fakedir := filepath.Join(tmpdir, "self")
		require.NoError(t, os.MkdirAll(fakedir, os.ModePerm))

		// create fake /proc/self/mountinfo
		mountinfo, err := os.ReadFile("testdata/mountinfo.all-ignored")
		require.NoError(t, err)

		mountfile := filepath.Join(fakedir, "mountinfo")
		t.Logf("mountfile: %q", mountfile)
		require.NoError(t, os.WriteFile(mountfile, mountinfo, os.ModePerm))

		disks, partitions, err := ipt.diskStats.FilterUsage()
		require.NoError(t, err)
		assert.Nil(t, disks)
		assert.Nil(t, partitions)
	})

	t.Run("no-merge-on-device", func(t *T.T) {
		ipt := defaultInput()
		ipt.MergeOnDevice = false

		tmpdir := filepath.Join(t.TempDir(), "proc")
		os.Setenv("HOST_PROC", tmpdir)
		t.Cleanup(func() {
			os.Unsetenv("HOST_PROC")
		})

		fakedir := filepath.Join(tmpdir, "self")
		require.NoError(t, os.MkdirAll(fakedir, os.ModePerm))

		// create fake /proc/self/mountinfo
		mountinfo, err := os.ReadFile("testdata/mountinfo.merged")
		require.NoError(t, err)

		mountfile := filepath.Join(fakedir, "mountinfo")
		t.Logf("mountfile: %q", mountfile)
		require.NoError(t, os.WriteFile(mountfile, mountinfo, os.ModePerm))

		_, partitions, err := ipt.diskStats.FilterUsage()
		require.NoError(t, err)
		assert.Len(t, partitions, 2)
	})

	t.Run("merge-on-device", func(t *T.T) {
		ipt := defaultInput()

		tmpdir := filepath.Join(t.TempDir(), "proc")
		os.Setenv("HOST_PROC", tmpdir)
		t.Cleanup(func() {
			os.Unsetenv("HOST_PROC")
		})

		fakedir := filepath.Join(tmpdir, "self")
		require.NoError(t, os.MkdirAll(fakedir, os.ModePerm))

		// create fake /proc/self/mountinfo
		mountinfo, err := os.ReadFile("testdata/mountinfo.merged")
		require.NoError(t, err)

		mountfile := filepath.Join(fakedir, "mountinfo")
		t.Logf("mountfile: %q", mountfile)
		require.NoError(t, os.WriteFile(mountfile, mountinfo, os.ModePerm))

		_, partitions, err := ipt.diskStats.FilterUsage()
		require.NoError(t, err)
		assert.Len(t, partitions, 1)
	})

	t.Run("case2", func(t *T.T) {
		ipt := defaultInput()

		tmpdir := filepath.Join(t.TempDir(), "proc")
		os.Setenv("HOST_PROC", tmpdir)
		t.Cleanup(func() {
			os.Unsetenv("HOST_PROC")
		})

		fakedir := filepath.Join(tmpdir, "self")
		require.NoError(t, os.MkdirAll(fakedir, os.ModePerm))

		// create fake /proc/self/mountinfo
		mountinfo, err := os.ReadFile("testdata/mountinfo.2")
		require.NoError(t, err)

		mountfile := filepath.Join(fakedir, "mountinfo")
		t.Logf("mountfile: %q", mountfile)
		require.NoError(t, os.WriteFile(mountfile, mountinfo, os.ModePerm))

		disks, partitions, err := ipt.diskStats.FilterUsage()
		require.NoError(t, err)

		t.Logf("usage info: %+#v", disks)

		for idx, part := range partitions {
			t.Logf("part[%d]: %+#v", idx, part)
		}
	})
}
