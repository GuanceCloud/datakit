package mongodb

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type mongodbMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

func (m *mongodbMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

func (m *mongodbMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{}
}

type MongodbData struct {
	StatLine      *StatLine
	Tags          map[string]string
	Fields        map[string]interface{}
	DbData        []DbData
	ColData       []ColData
	ShardHostData []DbData
	TopStatsData  []DbData
	collectCache  []inputs.Measurement
	collectCost   time.Duration
}

type DbData struct {
	Name   string
	Fields map[string]interface{}
}

type ColData struct {
	Name   string
	DbName string
	Fields map[string]interface{}
}

func NewMongodbData(statLine *StatLine, tags map[string]string, cost time.Duration) *MongodbData {
	return &MongodbData{
		StatLine:    statLine,
		Tags:        tags,
		Fields:      make(map[string]interface{}),
		DbData:      []DbData{},
		collectCost: cost,
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
		hostStatLine := reflect.ValueOf(&hostStat).Elem()
		newDbData := &DbData{
			Name:   host,
			Fields: make(map[string]interface{}),
		}
		newDbData.Fields["type"] = "shard_host_stat"
		for k, v := range shardHostStats {
			val := hostStatLine.FieldByName(v).Interface()
			newDbData.Fields[k] = val
		}
		d.ShardHostData = append(d.ShardHostData, *newDbData)
	}
}

func (d *MongodbData) AddDbStats() {
	for _, dbstat := range d.StatLine.DbStatsLines {
		dbStatLine := reflect.ValueOf(&dbstat).Elem()
		newDbData := &DbData{
			Name:   dbstat.Name,
			Fields: make(map[string]interface{}),
		}
		newDbData.Fields["type"] = "db_stat"
		for key, value := range dbDataStats {
			val := dbStatLine.FieldByName(value).Interface()
			newDbData.Fields[key] = val
		}
		d.DbData = append(d.DbData, *newDbData)
	}
}

func (d *MongodbData) AddColStats() {
	for _, colstat := range d.StatLine.ColStatsLines {
		colStatLine := reflect.ValueOf(&colstat).Elem()
		newColData := &ColData{
			Name:   colstat.Name,
			DbName: colstat.DbName,
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
		topStatLine := reflect.ValueOf(&topStat).Elem()
		newTopStatData := &DbData{
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
	now := time.Now()
	d.collectCache = append(d.collectCache, &mongodbMeasurement{
		name:   "mongodb",
		tags:   d.Tags,
		fields: d.Fields,
		ts:     now,
	})
	// d.Fields = make(map[string]interface{})

	for _, db := range d.DbData {
		d.Tags["db_name"] = db.Name
		d.collectCache = append(d.collectCache, &mongodbMeasurement{
			name:   "mongodb_db_stats",
			tags:   d.Tags,
			fields: db.Fields,
			ts:     now,
		})
		// db.Fields = make(map[string]interface{})
	}

	for _, col := range d.ColData {
		d.Tags["collection"] = col.Name
		d.Tags["db_name"] = col.DbName
		d.collectCache = append(d.collectCache, &mongodbMeasurement{
			name:   "mongodb_col_stats",
			tags:   d.Tags,
			fields: col.Fields,
			ts:     now,
		})
		// col.Fields = make(map[string]interface{})
	}

	for _, host := range d.ShardHostData {
		d.Tags["hostname"] = host.Name
		d.collectCache = append(d.collectCache, &mongodbMeasurement{
			name:   "mongodb_shard_stats",
			tags:   d.Tags,
			fields: host.Fields,
			ts:     now,
		})
		// host.Fields = make(map[string]interface{})
	}

	for _, col := range d.TopStatsData {
		d.Tags["collection"] = col.Name
		d.collectCache = append(d.collectCache, &mongodbMeasurement{
			name:   "mongodb_top_stats",
			tags:   d.Tags,
			fields: col.Fields,
			ts:     now,
		})
		// col.Fields = make(map[string]interface{})
	}
}

func (d *MongodbData) flush() {
	if len(d.collectCache) != 0 {
		if err := inputs.FeedMeasurement(inputName, io.Metric, d.collectCache, &io.Option{CollectCost: d.collectCost}); err != nil {
			l.Error(err)
		}
	}
}
