// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mongodb

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type DBData struct {
	Name   string
	Fields map[string]interface{}
}

type ColData struct {
	Name   string
	DBName string
	Fields map[string]interface{}
}

type MongodbData struct {
	StatLine      *StatLine
	Tags          map[string]string
	Fields        map[string]interface{}
	ShardHostData []DBData
	DBData        []DBData
	ColData       []ColData
	TopStatsData  []DBData
	collectCache  []*point.Point
	ipt           *Input
}

func NewMongodbData(statLine *StatLine, tags map[string]string, ipt *Input) *MongodbData {
	return &MongodbData{
		StatLine: statLine,
		Tags:     tags,
		Fields:   make(map[string]interface{}),
		ipt:      ipt,
	}
}

func (d *MongodbData) AddDefaultStats() {
	statLine := reflect.ValueOf(d.StatLine).Elem()
	d.addStat(statLine, defaultStats)
	if d.StatLine.NodeType != "" {
		d.addStat(statLine, defaultReplStats)
		d.Tags["node_type"] = d.StatLine.NodeType
	}

	if d.StatLine.ReadLatency > 0 {
		d.addStat(statLine, defaultLatencyStats)
	}

	if d.StatLine.ReplSetName != "" {
		d.Tags["rs_name"] = d.StatLine.ReplSetName
	}

	if d.StatLine.OplogStats != nil {
		d.add("repl_oplog_window_sec", d.StatLine.OplogStats.TimeDiff)
	}

	if d.StatLine.Version != "" {
		d.add("version", d.StatLine.Version)
	}

	d.addStat(statLine, defaultAssertsStats)
	d.addStat(statLine, defaultCommandsStats)
	d.addStat(statLine, defaultClusterStats)
	d.addStat(statLine, defaultShardStats)
	d.addStat(statLine, defaultStorageStats)
	d.addStat(statLine, defaultTCMallocStats)

	if d.StatLine.StorageEngine == "mmapv1" || d.StatLine.StorageEngine == "rocksdb" {
		d.addStat(statLine, mmapStats)
	} else if d.StatLine.StorageEngine == "wiredTiger" {
		for key, value := range wiredTigerStats {
			val := statLine.FieldByName(value).Interface()
			percentVal := fmt.Sprintf("%.1f", val.(float64)*100)
			floatVal, _ := strconv.ParseFloat(percentVal, 64)
			d.add(key, floatVal)
		}
		d.addStat(statLine, wiredTigerExtStats)
		d.add("page_faults", d.StatLine.FaultsCnt)
	}
}

func (d *MongodbData) AddShardHostStats() {
	for host, hostStat := range d.StatLine.ShardHostStatsLines {
		hostStatLine := reflect.ValueOf(hostStat)
		newDBData := &DBData{
			Name:   host,
			Fields: make(map[string]interface{}),
		}
		newDBData.Fields["type"] = "shard_host_stat"
		for k, v := range shardHostStats {
			val := hostStatLine.FieldByName(v).Interface()
			newDBData.Fields[k] = val
		}
		d.ShardHostData = append(d.ShardHostData, *newDBData)
	}
}

func (d *MongodbData) AddDBStats() {
	for _, dbstat := range d.StatLine.DBStatsLines {
		dbStatLine := reflect.ValueOf(dbstat)
		newDBData := &DBData{
			Name:   dbstat.Name,
			Fields: make(map[string]interface{}),
		}
		newDBData.Fields["type"] = "db_stat"
		for key, value := range dbDataStats {
			val := dbStatLine.FieldByName(value).Interface()
			newDBData.Fields[key] = val
		}
		d.DBData = append(d.DBData, *newDBData)
	}
}

func (d *MongodbData) AddColStats() {
	for _, colstat := range d.StatLine.ColStatsLines {
		colStatLine := reflect.ValueOf(colstat)
		newColData := &ColData{
			Name:   colstat.Name,
			DBName: colstat.DBName,
			Fields: make(map[string]interface{}),
		}
		newColData.Fields["type"] = "col_stat"
		for key, value := range colDataStats {
			val := colStatLine.FieldByName(value).Interface()
			newColData.Fields[key] = val
		}
		d.ColData = append(d.ColData, *newColData)
	}
}

func (d *MongodbData) AddTopStats() {
	for _, topStat := range d.StatLine.TopStatLines {
		topStatLine := reflect.ValueOf(topStat)
		newTopStatData := &DBData{
			Name:   topStat.CollectionName,
			Fields: make(map[string]interface{}),
		}
		newTopStatData.Fields["type"] = "top_stat"
		for key, value := range topDataStats {
			val := topStatLine.FieldByName(value).Interface()
			newTopStatData.Fields[key] = val
		}
		d.TopStatsData = append(d.TopStatsData, *newTopStatData)
	}
}

func (d *MongodbData) addStat(statLine reflect.Value, stats map[string]string) {
	for key, value := range stats {
		d.add(key, statLine.FieldByName(value).Interface())
	}
}

func (d *MongodbData) add(key string, val interface{}) {
	d.Fields[key] = val
}

func (d *MongodbData) append() {
	if d.ipt.Election {
		d.Tags = inputs.MergeTagsWrapper(d.Tags, d.ipt.Tagger.ElectionTags(), d.ipt.Tags, "")
	} else {
		d.Tags = inputs.MergeTagsWrapper(d.Tags, d.ipt.Tagger.HostTags(), d.ipt.Tags, "")
	}

	now := time.Now()
	metric := &mongodbMeasurement{
		name:   MongoDB,
		tags:   copyTags(d.Tags),
		fields: d.Fields,
		ts:     now,
	}
	d.collectCache = append(d.collectCache, metric.Point())

	for _, db := range d.DBData {
		tmp := copyTags(d.Tags)
		tmp["db_name"] = db.Name
		metric := &mongodbDBMeasurement{
			name:   MongoDBStats,
			tags:   tmp,
			fields: db.Fields,
			ts:     now,
		}
		d.collectCache = append(d.collectCache, metric.Point())
	}

	for _, col := range d.ColData {
		tmp := copyTags(d.Tags)
		tmp["collection"] = col.Name
		tmp["db_name"] = col.DBName
		metric := &mongodbColMeasurement{
			name:   MongoDBColStats,
			tags:   tmp,
			fields: col.Fields,
			ts:     now,
		}
		d.collectCache = append(d.collectCache, metric.Point())
	}

	for _, host := range d.ShardHostData {
		tmp := copyTags(d.Tags)
		tmp["hostname"] = host.Name
		metric := &mongodbShardMeasurement{
			name:   MongoDBShardStats,
			tags:   tmp,
			fields: host.Fields,
			ts:     now,
		}
		d.collectCache = append(d.collectCache, metric.Point())
	}

	for _, col := range d.TopStatsData {
		tmp := copyTags(d.Tags)
		tmp["collection"] = col.Name
		metric := &mongodbTopMeasurement{
			name:   MongoDBTopStats,
			tags:   tmp,
			fields: col.Fields,
			ts:     now,
		}
		d.collectCache = append(d.collectCache, metric.Point())
	}
}

func (d *MongodbData) flush(cost time.Duration) {
	if len(d.collectCache) > 0 {
		if err := d.ipt.feeder.Feed(inputName, point.Metric, d.collectCache, &dkio.Option{CollectCost: cost}); err != nil {
			log.Errorf(err.Error())
			d.ipt.feeder.FeedLastError(inputName, err.Error())
		}
	}
}

func copyTags(tags map[string]string) map[string]string {
	tmp := make(map[string]string)
	for k, v := range tags {
		tmp[k] = v
	}

	return tmp
}
