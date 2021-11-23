//go:build windows
// +build windows

package wmi

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/influxdata/toml"
)

func TestGenConfig(t *testing.T) {
	var cfg Instance

	q1 := &ClassQuery{
		Class: "Win32_LogicalDisk",
		Metrics: [][]string{
			{"DeviceID"},
			{"DriveType", "device_type"},
		},
	}

	q2 := &ClassQuery{
		Class: "Win32_Process",
		Metrics: [][]string{
			{"Name"},
		},
	}

	cfg.Queries = []*ClassQuery{
		q1, q2,
	}

	data, err := toml.Marshal(&cfg)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%s", string(data))
}

func TestLoadCfg(t *testing.T) {
	var cfg Instance
	err := toml.Unmarshal([]byte(sampleConfig), &cfg)
	if err != nil {
		t.Error(err)
	} else {
		log.Printf("ok")
	}
}

func TestQuerySysInfo(t *testing.T) {
	var q ClassQuery
	q.Class = `Win32_LogicalDisk`

	// q.Metrics = [][]string{
	// 	{`deviceid`},
	// 	{`size`},
	// 	{`Description`},
	// 	{`FileSystem`},
	// }

	sql, err := q.ToSQL()
	if err != nil {
		t.Error(err)
	}

	props := []string{}

	for _, ms := range q.Metrics {
		props = append(props, ms[0])
	}

	_, err = DefaultClient.QueryEx(sql, props)
	if err != nil {
		t.Error(err)
	}
}

func TestSvr(t *testing.T) {
	ag := NewAgent()

	data, err := ioutil.ReadFile("./test.conf")
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = toml.Unmarshal(data, ag)
	if err != nil {
		log.Fatalf("%s", err)
	}

	ag.Run()
}
