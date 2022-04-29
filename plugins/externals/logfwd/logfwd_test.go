//go:build linux
// +build linux

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForwardFunc(t *testing.T) {
	lg := &logging{
		Source: "t_source",
		Tags:   map[string]string{"service": "t_service"},
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
			out:        `{"type":"1","source":"t_source","tags":{"filename":"/tmp/111","service":"t_service"},"log":"hello,world"}`,
		},
		{
			inFilename: "/tmp/222",
			inText:     "hello,world,222",
			out:        `{"type":"1","source":"t_source","tags":{"filename":"/tmp/222","service":"t_service"},"log":"hello,world,222"}`,
		},
		{
			inFilename: "/tmp/333",
			inText:     "hello,world,333",
			out:        `{"type":"1","source":"t_source","tags":{"filename":"/tmp/333","service":"t_service"},"log":"hello,world,333"}`,
		},
	}

	for idx, tc := range cases {
		forwardFunc(lg, write)(tc.inFilename, tc.inText)
		assert.Equal(t, tc.out, writeRevicer[idx])
	}
}

func TestGetConfig(t *testing.T) {
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
	_, err = getConfig()
	assert.Error(t, err)

	// command arg file
	// level 3
	argConfig = &fname
	c3, err := getConfig()
	assert.NoError(t, err)

	// command arg
	// level 2
	argJSONConfig = &cfgstr
	c2, err := getConfig()
	assert.NoError(t, err)

	// env
	// level 1
	envMainJSONConfig = cfgstr
	c1, err := getConfig()
	assert.NoError(t, err)

	assert.Equal(t, c1, c2)
	assert.Equal(t, c1, c3)
	assert.Equal(t, c2, c3)
}

func TestParseConfig(t *testing.T) {
	cases := []struct {
		in   string
		out  config
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
			out: config{
				DataKitAddr: "192.168.0.20:9533",
				Loggings: loggings{
					{
						LogFiles:       []string{"/tmp/11", "/tmp/22"},
						Source:         "t_source",
						Service:        "t_service_new",
						Pipeline:       "t_pipeline_new",
						MultilineMatch: "t_multiline_new",
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
			// invalid config
			in:   ``,
			fail: true,
		},
	}

	wsHost = "192.168.0.20"
	wsPort = "9533"
	envAnnotationDataKitLogs = `
		[
		  {
		    "source":"t_source",
		    "service":"t_service_new",
		    "pipeline":"t_pipeline_new",
		    "multiline_match":"t_multiline_new"
		  }
		]
	`

	for _, tc := range cases {
		config, err := parseConfig(tc.in)
		if tc.fail && assert.Error(t, err) {
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, tc.out.DataKitAddr, config.DataKitAddr)

		for idx, cfg := range config.Loggings {
			assert.Equal(t, tc.out.Loggings[idx].Source, cfg.Source)
			assert.Equal(t, tc.out.Loggings[idx].Service, cfg.Service)
			assert.Equal(t, tc.out.Loggings[idx].Pipeline, cfg.Pipeline)
			assert.Equal(t, tc.out.Loggings[idx].MultilineMatch, cfg.MultilineMatch)
		}
	}
}

func TestLoggingMerge(t *testing.T) {
	cases := []struct {
		dst *logging
		src loggings
		res *logging
	}{
		{
			dst: &logging{
				LogFiles:       []string{"/tmp/11", "/tmp/22"},
				Source:         "t_source",
				Service:        "t_service",
				Pipeline:       "t_pipeline",
				MultilineMatch: "t_multiline",
			},
			src: loggings{
				{
					Source:         "t_source",
					Service:        "t_service2",
					Pipeline:       "t_pipeline2",
					MultilineMatch: "t_multiline2",
				},
			},
			res: &logging{
				LogFiles:       []string{"/tmp/11", "/tmp/22"},
				Source:         "t_source",
				Service:        "t_service2",
				Pipeline:       "t_pipeline2",
				MultilineMatch: "t_multiline2",
			},
		},
		{
			dst: &logging{
				LogFiles:       []string{"/tmp/11", "/tmp/22"},
				Source:         "t_source",
				Service:        "t_service",
				Pipeline:       "t_pipeline",
				MultilineMatch: "t_multiline",
			},
			src: loggings{
				{
					Source:         "t_source2",
					Service:        "t_service2",
					Pipeline:       "t_pipeline2",
					MultilineMatch: "t_multiline2",
				},
			},
			res: &logging{
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

func TestLoggingSetup(t *testing.T) {
	podName = "test_pod_name"
	podNamespace = "test_pod_namespace"

	cases := []struct {
		in  *logging
		out *logging
	}{
		{
			in: &logging{
				Source: "t_source",
			},
			out: &logging{
				Source:  "t_source",
				Service: "t_source", // use $source
				Tags: map[string]string{
					"pod_namespace": "test_pod_namespace",
					"pod_name":      "test_pod_name",
					"service":       "t_source",
				},
			},
		},
		{
			in: &logging{
				Service: "t_service",
			},
			out: &logging{
				Source:  "default", // use default
				Service: "t_service",
				Tags: map[string]string{
					"pod_namespace": "test_pod_namespace",
					"pod_name":      "test_pod_name",
					"service":       "t_service",
				},
			},
		},
		{
			in: &logging{},
			out: &logging{
				Source:  "default", // use default
				Service: "default",
				Tags: map[string]string{
					"pod_namespace": "test_pod_namespace",
					"pod_name":      "test_pod_name",
					"service":       "default",
				},
			},
		},
	}

	for _, tc := range cases {
		tc.in.setup()
		assert.Equal(t, tc.out, tc.in)
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
