// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"os"
	"path/filepath"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func Test_adjustPointTime(t *T.T) {
	t.Run(`adjust`, func(t *T.T) {
		recordTime := time.Now().Round(0).Add(-time.Hour) // 1h ago
		r := point.NewRander(point.WithRandTime(recordTime))
		pts := r.Rand(2) // all points with same time

		when := time.Now().Round(0).Add(-time.Minute)

		t.Logf("before pts: %s", pts[0].Pretty())

		pts = adjustPointTime(when, pts)

		t.Logf("after pts: %s", pts[0].Pretty())

		for _, pt := range pts {
			assert.Equal(t, when, pt.Time())
		}
	})

	t.Run(`adjust-to-older-time`, func(t *T.T) {
		recordTime := time.Now().Round(0).Add(-time.Hour) // 1h ago
		r := point.NewRander(point.WithRandTime(recordTime))
		pts := r.Rand(2) // all points with same time

		when := time.Now().Round(0).Add(-2 * time.Hour)

		t.Logf("before pts: %s", pts[0].Pretty())

		pts = adjustPointTime(when, pts)

		t.Logf("after pts: %s", pts[0].Pretty())

		for _, pt := range pts {
			assert.Equal(t, when, pt.Time())
		}
	})

	t.Run(`scattered-points`, func(t *T.T) {
		when := time.Now().Round(0)

		pts := []*point.Point{
			// 2 points with interval 1h
			point.NewPoint("p1", nil, point.WithTime(when.Add(-2*time.Hour))),
			point.NewPoint("p2", nil, point.WithTime(when.Add(-time.Hour))),
		}

		t.Logf("before pts: %s", pts[0].Pretty())
		t.Logf("before pts: %s", pts[1].Pretty())

		pts = adjustPointTime(when, pts)

		t.Logf("after pts: %s", pts[0].Pretty())
		t.Logf("after pts: %s", pts[1].Pretty())

		assert.Equal(t, when, pts[1].Time())
		assert.Equal(t, when.Add(-time.Hour), pts[0].Time())
	})
}

func Test_setupUploader(t *T.T) {
	origCfg := config.Cfg
	origInstallDir := datakit.InstallDir
	origImportURLs := append([]string(nil), *flagImportDatawayURL...)

	t.Cleanup(func() {
		config.Cfg = origCfg
		datakit.SetupWorkDir(origInstallDir)
		*flagImportDatawayURL = origImportURLs
	})

	t.Run("use dataway settings from main config", func(t *T.T) {
		tmpDir := t.TempDir()
		datakit.SetupWorkDir(tmpDir)
		require.NoError(t, os.MkdirAll(filepath.Dir(datakit.MainConfPath), 0o755))

		config.Cfg = config.DefaultConfig()
		*flagImportDatawayURL = nil

		conf := `
[dataway]
  urls = ["https://config.example.com?token=config-token"]
  max_raw_body_size = 2097152
  gzip = false
`
		require.NoError(t, os.WriteFile(datakit.MainConfPath, []byte(conf), 0o644))

		u, err := setupUploader()
		require.NoError(t, err)

		impl, ok := u.(*uploaderImpl)
		require.True(t, ok)
		assert.Equal(t, []string{"https://config.example.com?token=config-token"}, impl.dw.URLs)
		assert.Equal(t, 2*(1<<20), impl.dw.MaxRawBodySize)
		assert.False(t, impl.dw.GZip)
	})

	t.Run("override urls but keep dataway limits from main config", func(t *T.T) {
		tmpDir := t.TempDir()
		datakit.SetupWorkDir(tmpDir)
		require.NoError(t, os.MkdirAll(filepath.Dir(datakit.MainConfPath), 0o755))

		config.Cfg = config.DefaultConfig()
		*flagImportDatawayURL = []string{"https://override.example.com?token=override-token"}

		conf := `
[dataway]
  urls = ["https://config.example.com?token=config-token"]
  max_raw_body_size = 2097152
  gzip = false
`
		require.NoError(t, os.WriteFile(datakit.MainConfPath, []byte(conf), 0o644))

		u, err := setupUploader()
		require.NoError(t, err)

		impl, ok := u.(*uploaderImpl)
		require.True(t, ok)
		assert.Equal(t, []string{"https://override.example.com?token=override-token"}, impl.dw.URLs)
		assert.Equal(t, 2*(1<<20), impl.dw.MaxRawBodySize)
		assert.False(t, impl.dw.GZip)
	})

	t.Run("allow import with explicit dataway when main config missing", func(t *T.T) {
		tmpDir := t.TempDir()
		datakit.SetupWorkDir(tmpDir)

		config.Cfg = config.DefaultConfig()
		*flagImportDatawayURL = []string{"https://override.example.com?token=override-token"}

		u, err := setupUploader()
		require.NoError(t, err)

		impl, ok := u.(*uploaderImpl)
		require.True(t, ok)
		assert.Equal(t, []string{"https://override.example.com?token=override-token"}, impl.dw.URLs)
		assert.Equal(t, dataway.DefaultMaxRawBodySize, impl.dw.MaxRawBodySize)
	})
}
