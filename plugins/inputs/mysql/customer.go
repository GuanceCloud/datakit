package mysql

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
)

type customerMeasurement struct {
	name   string
	tags   map[string]string
	fields map[string]interface{}
	ts     time.Time
}

// 生成行协议
func (m *customerMeasurement) LineProto() (*io.Point, error) {
	return io.MakePoint(m.name, m.tags, m.fields, m.ts)
}

// 指定指标
func (m *customerMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: "mysql_customer",
		Tags: map[string]interface{}{
			"server": &inputs.TagInfo{
				Desc: "Server addr",
			},
		},
	}
}

func (i *Input) customSchemaMeasurement() ([]inputs.Measurement, error) {
	ms := []inputs.Measurement{}

	for _, item := range i.Query {
		resMap, err := i.query(item.sql)
		if err != nil {
			l.Errorf("custom sql %v query faild %v", item.sql, err)
			return nil, err
		}

		x, err := i.handleResponse(item, resMap)
		if err != nil {
			return nil, err
		}

		ms = append(ms, x...)
	}
	return ms, nil
}

func (i *Input) handleResponse(qy *customQuery,
	resMap []map[string]interface{}) ([]inputs.Measurement, error) {

	ms := []inputs.Measurement{}

	for _, item := range resMap {
		m := &customerMeasurement{
			name:   qy.metric,
			tags:   make(map[string]string),
			fields: make(map[string]interface{}),
		}

		for key, value := range i.Tags {
			m.tags[key] = value
		}

		if len(qy.tags) > 0 && len(qy.fields) == 0 {
			for _, tgKey := range qy.tags {
				if value, ok := item[tgKey]; ok {
					m.tags[tgKey] = cast.ToString(value)
					delete(item, tgKey)
				}
			}
			m.fields = item
		}

		if len(qy.tags) > 0 && len(qy.fields) > 0 {
			for _, tgKey := range qy.tags {
				if value, ok := item[tgKey]; ok {
					m.tags[tgKey] = cast.ToString(value)
					delete(item, tgKey)
				}
			}

			for _, fdKey := range qy.fields {
				if value, ok := item[fdKey]; ok {
					m.fields[fdKey] = value
				}
			}
		}

		if len(qy.tags) == 0 && len(qy.fields) == 0 {
			m.fields = item
		}
		m.ts = time.Now()

		if len(m.fields) > 0 {
			ms = append(ms, m)
		}
	}

	return ms, nil
}

func (i *Input) query(sql string) ([]map[string]interface{}, error) {
	rows, err := i.db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for idx, _ := range cache {
		var a interface{}
		cache[idx] = &a
	}
	var list []map[string]interface{}
	for rows.Next() {
		_ = rows.Scan(cache...)

		item := make(map[string]interface{})
		for i, data := range cache {
			key := strings.ToLower(columns[i])
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				switch vType.String() {
				case "int64":
					item[key] = val.(int64)
				case "string":
					var data interface{}
					data, err := strconv.ParseFloat(val.(string), 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					item[key] = val.(time.Time)
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					return nil, fmt.Errorf("unsupport data type '%s' now\n", vType)
				}
			}
		}

		list = append(list, item)
	}
	return list, nil
}
