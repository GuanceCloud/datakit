package httpPacket

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

const (
	pointsCallbackFnName   = "handle"
	pointsCallbackTypeName = "points"
)

type pointsData struct {
	name     string
	category string
	data     []map[string]interface{}
	intlog   map[string][]string
}

func NewPointsData(name string, category string, pts []*influxdb.Point) (*pointsData, error) {
	if name == "" {
		return nil, fmt.Errorf("invalid name, name is empty")
	}

	var p = pointsData{
		name:     name,
		category: category,
		data:     []map[string]interface{}{},
		intlog:   map[string][]string{},
	}

	for _, pt := range pts {
		f, err := pt.Fields()
		if err != nil {
			return nil, err
		}

		p.data = append(p.data, map[string]interface{}{
			"name":   pt.Name(),
			"tags":   pt.Tags(),
			"fields": f,
			"time":   pt.UnixNano(),
		})
		p.intlog[pt.Name()] = integerLog(f)
	}

	return &p, nil
}

func (p *pointsData) Name() string {
	return p.name
}

func (p *pointsData) DataToLua() interface{} {
	return p.data
}

func (*pointsData) CallbackFnName() string {
	return pointsCallbackFnName
}

func (*pointsData) CallbackTypeName() string {
	return pointsCallbackTypeName
}

func (p *pointsData) Handle(value string, err error) {
	if err != nil {
		l.Errorf("handle receive error: %s", err.Error())
		return
	}

	type pd struct {
		Name   string                 `json:"name"`
		Tags   map[string]string      `json:"tags,omitempty"`
		Fields map[string]interface{} `json:"fields"`
		Time   int64                  `json:"time,omitempty"`
	}

	x := []pd{}
	json.Unmarshal([]byte(value), &x)

	var buffer bytes.Buffer

	for _, m := range x {
		recoveType(p.intlog[m.Name], m.Fields)

		pt, err := influxdb.NewPoint(m.Name, m.Tags, m.Fields, time.Unix(0, m.Time))
		if err != nil {
			l.Error(err)
			return
		}

		buffer.WriteString(pt.String())
		buffer.WriteString("\n")
	}

	err = io.NamedFeed(buffer.Bytes(), p.category, inputName)
	if err != nil {
		l.Error(err)
	}
}

func integerLog(m map[string]interface{}) []string {
	var list []string
	for k, v := range m {
		switch v.(type) {
		case int, int8, int16, int32, int64,
			uint, uint8, uint16, uint32, uint64:
			list = append(list, k)
		}
	}
	return list
}

func recoveType(intlog []string, m map[string]interface{}) {
	for _, k := range intlog {
		switch m[k].(type) {
		case float64:
			m[k] = int64(m[k].(float64))
		}
	}
}
