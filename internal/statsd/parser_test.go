// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package statsd

import (
	"bytes"
	"os"
	"strings"
	T "testing"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseName(t *T.T) {
	cases := []struct {
		in, out, sep string
	}{
		{
			in:  `jvm.non_heap_memory_max`,
			sep: "_",
			out: "jvm_non_heap_memory_max",
		},

		{
			in:  `jvm.cpu_load.process`,
			sep: "_",
			out: "jvm_cpu_load_process",
		},

		{
			in:  `jvm.buffer_pool.direct.capacity`,
			sep: "_",
			out: "jvm_buffer_pool_direct_capacity",
		},

		{
			in:  `us.west.cpu.load`,
			sep: "_",
			out: "us_west_cpu_load",
		},
	}
	opt := option{}
	s := &Collector{opts: &opt}
	s.Templates = []string{}

	for _, tc := range cases {
		s.opts.metricSeparator = tc.sep

		name, fields, tags := s.parseName(tc.in)
		t.Logf("%s => name: %s, fields: %+#v, tags: %+#v", tc.in, name, fields, tags)
	}
}

func TestParseLine(t *T.T) {
	t.Run(`with-time(|T)`, func(t *T.T) {
		line := "namespace.test_gauge:21|g|#globalTags,globalTags2,tag1,tag2|T1234"
		col, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
		require.NoError(t, err)

		col.doJob(0, &job{
			Buffer: bytes.NewBuffer([]byte(line)),
			Time:   time.Unix(123, 0),
			Addr:   "1.2.3.4:4321",
		})
		pts, err := col.GetPoints()
		assert.NoError(t, err)

		require.Len(t, pts, 1)
		t.Logf("%s", pts[0].Pretty())
	})

	t.Run("type-set", func(t *T.T) {
		line := `auth.logins:user-a1|s|#tenant:a
auth.logins:user-b1|s|#tenant:b
auth.logins:user-b2|s|#tenant:b
auth.logins:user-b3|s|#tenant:b
auth.logins:user-b4|s|#tenant:b
auth.logins:user-b5|s|#tenant:b
auth.logins:user-a2|s|#tenant:a
auth.logins:user-a1|s|#tenant:a`
		col, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
		require.NoError(t, err)

		col.doJob(0, &job{
			Buffer: bytes.NewBuffer([]byte(line)),
			Time:   time.Unix(123, 0),
			Addr:   "1.2.3.4:4321",
		})
		pts, err := col.GetPoints()
		assert.NoError(t, err)

		for _, pt := range pts {
			t.Logf("%s", pt.Pretty())
		}

		require.Len(t, pts, 2)

		// for tenant a, there are 2 logings(for user a1/a2), same user(user-a1) only count 1 time
		// for tenant b, there are 1 logings(for user b1~b5)
		// Map iteration order is non-deterministic, so we need to check both points
		var tenantAPoint, tenantBPoint *point.Point
		for _, pt := range pts {
			if tenant := pt.GetTag("tenant"); tenant == "a" {
				tenantAPoint = pt
			} else if tenant == "b" {
				tenantBPoint = pt
			}
		}
		require.NotNil(t, tenantAPoint)
		require.NotNil(t, tenantBPoint)
		assert.Equal(t, int64(2), tenantAPoint.Get("logins"))
		assert.Equal(t, int64(5), tenantBPoint.Get("logins"))
	})

	t.Run(`with-dd-tags`, func(t *T.T) {
		line := "namespace.test_gauge:21|g|#globalTags,tag1,some:value"
		col, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
		require.NoError(t, err)

		col.doJob(0, &job{
			Buffer: bytes.NewBuffer([]byte(line)),
			Time:   time.Unix(123, 0),
			Addr:   "1.2.3.4:4321",
		})
		pts, err := col.GetPoints()
		assert.NoError(t, err)

		assert.Len(t, pts, 1)

		assert.Equal(t, "", pts[0].GetTag("globalTags"))
		assert.Equal(t, "", pts[0].GetTag("tag1"))
		assert.Equal(t, float64(21), pts[0].Get("test_gauge"))
		assert.Equal(t, "value", pts[0].GetTag("some"))

		for _, pt := range pts {
			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run(`with-service-check(_sc)`, func(t *T.T) {
		line := `_sc|jmxfetch-config.can_connect|0
		_sc|Redis Reachable|2|#host:cache-01,port:6379,env:e1|m:Connection timed out after 3s|d:12345`
		col, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true), WithDataDogServiceChecks(true))
		require.NoError(t, err)

		col.doJob(0, &job{
			Buffer: bytes.NewBuffer([]byte(line)),
			Time:   time.Unix(123, 0),
			Addr:   "1.2.3.4:4321",
		})
		pts := col.GetLoggings()
		assert.NoError(t, err)

		assert.Len(t, pts, 2)

		assert.Equal(t, serviceCheckMeasurementName, pts[0].Name())
		assert.Equal(t, "jmxfetch-config.can_connect", pts[0].GetTag("check_name"))
		assert.Equal(t, "ok", pts[0].Get("status"))

		assert.Equal(t, serviceCheckMeasurementName, pts[1].Name())
		assert.Equal(t, "Redis Reachable", pts[1].GetTag("check_name"))
		assert.Equal(t, "e1", pts[1].GetTag("env"))
		assert.Equal(t, "cache-01", pts[1].GetTag("host"))
		assert.Equal(t, "6379", pts[1].GetTag("port"))
		assert.Equal(t, "critical", pts[1].Get("status"))
		assert.Equal(t, "Connection timed out after 3s", pts[1].Get("message"))
		assert.Equal(t, time.Unix(12345, 0), pts[1].Time())

		for _, pt := range pts {
			t.Logf("%s", pt.Pretty())
		}
	})

	t.Run(`dogstatsd-logging-disabled-by-default`, func(t *T.T) {
		line := `_e{4,3}:test:msg|d:123
_sc|jmxfetch-config.can_connect|0`
		col, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
		require.NoError(t, err)

		col.doJob(0, &job{
			Buffer: bytes.NewBuffer([]byte(line)),
			Time:   time.Unix(123, 0),
			Addr:   "1.2.3.4:4321",
		})

		assert.Empty(t, col.GetLoggings())
	})
}

func TestParsePcap(t *T.T) {
	// test K6 statsd payload
	t.Run("k6-timing", func(t *T.T) {
		payload := []byte(`k6.http_req_waiting:10.303787|ms|#method:GET,test_type:constant_rate_rps,scenario:constant_rps,name:http://10.15.0.69/api/detail,expected_response:true,status:200,proto:HTTP/1.1
k6.http_req_waiting:6.914392|ms|#scenario:constant_rps,proto:HTTP/1.1,method:GET,test_type:constant_rate_rps,status:200,name:http://10.15.0.69/api/detail,expected_response:true
k6.http_req_waiting:6.137893|ms|#scenario:constant_rps,method:GET,expected_response:true,test_type:constant_rate_rps,status:200,proto:HTTP/1.1,name:http://10.15.0.69/api/detail
k6.http_req_waiting:6.251319|ms|#test_type:constant_rate_rps,status:200,name:http://10.15.0.69/api/detail,method:GET,expected_response:true,scenario:constant_rps,proto:HTTP/1.1
k6.http_req_waiting:5.615076|ms|#test_type:constant_rate_rps,status:200,scenario:constant_rps,method:GET,expected_response:true,proto:HTTP/1.1,name:http://10.15.0.69/api/detail
k6.http_req_waiting:5.918090|ms|#test_type:constant_rate_rps,scenario:constant_rps,proto:HTTP/1.1,name:http://10.15.0.69/api/detail,method:GET,expected_response:true,status:200
k6.http_req_waiting:5.855758|ms|#method:GET,test_type:constant_rate_rps,status:200,scenario:constant_rps,expected_response:true,proto:HTTP/1.1,name:http://10.15.0.69/api/detail
k6.http_req_waiting:5.634932|ms|#test_type:constant_rate_rps,scenario:constant_rps,method:GET,expected_response:true,status:200,proto:HTTP/1.1,name:http://10.15.0.69/api/detail
k6.http_req_waiting:5.714860|ms|#scenario:constant_rps,proto:HTTP/1.1,method:GET,expected_response:true,test_type:constant_rate_rps,status:200,name:http://10.15.0.69/api/detail
k6.http_req_waiting:6.397057|ms|#test_type:constant_rate_rps,proto:HTTP/1.1,status:200,scenario:constant_rps,name:http://10.15.0.69/api/detail,method:GET,expected_response:true
k6.http_req_waiting:5.587364|ms|#status:200,method:GET,expected_response:true,test_type:constant_rate_rps,scenario:constant_rps,proto:HTTP/1.1,name:http://10.15.0.69/api/detail
k6.http_req_waiting:6.559719|ms|#status:200,proto:HTTP/1.1,method:GET,expected_response:true,test_type:constant_rate_rps,scenario:constant_rps,name:http://10.15.0.69/api/detail
k6.http_req_waiting:5.653568|ms|#status:200,scenario:constant_rps,proto:HTTP/1.1,name:http://10.15.0.69/api/detail,test_type:constant_rate_rps,method:GET,expected_response:true
k6.http_req_waiting:6.669798|ms|#test_type:constant_rate_rps,scenario:constant_rps,status:200,proto:HTTP/1.1,name:http://10.15.0.69/api/detail,method:GET,expected_response:true
k6.http_req_waiting:5.703746|ms|#method:GET,expected_response:true,status:200,scenario:constant_rps,test_type:constant_rate_rps,proto:HTTP/1.1,name:http://10.15.0.69/api/detail`)

		col, err := NewCollector(nil, nil, WithProtocol("udp"),
			WithDataDogExtensions(true),
			WithPercentiles([]float64{50, 90, 99}),
		)
		require.NoError(t, err)

		col.doJob(0, &job{
			Buffer: bytes.NewBuffer(payload),
			Time:   time.Unix(123, 0),
			Addr:   "1.2.3.4:4321",
		})

		pts, err := col.GetPoints()
		assert.NoError(t, err)
		for _, pt := range pts {
			t.Logf("%s", pt.Pretty())

			assert.NotEmpty(t, pt.GetTag("expected_response"))
			assert.NotEmpty(t, pt.GetTag("method"))
			assert.NotEmpty(t, pt.GetTag("name"))
			assert.NotEmpty(t, pt.GetTag("proto"))
			assert.NotEmpty(t, pt.GetTag("scenario"))
			assert.NotEmpty(t, pt.GetTag("status"))
			assert.NotEmpty(t, pt.GetTag("test_type"))

			keyExists := false
			for _, kv := range pt.KVs() {
				if strings.HasSuffix(kv.Key, "_percentile") {
					keyExists = true
				}

				if strings.HasPrefix(kv.Key, "http_req_waiting_") {
					keyExists = true
				}
			}

			assert.True(t, keyExists)

			assert.Equal(t, "k6", pt.Name())
		}
	})
}

func TestParseMetricTypes(t *T.T) {
	testCases := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "counter",
			line:     "test.counter:1|c",
			expected: "counter",
		},
		{
			name:     "gauge",
			line:     "test.gauge:42|g",
			expected: "gauge",
		},
		{
			name:     "set",
			line:     "test.set:user123|s",
			expected: "set",
		},
		{
			name:     "timing",
			line:     "test.timing:123.45|ms",
			expected: "timing",
		},
		{
			name:     "histogram",
			line:     "test.histogram:5|h",
			expected: "histogram",
		},
		{
			name:     "distribution",
			line:     "test.distribution:3.14|d",
			expected: "distribution",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			// Distributions require DataDog extensions and distributions to be enabled
			var c *Collector
			var err error
			if tc.name == "distribution" {
				c, err = NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true), WithDataDogDistributions(true))
			} else {
				c, err = NewCollector(nil, nil, WithProtocol("udp"))
			}
			require.NoError(t, err)

			err = c.parseStatsdLine(tc.line)
			assert.NoError(t, err)

			pts, err := c.GetPoints()
			assert.NoError(t, err)

			// Timing, histogram, and distribution generate multiple fields (mean, stddev, sum, etc.)
			// so we need to check that we have at least one point with the correct metric type
			assert.Greater(t, len(pts), 0, "Should have at least one point")

			// Find a point with the expected metric type
			found := false
			for _, pt := range pts {
				if pt.GetTag("metric_type") == tc.expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find a point with metric_type %s", tc.expected)
		})
	}
}

func TestParseMetricTypesWithSampleRates(t *T.T) {
	testCases := []struct {
		name         string
		line         string
		sampleRate   float64
		expectedType string
	}{
		{
			name:         "counter with sample rate",
			line:         "test.counter:10|c|@0.1",
			sampleRate:   0.1,
			expectedType: "counter",
		},
		{
			name:         "gauge with sample rate",
			line:         "test.gauge:42|g|@0.5",
			sampleRate:   0.5,
			expectedType: "gauge",
		},
		{
			name:         "timing with sample rate",
			line:         "test.timing:100|ms|@0.2",
			sampleRate:   0.2,
			expectedType: "timing",
		},
		{
			name:         "histogram with sample rate",
			line:         "test.histogram:25|h|@0.3",
			sampleRate:   0.3,
			expectedType: "histogram",
		},
		{
			name:         "distribution with sample rate",
			line:         "test.distribution:7.5|d|@0.4",
			sampleRate:   0.4,
			expectedType: "distribution",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			// Distributions require DataDog extensions and distributions to be enabled
			var c *Collector
			var err error
			if strings.Contains(tc.name, "distribution") {
				c, err = NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true), WithDataDogDistributions(true))
			} else {
				c, err = NewCollector(nil, nil, WithProtocol("udp"))
			}
			require.NoError(t, err)

			err = c.parseStatsdLine(tc.line)
			assert.NoError(t, err)

			pts, err := c.GetPoints()
			assert.NoError(t, err)

			// Timing, histogram, and distribution generate multiple fields
			// so we need to check that we have at least one point with the correct metric type
			assert.Greater(t, len(pts), 0, "Should have at least one point")

			// Find a point with the expected metric type
			found := false
			for _, pt := range pts {
				if pt.GetTag("metric_type") == tc.expectedType {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find a point with metric_type %s", tc.expectedType)
		})
	}
}

func TestParseDataDogTags(t *T.T) {
	testCases := []struct {
		name     string
		line     string
		expected map[string]string
	}{
		{
			name: "single tag",
			line: "test.metric:1|c|#env:production",
			expected: map[string]string{
				"env":         "production",
				"metric_type": "counter",
			},
		},
		{
			name: "multiple tags",
			line: "test.metric:42|g|#env:production,service:api,region:us-west",
			expected: map[string]string{
				"env":         "production",
				"service":     "api",
				"region":      "us-west",
				"metric_type": "gauge",
			},
		},
		{
			name: "tag without value",
			line: "test.metric:123|s|#debug,env:staging",
			expected: map[string]string{
				"debug":       "",
				"env":         "staging",
				"metric_type": "set",
			},
		},
		{
			name: "complex tag values",
			line: "test.metric:5.5|d|#version:1.2.3,instance:i-1234567890abcdef0",
			expected: map[string]string{
				"version":     "1.2.3",
				"instance":    "i-1234567890abcdef0",
				"metric_type": "distribution",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			// Distributions require DataDog extensions and distributions to be enabled
			var c *Collector
			var err error
			if strings.Contains(tc.line, "|d") {
				c, err = NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true), WithDataDogDistributions(true))
			} else {
				c, err = NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
			}
			require.NoError(t, err)

			err = c.parseStatsdLine(tc.line)
			assert.NoError(t, err)

			pts, err := c.GetPoints()
			assert.NoError(t, err)
			assert.Greater(t, len(pts), 0, "Should have at least one point")

			// Check the first point for tags
			tags := pts[0].MapTags()
			for key, expectedValue := range tc.expected {
				actualValue, ok := tags[key]
				assert.True(t, ok, "tag %s should exist", key)
				assert.Equal(t, expectedValue, actualValue, "tag %s value mismatch", key)
			}
		})
	}
}

func TestParseBucketWithTags(t *T.T) {
	testCases := []struct {
		name     string
		line     string
		expected map[string]string
	}{
		{
			name: "bucket with single tag",
			line: "app.requests,endpoint=/api/users:1|c",
			expected: map[string]string{
				"endpoint":    "/api/users",
				"metric_type": "counter",
			},
		},
		{
			name: "bucket with multiple tags",
			line: "db.query.time,table=users,operation=select,host=db1:15.5|ms",
			expected: map[string]string{
				"table":       "users",
				"operation":   "select",
				"host":        "db1",
				"metric_type": "timing",
			},
		},
		{
			name: "bucket with tags and DataDog tags",
			line: "cache.hit,cache=redis:1|c|#env:production,cluster:main",
			expected: map[string]string{
				"cache":       "redis",
				"env":         "production",
				"cluster":     "main",
				"metric_type": "counter",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			// Some tests need DataDog extensions enabled
			var c *Collector
			var err error
			if strings.Contains(tc.line, "#") {
				c, err = NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
			} else {
				c, err = NewCollector(nil, nil, WithProtocol("udp"))
			}
			require.NoError(t, err)

			err = c.parseStatsdLine(tc.line)
			assert.NoError(t, err)

			pts, err := c.GetPoints()
			assert.NoError(t, err)
			assert.Greater(t, len(pts), 0, "Should have at least one point")

			// Check first point for tags
			tags := pts[0].MapTags()
			for key, expectedValue := range tc.expected {
				actualValue, ok := tags[key]
				assert.True(t, ok, "tag %s should exist", key)
				assert.Equal(t, expectedValue, actualValue, "tag %s value mismatch", key)
			}
		})
	}
}

func TestParseAdditiveValues(t *T.T) {
	testCases := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "gauge with positive increment",
			line:     "test.gauge:+5|g",
			expected: "gauge",
		},
		{
			name:     "gauge with negative increment",
			line:     "test.gauge:-3|g",
			expected: "gauge",
		},
		{
			name:     "counter with positive increment",
			line:     "test.counter:+1|c",
			expected: "counter",
		},
		{
			name:     "counter with negative increment",
			line:     "test.counter:-1|c",
			expected: "counter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			c, err := NewCollector(nil, nil, WithProtocol("udp"))
			require.NoError(t, err)

			err = c.parseStatsdLine(tc.line)
			assert.NoError(t, err)

			pts, err := c.GetPoints()
			assert.NoError(t, err)
			assert.Greater(t, len(pts), 0, "Should have at least one point")

			// Find a point with the expected metric type
			found := false
			for _, pt := range pts {
				if pt.GetTag("metric_type") == tc.expected {
					found = true
					break
				}
			}
			assert.True(t, found, "Should find a point with metric_type %s", tc.expected)
		})
	}
}

func TestParseMalformedMetrics(t *T.T) {
	c, err := NewCollector(nil, nil, WithProtocol("udp"))
	require.NoError(t, err)

	testCases := []struct {
		name        string
		line        string
		expectError bool
	}{
		{
			name:        "missing colon separator",
			line:        "test.metric.1|c",
			expectError: false, // Should not error, just warn
		},
		{
			name:        "missing pipe separator",
			line:        "test.metric:1c",
			expectError: false, // Should not error, just warn
		},
		{
			name:        "unsupported metric type",
			line:        "test.metric:1|x",
			expectError: false, // Should not error, just warn
		},
		{
			name:        "invalid sample rate",
			line:        "test.metric:1|c|@invalid",
			expectError: false, // Should not error, just warn
		},
		{
			name:        "invalid numeric value",
			line:        "test.metric:invalid|g",
			expectError: true, // Should error
		},
		{
			name:        "additive value on unsupported type",
			line:        "test.metric:+5|s",
			expectError: true, // Should error for set type
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *T.T) {
			err := c.parseStatsdLine(tc.line)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_parseEventMessage(t *T.T) {
	c, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
	require.NoError(t, err)

	testCases := []struct {
		name   string
		line   string
		onErr  bool
		expect *point.Point
	}{
		{
			name: "simple event",
			line: "_e{4,3}:test:msg|d:123",
			expect: func() *point.Point {
				var kvs point.KVs
				kvs = kvs.Add("message", "msg").
					Add("title", "test").
					AddTag("priority", priorityNormal).
					Add("status", eventInfo)

				return point.NewPoint(eventMeasurementName, kvs, point.WithTimestamp(123*int64(time.Second)))
			}(),
		},

		{
			name:  "invalid-title-text",
			line:  "_e{4,3}:test:|d:123",
			onErr: true,
		},
		{
			name:  "invalid-title-text",
			line:  "_e{4}:test:|d:123",
			onErr: true,
		},

		{
			name: "complex event",
			line: "_e{17,38}:Deployment v2.1.0|Deployment of service 'api' succeeded.|d:1678886400|h:web-prod-05|k:deployment-api-prod|p:normal|s:jenkins|t:success|#env:prod,service:api,version:2.1.0",
			expect: func() *point.Point {
				var kvs point.KVs
				kvs = kvs.Add("message", "Deployment of service 'api' succeeded.").
					Add("title", "Deployment v2.1.0").
					AddTag("priority", priorityNormal).
					Add("status", "success").
					AddTag("service", "api").
					AddTag("env", "prod").
					AddTag("version", "2.1.0").
					AddTag("source_type_name", "jenkins").
					AddTag("aggregation_key", "deployment-api-prod").
					AddTag("host", "web-prod-05")

				return point.NewPoint(eventMeasurementName, kvs, point.WithTimestamp(1678886400*int64(time.Second)))
			}(),
		},
	}

	for _, tc := range testCases {
		col, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true), WithDataDogEvents(true))
		require.NoError(t, err)

		t.Run(tc.name, func(t *T.T) {
			err := c.parseEventMessage(time.Now(), tc.line, "127.0.0.1:8125")
			if tc.onErr {
				assert.Error(t, err)
				t.Logf("expect error: %s", err)
				return
			}

			assert.NoError(t, err)

			col.doJob(0, &job{
				Buffer: bytes.NewBuffer([]byte(tc.line)),
				Time:   time.Unix(123, 0),
			})

			pts := col.GetLoggings()
			assert.NoError(t, err)
			assert.Len(t, pts, 1)

			assert.Equal(t, tc.expect.Pretty(), pts[0].Pretty())

			for _, pt := range pts {
				t.Logf("%s", pt.Pretty())
			}
		})
	}
}

func BenchmarkParse(b *T.B) {
	b.Run("k6", func(b *T.B) {
		payload, err := os.ReadFile("testdata/k6.udp")
		assert.NoError(b, err)

		col, err := NewCollector(nil, nil, WithProtocol("udp"),
			WithDataDogExtensions(true))
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			col.doJob(0, &job{
				Buffer: bytes.NewBuffer(payload),
				Time:   time.Unix(123, 0),
				Addr:   "1.2.3.4:4321",
			})
			// pts, err := col.GetPoints()
		}
	})

	b.Run("individual metrics", func(b *T.B) {
		metrics := []string{
			"test.counter:1|c",
			"test.gauge:42|g",
			"test.set:user123|s",
			"test.timing:123.45|ms",
			"test.histogram:5|h",
			"test.distribution:3.14|d",
		}

		col, err := NewCollector(nil, nil, WithProtocol("udp"))
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, metric := range metrics {
				col.parseStatsdLine(metric)
			}
		}
	})

	b.Run("metrics with tags", func(b *T.B) {
		metric := "test.metric:1|c|#env:production,service:api,region:us-west"

		col, err := NewCollector(nil, nil, WithProtocol("udp"), WithDataDogExtensions(true))
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			col.parseStatsdLine(metric)
		}
	})

	b.Run("metrics with sample rates", func(b *T.B) {
		metric := "test.counter:10|c|@0.1"

		col, err := NewCollector(nil, nil, WithProtocol("udp"))
		require.NoError(b, err)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			col.parseStatsdLine(metric)
		}
	})
}
