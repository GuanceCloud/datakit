// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	pprofile "github.com/google/pprof/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
)

func TestKubernetesProfileBoundedParsing(t *testing.T) {
	for _, fixture := range []string{"cpu", "goroutines", "delta-heap", "delta-mutex", "delta-block"} {
		t.Run(fixture, func(t *testing.T) {
			data, err := os.ReadFile("metrics/testdata/" + fixture + ".pprof")
			require.NoError(t, err)
			parsed, decoded, err := parseBoundedProfile(data, 32*MiB)
			require.NoError(t, err)
			require.NotEmpty(t, parsed.SampleType)
			require.NotEmpty(t, decoded)
		})
	}
	t.Run("gzip expansion", func(t *testing.T) {
		buffer := new(bytes.Buffer)
		writer := gzip.NewWriter(buffer)
		_, err := writer.Write(bytes.Repeat([]byte{0}, int(MiB+1)))
		require.NoError(t, err)
		require.NoError(t, writer.Close())
		require.Less(t, int64(buffer.Len()), int64(MiB))
		_, _, err = parseBoundedProfile(buffer.Bytes(), MiB)
		require.ErrorContains(t, err, "decompressed profile exceeds")
	})
	t.Run("many samples", func(t *testing.T) {
		var data []byte
		for index := 0; index <= maxKubernetesProfileFields; index++ {
			data = protowire.AppendTag(data, 2, protowire.BytesType)
			data = protowire.AppendBytes(data, nil)
		}
		_, _, err := parseBoundedProfile(data, 32*MiB)
		require.ErrorContains(t, err, "parsing budget")
	})
	t.Run("valid complex profile", func(t *testing.T) {
		parsed := &pprofile.Profile{SampleType: []*pprofile.ValueType{{Type: "samples", Unit: "count"}}}
		for index := 0; index < 25000; index++ {
			parsed.Sample = append(parsed.Sample, &pprofile.Sample{
				Value: []int64{1}, Label: map[string][]string{"work": {"request"}},
			})
		}
		require.NoError(t, parsed.CheckValid())
		buffer := new(bytes.Buffer)
		require.NoError(t, parsed.Write(buffer))
		_, _, err := parseBoundedProfile(buffer.Bytes(), 32*MiB)
		require.ErrorContains(t, err, "parsing budget")
	})
	t.Run("shared large labels", func(t *testing.T) {
		parsed := &pprofile.Profile{SampleType: []*pprofile.ValueType{{Type: "samples", Unit: "count"}}}
		for index := 0; index < 1024; index++ {
			parsed.Sample = append(parsed.Sample, &pprofile.Sample{
				Value: []int64{1}, Label: map[string][]string{"large": {strings.Repeat("x", 1024)}},
			})
		}
		buffer := new(bytes.Buffer)
		require.NoError(t, parsed.Write(buffer))
		_, _, err := parseBoundedProfile(buffer.Bytes(), MiB)
		require.ErrorContains(t, err, "expanded sample budget")
	})
}

func TestKubernetesProfileDeltaCacheBudgetAndChurn(t *testing.T) {
	data, err := os.ReadFile("metrics/testdata/delta-heap.pprof")
	require.NoError(t, err)
	_, decoded, err := parseBoundedProfile(data, 32*MiB)
	require.NoError(t, err)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, writeErr := writer.Write(data)
		assert.NoError(t, writeErr)
	}))
	defer server.Close()
	manager := newTestKubernetesProfileManager(t, []*KubernetesProfiler{{Port: "pprof"}})
	manager.deltaCache = newProfileDeltaCache(2 * (int64(len(decoded)) + profileDeltaEntryOverhead))
	var profilers []*GoProfiler
	for index := 0; index < 8; index++ {
		collector, createErr := manager.newGoCollector(manager.rules[0], server.URL, manager.input)
		require.NoError(t, createErr)
		profiler := collector.(*goProfileTargetCollector).profiler
		profilers = append(profilers, profiler)
		profileData, pullErr := profiler.pullProfileItem(context.Background(), "heap", profileConfigMap["heap"], time.Second)
		require.NoError(t, pullErr)
		assert.Nil(t, profileData)
		assert.Empty(t, profiler.deltas, "idle targets must not retain parsed profiles")
		assert.LessOrEqual(t, manager.deltaCache.used, manager.deltaCache.limit)
	}
	require.Len(t, manager.deltaCache.entries, 2)
	profileData, err := profilers[0].pullProfileItem(context.Background(), "heap", profileConfigMap["heap"], time.Second)
	require.NoError(t, err)
	assert.Nil(t, profileData, "an evicted baseline must be rebuilt without emitting a delta")
	profileData, err = profilers[0].pullProfileItem(context.Background(), "heap", profileConfigMap["heap"], time.Second)
	require.NoError(t, err)
	require.NotNil(t, profileData)
	_, err = pprofile.ParseData(profileData.buf.Bytes())
	require.NoError(t, err)
	for _, profiler := range profilers {
		manager.retireTarget(&kubernetesProfileTarget{collector: &goProfileTargetCollector{profiler: profiler}})
	}
	assert.Empty(t, manager.deltaCache.entries)
	assert.Zero(t, manager.deltaCache.used)
	_, err = profilers[0].pullProfileItem(context.Background(), "heap", profileConfigMap["heap"], time.Second)
	require.NoError(t, err)
	assert.Empty(t, manager.deltaCache.entries, "a retiring worker must not recreate its baseline")
}
