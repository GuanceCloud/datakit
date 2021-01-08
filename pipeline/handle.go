package pipeline

import (
	"net/url"
	"time"
	"fmt"
	"reflect"
	"github.com/GuilhermeCaruso/kair"
	"xojoc.pw/useragent"
	conv "github.com/spf13/cast"
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

	return nil, nil
}

func DateFormatHandle(data interface{}, precision int64, fmts string, tz int) (interface{}, error) {
	v, ok := data.(int64);
	if !ok {
		return nil, fmt.Errorf("timestamp is not expect %v", data)
	}

	t := time.Unix(v, precision)

	day := t.Day()
	year := t.Year()
	mounth := int(t.Month())
	hour := t.Hour()
	minute := t.Minute()
	sec := t.Second()

	datetime := kair.DateTime(day, mounth, year, hour, minute, sec)

	return datetime.CustomFormat(fmts), nil
}

func GroupHandle(value interface{}, start, end float64) bool {
	num := conv.ToFloat64(value)

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
