// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package flameshot

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_filterRegex(t *testing.T) {
	tmall := "^java\\b.*tmall\\.jar$"

	service1 := "java -jar tmall.jar"
	service2 := "java -javaagent:dd-java-agent-v1.55.0-ext.jar -jar tmall.jar"
	service3 := "java -javaagent:dd-java-agent-v1.55.0-ext.jar  -Ddd.service=tmall -Ddd.agent.port=9529 -jar tmall.jar"

	re, err := regexp.Compile(tmall)
	assert.NoError(t, err)
	assert.True(t, re.MatchString(service1))
	assert.True(t, re.MatchString(service2))
	assert.True(t, re.MatchString(service3))
}

func TestMonitorHttp(t *testing.T) {
	m := &monitor{
		config: &Config{
			HTTPConfig: &HTTPConfig{
				LocalHost: "localhost",
				LocalPort: "8989",
			},
		},
		cs:        make([]*processM, 0),
		csChan:    make(chan *processM, 1),
		statsChan: make(chan *triggerStats, 1),
	}
	cxt, cel := context.WithTimeout(context.Background(), time.Minute*2)
	go m.startHTTPServer()
	go func(cxt context.Context) {
		count := 0
		for {
			select {
			case <-cxt.Done():
				return
			case stats := <-m.statsChan:
				count++
				if stats != nil {
					t.Logf("stats:%+v", stats)
				}
			}
			if count == 2 {
				cel()
				return
			}
		}
	}(cxt)
	time.Sleep(time.Second * 2)

	req, err := http.NewRequest("GET", "http://127.0.0.1:8989/v1/profile?pid=1234&duration=10&events=all", nil)
	assert.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req2, err := http.NewRequest("GET", "http://127.0.0.1:8989/v1/profile?command=^java\\b.*tmall\\.jar$&duration=10s&events=cpu,alloc", nil)
	assert.NoError(t, err)
	resp2, err := http.DefaultClient.Do(req2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	<-cxt.Done()
}

func TestURLJoin(t *testing.T) {
	type args struct {
		addr string
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "case1",
			args: args{
				addr: "http://localhost",
				path: "/profiling/v1/input",
			},
			want: "http://localhost/profiling/v1/input",
		},
		{
			name: "case2",
			args: args{
				addr: "http://localhost:9529/",
				path: "/profiling/v1/input",
			},
			want: "http://localhost:9529/profiling/v1/input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.JoinPath(tt.args.addr, tt.args.path)
			assert.NoError(t, err)
			assert.Equalf(t, tt.want, u, "urlJoin(%v, %v)", tt.args.addr, tt.args.path)
		})
	}
}

func Test_filterProcessesByRegex(t *testing.T) {
	re, err := regexp.Compile("^.*app\\.jar$") //nolint
	if err != nil {
		assert.NoError(t, err)
	}
	f := re.MatchString("123app.jar")
	assert.Equal(t, f, true)

	re1, err := regexp.Compile("java") //nolint
	if err != nil {
		assert.NoError(t, err)
	}
	f = re1.MatchString("java")
	assert.Equal(t, f, true)
}
