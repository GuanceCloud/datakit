// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkm3db is for m3db
package sinkm3db

import (
	"context"
	"reflect"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

const (
	creatorID               = "m3db"
	defaulHTTPClientTimeout = 30 * time.Second
	defaultUserAgent        = "promremote-go/1.0.0"
)

var (
	l = logger.DefaultSLogger("m3db")
)

type SinkM3db struct {
	id            string
	promRemoteURL string

	client *client
}

func (s *SinkM3db) GetID() string {
	return s.id
}

func (s *SinkM3db) LoadConfig(mConf map[string]interface{}) error {
	l = logger.SLogger("m3db")

	if id, err := dkstring.GetMapAssertString("id", mConf); err != nil {
		return err
	} else {
		idNew, err := dkstring.CheckNotEmpty(id, "id")
		if err != nil {
			return err
		}
		s.id = idNew
	}

	if addr, err := dkstring.GetMapAssertString("addr", mConf); err != nil {
		return err
	} else {
		addrNew, err := dkstring.CheckNotEmpty(addr, "addr")
		if err != nil {
			return err
		}
		s.promRemoteURL = addrNew
	}
	// 其他字段

	// 初始化 prom client
	cfg := NewConfig(
		WriteURLOption(s.promRemoteURL),
		HTTPClientTimeoutOption(defaulHTTPClientTimeout),
		UserAgent(defaultUserAgent),
	)
	if err := cfg.validate(); err != nil {
		l.Errorf("config err = %v", err)
		return err
	}
	client, err := NewClient(cfg)
	if err != nil {
		l.Errorf("unable to construct client: %v", err)
		return err
	}
	s.client = client
	l.Infof("init m3db client ok")
	sinkcommon.AddImpl(s)
	return nil
}

func (s *SinkM3db) Write(pts []sinkcommon.ISinkPoint) error {
	var ctx = context.Background()
	var writeOpts WriteOptions
	ts := pointToPromData(pts)
	prompbReq := toPromWriteRequest(ts)
	if len(ts) > 0 {
		result, err := s.client.WriteProto(ctx, prompbReq, writeOpts)
		if err != nil {
			l.Errorf("write err=%v", err)
			return err
		}
		l.Debugf("Status code: %d", result.StatusCode) // 此处使用 debug 级别日志，方便查询问题
	} else {
		l.Debugf("from points to make PromWriteRequest data, len is 0")
	}
	return nil
}

func pointToPromData(pts []sinkcommon.ISinkPoint) []*TimeSeries {
	tslist := make([]*TimeSeries, 0)
	for _, pt := range pts {
		jsonPrint, err := pt.ToJSON()
		if err != nil {
			l.Errorf("to json is err=%v", err)
			continue
		}
		for key, val := range jsonPrint.Fields {
			res := makeSeries(jsonPrint.Tags, key, val, jsonPrint.Time)
			tslist = append(tslist, res...)
		}
		// todo 其他数据
	}
	return tslist
}

func makeSeries(tags map[string]string, key string, i interface{}, dataTime time.Time) []*TimeSeries {
	defer func() {
		if err := recover(); err != nil {
			l.Infof("invalid data type")
		}
	}()

	labels := make([]Label, 0)
	for key, val := range tags {
		labels = append(labels, Label{
			Name:  key,
			Value: val,
		})
	}
	labels = append(labels, Label{Name: model.MetricNameLabel, Value: key})
	switch i.(type) {
	case int, int32, int64:
		if val, ok := i.(int64); ok { // todo test
			return []*TimeSeries{{Labels: labels, Datapoint: Datapoint{
				Timestamp: dataTime,
				Value:     float64(val),
			}}}
		}
	case uint32, uint64:
		val := i.(uint64)
		return []*TimeSeries{{Labels: labels, Datapoint: Datapoint{
			Timestamp: dataTime,
			Value:     float64(val),
		}}}
	case float32, float64:
		val := i.(float64)
		return []*TimeSeries{{Labels: labels, Datapoint: Datapoint{
			Timestamp: dataTime,
			Value:     float64(val),
		}}}
	default:
		// 不能使用 map[]interface{} 去接收 map[string]int 或者 map[string]int64 等类型。
		// 也不能使用 []interface{} 去接收数组 []int []int64 等。
		// 这里使用反射 只处理 map 和 array/slice 类型。
		ts := make([]*TimeSeries, 0)
		v := reflect.ValueOf(i)

		// map
		if v.Kind() == reflect.Map {
			// v = v.Elem()
			iter := v.MapRange()
			for iter.Next() {
				k := iter.Key()
				if k.Kind() != reflect.String {
					return []*TimeSeries{}
				}
				key := k.String()

				v := iter.Value()
				var val interface{}
				val = v.Interface()

				res := makeSeries(tags, key, val, dataTime)
				if len(res) > 0 {
					ts = append(ts, res[0])
				}
			}
			return ts
		}

		// array
		if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				val := v.Index(i).Interface()
				res := makeSeries(tags, key, val, dataTime)
				if len(res) > 0 {
					ts = append(ts, res[0])
				}
			}
			return ts
		}
		l.Infof("default metric data")
	}
	return []*TimeSeries{}
}

// toPromWriteRequest converts a list of timeseries to a Prometheus proto write request.
func toPromWriteRequest(promts []*TimeSeries) *prompb.WriteRequest {
	promPbTS := make([]*prompb.TimeSeries, len(promts))
	for i, ts := range promts {
		labels := make([]*prompb.Label, len(ts.Labels))
		for j, label := range ts.Labels {
			labels[j] = &prompb.Label{Name: label.Name, Value: label.Value}
		}

		sample := []prompb.Sample{
			{
				// Timestamp is int milliseconds for remote write.
				Timestamp: ts.Datapoint.Timestamp.UnixNano() / int64(time.Millisecond),
				Value:     ts.Datapoint.Value,
			},
		}
		promPbTS[i] = &prompb.TimeSeries{Labels: labels, Samples: sample}
	}

	return &prompb.WriteRequest{
		Timeseries: promPbTS,
	}
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorID, func() sinkcommon.ISink {
		return &SinkM3db{id: creatorID}
	})
}
