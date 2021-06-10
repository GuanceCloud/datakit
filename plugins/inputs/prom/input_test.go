package prom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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
		_, err := PromText2Metrics(promText, prom)
		assert.NoError(t, err)
	})

	t.Run("name filter", func(t *testing.T) {
		prom := &Input{
			MetricNameFilter: []string{"^go_.*"},
		}
		points, err := PromText2Metrics(promText, prom)
		assert.NoError(t, err)
		assert.Greater(t, len(points), 0)
		for _, point := range points {
			p, err := point.LineProto()
			assert.NoError(t, err)
			name := p.Name()
			assert.Equal(t, name, "go")
		}
	})

	t.Run("metric types filter", func(t *testing.T) {
		prom := &Input{
			MetricTypes: []string{"gauge"},
		}
		points, err := PromText2Metrics(promText, prom)
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
		MetricNameFilter: []string{"node_network_transmit_packets_total"},
	}
	prom.client = &http.Client{}
	prom.client.Transport = newTransportMock(promText)
	err := prom.Collect()

	assert.NoError(t, err)
	assert.Equal(t, len(prom.collectCache), 4)

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
