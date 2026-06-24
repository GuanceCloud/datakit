// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io/dataway"
)

func TestDoUploadHostStatusSkipsInvalidInputs(t *testing.T) {
	db := withTestDatakitDB(t)

	doUploadHostStatus(nil)

	badJSON := newTestDataKit("upload-bad-json")
	badJSON.Config = "{bad-json"
	require.NoError(t, db.Insert(badJSON))

	emptyConfig := newTestDataKit("upload-empty-config")
	emptyConfig.Config = "null"
	require.NoError(t, db.Insert(emptyConfig))

	emptyDataway := newTestDataKit("upload-empty-dataway")
	emptyDataway.Config = `{"dataway_url":""}`
	require.NoError(t, db.Insert(emptyDataway))

	invalidURL := newTestDataKit("upload-invalid-url")
	invalidURL.Config = `{"dataway_url":"://bad"}`
	invalidURL.Status = ws.StatusOffline
	require.NoError(t, db.Insert(invalidURL))

	doUploadHostStatus(&dataway.DialtestingSender{})
}

func TestDoUploadHostStatusWritesMetric(t *testing.T) {
	db := withTestDatakitDB(t)
	requests := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dk := newTestDataKit("upload-ok")
	dk.Config = `{"dataway_url":"` + server.URL + `/v1/write"}`
	dk.Status = ws.StatusRunning
	require.NoError(t, db.Insert(dk))

	sender := &dataway.DialtestingSender{}
	require.NoError(t, sender.Init(&dataway.DialtestingSenderOpt{HTTPTimeout: time.Second}))
	doUploadHostStatus(sender)

	select {
	case path := <-requests:
		require.Contains(t, path, "metric")
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for dataway metric upload")
	}
}

func TestUploadHostStatusCanStopBeforeTick(t *testing.T) {
	closeCh := make(chan struct{})
	close(closeCh)
	UploadHostStatus(time.Hour, closeCh)
}
