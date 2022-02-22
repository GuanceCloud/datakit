package funcs

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	conv "github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func fexpr(node parser.Node) *parser.FuncStmt {
	if x, ok := node.(*parser.FuncStmt); ok {
		return x
	}
	return nil
}

func arglistForIndexOne(fe *parser.FuncStmt) parser.FuncArgList {
	return arglist(fe, 1)
}

func arglist(fe *parser.FuncStmt, idx int) parser.FuncArgList {
	if x, ok := fe.Param[idx].(parser.FuncArgList); ok {
		return x
	}
	return nil
}

const (
	InvalidInt   = math.MinInt32
	DefaultInt   = int64(0xdeadbeef)
	DefaultStr   = ""
	InvalidStr   = "deadbeaf"
	InvalidFloat = math.SmallestNonzeroFloat64
	DefaultFloat = float64(0.0)
)

func fixYear(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Year(), nil
	}

	if x < 0 {
		return -1, fmt.Errorf("year should larger than 0")
	}

	// year like 02 -> 2002, 21 -> 2021
	if x < int64(100) { //nolint:gomnd
		x += int64(now.Year() / 100 * 100) //nolint:gomnd
		return int(x), nil
	}

	return int(x), nil
}

func fixMonth(now time.Time, x string) (time.Month, error) {
	if x == DefaultStr {
		return now.Month(), nil
	} else {
		if v, err := strconv.ParseInt(x, 10, 64); err == nil { // digital month: i.e., 01, 11
			if v < 1 || v > 12 {
				return time.Month(-1), fmt.Errorf("month should between [1,12], got %x,", x)
			}
			return time.Month(v), nil
		} else { // month like aug/august, january/jan
			v, ok := monthMaps[strings.ToLower(x)]
			if !ok {
				return InvalidInt, fmt.Errorf("unknown month %s", x)
			}
			return v, nil
		}
	}
}

func fixDay(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Day(), nil
	}

	if x < 1 || x > 31 {
		return -1, fmt.Errorf("month day should between [1,31], got %d", x)
	}

	return int(x), nil
}

func fixHour(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Hour(), nil
	}

	if x < 0 || x > 23 {
		return -1, fmt.Errorf("hour should between [0,24], got %d", x)
	}

	return int(x), nil
}

func fixMinute(now time.Time, x int64) (int, error) {
	if x == DefaultInt {
		return now.Minute(), nil
	}

	if x < 0 || x > 59 {
		return -1, fmt.Errorf("minute should between [0,59], got %d", x)
	}

	return int(x), nil
}

func fixSecond(x int64) (int, error) {
	if x == DefaultInt {
		return 0, nil
	}

	if x < 0 || x > 59 {
		return -1, fmt.Errorf("second should between [0,59], got %d", x)
	}

	return int(x), nil
}

func tz(s string) (z *time.Location, err error) {
	z, err = time.LoadLocation(s)
	if err != nil {
		if _, ok := timezoneList[s]; !ok {
			return nil, fmt.Errorf("unknown timezone %s", s)
		}

		z, err = time.LoadLocation(timezoneList[s])
		if err != nil {
			return nil, err
		}
	}

	return z, nil
}

func doCast(result interface{}, tInfo string) interface{} {
	switch strings.ToLower(tInfo) {
	case "bool":
		return conv.ToBool(result)

	case "int":
		return conv.ToInt64(conv.ToFloat64(result))

	case "float":
		return conv.ToFloat64(result)

	case "str":
		return conv.ToString(result)
	}

	return nil
}
