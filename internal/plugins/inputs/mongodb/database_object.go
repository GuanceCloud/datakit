// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	mongodbType                  = "MongoDB"
	mongodbObjectMeasurementName = "database"
)

type mongodbObjectMeasurement struct{}

//nolint:lll
func (*mongodbObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: mongodbObjectMeasurementName,
		Cat:  point.Object,
		Desc: "MongoDB object metrics.",
		Tags: map[string]interface{}{
			"database_instance": &inputs.TagInfo{Desc: mongodbDatabaseInstanceDesc},
			"database_type":     &inputs.TagInfo{Desc: "The type of the database. The value is `MongoDB`"},
			"host":              &inputs.TagInfo{Desc: "The hostname of the MongoDB server"},
			"name":              &inputs.TagInfo{Desc: "The object identifier. The value is `<server>-<database_instance>` when `server` and `database_instance` are different, otherwise `server`."},
			"port":              &inputs.TagInfo{Desc: "The port of the MongoDB server"},
			"server":            &inputs.TagInfo{Desc: mongodbServerDesc},
			"version":           &inputs.TagInfo{Desc: "The version of the MongoDB server"},
		},
		Fields: map[string]interface{}{
			"message":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "MongoDB startup/configuration settings from the parsed getCmdLineOpts output."},
			"uptime":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "The number of seconds that the server has been up"},
			"avg_query_time": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.TimestampUS, Desc: "The average time taken by a query to execute"},
			"qps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of operations executed by the database per second."},
		},
	}
}

type mongodbObjectMessage struct {
	Setting map[string]interface{} `json:"setting"`
}

func (m *mongodbObjectMessage) String() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

func (svr *MongodbServer) feedDatabaseObject(status *MongoStatus, statLine *StatLine, ptTS int64) {
	if !svr.ipt.Object.Enable || status == nil || status.ServerStatus == nil {
		return
	}

	now := time.Now()
	if !svr.lastObjectCollectionTime.IsZero() &&
		svr.lastObjectCollectionTime.Add(svr.ipt.Object.Interval.Duration).After(now) {
		log.Debugf("skip mongodb_object collection, time interval not reached")
		return
	}
	svr.lastObjectCollectionTime = now

	// configuration settings, not runtime counters from serverStatus.
	settings, err := svr.getMongoSettings()
	if err != nil {
		log.Warnf("get mongodb settings failed: %s", err.Error())
	}

	pt := svr.buildDatabaseObjectPoint(status, statLine, settings, ptTS)
	if pt == nil {
		return
	}

	if err := svr.ipt.feeder.Feed(point.Object, []*point.Point{pt},
		dkio.WithCollectCost(time.Since(now)),
		dkio.WithElection(svr.ipt.Election),
		dkio.WithSource(objectFeedName),
	); err != nil {
		svr.ipt.feeder.FeedLastError(err.Error(),
			metrics.WithLastErrorInput(inputName),
			metrics.WithLastErrorCategory(point.Object),
		)
		log.Errorf("feed mongodb object failed: %s", err.Error())
	}
}

func (svr *MongodbServer) buildDatabaseObjectPoint(
	status *MongoStatus,
	statLine *StatLine,
	settings map[string]interface{},
	ptTS int64,
) *point.Point {
	if status == nil || status.ServerStatus == nil {
		return nil
	}

	serverStatus := status.ServerStatus
	tags := svr.getDefaultTags()
	server := tags["server"]
	databaseInstance := tags["database_instance"]

	objectName := server
	if server != databaseInstance {
		objectName = fmt.Sprintf("%s-%s", server, databaseInstance)
	}
	kvs := point.NewTags(tags).
		AddTag("version", serverStatus.Version).
		AddTag("database_type", mongodbType).
		AddTag("name", objectName)

	_, port, _ := splitMongoAddr(server)
	if port != "" {
		kvs = kvs.AddTag("port", port)
	}

	qps := mongodbObjectQPS(statLine)
	kvs = kvs.Set("uptime", serverStatus.Uptime).
		Set("message", buildMongoDBObjectMessage(settings).String()).
		Set("qps", qps)
	if avgQueryTime, ok := mongodbAvgQueryTime(status); ok {
		kvs = kvs.Set("avg_query_time", avgQueryTime)
	}

	opts := append(point.DefaultObjectOptions(), point.WithTimestamp(ptTS))

	return point.NewPoint(mongodbObjectMeasurementName, kvs, opts...)
}

func (svr *MongodbServer) getMongoSettings() (map[string]interface{}, error) {
	var opts struct {
		Parsed bson.M `bson:"parsed"`
	}

	// parsed contains the effective startup/config file options; argv is
	// intentionally skipped to avoid publishing the raw command line.
	rslt := svr.cli.Database("admin").RunCommand(context.TODO(), bson.M{"getCmdLineOpts": 1})
	if err := rslt.Err(); err != nil {
		return nil, err
	}

	if err := rslt.Decode(&opts); err != nil {
		return nil, err
	}

	if opts.Parsed == nil {
		return map[string]interface{}{}, nil
	}

	return normalizeMongoSettings(opts.Parsed), nil
}

func normalizeMongoSettings(settings bson.M) map[string]interface{} {
	normalized := make(map[string]interface{}, len(settings))
	for k, v := range settings {
		normalized[k] = normalizeMongoSettingValue(v)
	}

	return normalized
}

func normalizeMongoSettingValue(v interface{}) interface{} {
	switch x := v.(type) {
	case bson.M:
		return normalizeMongoSettings(x)
	case bson.D:
		m := make(map[string]interface{}, len(x))
		for _, elem := range x {
			m[elem.Key] = normalizeMongoSettingValue(elem.Value)
		}
		return m
	case bson.A:
		arr := make([]interface{}, 0, len(x))
		for _, elem := range x {
			arr = append(arr, normalizeMongoSettingValue(elem))
		}
		return arr
	case []interface{}:
		arr := make([]interface{}, 0, len(x))
		for _, elem := range x {
			arr = append(arr, normalizeMongoSettingValue(elem))
		}
		return arr
	default:
		return v
	}
}

func mongodbObjectQPS(statLine *StatLine) float64 {
	if statLine == nil {
		return 0
	}

	return float64(statLine.Insert + statLine.Query + statLine.Update +
		statLine.Delete + statLine.GetMore + statLine.Command)
}

func mongodbAvgQueryTime(status *MongoStatus) (float64, bool) {
	readLatency := mongoReadLatency(status)
	if readLatency == nil || readLatency.Ops <= 0 {
		return 0, false
	}

	// Other database object collectors expose cumulative average query time.
	return float64(readLatency.Latency) / float64(readLatency.Ops), true
}

func mongoReadLatency(status *MongoStatus) *LatencyStats {
	if status == nil || status.ServerStatus == nil ||
		status.ServerStatus.OpLatencies == nil {
		return nil
	}

	return status.ServerStatus.OpLatencies.Reads
}

func buildMongoDBObjectMessage(settings map[string]interface{}) *mongodbObjectMessage {
	message := &mongodbObjectMessage{
		Setting: map[string]interface{}{},
	}

	for k, v := range settings {
		message.Setting[k] = v
	}
	return message
}
