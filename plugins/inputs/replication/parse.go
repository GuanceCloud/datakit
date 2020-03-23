package replication

import (
	"strconv"
	"time"
	"unsafe"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	"github.com/jackc/pgx"
	"github.com/nickelser/parselogical"
)

// Parse test_decoding format wal to WalData
func parse(sub *Subscribe, msg *pgx.WalMessage) (*influxdb.Point, error) {
	result := parselogical.NewParseResult(Bytes2String(msg.WalData))
	if err := result.Parse(); err != nil {
		return nil, err
	}

	if _, ok := sub.eventsOperation[result.Operation]; !ok {
		// INSERT, UPDATE, DELETE, BEGIN, COMMIT
		return nil, nil
	}

	var tags = make(map[string]string, len(sub.Tags))
	var fields = make(map[string]interface{}, len(sub.Fields))

	for key, column := range result.Columns {
		if _, ok := sub.pointConfig.tags[key]; ok {
			tags[key] = column.Value
		}

		if _, ok := sub.pointConfig.fields[key]; ok {
			fields[key] = fieldsTransform(column.Type, column.Value)
		}
	}

	return influxdb.NewPoint(sub.pointConfig.measurement, tags, fields, time.Now())
}

func fieldsTransform(typer, value string) interface{} {

	switch typer {
	case "smallint", "integer", "bigint", "real", "double precision", "smallserial", "serial", "bigserial", "money":
		if val, err := strconv.ParseInt(value, 10, 64); err == nil {
			return val
		}
	case "decimal", "numeric":
		if val, err := strconv.ParseFloat(value, 64); err == nil {
			return val
		}
	case "boolean":
		// FIXME: pg boolean has 'true','false','unknow'
		if val, err := strconv.ParseBool(value); err == nil {
			return val
		}
	default:
		return value
	}

	return nil
}

// String2Bytes convert string to bytes
func String2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

// Bytes2String convert bytes to string
func Bytes2String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
