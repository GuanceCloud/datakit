package nginx

import (
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type ServerZoneMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *ServerZoneMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *ServerZoneMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: ServerZone,
		Fields: map[string]interface{}{
			"request_count": newCountFieldInfo("The total number of client requests received from clients."),
			"received":      newByteFieldInfo("The total amount of data received from clients."),
			"sent":          newByteFieldInfo("The total amount of data sent to clients."),
			"response_1xx":  newCountFieldInfo("The number of responses with status codes 1xx"),
			"response_2xx":  newCountFieldInfo("The number of responses with status codes 2xx"),
			"response_3xx":  newCountFieldInfo("The number of responses with status codes 3xx"),
			"response_4xx":  newCountFieldInfo("The number of responses with status codes 4xx"),
			"response_5xx":  newCountFieldInfo("The number of responses with status codes 5xx"),
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("nginx server host"),
			"nginx_port":    inputs.NewTagInfo("nginx server port"),
			"server_zone":   inputs.NewTagInfo("server zone"),
			"host":          inputs.NewTagInfo("host mame which installed nginx"),
			"nginx_version": inputs.NewTagInfo("nginx version"),
		},
	}
}

type UpstreamZoneMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *UpstreamZoneMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *UpstreamZoneMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: UpstreamZone,
		Fields: map[string]interface{}{
			"request_count": newCountFieldInfo("The total number of client requests received from server."),
			"received":      newByteFieldInfo("The total number of bytes received from this server."),
			"sent":          newByteFieldInfo("The total number of bytes sent to clients."),
			"response_1xx":  newCountFieldInfo("The number of responses with status codes 1xx"),
			"response_2xx":  newCountFieldInfo("The number of responses with status codes 2xx"),
			"response_3xx":  newCountFieldInfo("The number of responses with status codes 3xx"),
			"response_4xx":  newCountFieldInfo("The number of responses with status codes 4xx"),
			"response_5xx":  newCountFieldInfo("The number of responses with status codes 5xx"),
		},
		Tags: map[string]interface{}{
			"nginx_server":    inputs.NewTagInfo("nginx server host"),
			"nginx_port":      inputs.NewTagInfo("nginx server port"),
			"upstream_zone":   inputs.NewTagInfo("upstream zone"),
			"upstream_server": inputs.NewTagInfo("upstream server"),
			"host":            inputs.NewTagInfo("host mame which installed nginx"),
			"nginx_version":   inputs.NewTagInfo("nginx version"),
		},
	}
}

type CacheZoneMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *CacheZoneMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

//nolint:lll
func (m *CacheZoneMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: CacheZone,
		Fields: map[string]interface{}{
			"max_size":              newByteFieldInfo("The limit on the maximum size of the cache specified in the configuration"),
			"used_size":             newByteFieldInfo("The current size of the cache."),
			"receive":               newByteFieldInfo("The total number of bytes received from the cache."),
			"sent":                  newByteFieldInfo("The total number of bytes sent from the cache."),
			"responses_miss":        newCountFieldInfo("The number of cache miss"),
			"responses_bypass":      newCountFieldInfo("The number of cache bypass"),
			"responses_expired":     newCountFieldInfo("The number of cache expired"),
			"responses_stale":       newCountFieldInfo("The number of cache stale"),
			"responses_updating":    newCountFieldInfo("The number of cache updating"),
			"responses_revalidated": newCountFieldInfo("The number of cache revalidated"),
			"responses_hit":         newCountFieldInfo("The number of cache hit"),
			"responses_scarce":      newCountFieldInfo("The number of cache scarce"),
		},
		Tags: map[string]interface{}{
			"nginx_server":  inputs.NewTagInfo("nginx server host"),
			"nginx_port":    inputs.NewTagInfo("nginx server port"),
			"cache_zone":    inputs.NewTagInfo("cache zone"),
			"host":          inputs.NewTagInfo("host mame which installed nginx"),
			"nginx_version": inputs.NewTagInfo("nginx version"),
		},
	}
}
