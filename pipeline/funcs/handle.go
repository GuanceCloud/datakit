package funcs

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mssola/user_agent"
	conv "github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
)

var datePattern = func() []struct {
	desc        string
	pattern     *regexp.Regexp
	goFmt       string
	defaultYear bool
} {
	dataPatternSource := []struct {
		desc        string
		pattern     string
		goFmt       string
		defaultYear bool
	}{
		{
			desc:    "nginx log datetime, 02/Jan/2006:15:04:05 -0700",
			pattern: `\d{2}/\w+/\d{4}:\d{2}:\d{2}:\d{2} \+\d{4}`,
			goFmt:   "02/Jan/2006:15:04:05 -0700",
		},
		{
			desc:    "redis log datetime, 14 May 2019 19:11:40.164",
			pattern: `\d{2} \w+ \d{4} \d{2}:\d{2}:\d{2}.\d{3}`,
			goFmt:   "02 Jan 2006 15:04:05.000",
		},
		{
			desc:        "redis log datetime, 14 May 19:11:40.164",
			pattern:     `\d{2} \w+ \d{2}:\d{2}:\d{2}.\d{3}`,
			goFmt:       "02 Jan 15:04:05.000 2006",
			defaultYear: true,
		},
		{
			desc:    "mysql, 171113 14:14:20",
			pattern: `\d{6} \d{2}:\d{2}:\d{2}`,
			goFmt:   "060102 15:04:05",
		},

		{
			desc:    "gin, 2021/02/27 - 14:14:20",
			pattern: `\d{4}/\d{2}/\d{2} - \d{2}:\d{2}:\d{2}`,
			goFmt:   "2006/01/02 - 15:04:05",
		},
		{
			desc:    "apache,  Tue May 18 06:25:05.176170 2021",
			pattern: `\w+ \w+ \d{2} \d{2}:\d{2}:\d{2}.\d{6} \d{4}`,
			goFmt:   "Mon Jan 2 15:04:05.000000 2006",
		},
		{
			desc:    "postgresql, 2021-05-27 06:54:14.760 UTC",
			pattern: `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3} UTC`,
			goFmt:   "2006-01-02 15:04:05.000 UTC",
		},
	}

	dst := []struct {
		desc        string
		pattern     *regexp.Regexp
		goFmt       string
		defaultYear bool
	}{}

	for _, p := range dataPatternSource {
		if c, err := regexp.Compile(p.pattern); err != nil {
			l.Errorf("Compile `%s` failed!", p.goFmt)
		} else {
			dst = append(dst, struct {
				desc        string
				pattern     *regexp.Regexp
				goFmt       string
				defaultYear bool
			}{
				desc:        p.desc,
				pattern:     c,
				goFmt:       p.goFmt,
				defaultYear: p.defaultYear,
			})
		}
	}
	return dst
}()

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

func GeoIPHandle(ip string) (map[string]string, error) {
	record, err := Geo(ip)
	if err != nil {
		return nil, err
	}

	res := make(map[string]string)

	res["city"] = record.City
	res["province"] = record.Region
	res["country"] = record.Country_short
	res["isp"] = ip2isp.SearchIsp(ip)

	return res, nil
}

func DateFormatHandle(data interface{}, precision string, fmts string) (interface{}, error) {
	v := conv.ToInt64(data)

	var t time.Time
	switch precision {
	case "s":
		t = time.Unix(v, 0)
	case "ms":
		num := v * int64(time.Millisecond)
		t = time.Unix(0, num)
	}

	for key, value := range dateFormatStr {
		if key == fmts {
			return t.Format(value), nil
		}
	}

	return "", fmt.Errorf("format pattern %v no support", fmts)
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

func parseDatePattern(value string, loc *time.Location) (int64, error) {
	valueCpy := value
	for _, p := range datePattern {
		if p.defaultYear {
			ty := time.Now()
			year := ty.Year()
			value = fmt.Sprintf("%s %d", value, year)
		} else {
			value = valueCpy
		}

		// 默认定义的规则能匹配，不匹配的则由 dataparse 处理
		if tm, err := time.ParseInLocation(p.goFmt, value, loc); err != nil {
			continue
		} else {
			unixTime := tm.UnixNano()
			return unixTime, nil
		}
	}
	return 0, fmt.Errorf("no match")
}

func TimestampHandle(value, tz string) (int64, error) {
	var t time.Time
	var err error
	timezone := time.Local

	if tz != "" {
		if timezone, err = time.LoadLocation(tz); err != nil {
			return 0, err
		}
	}

	// pattern match first
	unixTime, err := parseDatePattern(value, timezone)

	if unixTime > 0 && err == nil {
		return unixTime, nil
	}

	if t, err = dateparse.ParseIn(value, timezone); err != nil {
		return 0, err
	}

	// l.Debugf("parse `%s' -> %v(nano: %d)", value, t, t.UnixNano())

	return t.UnixNano(), nil
}

var dateFormatStr = map[string]string{
	"ANSIC":       time.ANSIC,
	"UnixDate":    time.UnixDate,
	"RubyDate":    time.RubyDate,
	"RFC822":      time.RFC822,
	"RFC822Z":     time.RFC822Z,
	"RFC850":      time.RFC850,
	"RFC1123":     time.RFC1123,
	"RFC1123Z":    time.RFC1123Z,
	"RFC3339":     time.RFC3339,
	"RFC3339Nano": time.RFC3339Nano,
	"Kitchen":     time.Kitchen,
	"Stamp":       time.Stamp,
	"StampMilli":  time.StampMilli,
	"StampMicro":  time.StampMicro,
	"StampNano":   time.StampNano,
}

func JSONParse(jsonStr string) map[string]interface{} {
	res := make(map[string]interface{})
	jsonObj := gjson.Parse(jsonStr)

	if isObject(jsonObj) {
		parseJSON2Map(jsonObj, res, "")
	} else if isArray(jsonObj) {
		for idx, obj := range jsonObj.Array() {
			key := fmt.Sprintf("[%d]", idx)
			parseJSON2Map(obj, res, key)
		}
	}

	return res
}

func parseJSON2Map(obj gjson.Result, res map[string]interface{}, prefix string) {
	if isObject(obj) {
		for key, value := range obj.Map() {
			if prefix != "" {
				key = prefix + "." + key
			}

			switch {
			case isObject(value):
				parseJSON2Map(value, res, key)
			case isArray(value):
				for idx, v := range value.Array() {
					fullkey := key + "[" + fmt.Sprintf("%d", idx) + "]"
					parseJSON2Map(v, res, fullkey)
				}
			default:
				res[key] = value.Value()
				continue
			}
		}
	} else {
		res[prefix] = obj.Value()
	}
}

func isObject(obj gjson.Result) bool {
	return obj.IsObject()
}

func isArray(obj gjson.Result) bool {
	return obj.IsArray()
}

var monthMaps = map[string]time.Month{
	"january":   time.January,
	"february":  time.February,
	"march":     time.March,
	"april":     time.April,
	"june":      time.June,
	"july":      time.July,
	"august":    time.August,
	"september": time.September,
	"october":   time.October,
	"november":  time.November,
	"december":  time.December,

	"jan": time.January,
	"feb": time.February,
	"mar": time.March,
	"apr": time.April,
	"may": time.May,
	"jun": time.June,
	"jul": time.July,
	"aug": time.August,
	"sep": time.September,
	"oct": time.October,
	"nov": time.November,
	"dec": time.December,
}

var timezoneList = map[string]string{
	"-11":    "Pacific/Midway",
	"-10":    "Pacific/Honolulu",
	"-9:30":  "Pacific/Marquesas",
	"-9":     "America/Anchorage",
	"-8":     "America/Los_Angeles",
	"-7":     "America/Phoenix",
	"-6":     "America/Chicago",
	"-5":     "America/New_York",
	"-4":     "America/Santiago",
	"-3:30":  "America/St_Johns",
	"-3":     "America/Sao_Paulo",
	"-2":     "America/Noronha",
	"-1":     "America/Scoresbysund",
	"+0":     "Europe/London",
	"+1":     "Europe/Vatican",
	"+2":     "Europe/Kiev",
	"+3":     "Europe/Moscow",
	"+3:30":  "Asia/Tehran",
	"+4":     "Asia/Dubai",
	"+4:30":  "Asia/Kabul",
	"+5":     "Asia/Samarkand",
	"+5:30":  "Asia/Kolkata",
	"+5:45":  "Asia/Kathmandu",
	"+6":     "Asia/Almaty",
	"+6:30":  "Asia/Yangon",
	"+7":     "Asia/Jakarta",
	"+8":     "Asia/Shanghai",
	"+8:45":  "Australia/Eucla",
	"+9":     "Asia/Tokyo",
	"+9:30":  "Australia/Darwin",
	"+10":    "Australia/Sydney",
	"+10:30": "Australia/Lord_Howe",
	"+11":    "Pacific/Guadalcanal",
	"+12":    "Pacific/Auckland",
	"+12:45": "Pacific/Chatham",
	"+13":    "Pacific/Apia",
	"+14":    "Pacific/Kiritimati",

	"CST": "Asia/Shanghai",
	"UTC": "Europe/London",
	// TODO: add more...
}
