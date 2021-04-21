package mysql

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"math"
	"strconv"
	"strings"
)

type ConversionFunc func(value sql.RawBytes) (interface{}, error)

func ParseInt(value sql.RawBytes) (interface{}, error) {
	v, err := strconv.ParseInt(string(value), 10, 64)

	// Ignore ErrRange.  When this error is set the returned value is "the
	// maximum magnitude integer of the appropriate bitSize and sign."
	if err, ok := err.(*strconv.NumError); ok && err.Err == strconv.ErrRange {
		return v, nil
	}

	return v, err
}

func ParseBoolAsInteger(value sql.RawBytes) (interface{}, error) {
	if bytes.EqualFold(value, []byte("YES")) || bytes.EqualFold(value, []byte("ON")) {
		return int64(1), nil
	}

	return int64(0), nil
}

func ParseGTIDMode(value sql.RawBytes) (interface{}, error) {
	// https://dev.mysql.com/doc/refman/8.0/en/replication-mode-change-online-concepts.html
	v := string(value)
	switch v {
	case "OFF":
		return int64(0), nil
	case "ON":
		return int64(1), nil
	case "OFF_PERMISSIVE":
		return int64(0), nil
	case "ON_PERMISSIVE":
		return int64(1), nil
	default:
		return nil, fmt.Errorf("unrecognized gtid_mode: %q", v)
	}
}

func ParseValue(value sql.RawBytes) (interface{}, error) {
	if bytes.EqualFold(value, []byte("YES")) || bytes.Compare(value, []byte("ON")) == 0 {
		return 1, nil
	}

	if bytes.EqualFold(value, []byte("NO")) || bytes.Compare(value, []byte("OFF")) == 0 {
		return 0, nil
	}

	if val, err := strconv.ParseInt(string(value), 10, 64); err != nil {
		if err, ok := err.(*strconv.NumError); ok && err.Err == strconv.ErrRange {
			return int64(math.MaxInt64), nil
		}
	} else {
		return val, nil
	}

	if val, err := strconv.ParseFloat(string(value), 64); err == nil {
		return val, nil
	}

	if len(string(value)) > 0 {
		return string(value), nil
	}

	return nil, fmt.Errorf("unconvertible value: %q", string(value))
}

var GlobalStatusConversions = map[string]ConversionFunc{
	"ssl_ctx_verify_depth": ParseInt,
	"ssl_verify_depth":     ParseInt,
}

var GlobalVariableConversions = map[string]ConversionFunc{
	// "gtid_mode": ParseGTIDMode,
}

func ConvertGlobalStatus(key string, value sql.RawBytes) (interface{}, error) {
	if bytes.Equal(value, []byte("")) {
		return nil, nil
	}

	if conv, ok := GlobalStatusConversions[key]; ok {
		return conv(value)
	}

	return ParseValue(value)
}

func ConvertGlobalVariables(key string, value sql.RawBytes) (interface{}, error) {
	if bytes.Equal(value, []byte("")) {
		return nil, nil
	}

	if conv, ok := GlobalVariableConversions[key]; ok {
		return conv(value)
	}

	return ParseValue(value)
}

func atof(str string) float64 {
	str = strings.Replace(str, ",", "", -1)
	str = strings.Replace(str, ";", "", -1)
	str = strings.Replace(str, "/s", "", -1)
	str = strings.Trim(str, " ")
	val, err := strconv.ParseFloat(str, 64)
	if err != nil {
		l.Warnf("Error occurred when parse %s to float. %s", str, err)
		val = 0
	}
	return val
}

func makeBigint(hi string, lo string) int64 {
	if lo == "" {
		val, _ := strconv.ParseInt(hi, 16, 64)
		return val
	}

	var hiVal int64
	var loVal int64
	if hi != "" {
		hiVal, _ = strconv.ParseInt(hi, 10, 64)
	}
	if lo != "" {
		loVal, _ = strconv.ParseInt(lo, 10, 64)
	}

	val := hiVal * loVal

	return val
}

func Conv(val interface{}, Datatype string) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	switch Datatype {
	case inputs.Float:
		res, err = cast.ToFloat64E(val)
	case inputs.Int:
		res, err = cast.ToInt64E(val)
	case inputs.Bool:
		res, err = cast.ToBoolE(val)
	case inputs.String:
		res, err = cast.ToStringE(val)
	}

	return res, err
}
