// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package ccommon

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GuanceCloud/cliutils/logger"
	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/require"
)

func TestOptionParsingSupportsODBCFlagsAndLists(t *testing.T) {
	var opt Option
	_, err := flags.ParseArgs(&opt, []string{
		"--host", "ibmi.example.com",
		"--username", "datakit",
		"--password", "secret",
		"--driver", "IBM i Access ODBC Driver 64-bit",
		"--metric-enabled=false",
		"--query", "cpu_usage",
		"--query", "message_queue_info",
		"--severity-threshold", "80",
		"--message-queue", "QSYSOPR",
		"--message-queue", "QSYSMSG",
	})
	require.NoError(t, err)
	require.Equal(t, "ibmi.example.com", opt.Host)
	require.Equal(t, "datakit", opt.Username)
	require.Equal(t, "secret", opt.Password)
	require.Equal(t, "IBM i Access ODBC Driver 64-bit", opt.Driver)
	require.Equal(t, "false", opt.MetricEnabled)
	require.Equal(t, []string{"cpu_usage", "message_queue_info"}, opt.Queries)
	require.Equal(t, 80, opt.SeverityThreshold)
	require.Equal(t, []string{"QSYSOPR", "QSYSMSG"}, opt.MessageQueues)
}

func TestGetPostURLWithElection(t *testing.T) {
	got := GetPostURL(true, CategoryLogging, "ibm_i", "127.0.0.1", 9529)
	require.Contains(t, got, "/v1/write/logging?input=ibm_i")
	require.Contains(t, got, "ignore_global_host_tags=true")
	require.Contains(t, got, "global_env_tags=true")
}

func TestWriteDataAndFeedLastError(t *testing.T) {
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		bodies = append(bodies, string(body))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	l := logger.DefaultSLogger("ibm_i-test")
	require.NoError(t, WriteData(l, []byte("ibm_i up=1i"), server.URL))
	DatakitLastErrURL = server.URL
	defer func() { DatakitLastErrURL = "" }()
	require.NoError(t, FeedLastError("ibm_i", l, "connection failed"))

	require.Len(t, bodies, 2)
	require.Equal(t, "ibm_i up=1i", bodies[0])
	require.True(t, strings.Contains(bodies[1], `"input":"ibm_i"`))
	require.True(t, strings.Contains(bodies[1], `"err_content":"connection failed"`))
}
