// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package logstreaming

import (
	"bytes"
	"errors"
	"io"
	"net/url"
	"sync"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
)

type capturedLogstreamingFeed struct {
	points       []*point.Point
	storageIndex string
	feedSource   string
}

type logstreamingCaptureFeeder struct {
	feeds     []capturedLogstreamingFeed
	failIndex string
	feedErr   error
}

func (f *logstreamingCaptureFeeder) Feed(
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
	f.feeds = append(f.feeds, capturedLogstreamingFeed{
		points:       points,
		storageIndex: feedData.GetStorageIndex(),
		feedSource:   feedData.GetFeedSource(),
	})
	if f.feedErr != nil && feedData.GetStorageIndex() == f.failIndex {
		return f.feedErr
	}
	return nil
}

func (*logstreamingCaptureFeeder) FeedLastError(string, ...metrics.LastErrorOption) {}

func TestJSONAsFieldsConfig(t *testing.T) {
	input := defaultInput()
	_, err := toml.Decode(`json_as_fields = true`, input)
	require.NoError(t, err)
	assert.True(t, input.JSONAsFields)

	_, err = toml.Decode(`json_as_fields = "true"`, defaultInput())
	require.Error(t, err)
}

func TestProcessLogBodyJSONAsFields(t *testing.T) {
	feeder := &logstreamingCaptureFeeder{}
	input := &Input{
		JSONAsFields:   true,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}
	parameter := &parameters{
		url: &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/v1/write/logstreaming"},
		queryValues: url.Values{
			"source":        []string{"testing"},
			"service":       []string{"collector-service"},
			"storage_index": []string{"configured"},
		},
		body: io.NopCloser(bytes.NewBufferString(
			"{\"service\":\"json-service\",\"value\":1,\"storage_index\":\"a\"}\n" +
				"{\"value\":2,\"storage_index\":\"b\"}\n" +
				"{\"value\":3,\"storage_index\":\"a\"}\n" +
				"not-json\n{}\n")),
	}

	require.NoError(t, input.processLogBody(parameter))
	require.Len(t, feeder.feeds, 1)
	assert.Equal(t, "configured", feeder.feeds[0].storageIndex)
	assert.Equal(t, "logstreaming.testing.configured", feeder.feeds[0].feedSource)

	require.Len(t, feeder.feeds[0].points, 5)
	first := feeder.feeds[0].points[0]
	assert.Empty(t, first.GetTag("service"))
	assert.Equal(t, "json-service", first.Get("service"))
	assert.Equal(t, int64(1), first.Get("value"))
	assert.Equal(t, "a", first.Get("json_storage_index"))
	assert.Nil(t, first.Get("storage_index"))
	assert.Nil(t, first.Get("message"))

	assert.Equal(t, "b", feeder.feeds[0].points[1].Get("json_storage_index"))
	assert.Equal(t, "not-json", feeder.feeds[0].points[3].Get("message"))
	assert.Equal(t, "{}", feeder.feeds[0].points[4].Get("message"))
}

func TestProcessLogBodyJSONFeedFailure(t *testing.T) {
	feeder := &logstreamingCaptureFeeder{
		failIndex: "configured",
		feedErr:   errors.New("feed failed"),
	}
	input := &Input{
		JSONAsFields:   true,
		feeder:         feeder,
		scanbufPool:    &sync.Pool{},
		ScanBufferSize: 4 * 1024,
	}
	parameter := &parameters{
		url: &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/v1/write/logstreaming"},
		queryValues: url.Values{
			"source":        []string{"testing"},
			"storage_index": []string{"configured"},
		},
		body: io.NopCloser(bytes.NewBufferString(
			"{\"value\":1,\"storage_index\":\"a\"}\n" +
				"{\"value\":2,\"storage_index\":\"b\"}\n")),
	}

	require.ErrorIs(t, input.processLogBody(parameter), feeder.feedErr)
	require.Len(t, feeder.feeds, 1)
	assert.Equal(t, "configured", feeder.feeds[0].storageIndex)
	require.Len(t, feeder.feeds[0].points, 2)
}
