// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package dk

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils"
	pprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/storage"
	"google.golang.org/protobuf/proto"
)

func TestNormalizeSelfProfilingConfig(t *testing.T) {
	cfg := normalizeSelfProfilingConfig(SelfProfilingConfig{
		Enabled:  true,
		CPUCores: -1,
		MEMMaxMB: -1,
	})

	assert.Equal(t, defaultSelfProfileInterval, cfg.Interval)
	assert.Equal(t, defaultSelfProfileDuration, cfg.Duration)
	assert.Equal(t, defaultSelfProfileEmergencyDuration, cfg.EmergencyDuration)
	assert.Equal(t, defaultSelfProfileCooldown, cfg.Cooldown)
	assert.Equal(t, defaultSelfProfileRecentPoints, cfg.RecentPoints)
	assert.Equal(t, defaultSelfProfileTypes, cfg.EnabledTypes)
	assert.Equal(t, defaultSelfProfileCPUCores, cfg.CPUCores)
	assert.Equal(t, int64(defaultSelfProfileMEMMaxMB), cfg.MEMMaxMB)
	assert.Equal(t, defaultSelfProfileCachePath, cfg.CachePath)
	assert.Equal(t, defaultSelfProfileCacheCapacityMB, cfg.CacheCapacityMB)
	assert.Equal(t, defaultSelfProfileSendTimeout, cfg.SendTimeout)
	assert.Equal(t, defaultSelfProfileSendRetryCount, cfg.SendRetryCount)

	cfg = normalizeSelfProfilingConfig(SelfProfilingConfig{Enabled: true})
	assert.Equal(t, float64(0), cfg.CPUCores)
	assert.Equal(t, int64(0), cfg.MEMMaxMB)
}

func TestSelfProfilerTriggerDecision(t *testing.T) {
	t.Run("thresholds-disabled", func(t *testing.T) {
		p := newTestSelfProfiler(SelfProfilingConfig{})
		p.samples = append(p.samples, &resourceSample{
			HasCPU:          true,
			CPUUsagePercent: 100,
			MEMUsageMB:      1024,
			MEMUsagePercent: 90,
		})

		decision := p.triggerDecision(&resourceSample{
			HasCPU:          true,
			CPUUsagePercent: 100,
			MEMUsageMB:      1024,
			MEMUsagePercent: 90,
		})
		assert.Empty(t, decision.reasons)
	})

	t.Run("normal-threshold-uses-recent-average", func(t *testing.T) {
		p := newTestSelfProfiler(SelfProfilingConfig{
			RecentPoints:      5,
			CPUUsagePercent:   80,
			MEMUsageMB:        512,
			MEMUsagePercent:   70,
			Duration:          time.Second,
			EmergencyDuration: time.Second,
			Cooldown:          time.Minute,
		})

		for i := 0; i < 5; i++ {
			p.samples = append(p.samples, &resourceSample{
				HasCPU:          true,
				CPUUsagePercent: 90,
				MEMUsageMB:      600,
				MEMUsagePercent: 75,
			})
		}

		decision := p.triggerDecision(&resourceSample{
			HasCPU:          true,
			CPUUsagePercent: 90,
			MEMUsageMB:      600,
			MEMUsagePercent: 75,
		})
		require.NotEmpty(t, decision.reasons)
		assert.False(t, decision.emergency)
		assert.Len(t, decision.reasons, 3)
	})

	t.Run("emergency-threshold-uses-current-point", func(t *testing.T) {
		p := newTestSelfProfiler(SelfProfilingConfig{
			MEMUsageMBEmergency:      1024,
			MEMUsagePercentEmergency: 95,
		})

		decision := p.triggerDecision(&resourceSample{
			MEMUsageMB:      2048,
			MEMUsagePercent: 96,
		})
		require.NotEmpty(t, decision.reasons)
		assert.True(t, decision.emergency)
		assert.Len(t, decision.reasons, 2)
	})
}

func TestSelfProfilerReserveTrigger(t *testing.T) {
	p := newTestSelfProfiler(SelfProfilingConfig{
		Cooldown: time.Minute,
	})

	require.True(t, p.reserveTrigger())
	assert.True(t, p.reserveTrigger())

	p.lastProfileTime = time.Now()
	assert.False(t, p.reserveTrigger())
	p.lastProfileTime = time.Now().Add(-2 * time.Minute)
	assert.True(t, p.reserveTrigger())
}

func TestSelfProfilerEnabledTypes(t *testing.T) {
	p := newTestSelfProfiler(SelfProfilingConfig{
		EnabledTypes: []string{" CPU ", "heap", "cpu", "", "goroutine"},
	})

	assert.Equal(t, []string{"cpu", "heap", "goroutine"}, p.enabledTypes())
}

func TestSelfProfilerCollectCPUProfile(t *testing.T) {
	p := newTestSelfProfiler(SelfProfilingConfig{
		Duration: 100 * time.Millisecond,
	})

	file, err := p.collectCPUProfile(selfProfileDecision{})
	require.NoError(t, err)
	assert.Equal(t, "cpu.pprof", file.fileName)
	require.NotEmpty(t, file.data)

	_, err = pprofile.ParseData(file.data)
	require.NoError(t, err)
}

func TestBuildProfileUploadRequest(t *testing.T) {
	start := time.Unix(100, 0)
	end := start.Add(time.Second)
	pbBytes, err := buildProfileUploadRequest(start, end,
		map[string]string{"env": "testing", "team": "datakit"},
		[]profileFile{{fileName: "cpu.pprof", data: []byte("profile-data")}},
	)
	require.NoError(t, err)

	var reqPB storage.Request
	require.NoError(t, proto.Unmarshal(pbBytes, &reqPB))
	assert.Equal(t, http.MethodPost, reqPB.Method)
	assert.Equal(t, int64(len(reqPB.Body)), reqPB.ContentLength)

	headers := storage.ConvertMapEntriesToMap(reqPB.Header)
	require.NotEmpty(t, headers["Content-Type"])
	assert.Contains(t, headers["Content-Type"][0], "multipart/form-data")

	req, err := http.NewRequest(http.MethodPost, "http://example.com", bytes.NewReader(reqPB.Body))
	require.NoError(t, err)
	for k, vals := range headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	require.NoError(t, req.ParseMultipartForm(32<<20))
	require.Contains(t, req.MultipartForm.File, "cpu.pprof")

	eventFile, _, err := req.FormFile(profileEventFile)
	require.NoError(t, err)
	defer eventFile.Close() //nolint:errcheck

	eventBytes, err := io.ReadAll(eventFile)
	require.NoError(t, err)

	var event profileEvent
	require.NoError(t, json.Unmarshal(eventBytes, &event))
	assert.Equal(t, "pprof", event.Format)
	assert.Equal(t, "pprof", event.Profiler)
	assert.Equal(t, "golang", event.Language)
	assert.Equal(t, []string{"cpu.pprof"}, event.Attachments)
	assert.Contains(t, event.TagsProfiler, "service:dk")
	assert.Contains(t, event.TagsProfiler, "env:testing")
	assert.Contains(t, event.TagsProfiler, "team:datakit")
	assert.Contains(t, event.SubCustomTags, "env:testing")
	assert.Contains(t, event.SubCustomTags, "team:datakit")

	var eventMap map[string]interface{}
	require.NoError(t, json.Unmarshal(eventBytes, &eventMap))
	assert.NotContains(t, eventMap, "trigger_reason")
}

func TestBuildProfileUploadRequestWithSinkerV2(t *testing.T) {
	oldCfg := config.Cfg
	t.Cleanup(func() {
		config.Cfg = oldCfg
	})

	dw := dataway.NewDefaultDataway(dataway.WithGlobalTags(map[string]string{
		"env": "",
	}))
	dw.EnableSinker = true
	dw.SinkerHeaderVersion = "v2"
	config.Cfg = &config.Config{Dataway: dw}

	pbBytes, err := buildProfileUploadRequest(time.Unix(100, 0), time.Unix(101, 0),
		map[string]string{"env": "prod,cn"},
		[]profileFile{{fileName: "cpu.pprof", data: []byte("profile-data")}},
	)
	require.NoError(t, err)

	var reqPB storage.Request
	require.NoError(t, proto.Unmarshal(pbBytes, &reqPB))

	headers := storage.ConvertMapEntriesToMap(reqPB.Header)
	assert.Empty(t, headers[dataway.HeaderXGlobalTags])
	require.Equal(t, []string{"env=prod%2Ccn"}, headers[dataway.HeaderXGlobalTagsV2])
}

func newTestSelfProfiler(cfg SelfProfilingConfig) *selfProfiler {
	cfg.Enabled = true
	cfg = normalizeSelfProfilingConfig(cfg)
	return &selfProfiler{
		ipt: &Input{
			SelfProfiling: &cfg,
			semStop:       cliutils.NewSem(),
		},
		samples: make([]*resourceSample, 0, cfg.RecentPoints),
	}
}
