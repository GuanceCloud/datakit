// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build !windows
// +build !windows

package logfwdserver

import (
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type capturedLogfwdFeed struct {
	points       []*point.Point
	storageIndex string
	feedSource   string
}

type logfwdCaptureFeeder struct {
	feeds []capturedLogfwdFeed
}

func (f *logfwdCaptureFeeder) Feed(
	_ point.Category,
	points []*point.Point,
	opts ...dkio.FeedOption,
) error {
	feedData := dkio.GetFeedData()
	for _, opt := range opts {
		if opt != nil {
			opt(feedData)
		}
	}
	f.feeds = append(f.feeds, capturedLogfwdFeed{
		points:       points,
		storageIndex: feedData.GetStorageIndex(),
		feedSource:   feedData.GetFeedSource(),
	})
	return nil
}

func (*logfwdCaptureFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestProcessLogMessageJSONAsFields(t *testing.T) {
	feeder := &logfwdCaptureFeeder{}
	input := &Input{
		Tags:   map[string]string{"environment": "testing"},
		feeder: feeder,
	}
	message := logMessage{
		Source:       "source",
		StorageIndex: "configured",
		JSONAsFields: true,
		Tags:         map[string]string{"host": "collector-host"},
		Fields:       map[string]interface{}{"filepath": "/var/log/app.log"},
		Log:          `{"host":"json-host","value":1,"time":123,"storage_index":"dynamic"}`,
	}

	require.NoError(t, input.processLogMessage(message))
	require.Len(t, feeder.feeds, 1)
	feed := feeder.feeds[0]
	assert.Equal(t, "configured", feed.storageIndex)
	assert.Equal(t, "logfwd.source.configured", feed.feedSource)
	require.Len(t, feed.points, 1)

	pt := feed.points[0]
	assert.Empty(t, pt.GetTag("host"))
	assert.Equal(t, "json-host", pt.Get("host"))
	assert.Equal(t, int64(1), pt.Get("value"))
	assert.Equal(t, int64(123), pt.Get("json_time"))
	assert.Equal(t, "dynamic", pt.Get("json_storage_index"))
	assert.Nil(t, pt.Get("storage_index"))
	assert.Equal(t, "/var/log/app.log", pt.Get("filepath"))
	assert.Nil(t, pt.Get("message"))
}

func TestProcessLogMessageJSONFallback(t *testing.T) {
	feeder := &logfwdCaptureFeeder{}
	input := &Input{Tags: map[string]string{}, feeder: feeder}

	require.NoError(t, input.processLogMessage(logMessage{
		Source:       "source",
		StorageIndex: "configured",
		JSONAsFields: true,
		Log:          `not-json`,
	}))

	require.Len(t, feeder.feeds, 1)
	assert.Equal(t, "configured", feeder.feeds[0].storageIndex)
	assert.Equal(t, "not-json", feeder.feeds[0].points[0].Get("message"))
}
