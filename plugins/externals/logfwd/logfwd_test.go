// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	gws "github.com/gobwas/ws"
	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/network/ws"
)

func TestForwardFunc(t *testing.T) {
	lc := &logConfig{
		Source:  "t_source",
		TagsStr: "service=t_service",
	}

	var writeRevicer []string

	write := func(b []byte) error {
		writeRevicer = append(writeRevicer, string(b))
		return nil
	}

	cases := []struct {
		inFilename, inText string
		out                string
	}{
		{
			inFilename: "/tmp/111",
			inText:     "hello,world",
			out:        `{"type":"1","source":"t_source","tags_str":"filename=/tmp/111,service=t_service","log":"hello,world"}`,
		},
		{
			inFilename: "/tmp/222",
			inText:     "hello,world,222",
			out:        `{"type":"1","source":"t_source","tags_str":"filename=/tmp/222,service=t_service","log":"hello,world,222"}`,
		},
		{
			inFilename: "/tmp/333",
			inText:     "hello,world,333",
			out:        `{"type":"1","source":"t_source","tags_str":"filename=/tmp/333,service=t_service","log":"hello,world,333"}`,
		},
	}

	for idx, tc := range cases {
		forwardFunc(lc, write)(tc.inFilename, tc.inText)
		assert.Equal(t, tc.out, writeRevicer[idx])
	}
}

func TestGetFwdConfig(t *testing.T) {
	cfgstr := `
		[
		  {
		    "datakit_addr":"192.168.0.10:8080",
		    "loggings": [
		      {
		        "logfiles": ["/tmp/11", "/tmp/22"],
		        "source":"t_source",
		        "service":"t_service",
		        "pipeline":"t_pipeline",
		        "multiline_match":"t_multiline"
		      }
		    ]
		  }
		]
	`
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)

	f.WriteString(cfgstr)
	fname := f.Name()

	defer func() {
		err := f.Close()
		assert.NoError(t, err)

		err = os.Remove(fname)
		assert.NoError(t, err)
	}()

	// fail
	_, err = getFwdConfig()
	assert.Error(t, err)

	// command arg file
	// level 3
	argConfig = &fname
	c3, err := getFwdConfig()
	assert.NoError(t, err)

	// command arg
	// level 2
	argConfigJSON = &cfgstr
	c2, err := getFwdConfig()
	assert.NoError(t, err)

	// env
	// level 1
	err = os.Setenv(envFwdConfigKey, cfgstr)
	assert.NoError(t, err)
	c1, err := getFwdConfig()
	assert.NoError(t, err)
	os.Unsetenv(envFwdConfigKey)

	assert.Equal(t, c1, c2)
	assert.Equal(t, c1, c3)
	assert.Equal(t, c2, c3)
}

func TestParseFwdConfig(t *testing.T) {
	cases := []struct {
		in   string
		out  fwdConfig
		fail bool
	}{
		{
			in: `
				[
				  {
				    "datakit_addr":"192.168.0.10:8080",
				    "loggings": [
				      {
				        "logfiles": ["/tmp/11", "/tmp/22"],
				        "source":"t_source",
				        "service":"t_service",
				        "pipeline":"t_pipeline",
				        "multiline_match":"t_multiline"
				      }
				    ]
				  }
				]
			`,
			out: fwdConfig{
				DataKitAddr: "192.168.0.20:9090",
				LogConfigs: logConfigs{
					{
						LogFiles:       []string{"/tmp/11", "/tmp/22"},
						Source:         "t_source",
						Service:        "t_service",
						Pipeline:       "t_pipeline",
						MultilineMatch: "t_multiline",
					},
				},
			},
		},

		{
			// invalid fwd config
			in:   `[]`,
			fail: true,
		},

		{
			// invalid fwdConfig
			in:   ``,
			fail: true,
		},
	}

	err := os.Setenv(envWsHostKey, "192.168.0.20")
	assert.NoError(t, err)
	err = os.Setenv(envWsPortKey, "9090")
	assert.NoError(t, err)
	defer func() {
		err := os.Unsetenv(envWsHostKey)
		assert.NoError(t, err)
		err = os.Unsetenv(envWsPortKey)
		assert.NoError(t, err)
	}()

	for _, tc := range cases {
		config, err := parseFwdConfig(tc.in)
		if tc.fail && assert.Error(t, err) {
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, tc.out.DataKitAddr, config.DataKitAddr)
		assert.Equal(t, tc.out.LogPath, config.LogPath)
		assert.Equal(t, tc.out.LogLevel, config.LogLevel)

		for idx, cfg := range config.LogConfigs {
			assert.Equal(t, tc.out.LogConfigs[idx].Source, cfg.Source)
			assert.Equal(t, tc.out.LogConfigs[idx].Service, cfg.Service)
			assert.Equal(t, tc.out.LogConfigs[idx].Pipeline, cfg.Pipeline)
			assert.Equal(t, tc.out.LogConfigs[idx].MultilineMatch, cfg.MultilineMatch)
		}
	}
}

func TestGetEnvLogConfigs(t *testing.T) {
	var (
		envKey = envLogConfigKey

		cases = []struct {
			in  string
			out logConfigs
		}{
			{
				in: `
					[
					  {
					    "source":"t_source",
					    "service":"t_service",
					    "pipeline":"t_pipeline",
					    "multiline_match":"t_multiline"
					  }
					]
				`,
				out: logConfigs{
					{
						Source:         "t_source",
						Service:        "t_service",
						Pipeline:       "t_pipeline",
						MultilineMatch: "t_multiline",
					},
				},
			},
			{
				in: `
				        ## failed to unmarshal
					[
					  {
					    "source":"t_source",
					    "service":"t_service",
					    "pipeline":"t_pipeline",
					    "multiline_match":"t_multiline"
					  }
					]
				`,
				out: nil,
			},
		}
	)

	for _, tc := range cases {
		err := os.Setenv(envKey, tc.in)
		assert.NoError(t, err)

		configs := getEnvLogConfigs(envKey)
		assert.Equal(t, tc.out, configs)

		err = os.Unsetenv(envKey)
		assert.NoError(t, err)
	}
}

func TestLogConfigMerge(t *testing.T) {
	cases := []struct {
		dst *logConfig
		src logConfigs
		res *logConfig
	}{
		{
			dst: &logConfig{
				LogFiles:       []string{"/tmp/11", "/tmp/22"},
				Source:         "t_source",
				Service:        "t_service",
				Pipeline:       "t_pipeline",
				MultilineMatch: "t_multiline",
			},
			src: logConfigs{
				{
					Source:         "t_source",
					Service:        "t_service2",
					Pipeline:       "t_pipeline2",
					MultilineMatch: "t_multiline2",
				},
			},
			res: &logConfig{
				LogFiles:       []string{"/tmp/11", "/tmp/22"},
				Source:         "t_source",
				Service:        "t_service2",
				Pipeline:       "t_pipeline2",
				MultilineMatch: "t_multiline2",
			},
		},
		{
			dst: &logConfig{
				LogFiles:       []string{"/tmp/11", "/tmp/22"},
				Source:         "t_source",
				Service:        "t_service",
				Pipeline:       "t_pipeline",
				MultilineMatch: "t_multiline",
			},
			src: logConfigs{
				{
					Source:         "t_source2",
					Service:        "t_service2",
					Pipeline:       "t_pipeline2",
					MultilineMatch: "t_multiline2",
				},
			},
			res: &logConfig{
				LogFiles:       []string{"/tmp/11", "/tmp/22"},
				Source:         "t_source",
				Service:        "t_service",
				Pipeline:       "t_pipeline",
				MultilineMatch: "t_multiline",
			},
		},
	}

	for _, tc := range cases {
		tc.dst.merge(tc.src)
		assert.Equal(t, tc.dst, tc.res)
	}
}

func TestLogConfigSetup(t *testing.T) {
	err := os.Setenv(envPodNameKey, "test_pod_name")
	assert.NoError(t, err)
	err = os.Setenv(envPodNamespaceKey, "test_pod_namespace")
	assert.NoError(t, err)
	defer func() {
		err := os.Unsetenv(envPodNameKey)
		assert.NoError(t, err)
		err = os.Unsetenv(envPodNamespaceKey)
		assert.NoError(t, err)
	}()

	cases := []struct {
		in  *logConfig
		out *logConfig
	}{
		{
			in: &logConfig{
				Source: "t_source",
			},
			out: &logConfig{
				Source:  "t_source",
				Service: "t_source", // use $source
				TagsStr: "pod_namespace=test_pod_namespace,pod_name=test_pod_name,service=t_source",
			},
		},
		{
			in: &logConfig{
				Service: "t_service",
			},
			out: &logConfig{
				Source:  "default", // use default
				Service: "t_service",
				TagsStr: "pod_namespace=test_pod_namespace,pod_name=test_pod_name,service=t_service",
			},
		},
		{
			in: &logConfig{},
			out: &logConfig{
				Source:  "default", // use default
				Service: "default",
				TagsStr: "pod_namespace=test_pod_namespace,pod_name=test_pod_name,service=default",
			},
		},
	}

	for _, tc := range cases {
		tc.in.setup()
		assert.NoError(t, err)
		assert.Equal(t, tc.out, tc.in)
	}
}

func TestAddTag(t *testing.T) {
	cases := []struct {
		in             *message
		inKey, inValue string
		out            string
	}{
		{
			in:      &message{},
			inKey:   "key",
			inValue: "value",
			out:     "key=value",
		},
		{
			in: &message{
				TagsStr: "key1=value1",
			},
			inKey:   "key2",
			inValue: "value2",
			out:     "key2=value2,key1=value1",
		},
	}

	for _, tc := range cases {
		out := tc.in.appendToTagsStr(tc.inKey, tc.inValue)
		assert.Equal(t, tc.out, out)
	}
}

func TestMessageToJSON(t *testing.T) {
	cases := []struct {
		in  *message
		out []byte
	}{
		{
			in: &message{
				Type:   "1",
				Source: "source1",
				Log:    "log1",
			},
			out: []byte(`{"type":"1","source":"source1","log":"log1"}`),
		},
	}

	for _, tc := range cases {
		out, err := tc.in.json()
		assert.NoError(t, err)
		assert.Equal(t, tc.out, out)
	}
}

func TestStartLog(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)

	defer func() {
		err := f.Close()
		assert.NoError(t, err)

		err = os.Remove(f.Name())
		assert.NoError(t, err)
	}()

	const addr = "0.0.0.0:9090"

	cfg := &fwdConfig{
		DataKitAddr: addr,
		LogConfigs: logConfigs{
			{
				LogFiles: []string{f.Name()},
				Source:   "t_source",
				Service:  "t_service",
				Pipeline: "t_pipeline",
			},
		},
	}

	go func() {
		srv, err := ws.NewServer(addr, "/logfwd")
		assert.NoError(t, err)
		defer srv.Stop()

		srv.MsgHandler = func(s *ws.Server, c net.Conn, data []byte, op gws.OpCode) error {
			t.Logf("ws-reciver: %s", string(data))
			return nil
		}
		srv.AddCli = func(w http.ResponseWriter, r *http.Request) {
			conn, _, _, err := gws.UpgradeHTTP(r, w)
			if err != nil {
				l.Error("ws.UpgradeHTTP error: %s", err.Error())
				return
			}

			if err := srv.AddConnection(conn); err != nil {
				l.Error(err)
			}
		}

		go srv.Start()
		<-time.Tick(time.Second * 10)
	}()

	go func() {
		idx := 0
		for {
			select {
			case <-time.Tick(time.Second * 10):
				return
			default:
				// nil
			}
			f.WriteString(fmt.Sprintf("[%s] logfwd - %d\n", time.Now(), idx))
			idx++
			time.Sleep(time.Second)
		}
	}()

	quitChannel := make(chan struct{})
	go startLog(cfg, quitChannel)

	<-time.Tick(time.Second * 11)
	quitChannel <- struct{}{}
}
