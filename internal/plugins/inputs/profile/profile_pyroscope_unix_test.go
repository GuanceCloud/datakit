// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows && !arm && !386
// +build !windows,!arm,!386

package profile

import (
	"os"
	"testing"

	"github.com/GuanceCloud/cliutils"
	"github.com/pyroscope-io/pyroscope/pkg/storage"
	"github.com/pyroscope-io/pyroscope/pkg/storage/segment"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

// 检查是不是开发机，如果不是开发机，则直接退出。开发机上需要定义 LOCAL_UNIT_TEST 环境变量。
func checkDevHost() bool {
	if envs := os.Getenv("LOCAL_UNIT_TEST"); envs == "" {
		return false
	}
	return true
}

// go test -v -timeout 30s -run ^TestPyroscopeRun$ gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs/profile
func TestPyroscopeRun(t *testing.T) {
	if !checkDevHost() {
		return
	}

	pyro := pyroscopeOpts{
		URL: "0.0.0.0:4040",
	}

	config.Cfg.Dataway = &dataway.Dataway{URLs: []string{"http://<GATEWAY>?token=<TOKEN>"}}

	err := config.Cfg.SetupDataway()
	if err != nil {
		panic(err)
	}
	ipt := &Input{
		semStop: cliutils.NewSem(),
	}
	err = pyro.run(ipt)
	if err != nil {
		panic(err)
	}
}

// go test -v -timeout 30s -run ^Test_getReportCacheKeyName$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/profile
func Test_getReportCacheKeyName(t *testing.T) {
	cases := []struct {
		name string
		out  string
	}{
		{
			name: "myNodeService.cpu",
			out:  "myNodeService{}nodespy",
		},
		{
			name: "myNodeService.inuse_objects",
			out:  "myNodeService{}nodespy",
		},
		{
			name: "myNodeService.inuse_space",
			out:  "myNodeService{}nodespy",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &storage.PutInput{SpyName: nodeSpyName}
			in.Key = segment.NewKey(map[string]string{"__name__": tc.name})

			out := getReportCacheKeyName(in)
			require.Equal(t, tc.out, out)
		})
	}
}

// go test -v -timeout 30s -run ^Test_LoadAndStore$ gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/profile
func Test_LoadAndStore(t *testing.T) {
	cases := []struct {
		name     string
		in       *storage.PutInput
		inCache  map[string]interface{}
		out      *cacheDetail
		outCache *cacheDetail
	}{
		{
			name: "not found",
			in: &storage.PutInput{
				Units: "samples",
			},
			outCache: &cacheDetail{
				CPU: &storage.PutInput{
					Units: "samples",
				},
			},
		},

		{
			name: "impossible",
			in:   &storage.PutInput{},
			inCache: map[string]interface{}{
				"impossible": 1,
			},
			outCache: &cacheDetail{},
		},

		{
			name: "new cpu",
			in: &storage.PutInput{
				Units: "samples",
			},
			outCache: &cacheDetail{
				CPU: &storage.PutInput{
					Units: "samples",
				},
			},
		},

		{
			name: "has cpu, new inuse_objects",
			in: &storage.PutInput{
				Units: "objects",
			},
			inCache: map[string]interface{}{
				"has cpu, new inuse_objects": &cacheDetail{
					CPU: &storage.PutInput{
						Units: "samples",
					},
				},
			},
			outCache: &cacheDetail{
				CPU: &storage.PutInput{
					Units: "samples",
				},
				InuseObjects: &storage.PutInput{
					Units: "objects",
				},
			},
		},

		{
			name: "has cpu and inuse_objects, new inuse_space",
			in: &storage.PutInput{
				Units: "bytes",
			},
			inCache: map[string]interface{}{
				"has cpu and inuse_objects, new inuse_space": &cacheDetail{
					CPU: &storage.PutInput{
						Units: "samples",
					},
					InuseObjects: &storage.PutInput{
						Units: "objects",
					},
				},
			},
			out: &cacheDetail{
				CPU: &storage.PutInput{
					Units: "samples",
				},
				InuseObjects: &storage.PutInput{
					Units: "objects",
				},
				InuseSpace: &storage.PutInput{
					Units: "bytes",
				},
			},
			outCache: &cacheDetail{
				CPU: &storage.PutInput{
					Units: "samples",
				},
				InuseObjects: &storage.PutInput{
					Units: "objects",
				},
				InuseSpace: &storage.PutInput{
					Units: "bytes",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			report := &pyroscopeDatakitReport{pyrs: &pyroscopeOpts{}}

			// initialize data cache.
			if tc.inCache != nil {
				for k, v := range tc.inCache {
					report.pyrs.cacheData.Store(k, v)
				}
			}

			out := report.LoadAndStore(tc.name, tc.in)

			// compare out.
			require.Equal(t, tc.out, out)

			// compare cache.
			{
				outCache, ok := report.pyrs.cacheData.Load(tc.name)
				require.True(t, ok)

				val, ok := outCache.(*cacheDetail)
				require.True(t, ok)

				require.Equal(t, tc.outCache, val)
			}
		})
	}
}
