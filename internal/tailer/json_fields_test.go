// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package tailer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type capturedJSONFeed struct {
	points       []*point.Point
	storageIndex string
	feedSource   string
}

type jsonCaptureFeeder struct {
	feeds []capturedJSONFeed
}

func newJSONTestFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "app.log")
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, file.Close())
	return path
}

func (f *jsonCaptureFeeder) Feed(_ point.Category, points []*point.Point, opts ...dkio.FeedOption) error {
	feedData := dkio.GetFeedData()
	for _, opt := range opts {
		if opt != nil {
			opt(feedData)
		}
	}

	f.feeds = append(f.feeds, capturedJSONFeed{
		points:       points,
		storageIndex: feedData.GetStorageIndex(),
		feedSource:   feedData.GetFeedSource(),
	})
	return nil
}

func (*jsonCaptureFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestSingleJSONAsFields(t *testing.T) {
	feeder := &jsonCaptureFeeder{}
	single, err := NewTailerSingle(newJSONTestFile(t),
		WithSource("testing"),
		WithStorageIndex("configured"),
		WithExtraTags(map[string]string{"host": "collector-host"}),
		WithJSONAsFields(true),
		WithFeeder(feeder),
	)
	require.NoError(t, err)

	single.feedToIO([][]byte{
		[]byte(`{"host":"json-host","status":1,"value":1,"http.status":200,"time":123,"storage_index":"a"}`),
		[]byte(`{"value":2,"storage_index":"b"}`),
		[]byte(`{"value":3,"storage_index":"a"}`),
		[]byte(`not-json`),
		[]byte(`{}`),
	})

	require.Len(t, feeder.feeds, 1)
	assert.Equal(t, "configured", feeder.feeds[0].storageIndex)
	assert.Equal(t, "logging.testing.configured", feeder.feeds[0].feedSource)

	require.Len(t, feeder.feeds[0].points, 5)
	first := feeder.feeds[0].points[0]
	assert.Empty(t, first.GetTag("host"))
	assert.Equal(t, "json-host", first.Get("host"))
	assert.Empty(t, first.GetTag("status"))
	assert.Equal(t, int64(1), first.Get("status"))
	assert.Equal(t, int64(1), first.Get("value"))
	assert.Equal(t, int64(200), first.Get("http_status"))
	assert.Nil(t, first.Get("http.status"))
	assert.Equal(t, int64(123), first.Get("json_time"))
	assert.Equal(t, "a", first.Get("json_storage_index"))
	assert.Nil(t, first.Get("storage_index"))
	assert.Nil(t, first.Get("message"))

	assert.Equal(t, "b", feeder.feeds[0].points[1].Get("json_storage_index"))
	assert.Equal(t, "not-json", feeder.feeds[0].points[3].Get("message"))
	assert.Equal(t, "{}", feeder.feeds[0].points[4].Get("message"))
}

func TestSingleJSONFieldsBypassFieldWhitelist(t *testing.T) {
	feeder := &jsonCaptureFeeder{}
	single, err := NewTailerSingle(newJSONTestFile(t),
		WithSource("testing"),
		WithFieldWhitelist([]string{"no_collector_field_matches"}),
		WithJSONAsFields(true),
		WithFeeder(feeder),
	)
	require.NoError(t, err)

	single.feedToIO([][]byte{
		[]byte(`{"value":1}`),
		[]byte(`plain text`),
	})

	require.Len(t, feeder.feeds, 1)
	require.Len(t, feeder.feeds[0].points, 1)
	assert.Equal(t, int64(1), feeder.feeds[0].points[0].Get("value"))
}

func TestSocketJSONAsFields(t *testing.T) {
	feeder := &jsonCaptureFeeder{}
	socket, err := NewSocketLogging(
		WithSource("testing"),
		WithStorageIndex("configured"),
		WithJSONAsFields(true),
		WithFeeder(feeder),
	)
	require.NoError(t, err)

	socket.feedMessages([]socketMessage{
		{content: []byte(`{"collector_source_ip":"json-ip","value":1,"storage_index":"a"}`), sourceIP: "collector-ip"},
		{content: []byte(`plain text`)},
	})

	require.Len(t, feeder.feeds, 1)
	assert.Equal(t, "configured", feeder.feeds[0].storageIndex)
	require.Len(t, feeder.feeds[0].points, 2)
	first := feeder.feeds[0].points[0]
	assert.Empty(t, first.GetTag(collectorSourceIPTag))
	assert.Equal(t, "json-ip", first.Get(collectorSourceIPTag))
	assert.Equal(t, int64(1), first.Get("value"))
	assert.Equal(t, "a", first.Get("json_storage_index"))
	assert.Nil(t, first.Get("storage_index"))
	assert.Nil(t, first.Get("message"))
	assert.Equal(t, "plain text", feeder.feeds[0].points[1].Get("message"))
}
