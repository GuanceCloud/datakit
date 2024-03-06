// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package healthcheck

import (
	"net"
	h "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	var pid int32 = 888888
	input := defaultInput()
	input.Interval = "1s"
	input.Process = []*process{{
		NamesRegex: []string{"datakit", "sleep"},
		MinRunTime: "2s",
		processes: map[int32]*processInfo{
			pid: {
				pid:  pid,
				name: "test",
			},
		},
	}}

	input.initConfig()

	assert.NoError(t, input.Collect())
	assert.NotEmpty(t, input.collectCache)
	assert.Equal(t, input.collectCache[0].Name(), processMetricName)
}

func TestTCP(t *testing.T) {
	cases := []struct {
		Title    string
		IsFail   bool
		TCP      []*tcp
		FailType string
		Server   func() (net.Listener, error)
	}{
		{
			Title: "test tcp ok",
			TCP: []*tcp{
				{
					ConnectionTimeOut: "3m",
				},
			},
			Server: tcpServer,
		},
		{
			Title: "test tcp timeout",
			TCP: []*tcp{
				{
					ConnectionTimeOut: "1us",
				},
			},
			IsFail:   true,
			FailType: "connection-timeout",
			Server:   tcpServer,
		},
		{
			Title:    "test tcp refused",
			IsFail:   true,
			FailType: "connection-refused",
			Server: func() (net.Listener, error) {
				server, err := tcpServer()
				server.Close() //nolint:err
				return server, err
			},
		},
	}

	for _, cs := range cases {
		input := defaultInput()
		if cs.Server != nil {
			server, err := cs.Server()
			assert.NoError(t, err)
			defer func() {
				server.Close()
			}()
			if len(cs.TCP) > 0 {
				cs.TCP[0].HostPorts = []string{server.Addr().String()}
			} else {
				cs.TCP = []*tcp{
					{
						HostPorts: []string{server.Addr().String()},
					},
				}
			}
		}
		input.TCP = cs.TCP
		input.initConfig()

		assert.NoError(t, input.Collect())

		if !cs.IsFail {
			assert.Empty(t, input.collectCache)
		} else {
			assert.NotEmpty(t, input.collectCache)
			p := input.collectCache[0]
			assert.Equal(t, p.Name(), tcpMetricName)
			if len(cs.FailType) > 0 {
				failType := p.GetTag("type")
				assert.Equal(t, cs.FailType, failType)
			}
		}
	}
}

func TestHTTP(t *testing.T) {
	cases := []struct {
		Title   string
		IsFail  bool
		HTTP    []*http
		Handler h.HandlerFunc
	}{
		{
			Title: "test http ok",
			HTTP: []*http{
				{
					Method:       "GET",
					ExpectStatus: 200,
				},
			},
		},
		{
			Title:  "test http failed",
			IsFail: true,
			HTTP: []*http{
				{
					Method:       "GET",
					ExpectStatus: 200,
				},
			},
			Handler: func(w h.ResponseWriter, r *h.Request) {
				w.WriteHeader(404)
			},
		},
		{
			Title:  "test http timeout failed",
			IsFail: true,
			HTTP: []*http{
				{
					Method:  "GET",
					Timeout: "1ms",
				},
			},
			Handler: func(w h.ResponseWriter, r *h.Request) {
				time.Sleep(5 * time.Millisecond)
				w.WriteHeader(404)
			},
		},
		{
			Title: "test http timeout ok",
			HTTP: []*http{
				{
					Method:       "GET",
					Timeout:      "1s",
					ExpectStatus: 200,
				},
			},
			Handler: func(w h.ResponseWriter, r *h.Request) {
				w.WriteHeader(200)
			},
		},
	}

	for _, cs := range cases {
		input := defaultInput()
		if cs.HTTP == nil {
			cs.HTTP = []*http{
				{},
			}
		}
		input.HTTP = cs.HTTP

		handerFunc := cs.Handler
		if handerFunc == nil {
			handerFunc = func(w h.ResponseWriter, r *h.Request) {
				w.WriteHeader(h.StatusOK)
			}
		}

		server := httptest.NewServer(handerFunc)

		cs.HTTP[0].HTTPURLs = []string{server.URL}

		input.initConfig()

		assert.NoError(t, input.Collect())

		if !cs.IsFail {
			assert.Empty(t, input.collectCache)
		} else {
			assert.NotEmpty(t, input.collectCache)
		}
		server.Close()
	}
}

func tcpServer() (server net.Listener, err error) {
	server, err = net.Listen("tcp4", "")
	if err != nil {
		return
	}

	go func() {
		if conn, err := server.Accept(); err != nil {
			return
		} else {
			defer conn.Close()
			conn.SetDeadline(time.Now().Add(5 * time.Second))
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				return
			}

			_, _ = conn.Write(buf[:n])
		}
	}()

	return
}
