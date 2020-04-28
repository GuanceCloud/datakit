package dataclean

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
)

type JsonTag struct {
	Tagk string `json:"k"`
	Tagv string `json:"v"`
}

type JsonField struct {
	Fieldk string      `json:"k"`
	Fieldv interface{} `json:"v"`
	Fieldt string      `json:"t"`
}
type JsonPoint struct {
	Measurement string      `json:"M"`
	Tags        []JsonTag   `json:"T"`
	Fields      []JsonField `json:"F"`
	Timestamp   int64       `json:"TS"`
	Precision   string      `json:"TP"`
}

var (
	// 1677-09-21 00:12:43.145224194 +0000 UTC
	minNanoTimes = time.Unix(0, (int64(math.MinInt64) + 2)).UTC()
	// 2262-04-11 23:47:16.854775806 +0000 UTC
	maxNanoTimes = time.Unix(0, (int64(math.MaxInt64) - 1)).UTC()
)

func ParseJsonInflux(data []byte, template string) ([]*influxdb.Point, error) {
	points := []JsonPoint{}
	pts := []*influxdb.Point{}

	if err := json.Unmarshal(data, &points); err != nil {
		return nil, err
	}

	for index, point := range points {
		p, err := GenNewPoint(index, &point)
		if err != nil {
			return nil, err
		}
		pts = append(pts, p)
	}

	return pts, nil
}

func GenNewPoint(index int, p *JsonPoint) (*influxdb.Point, error) {
	if p.Measurement == "" {
		return nil, fmt.Errorf("point[%d] missed M", index)
	}

	if len(p.Fields) == 0 {
		return nil, fmt.Errorf("point[%d] missed F", index)
	}

	fields, err := GenFields(index, p.Fields)
	if err != nil {
		return nil, err
	}

	tags, err := GenTags(p.Tags)
	if err != nil {
		return nil, err
	}

	ts, err := GenTime(index, p.Timestamp, p.Precision)
	if err != nil {
		return nil, err
	}

	pt, err := influxdb.NewPoint(p.Measurement, tags, fields, ts)
	if err != nil {
		return nil, err
	}

	return pt, nil
}

func GenTags(tags []JsonTag) (map[string]string, error) {
	ts := make(map[string]string)

	for _, v := range tags {
		ts[v.Tagk] = v.Tagv
	}

	return ts, nil
}

func GenFields(index int, fields []JsonField) (map[string]interface{}, error) {
	fs := make(map[string]interface{})
	valid := 0

	for _, v := range fields {
		t := strings.ToLower(v.Fieldt)
		switch t {
		case "i":
			switch v.Fieldv.(type) {
			case float32, float64:
			default:
				return nil, fmt.Errorf("point[%d] `%v:%v` can not covert value to integer",
					index, v.Fieldk, v.Fieldv)
			}
			// 1234.456 ==> "1234"
			istr := fmt.Sprintf("%.0f", v.Fieldv)
			// "1234" ==> 1234
			val, err := strconv.ParseInt(istr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("point[%d] `%v:%v` can not covert value to integer",
					index, v.Fieldk, v.Fieldv)
			}
			fs[v.Fieldk] = val
		case "f":
			switch v.Fieldv.(type) {
			case float32, float64:
			default:
				return nil, fmt.Errorf("point[%d] `%v:%v` can not covert value to float",
					index, v.Fieldk, v.Fieldv)
			}
			fs[v.Fieldk] = v.Fieldv
		case "b":
			switch v.Fieldv.(type) {
			case bool:
			default:
				return nil, fmt.Errorf("point[%d] `%v:%v` can not covert value to bool",
					index, v.Fieldk, v.Fieldv)
			}
			fs[v.Fieldk] = v.Fieldv
		case "s":
			switch v.Fieldv.(type) {
			case string:
			default:
				return nil, fmt.Errorf("point[%d] `%v:%v` can not covert value to string",
					index, v.Fieldk, v.Fieldv)
			}
			fs[v.Fieldk] = v.Fieldv
		case "":
			fs[v.Fieldk] = v.Fieldv
		default:
			return nil, fmt.Errorf("point[%d] `%v:%v` unknown filed type %s",
				index, v.Fieldk, v.Fieldv, v.Fieldt)
		}
		valid++
	}

	if valid == 0 {
		return nil, fmt.Errorf("point[%d] no valid field", index)
	}

	return fs, nil
}

func GenTime(index int, timestamp int64, precision string) (time.Time, error) {
	p := strings.ToLower(precision)
	t := time.Time{}
	if timestamp == 0 {
		return time.Now(), nil
	}
	switch p {
	case "s":
		t = time.Unix(timestamp, 0)
	case "ms":
		t = time.Unix(timestamp/1e3, (timestamp%1e3)*1e6)
	case "us":
		t = time.Unix(timestamp/1e6, (timestamp%1e6)*1e3)
	case "ns", "":
		t = time.Unix(timestamp/1e9, (timestamp % 1e9))
	default:
		return t, fmt.Errorf("point[%d] unknown TP:%v", index, precision)
	}

	if t.Before(minNanoTimes) || t.After(maxNanoTimes) {
		return t, fmt.Errorf("point[%d] TS:%v does not match TP:%v", index, timestamp, precision)
	}
	return t, nil
}
