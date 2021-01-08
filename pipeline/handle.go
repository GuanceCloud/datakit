package pipeline

import (
	"net/url"
	"time"
	"fmt"
	"github.com/GuilhermeCaruso/kair"
	"xojoc.pw/useragent"
	"github.com/spf13/cast"
)

func UrldecodeHandle(path string) (interface{}, error) {
	params, err := url.QueryUnescape(path)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func UserAgentHandle(str string) interface{} {
	ua := useragent.Parse(str)

	return ua
}

func GeoIpHandle(str string) (interface{}, error) {
	// todo

	return res, nil
}

func DateFormatHandle(data int64, precision int64, fmts string, tz int) (interface{}, error) {
	if v, ok := data.(int64); !ok {
		return nil, fmt.Error("timestamp is not expect %v", data)
	}

	t := time.Unix(data, precision)

	day := t.Day()
	year := t.Year()
	mounth := int(t.Month())
	hour := t.Hour()
	minute := t.Minute()
	sec := t.Second()

	datetime := kair.DateTime(day, mounth, year, hour, minute, sec)

	return datetime.CustomFormat(pattern), nil
}

func GroupHandle(value interface{}, start, end float64) bool {
	num := cast.ToFloat64(value)

	if  num >= start && num <= end {
		return true
	}

	return false
}

func GroupInHandle(value interface{}, set []interface{}) bool {
	for _, val := range set {
		if reflect.DeepEqual(value, val) {
			return true
		}
	}

	return false
}
