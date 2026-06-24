// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckWebsocketOrigin(t *testing.T) {
	cases := []struct {
		name   string
		host   string
		origin string
		want   bool
	}{
		{
			name: "missing origin allowed for non browser clients",
			host: "dca.example.com",
			want: true,
		},
		{
			name:   "same host allowed",
			host:   "dca.example.com",
			origin: "https://dca.example.com",
			want:   true,
		},
		{
			name:   "same host with port allowed",
			host:   "dca.example.com:8080",
			origin: "https://dca.example.com:8080",
			want:   true,
		},
		{
			name:   "localhost development origin allowed",
			host:   "127.0.0.1:7001",
			origin: "http://localhost:3000",
			want:   true,
		},
		{
			name:   "cross site origin rejected",
			host:   "dca.example.com",
			origin: "https://evil.example.com",
			want:   false,
		},
		{
			name:   "invalid origin rejected",
			host:   "dca.example.com",
			origin: "://bad-origin",
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "http://"+tc.host+"/api/datakit/ws", nil)
			require.NoError(t, err)
			req.Host = tc.host
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}

			require.Equal(t, tc.want, checkWebsocketOrigin(req))
		})
	}
}
