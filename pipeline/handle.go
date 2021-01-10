package pipeline

import (
	"net/url"
	"time"
	"reflect"
	"github.com/GuilhermeCaruso/kair"
	"github.com/mssola/user_agent"
	conv "github.com/spf13/cast"
)

func UrldecodeHandle(path string) (interface{}, error) {
	params, err := url.QueryUnescape(path)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func UserAgentHandle(str string) (res map[string]interface{}) {
	res = make(map[string]interface{})
	ua := user_agent.New(str)

	res["isMobile"] = ua.Mobile()
	res["isBot"] = ua.Bot()
	res["os"] = ua.OS()

	name, version := ua.Browser()
	res["browser"] = name
	res["browserVer"] = version

	en, v := ua.Engine()
	res["engine"] = en
	res["engineVer"] = v

	res["ua"] = ua.Platform()

	return res
}

func GeoIpHandle(str string) (interface{}, error) {
	// todo

	return nil, nil
}

func DateFormatHandle(data interface{}, precision string, fmts string, tz int) (interface{}, error) {
	v := conv.ToInt64(data)
	var num int64 = 0
	switch precision {
	case "s":
		num = 0
	case "ms":
		num = v * int64(time.Millisecond)
	}


	t := time.Unix(v, num)

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
