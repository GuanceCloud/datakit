// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package sinkm3db is for m3db
package sinkm3db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/m3db/prometheus_remote_client_golang/promremote"
	"github.com/prometheus/common/model"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/dkstring"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io/sink/sinkcommon"
)

const (
	creatorID = "m3db"
)

type SinkM3db struct {
	ID            string
	promRemoteURL string

	client promremote.Client
}

func (s *SinkM3db) GetID() string {
	return s.ID
}

func (s *SinkM3db) LoadConfig(mConf map[string]interface{}) error {
	if id, err := dkstring.GetMapAssertString("id", mConf); err != nil {
		return err
	} else {
		idNew, err := dkstring.CheckNotEmpty(id, "id")
		if err != nil {
			return err
		}
		s.ID = idNew
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
	cfg := promremote.NewConfig(
		promremote.WriteURLOption("PROM_WRITE_URL"),
		promremote.HTTPClientTimeoutOption(60*time.Second),
	)
	client, err := promremote.NewClient(cfg)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to construct client: %v", err))
	}
	s.client = client
	return nil
}

func (s *SinkM3db) Write(pts []sinkcommon.ISinkPoint) error {
	var ctx = context.Background()
	var writeOpts promremote.WriteOptions
	ts := pointToPromData(pts)
	if len(ts) > 0 {
		result, err := s.client.WriteTimeSeries(ctx, ts, writeOpts)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Status code: %d\n", result.StatusCode)
	}
	return nil
}

func pointToPromData(pts []sinkcommon.ISinkPoint) []promremote.TimeSeries {
	ms := make([]promremote.TimeSeries, 0)
	for _, pt := range pts {
		jsonPrint, err := pt.ToJSON()
		if err != nil {
			continue
		}
		for key, val := range jsonPrint.Fields {
			res := getSeries(jsonPrint.Tags, key, val, jsonPrint.Time)
			ms = append(ms, res...)
		}
		// todo 其他数据
	}
	return ms
}

func getSeries(tags map[string]string, key string, i interface{}, dataTime time.Time) []promremote.TimeSeries {
	labels := make([]promremote.Label, 0)
	for key, val := range tags {
		labels = append(labels, promremote.Label{
			Name:  key,
			Value: val,
		})
	}
	labels = append(labels, promremote.Label{Name: model.MetricNameLabel, Value: key})
	switch i.(type) {
	case int, int32, int64:
		if val, ok := i.(int64); ok { // todo test
			return []promremote.TimeSeries{{Labels: labels, Datapoint: promremote.Datapoint{
				Timestamp: dataTime,
				Value:     float64(val),
			}}}
		}
	case uint32, uint64:
		val := i.(uint64)
		return []promremote.TimeSeries{{Labels: labels, Datapoint: promremote.Datapoint{
			Timestamp: dataTime,
			Value:     float64(val),
		}}}
	case float32, float64:
		val := i.(float64)
		return []promremote.TimeSeries{{Labels: labels, Datapoint: promremote.Datapoint{
			Timestamp: dataTime,
			Value:     float64(val),
		}}}
	case map[string]interface{}:
		maps := i.(map[string]interface{})
		ts := make([]promremote.TimeSeries, 0)
		for keyi, i2 := range maps {
			res := getSeries(tags, keyi, i2, dataTime)
			if len(res) > 0 {
				ts = append(ts, res[0])
			}
		}
		return ts
	case []interface{}: // todo 有没有 数组形式？
	default:
	}
	return []promremote.TimeSeries{}
}

func init() { //nolint:gochecknoinits
	sinkcommon.AddCreator(creatorID, func() sinkcommon.ISink {
		return &SinkM3db{}
	})
}

/*

const (
	namespace = "default"
)

var (
	namespaceID = ident.StringID(namespace)
)

type config struct {
	Client client.Configuration `yaml:"m3db_client"`
}

var configFile = flag.String("f", "", "configuration file")

func main() {
	flag.Parse()
	if *configFile == "" {
		flag.Usage()
		return
	}

	cfgBytes, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("unable to read config file: %s, err: %v", *configFile, err)
	}

	cfg := &config{}
	if err := yaml.UnmarshalStrict(cfgBytes, cfg); err != nil {
		log.Fatalf("unable to parse YAML: %v", err)
	}

	// TODO(rartoul): Provide guidelines on reducing memory usage by tuning pooling options.
	client, err := cfg.Client.NewClient(client.ConfigurationParameters{})
	if err != nil {
		log.Fatalf("unable to create new M3DB client: %v", err)
	}

	session, err := client.DefaultSession()
	if err != nil {
		log.Fatalf("unable to create new M3DB session: %v", err)
	}
	defer session.Close()

	schemaConfig, ok := cfg.Client.Proto.SchemaRegistry[namespace]
	if !ok {
		log.Fatalf("schema path for namespace: %s not found", namespace)
	}

	// NB(rartoul): Using dynamic / reflection based library for marshaling and unmarshaling protobuf
	// messages for simplicity, use generated message-specific bindings in production.
	schema, err := proto.ParseProtoSchema(schemaConfig.SchemaFilePath, schemaConfig.MessageName)
	if err != nil {
		log.Fatalf("could not parse proto schema: %v", err)
	}

	runUntaggedExample(session, schema)
	runTaggedExample(session, schema)
	// TODO(rartoul): Add an aggregations query example.
}

// runUntaggedExample demonstrates how to write "untagged" (unindexed) data into M3DB with a given
// protobuf schema and then read it back out again.
func runUntaggedExample(session client.Session, schema *desc.MessageDescriptor) {
	log.Printf("------ run untagged example ------")
	var (
		untaggedSeriesID = ident.StringID("untagged_seriesID")
		m                = newTestValue(schema)
	)
	marshaled, err := m.Marshal()
	if err != nil {
		log.Fatalf("error marshaling protobuf message: %v", err)
	}

	// Write an untagged series ID. Pass 0 for value since it is ignored.
	if err := session.Write(namespaceID, untaggedSeriesID, xtime.Now(), 0, xtime.Nanosecond, marshaled); err != nil {
		log.Fatalf("unable to write untagged series: %v", err)
	}

	// Fetch data for the untagged seriesID written within the last minute.
	seriesIter, err := session.Fetch(namespaceID, untaggedSeriesID, xtime.Now().Add(-time.Minute), xtime.Now())
	if err != nil {
		log.Fatalf("error fetching data for untagged series: %v", err)
	}
	for seriesIter.Next() {
		m = dynamic.NewMessage(schema)
		dp, _, marshaledProto := seriesIter.Current()
		if err := m.Unmarshal(marshaledProto); err != nil {
			log.Fatalf("error unmarshaling protobuf message: %v", err)
		}
		log.Printf("%s: %s", dp.TimestampNanos.String(), m.String())
	}
	if err := seriesIter.Err(); err != nil {
		log.Fatalf("error in series iterator: %v", err)
	}
}

// runTaggedExample demonstrates how to write "tagged" (indexed) data into M3DB with a given protobuf
// schema and then read it back out again by either:
//
//   1. Querying for a specific time series by its ID directly
//   2. TODO(rartoul): Querying for a set of time series using an inverted index query
func runTaggedExample(session client.Session, schema *desc.MessageDescriptor) {
	log.Printf("------ run tagged example ------")
	var (
		seriesID = ident.StringID("vehicle_id_1")
		tags     = []ident.Tag{
			{Name: ident.StringID("type"), Value: ident.StringID("sedan")},
			{Name: ident.StringID("city"), Value: ident.StringID("san_francisco")},
		}
		tagsIter = ident.NewTagsIterator(ident.NewTags(tags...))
		m        = newTestValue(schema)
	)
	marshaled, err := m.Marshal()
	if err != nil {
		log.Fatalf("error marshaling protobuf message: %v", err)
	}

	// Write a tagged series ID. Pass 0 for value since it is ignored.
	if err := session.WriteTagged(namespaceID, seriesID, tagsIter, xtime.Now(), 0, xtime.Nanosecond, marshaled); err != nil {
		log.Fatalf("error writing series %s, err: %v", seriesID.String(), err)
	}

	// Fetch data for the tagged seriesID using a direct ID lookup (only data written within the last minute).
	seriesIter, err := session.Fetch(namespaceID, seriesID, xtime.Now().Add(-time.Minute), xtime.Now())
	if err != nil {
		log.Fatalf("error fetching data for untagged series: %v", err)
	}
	for seriesIter.Next() {
		m = dynamic.NewMessage(schema)
		dp, _, marshaledProto := seriesIter.Current()
		if err := m.Unmarshal(marshaledProto); err != nil {
			log.Fatalf("error unamrshaling protobuf message: %v", err)
		}
		log.Printf("%s: %s", dp.TimestampNanos.String(), m.String())
	}
	if err := seriesIter.Err(); err != nil {
		log.Fatalf("error in series iterator: %v", err)
	}

	// TODO(rartoul): Show an example of how to execute a FetchTagged() call with an index query.
}

var (
	testValueLock  sync.Mutex
	testValueCount = 1
)

func newTestValue(schema *desc.MessageDescriptor) *dynamic.Message {
	testValueLock.Lock()
	defer testValueLock.Unlock()

	m := dynamic.NewMessage(schema)
	m.SetFieldByName("latitude", float64(testValueCount))
	m.SetFieldByName("longitude", float64(testValueCount))
	m.SetFieldByName("fuel_percent", 0.75)
	m.SetFieldByName("status", "active")

	testValueCount++

	return m
}

*/

//
