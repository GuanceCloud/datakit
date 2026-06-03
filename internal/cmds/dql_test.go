// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cmds

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/httpapi"
)

func TestRunDQLQueriesMockServerAndWritesCSV(t *testing.T) {
	origTemporaryToken := temporaryToken
	origEchoExplain := echoExplain
	temporaryToken = ""
	echoExplain = false
	t.Cleanup(func() {
		temporaryToken = origTemporaryToken
		echoExplain = origEchoExplain
	})

	const query = "M::cpu:(last(`usage`))"
	const token = "test-token"

	var gotReq httpapi.QueryRaw
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, dqlraw, r.URL.Path)

		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotReq))
		assert.Equal(t, token, gotReq.Token)
		require.Len(t, gotReq.Queries, 1)
		assert.Equal(t, query, gotReq.Queries[0].Query)
		assert.False(t, gotReq.EchoExplain)

		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte(`{
			"content": [{
				"series": [{
					"name": "cpu",
					"columns": ["time", "usage", "enabled"],
					"values": [[1710000000000, 0.9, true]]
				}]
			}]
		}`))
		require.NoError(t, err)
	}))
	defer srv.Close()

	csvPath := filepath.Join(t.TempDir(), "dql.csv")
	err := RunDQL(DQLOptions{
		Run:     query,
		Token:   token,
		Host:    strings.TrimPrefix(srv.URL, "http://"),
		CSV:     csvPath,
		Force:   true,
		LogPath: filepath.Join(t.TempDir(), "dql.log"),
	})
	require.NoError(t, err)

	f, err := http.Dir(filepath.Dir(csvPath)).Open(filepath.Base(csvPath))
	require.NoError(t, err)
	defer f.Close() //nolint:errcheck

	records, err := csv.NewReader(f).ReadAll()
	require.NoError(t, err)
	require.Equal(t, [][]string{
		{"name", "enabled", "time", "usage"},
		{"cpu", "true", "1710000000000", "0.9"},
	}, records)
}

// func TestConvertToString(t *testing.T) {
// 	series := []*models.Row{}
// 	body, _ := os.ReadFile("test.json")
// 	_ = json.Unmarshal(body, &series)

// 	for _, v := range series[0].Values {
// 		res := convertStrings(v)
// 		for _, value := range res {
// 			t.Logf(value)
// 			tu.Equals(t, "string", reflect.TypeOf(value).String())
// 		}
// 	}
// }
