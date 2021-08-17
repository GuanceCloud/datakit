package cmds

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/prom"
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

func TestGetPromInput(t *testing.T) {
	toml := `
[[inputs.prom]]
  ## Exporter 地址
  url = "http://127.0.0.1:9100/metrics"

  ## 指标类型过滤, 可选值为 counter, gauge, histogram, summary
  # 默认只采集 counter 和 gauge 类型的指标
  # 如果为空，则不进行过滤
  metric_types = []

  ## 指标名称过滤
  # 支持正则，可以配置多个，即满足其中之一即可
  # 如果为空，则不进行过滤
  # metric_name_filter = ["cpu"]

  ## 指标集名称前缀
  # 配置此项，可以给指标集名称添加前缀
  measurement_prefix = "prom_"

  ## 指标集名称
  # 默认会将指标名称以下划线"_"进行切割，切割后的第一个字段作为指标集名称，剩下字段作为当前指标名称
  # 如果配置measurement_name, 则不进行指标名称的切割
  # 最终的指标集名称会添加上measurement_prefix前缀
  # measurement_name = "prom"

  ## 自定义指标集名称
  # 可以将名称满足指定pattern的指标归为一类指标集
  # 自定义指标集名称配置优先measurement_name配置项
  #[[inputs.prom.measurements]]
  #  名称匹配, 支持正则
  #  pattern = "cpu"
  #  指标集名称
  #  name = "prom_cpu"

  # [[inputs.prom.measurements]]
  # disable_prefix = 0
  # pattern = "mem"
  # name = "mem"

  ## 采集间隔 "ns", "us" (or "µs"), "ms", "s", "m", "h"
  interval = "10s"

  ## TLS 配置
  tls_open = false
  # tls_ca = "/tmp/ca.crt"
  # tls_cert = "/tmp/peer.crt"
  # tls_key = "/tmp/peer.key"

  ## 自定义Tags
  [inputs.prom.tags]
  # some_tag = "some_value"
  # more_tag = "some_other_value"
	`
	f := "test.toml"
	err := ioutil.WriteFile(f, []byte(toml), 0666)
	assert.NoError(t, err)
	defer func() { os.Remove(f) }()

	pInput, err := getPromInput(f)

	assert.NoError(t, err)

	assert.Equal(t, pInput.Catalog(), "prom")

	err = pInput.InitClient()
	assert.NoError(t, err)

	client := &http.Client{}
	mockResult := `
# HELP promhttp_metric_handler_errors_total Total number of internal errors encountered by the promhttp metric handler.
# TYPE promhttp_metric_handler_errors_total counter
promhttp_metric_handler_errors_total{cause="encoding"} 0
promhttp_metric_handler_errors_total{cause="gathering"} 0
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
	`
	client.Transport = newTransportMock(mockResult)
	pInput.SetClient(client)

	err = pInput.Collect()
	assert.NoError(t, err)

	points := pInput.GetCachedPoints()

	assert.NotEmpty(t, points)

	t.Run("invalid path", func(t *testing.T) {
		_, err := getPromInput("invalid_path")
		assert.Error(t, err)
	})

	t.Run("invalid toml", func(t *testing.T) {
		toml := `

	`
		f := "test.toml"
		err := ioutil.WriteFile(f, []byte(toml), 0666)
		assert.NoError(t, err)
		defer func() { os.Remove(f) }()

		_, err = getPromInput(f)
		assert.Error(t, err)

	})

}

func TestShowInput(t *testing.T) {
	input := &prom.Input{}
	client := &http.Client{}
	mockResult := `
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 15143
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0
	`
	client.Transport = newTransportMock(mockResult)
	input.SetClient(client)

	err := showInput(input)

	assert.NoError(t, err)

}
