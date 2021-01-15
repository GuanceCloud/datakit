package pipeline

import (
	"net/url"
	"reflect"
	"time"
	"regexp"

	"github.com/GuilhermeCaruso/kair"
	"github.com/araddon/dateparse"
	"github.com/mssola/user_agent"
	conv "github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
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

func GeoIpHandle(ip string) (map[string]string, error) {
	record, err := geo.Geo(ip)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)

	res["city"] = record.City
	res["region"] = record.Region
	res["country"] = record.Country_short
	res["city"] = record.City
	res["isp"] = record.Isp

	return res, nil
}

func DateFormatHandle(data interface{}, precision string, fmts string, tz int) (interface{}, error) {
	v := conv.ToInt64(data)

	var t time.Time
	switch precision {
	case "s":
		t = time.Unix(v, 0)
	case "ms":
		num := v * int64(time.Millisecond)
		t = time.Unix(0, num)
	}

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

	if num >= start && num <= end {
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

func TimestampHandle(value string) (int64, error) {
	if match, err := regexp.MatchString(`\d{2}/\w+/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4}`, value); err != nil {
		return 0, err
	} else if match {
		// 06/Jan/2017:16:16:37 +0000
		if tm, err := time.Parse("02/Jan/2006:15:04:05 -0700", "06/Jan/2017:16:16:37 +0000"); err != nil {
			return 0, err
		} else {
			unix_time := tm.UnixNano()
			return unix_time, nil
		}
	}

	t, err := dateparse.ParseLocal(value)
	if err != nil {
		return 0, err
	}

	unix_time := t.UnixNano()

	return unix_time, nil
}
