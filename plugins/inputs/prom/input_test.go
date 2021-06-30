package prom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
)

type transportMock struct {
	statusCode int
	body       string
}

func (t *transportMock) RoundTrip(r *http.Request) (*http.Response, error) {
	res := &http.Response{
		Header:     make(http.Header),
		Request:    r,
		StatusCode: t.statusCode,
	}
	res.Header.Set("Content-Type", "application/json")
	res.Body = ioutil.NopCloser(strings.NewReader(t.body))
	return res, nil
}

func (t *transportMock) CancelRequest(_ *http.Request) {
}

func newTransportMock(body string) http.RoundTripper {
	return &transportMock{
		statusCode: http.StatusOK,
		body:       body,
	}
}

func TestInput(t *testing.T) {
	input := NewProm(sampleCfg)

	assert.Equal(t, input.Catalog(), catalogName)
	assert.Equal(t, input.SampleConfig(), sampleCfg)
	assert.Equal(t, input.AvailableArchs(), datakit.AllArch)

	_, err := input.createHTTPClient()
	assert.NoError(t, err)

	m := Measurement{}
	info := m.Info()

	assert.Equal(t, info.Name, inputName)
}

func TestPromText2Metrics(t *testing.T) {
	var mockProm = make(map[string]MockMetric)
	json.Unmarshal([]byte(promJsonStr), &mockProm)

	t.Run("should work", func(t *testing.T) {
		prom := &Input{
			MetricTypes:       []string{},
			MetricNameFilter:  []string{},
			MeasurementPrefix: "",
			Measurements: []Rule{
				Rule{
					Pattern: "node",
					Name:    "node-network",
				},
			},
		}
		_, err := PromText2Metrics(promText, prom, map[string]string{})
		assert.NoError(t, err)
	})

	t.Run("name filter", func(t *testing.T) {
		prom := &Input{
			MetricNameFilter: []string{"^go_.*"},
		}
		points, err := PromText2Metrics(promText, prom, map[string]string{})
		assert.NoError(t, err)
		assert.Greater(t, len(points), 0)
		for _, point := range points {
			p, err := point.LineProto()
			assert.NoError(t, err)
			name := p.Name()
			assert.Equal(t, name, "go")
		}
	})

	t.Run("Inf", func(t *testing.T) {
		text := `
# HELP rest_client_exec_plugin_ttl_seconds [ALPHA] Gauge of the shortest TTL (time-to-live) of the client certificate(s) managed by the auth exec plugin. The value is in seconds until certificate expiry (negative if already expired). If auth exec plugins are unused or manage no TLS certificates, the value will be +INF.
# TYPE rest_client_exec_plugin_ttl_seconds gauge
rest_client_exec_plugin_ttl_seconds +Inf
		`
		prom := &Input{
			MetricNameFilter: []string{"rest_client_exec_plugin_ttl_seconds"},
		}
		points, err := PromText2Metrics(text, prom, map[string]string{})
		assert.NoError(t, err)
		assert.Empty(t, points)
	})

	t.Run("metric types filter", func(t *testing.T) {
		prom := &Input{
			MetricTypes: []string{"gauge"},
		}
		points, err := PromText2Metrics(promText, prom, map[string]string{})
		assert.NoError(t, err)
		assert.NotEmpty(t, points)
		for _, point := range points {
			p, err := point.LineProto()
			assert.NoError(t, err)
			name := p.Name()
			fields, err := p.Fields()
			assert.NoError(t, err)

			for fName, _ := range fields {
				oriName := fmt.Sprintf("%v_%v", name, fName)
				originProm, ok := mockProm[oriName]
				assert.True(t, ok)
				assert.Equal(t, originProm.Type, "gauge")
			}
		}
	})

}

func TestCollect(t *testing.T) {
	prom := &Input{
		URL:              "http://xxxxx",
		MetricNameFilter: []string{"node_network_transmit_packets_total"},
	}
	prom.client = &http.Client{}
	prom.client.Transport = newTransportMock(promText)
	err := prom.Collect()

	assert.NoError(t, err)
	assert.Equal(t, len(prom.collectCache), 4)

	t.Run("when url is array", func(t *testing.T) {
		prom := &Input{
			URL:              []string{"http://1", "http://2"},
			MetricNameFilter: []string{"node_network_transmit_packets_total"},
		}

		prom.client = &http.Client{}
		prom.client.Transport = newTransportMock(promText)
		err := prom.Collect()
		assert.NoError(t, err)
	})

	t.Run("when error occurr", func(t *testing.T) {
		prom := NewProm(sampleCfg)
		prom.URL = "invalid"
		prom.client = &http.Client{}

		err := prom.Collect()
		assert.Error(t, err)

		prom.URL = "\n\r"
		err = prom.Collect()
		assert.Error(t, err)

		prom.URL = "localhost"
		prom.client.Transport = newTransportMock("xxxxxxxxx")
		err = prom.Collect()
		assert.Error(t, err)
	})
}

func testK8sFile(jsonContent string, prom *Input, t *testing.T) {
	f := "/tmp/__test__.json"
	err := ioutil.WriteFile(f, []byte(jsonContent), 0666)
	assert.NoError(t, err)

	defer func() {
		err := os.Remove(f)
		assert.NoError(t, err)
	}()

	prom.URL = f

	prom.client = &http.Client{}
	prom.client.Transport = newTransportMock(promText)

	err = prom.Collect()

	assert.NoError(t, err)

}

func TestK8s(t *testing.T) {
	prom := &Input{}
	jsonContent := `
	[
		{
			"pod": "dummy-exporter-deployment-f59677c-577wr",
			"namespace": "default",
			"status": "Running",
			"podIp": "10.1.0.175",
			"labels": {
				"app": "dummy-exporter",
				"datakit": "prom-dev",
				"pod-template-hash": "f59677c"
			},
			"nodeName": "df-idc-qa-001",
			"targets": [
				"http://10.1.0.175/metric"
			]
		},
		{
			"pod": "dummy-exporter-deployment-f59677c-b8jpw",
			"namespace": "default",
			"status": "Running",
			"podIp": "10.1.0.169",
			"labels": {
				"app": "dummy-exporter",
				"datakit": "prom-dev",
				"pod-template-hash": "f59677c"
			},
			"nodeName": "df-idc-qa-001",
			"targets": [
				"http://10.1.0.169/metric"
			]
		},
		{
			"pod": "dummy-exporter-deployment-f59677c-rlzww",
			"namespace": "default",
			"status": "Running",
			"podIp": "10.1.0.150",
			"labels": {
				"app": "dummy-exporter",
				"datakit": "prom-dev",
				"pod-template-hash": "f59677c"
			},
			"nodeName": "df-idc-qa-001",
			"targets": [
				"http://10.1.0.150/metric"
			]
		}
	]
	`
	prom.MetricNameFilter = []string{"node_network_transmit_packets_untyped"}

	testK8sFile(jsonContent, prom, t)

	assert.Equal(t, len(prom.collectCache), 3)

	t.Run("should ignore status not Running", func(t *testing.T) {
		prom := &Input{}
		jsonContent := `
		[
			{
				"pod": "dummy-exporter-deployment-f59677c-577wr",
				"namespace": "default",
				"status": "Running",
				"podIp": "10.1.0.175",
				"labels": {
					"app": "dummy-exporter",
					"datakit": "prom-dev",
					"pod-template-hash": "f59677c"
				},
				"nodeName": "df-idc-qa-001",
				"targets": [
					"http://10.1.0.175/metric"
				]
			},
			{
				"pod": "dummy-exporter-deployment-f59677c-b8jpw",
				"namespace": "default",
				"status": "Stopped",
				"podIp": "10.1.0.169",
				"labels": {
					"app": "dummy-exporter",
					"datakit": "prom-dev",
					"pod-template-hash": "f59677c"
				},
				"nodeName": "df-idc-qa-001",
				"targets": [
					"http://10.1.0.169/metric"
				]
			},
			{
				"pod": "dummy-exporter-deployment-f59677c-rlzww",
				"namespace": "default",
				"status": "Stopped",
				"podIp": "10.1.0.150",
				"labels": {
					"app": "dummy-exporter",
					"datakit": "prom-dev",
					"pod-template-hash": "f59677c"
				},
				"nodeName": "df-idc-qa-001",
				"targets": [
					"http://10.1.0.150/metric"
				]
			}
		]
	`
		prom.MetricNameFilter = []string{"node_network_transmit_packets_untyped"}
		testK8sFile(jsonContent, prom, t)
		assert.Equal(t, len(prom.collectCache), 1)

		m := prom.collectCache[0]
		p, err := m.LineProto()
		assert.NoError(t, err)
		pTags := p.Tags()
		assert.Equal(t, "dummy-exporter-deployment-f59677c-577wr", pTags["pod"])
		assert.Equal(t, "f59677c", pTags["pod-template-hash"])
	})

	t.Run("ignore tags", func(t *testing.T) {
		prom := &Input{}
		jsonContent := `
		[
			{
				"pod": "dummy-exporter-deployment-f59677c-577wr",
				"namespace": "default",
				"status": "Running",
				"podIp": "10.1.0.175",
				"labels": {
					"app": "dummy-exporter",
					"datakit": "prom-dev",
					"pod-template-hash": "f59677c"
				},
				"nodeName": "df-idc-qa-001",
				"targets": [
					"http://10.1.0.175/metric"
				]
			}
		]
		`
		ignoreTag := "pod-template-hash"
		prom.MetricNameFilter = []string{"node_network_transmit_packets_untyped"}
		prom.TagsIgnore = []string{ignoreTag}
		testK8sFile(jsonContent, prom, t)
		assert.Equal(t, 1, len(prom.collectCache))
		m := prom.collectCache[0]
		p, err := m.LineProto()
		assert.NoError(t, err)
		_, ok := p.Tags()[ignoreTag]
		assert.False(t, ok)

		_, ok = p.Tags()["app"]
		assert.True(t, ok)
	})
}
