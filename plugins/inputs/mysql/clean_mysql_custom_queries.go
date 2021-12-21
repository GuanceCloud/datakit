package mysql

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getCleanMysqlCustomQueries(r rows) []map[string]interface{} {
	if r == nil {
		return nil
	}

	defer closeRows(r)

	var list []map[string]interface{}

	columns, _ := r.Columns()
	columnLength := len(columns)
	cache := make([]interface{}, columnLength)
	for idx := range cache {
		var a interface{}
		cache[idx] = &a
	}
	for r.Next() {
		_ = r.Scan(cache...)

		item := make(map[string]interface{})
		for i, data := range cache {
			key := strings.ToLower(columns[i])
			val := *data.(*interface{})

			if val != nil {
				vType := reflect.TypeOf(val)

				switch vType.String() {
				case "int64":
					if v, ok := val.(int64); ok {
						item[key] = v
					} else {
						l.Warn("expect int64, ignored")
					}
				case "string":
					var data interface{}
					data, err := strconv.ParseFloat(val.(string), 64)
					if err != nil {
						data = val
					}
					item[key] = data
				case "time.Time":
					if v, ok := val.(time.Time); ok {
						item[key] = v
					} else {
						l.Warn("expect time.Time, ignored")
					}
				case "[]uint8":
					item[key] = string(val.([]uint8))
				default:
					l.Warn("unsupport data type '%s', ignored", vType)
				}
			}
		}

		list = append(list, item)
	}

	return list
}
