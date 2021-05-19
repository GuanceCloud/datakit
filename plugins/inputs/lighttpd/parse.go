package lighttpd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	yaml "gopkg.in/yaml.v2"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

type Version int

const (
	v1 Version = iota
	v2
)

type StatusV1 struct {
	RequestsTotal    int `json:"RequestsTotal"`
	TrafficTotal     int `json:"TrafficTotal"`
	Uptime           int `json:"Uptime"`
	BusyServers      int `json:"BusyServers"`
	IdleServers      int `json:"IdleServers"`
	RequestAverage5s int `json:"RequestAverage5s"`
	TrafficAverage5s int `json:"TrafficAverage5s"`
}

type StatusV2 struct {
	// Absolute Values
	Uptime         int `yaml:"uptime"`
	MemoryUsage    int `yaml:"memory_usage"`
	RequestsAbs    int `yaml:"requests_abs"`
	TrafficOutAbs  int `yaml:"traffic_out_abs"`
	TrafficInAbs   int `yaml:"traffic_in_abs"`
	ConnectionsAbs int `yaml:"connections_abs"`

	// Average Values (since start)
	RequestsAvg    int `yaml:"requests_avg"`
	TrafficOutAvg  int `yaml:"traffic_out_avg"`
	TrafficInAvg   int `yaml:"traffic_in_avg"`
	ConnectionsAvg int `yaml:"connections_avg"`

	// Average Values (5 seconds)
	RequestsAvg5sec    int `yaml:"requests_avg_5sec"`
	TrafficOutAvg5sec  int `yaml:"traffic_out_avg_5sec"`
	TrafficInAvg5sec   int `yaml:"traffic_in_avg_5sec"`
	ConnectionsAvg5sec int `yaml:"connections_avg_5sec"`

	// Connection States
	ConnectionStateStart         int `yaml:"connection_state_start"`
	ConnectionStateReadHeader    int `yaml:"connection_state_read_header"`
	ConnectionStateHandleRequest int `yaml:"connection_state_handle_request"`
	ConnectionStateWriteResponse int `yaml:"connection_state_write_response"`
	ConnectionStateKeepAlive     int `yaml:"connection_state_keep_alive"`
	ConnectionStateUpgraded      int `yaml:"connection_state_upgraded"`

	// Status Codes (since start)
	Status1xx int `yaml:"status_1xx"`
	Status2xx int `yaml:"status_2xx"`
	Status3xx int `yaml:"status_3xx"`
	Status4xx int `yaml:"status_4xx"`
	Status5xx int `yaml:"status_5xx"`
}

func (h *Lighttpd) getMetrics() ([]byte, error) {
	resp, err := http.Get(h.statusURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var value reflect.Value

	switch h.statusVersion {
	case v1:
		status := StatusV1{}
		if err := json.Unmarshal(body, &status); err != nil {
			return nil, err
		}
		value = reflect.ValueOf(status)

	case v2:
		status := StatusV2{}
		if err := yaml.Unmarshal(body, &status); err != nil {
			return nil, err
		}
		value = reflect.ValueOf(status)

	default:
		return nil, errors.New("invalid lighttpd version")
	}

	var fields = make(map[string]interface{}, value.NumField())

	for i := 0; i < value.NumField(); i++ {
		fields[value.Type().Field(i).Name] = value.Field(i).Int()
	}

	return io.MakeMetric(defaultMeasurement, h.Tags, fields, time.Now())
}
