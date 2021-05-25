package mysql

import (
	"database/sql"
	"time"

	"github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type schemaMeasurement struct {
	client *sql.DB
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议
func (m *schemaMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *schemaMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_schema",
		Fields: map[string]interface{}{
			"schema_size": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.SizeMiB,
				Desc:     "Size of schemas(MiB)",
			},
			"query_run_time_avg": &inputs.FieldInfo{
				DataType: inputs.Float,
				Type:     inputs.Gauge,
				Unit:     inputs.DurationNS,
				Desc:     "Avg query response time per schema.",
			},
		},
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
			"schema_name": &inputs.TagInfo{
				Desc: "Schema name",
			},
		},
	}
}

// 数据源获取数据
func (i *Input) getSchemaSize() ([]inputs.Measurement, error) {
	querySizePerschemaSql := `
		SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb
		FROM     information_schema.tables
		GROUP BY table_schema;
	`
	rows, err := i.db.Query(querySizePerschemaSql)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	defer rows.Close()

	ms := []inputs.Measurement{}

	for rows.Next() {
		m := &schemaMeasurement{
			name:   "mysql_schema",
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			l.Error(err)
			return nil, err
		}

		size := cast.ToFloat64(string(*val))

		m.fields["schema_size"] = size
		m.tags["schema_name"] = key
		m.ts = time.Now()

		if len(m.fields) > 0 {
			ms = append(ms, m)
		}
	}

	return ms, nil
}

func (i *Input) getQueryExecTimePerSchema() ([]inputs.Measurement, error) {
	queryExecPerTimeSql := `
	SELECT schema_name, ROUND((SUM(sum_timer_wait) / SUM(count_star)) / 1000000) AS avg_us
	FROM performance_schema.events_statements_summary_by_digest
	WHERE schema_name IS NOT NULL
	GROUP BY schema_name;
	`
	rows, err := i.db.Query(queryExecPerTimeSql)
	if err != nil {
		l.Error(err)
		return nil, err
	}
	defer rows.Close()

	ms := []inputs.Measurement{}

	for rows.Next() {
		m := &schemaMeasurement{
			name:   "mysql_schema",
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		var key string
		var val *sql.RawBytes = new(sql.RawBytes)

		if err = rows.Scan(&key, val); err != nil {
			l.Error(err)
			return nil, err
		}

		size := cast.ToInt64(string(*val))

		m.fields["query_run_time_avg"] = size
		m.tags["schema_name"] = key
		m.ts = time.Now()

		if len(m.fields) > 0 {
			ms = append(ms, m)
		}
	}

	return ms, nil
}
