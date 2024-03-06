// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	uhttp "github.com/GuanceCloud/cliutils/network/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/config"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

var (
	router = gin.New()
	ts     *httptest.Server
)

type args struct {
	globalElection     bool
	isDocker           bool
	hostName           string
	globalHostTags     map[string]string
	globalElectionTags map[string]string
	query              map[string][]string
}

func Test_All(t *testing.T) {
	router.Use(uhttp.RequestLoggerMiddleware)
	router.POST("/v1/global/host/tags", ginLimiter(reqLimiter), postHostTags)
	router.DELETE("/v1/global/host/tags", ginLimiter(reqLimiter), deleteHostTags)
	router.POST("/v1/global/election/tags", ginLimiter(reqLimiter), postElectionTags)
	router.DELETE("/v1/global/election/tags", ginLimiter(reqLimiter), deleteElectionTags)

	ts = httptest.NewServer(router)
	defer ts.Close()
	time.Sleep(time.Second)

	case_postHostTags(t)
	case_deleteHostTags(t)
	case_postElectionTags(t)
	case_deleteElectionTags(t)
}

func case_postHostTags(t *testing.T) {
	t.Helper()
	tests := []struct {
		name                 string
		args                 args
		wantStatusCode       int
		wantTags             map[string](map[string]string)
		wantDumpHostTags     map[string]string
		wantDumpElectionTags map[string]string
	}{
		{
			name: "ele=true post",
			args: args{
				globalElection:     true,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				query:              map[string][]string{"h2": ([]string{"h22"}), "h3": ([]string{"h3"}), "host": ([]string{"no_useful"})},
			},
			wantStatusCode: 200,
			wantTags: map[string](map[string]string){
				"host-tags":     map[string]string{"host": "hostMock", "h1": "h1", "h2": "h22", "h3": "h3"},
				"election-tags": map[string]string{"e1": "e1", "e2": "e2"},
				"dataway-tags":  map[string]string{"host": "hostMock", "h1": "h1", "h2": "h22", "h3": "h3", "e1": "e1", "e2": "e2"},
			},
			wantDumpHostTags:     map[string]string{"h1": "h1", "h2": "h22", "h3": "h3"},
			wantDumpElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
		},
		{
			name: "ele=false post",
			args: args{
				globalElection:     false,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				// host will no useful, h2:h22 will recover h2:h2
				query: map[string][]string{"h2": ([]string{"h22"}), "h3": ([]string{"h3"}), "host": ([]string{"no_useful"})},
			},
			wantStatusCode: 200,
			wantTags: map[string](map[string]string){
				"host-tags":     map[string]string{"host": "hostMock", "h1": "h1", "h2": "h22", "h3": "h3"},
				"election-tags": map[string]string{"host": "hostMock", "h1": "h1", "h2": "h22", "h3": "h3"},
				"dataway-tags":  map[string]string{"host": "hostMock", "h1": "h1", "h2": "h22", "h3": "h3", "e1": "e1", "e2": "e2"},
			},
			wantDumpHostTags:     map[string]string{"h1": "h1", "h2": "h22", "h3": "h3"},
			wantDumpElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := &mockHandle{args: tt.args}
			handle.initCfgAndTags()
			newHandle = func() handler {
				return handle
			}

			str := "/v1/global/host/tags?" + getQueryString(tt.args.query)
			req, err := http.NewRequest("POST", ts.URL+str, nil)
			if err != nil {
				t.Error(err)
			}

			c := &http.Client{}

			resp, err := c.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			// check dump datakit.conf
			h, e := handle.getTags()
			assert.Equal(t, tt.wantDumpHostTags, h)
			assert.Equal(t, tt.wantDumpElectionTags, e)

			// check StatusCode
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			if resp.StatusCode != 200 {
				return
			}

			// check response
			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			resTags := map[string](map[string]string){}
			err = json.Unmarshal(respBody, &resTags)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantTags, resTags)
		})
	}
}

func case_deleteHostTags(t *testing.T) {
	t.Helper()
	tests := []struct {
		name                 string
		args                 args
		wantStatusCode       int
		wantTags             map[string](map[string]string)
		wantDumpHostTags     map[string]string
		wantDumpElectionTags map[string]string
	}{
		{
			name: "ele=true delete",
			args: args{
				globalElection:     true,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				// host will no useful
				query: map[string][]string{"tags": ([]string{"host", "h1", "e1"})},
			},
			wantStatusCode: 200,
			wantTags: map[string](map[string]string){
				"host-tags":     map[string]string{"host": "hostMock", "h2": "h2"},
				"election-tags": map[string]string{"e1": "e1", "e2": "e2"},
				"dataway-tags":  map[string]string{"host": "hostMock", "h2": "h2", "e2": "e2"},
			},
			wantDumpHostTags:     map[string]string{"h2": "h2"},
			wantDumpElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
		},
		{
			name: "ele=false delete",
			args: args{
				globalElection:     false,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				// host will no useful
				query: map[string][]string{"tags": ([]string{"host", "h1", "e1"})},
			},
			wantStatusCode: 200,
			wantTags: map[string](map[string]string){
				"host-tags":     map[string]string{"host": "hostMock", "h2": "h2"},
				"election-tags": map[string]string{"host": "hostMock", "h2": "h2"},
				"dataway-tags":  map[string]string{"host": "hostMock", "h2": "h2", "e2": "e2"},
			},
			wantDumpHostTags:     map[string]string{"h2": "h2"},
			wantDumpElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := &mockHandle{args: tt.args}
			handle.initCfgAndTags()
			newHandle = func() handler {
				return handle
			}

			str := "/v1/global/host/tags?" + getQueryString(tt.args.query)
			req, err := http.NewRequest("DELETE", ts.URL+str, nil)
			if err != nil {
				t.Error(err)
			}

			c := &http.Client{}

			resp, err := c.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			// check dump datakit.conf
			h, e := handle.getTags()
			assert.Equal(t, tt.wantDumpHostTags, h)
			assert.Equal(t, tt.wantDumpElectionTags, e)

			// check StatusCode
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			if resp.StatusCode != 200 {
				return
			}

			// check response
			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			resTags := map[string](map[string]string){}
			err = json.Unmarshal(respBody, &resTags)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantTags, resTags)
		})
	}
}

func case_postElectionTags(t *testing.T) {
	t.Helper()
	tests := []struct {
		name                 string
		args                 args
		wantStatusCode       int
		wantTags             map[string](map[string]string)
		wantDumpHostTags     map[string]string
		wantDumpElectionTags map[string]string
	}{
		{
			name: "ele=true post",
			args: args{
				globalElection:     true,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				query:              map[string][]string{"e2": ([]string{"e22"}), "e3": ([]string{"e3"}), "host": ([]string{"no_useful"})},
			},
			wantStatusCode: 200,
			wantTags: map[string](map[string]string){
				"host-tags":     map[string]string{"host": "hostMock", "h1": "h1", "h2": "h2"},
				"election-tags": map[string]string{"e1": "e1", "e2": "e22", "e3": "e3"},
				"dataway-tags":  map[string]string{"host": "hostMock", "h1": "h1", "h2": "h2", "e1": "e1", "e2": "e22", "e3": "e3"},
			},
			wantDumpHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
			wantDumpElectionTags: map[string]string{"e1": "e1", "e2": "e22", "e3": "e3"},
		},
		{
			name: "ele=false post",
			args: args{
				globalElection:     false,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				// any is error
				query: map[string][]string{"h2": ([]string{"h22"}), "h3": ([]string{"h3"}), "host": ([]string{"no_useful"})},
			},
			wantStatusCode: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := &mockHandle{args: tt.args}
			handle.initCfgAndTags()
			newHandle = func() handler {
				return handle
			}

			str := "/v1/global/election/tags?" + getQueryString(tt.args.query)
			req, err := http.NewRequest("POST", ts.URL+str, nil)
			if err != nil {
				t.Error(err)
			}

			c := &http.Client{}

			resp, err := c.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			// check StatusCode
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			if resp.StatusCode != 200 {
				return
			}

			// check dump datakit.conf
			h, e := handle.getTags()
			assert.Equal(t, tt.wantDumpHostTags, h)
			assert.Equal(t, tt.wantDumpElectionTags, e)

			// check response
			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			resTags := map[string](map[string]string){}
			err = json.Unmarshal(respBody, &resTags)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantTags, resTags)
		})
	}
}

func case_deleteElectionTags(t *testing.T) {
	t.Helper()
	tests := []struct {
		name                 string
		args                 args
		wantStatusCode       int
		wantTags             map[string](map[string]string)
		wantDumpHostTags     map[string]string
		wantDumpElectionTags map[string]string
	}{
		{
			name: "ele=true delete",
			args: args{
				globalElection:     true,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				// host will no useful
				query: map[string][]string{"tags": ([]string{"host", "h1", "e1"})},
			},
			wantStatusCode: 200,
			wantTags: map[string](map[string]string){
				"host-tags":     map[string]string{"host": "hostMock", "h1": "h1", "h2": "h2"},
				"election-tags": map[string]string{"e2": "e2"},
				"dataway-tags":  map[string]string{"host": "hostMock", "h2": "h2", "e2": "e2"},
			},
			wantDumpHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
			wantDumpElectionTags: map[string]string{"e2": "e2"},
		},
		{
			name: "ele=false delete",
			args: args{
				globalElection:     false,
				isDocker:           false,
				hostName:           "hostMock",
				globalHostTags:     map[string]string{"h1": "h1", "h2": "h2"},
				globalElectionTags: map[string]string{"e1": "e1", "e2": "e2"},
				// any is error
				query: map[string][]string{"tags": ([]string{"host", "h1", "e1"})},
			},
			wantStatusCode: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handle := &mockHandle{args: tt.args}
			handle.initCfgAndTags()
			newHandle = func() handler {
				return handle
			}

			str := "/v1/global/election/tags?" + getQueryString(tt.args.query)
			req, err := http.NewRequest("DELETE", ts.URL+str, nil)
			if err != nil {
				t.Error(err)
			}

			c := &http.Client{}

			resp, err := c.Do(req)
			if err != nil {
				t.Error(err)
			}
			defer resp.Body.Close()

			// check StatusCode
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode)
			if resp.StatusCode != 200 {
				return
			}

			// check dump datakit.conf
			h, e := handle.getTags()
			assert.Equal(t, tt.wantDumpHostTags, h)
			assert.Equal(t, tt.wantDumpElectionTags, e)

			// check response
			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			resTags := map[string](map[string]string){}
			err = json.Unmarshal(respBody, &resTags)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantTags, resTags)
		})
	}
}

type mockHandle struct {
	hostTags     map[string]string
	electionTags map[string]string
	args         args
}

func (m *mockHandle) getDuplicateCfg() (*config.Config, bool) {
	c := config.DefaultConfig()
	c.Election.Enable = m.args.globalElection
	c.GlobalHostTags = internal.CopyMapString(m.args.globalHostTags)
	c.Election.Tags = internal.CopyMapString(m.args.globalElectionTags)

	return c, true
}

func (m *mockHandle) dumpMainCfgTOML(c *config.Config) {
	m.hostTags = c.GlobalHostTags
	m.electionTags = c.Election.Tags
}

func (m *mockHandle) initCfgAndTags() {
	datakit.ClearGlobalTags()
	config.Cfg = config.DefaultConfig()
	config.Cfg.Election.Enable = m.args.globalElection
	config.Cfg.Hostname = m.args.hostName

	// set config.Cfg and datakit global tags
	config.Cfg.GlobalHostTags = internal.CopyMapString(m.args.globalHostTags)
	config.Cfg.GlobalHostTags["host"] = config.Cfg.Hostname
	datakit.SetGlobalHostTagsByMap(config.Cfg.GlobalHostTags)
	if config.Cfg.Election.Enable {
		config.Cfg.Election.Tags = internal.CopyMapString(m.args.globalElectionTags)
		datakit.SetGlobalElectionTagsByMap(m.args.globalElectionTags)
	} else {
		config.Cfg.Election.Tags = internal.MergeMapString(config.Cfg.GlobalHostTags, m.args.globalElectionTags)
		datakit.SetGlobalElectionTagsByMap(config.Cfg.GlobalHostTags)
	}

	// set dataway global tags
	config.Cfg.Dataway.UpdateGlobalTags(internal.MergeMapString(config.Cfg.GlobalHostTags, config.Cfg.Election.Tags))
}

func (m *mockHandle) getTags() (map[string]string, map[string]string) {
	return m.hostTags, m.electionTags
}

func getQueryString(m map[string][]string) string {
	s := ""
	for k, v := range m {
		s += "&" + k + "="
		for i, vv := range v {
			if i != 0 {
				s += ","
			}
			s += vv
		}
	}
	s = strings.TrimPrefix(s, "&")
	return s
}

func Test_getQueryString(t *testing.T) {
	tests := []struct {
		name string
		m    map[string][]string
		want string
	}{
		{
			name: "01",
			m:    map[string][]string{"h2": ([]string{"h22"}), "h3": ([]string{"h3"})},
			want: "h2=h22&h3=h3",
		},
		{
			name: "02",
			m:    map[string][]string{"tags": ([]string{"e2", "e3"})},
			want: "tags=e2,e3",
		},
		{
			name: "03",
			m:    map[string][]string{"tags": ([]string{"e2", "e3"}), "h3": ([]string{"h3"})},
			want: "tags=e2,e3&h3=h3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getQueryString(tt.m); got != tt.want {
				t.Errorf("name = %s ,getQueryString() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
