package pipeline

import (
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"time"

	// "github.com/GuilhermeCaruso/kair"
	"github.com/araddon/dateparse"
	"github.com/mssola/user_agent"
	conv "github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/geo"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/ip2isp"
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

var datePattern = []struct {
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

func parseDatePattern(value string) (int64, error) {
	for _, p := range datePattern {
		if match, err := regexp.MatchString(p.pattern, value); err != nil {
			l.Errorf("regexp.MatchString: %s", err)
			return 0, err
		} else if match {
			if p.defaultYear {
				ty := time.Now()
				year := ty.Year()
				value = fmt.Sprintf("%s %d", value, year)
			}

			if tm, err := time.Parse(p.goFmt, value); err != nil {
				l.Errorf("time.Parse(): %s", err)
				return 0, err
			} else {
				unix_time := tm.UnixNano()
				l.Debugf("parse `%s` -> %v(nano: %d)", value, tm, tm.UnixNano())
				return unix_time, nil
			}
		}
	}

	return 0, nil
}

func TimestampHandle(value, tz string) (int64, error) {
	var t time.Time
	var err error
	var timezone = time.Local

	// pattern match first
	unix_time, err := parseDatePattern(value)
	if unix_time > 0 && err == nil {
		return unix_time, err
	}

	if tz != "" {
		timezone, err = time.LoadLocation(tz)
	}

	if err == nil {
		t, err = dateparse.ParseIn(value, timezone)
	}

	if err != nil {
		return 0, err
	} else {
		l.Debugf("parse `%s' -> %v(nano: %d)", value, t, t.UnixNano())
	}

	unix_time = t.UnixNano()
	return unix_time, nil
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

func JsonParse(jsonStr string) map[string]interface{} {
	res := make(map[string]interface{})
	jsonObj := gjson.Parse(jsonStr)

	if isObject(jsonObj) {
		parseJson2Map(jsonObj, res, "")
	} else if isArray(jsonObj) {
		for idx, obj := range jsonObj.Array() {
			key := fmt.Sprintf("[%d]", idx)
			parseJson2Map(obj, res, key)
		}
	}

	return res
}

func parseJson2Map(obj gjson.Result, res map[string]interface{}, prefix string) {
	if isObject(obj) {
		for key, value := range obj.Map() {
			if prefix != "" {
				key = prefix + "." + key
			}
			if isObject(value) {
				parseJson2Map(value, res, key)
			} else if isArray(value) {
				for idx, v := range value.Array() {
					fullkey := key + "[" + fmt.Sprintf("%d", idx) + "]"
					parseJson2Map(v, res, fullkey)
				}
			} else {
				res[key] = value.Value()
				continue
			}
		}
	} else {
		res[prefix] = obj.Value()
	}
}

func isObject(obj gjson.Result) bool {
	if obj.IsObject() {
		return true
	}
	return false
}

func isArray(obj gjson.Result) bool {
	if obj.IsArray() {
		return true
	}
	return false
}

func parseDate(yy, mm, dd, hh, mi, ss, ns, zone string) int64 {
	// 参数类型判断及转化(todo)
	var yyi, ddi, hi, mii, ssi, nsi int
	var mmi time.Month
	if len(yy) == 2 {
		year := "20" + yy
		yyi = conv.ToInt(year)
	} else {
		yyi = conv.ToInt(yy)
	}

	// day (todo)
	if mm != "" {
		fc := int32(mm[0])
		if isDigit(fc) {
			mi := conv.ToInt(mm)
			mmi = time.Month(mi)
		} else if isAlpha(fc) {
			if len(mm) < 5 {
				mmi = monthShort[mm]
			} else {
				mmi = monthLong[mm]
			}
		}
	}

	// day
	if dd != "" {
		ddi = conv.ToInt(dd)
	}

	// hour
	if hh != "" {
		hi = conv.ToInt(hh)
	}

	// minute
	if mi != "" {
		mii = conv.ToInt(mi)
	}

	// second
	if ss != "" {
		ssi = conv.ToInt(ss)
	}

	// millisecond
	if ns != "" {
		nsi = conv.ToInt(ns)
	}

	tz, err := time.LoadLocation(zone)
	if err == nil {
	} else {
		if zz, ok := timezoneList[zone]; ok {
			tz, err = time.LoadLocation(zz)
			if err != nil {
				l.Errorf("location time zone error %v", err)
			}
		}
	}

	t := time.Date(yyi, mmi, ddi, hi, mii, ssi, nsi, tz)
	res := t.UnixNano()

	fmt.Println("result =====>", res)

	return res
}

func isAlpha(ch int32) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}
func isDigit(ch int32) bool {
	return ch >= '0' && ch <= '9'
}

var monthLong = map[string]time.Month{
	"January":   time.January,
	"February":  time.February,
	"March":     time.March,
	"April":     time.April,
	"May":       time.May,
	"June":      time.June,
	"July":      time.July,
	"August":    time.August,
	"September": time.September,
	"October":   time.October,
	"November":  time.November,
	"December":  time.December,
}

var monthShort = map[string]time.Month{
	"jan":  time.January,
	"Feb":  time.February,
	"Mar":  time.March,
	"Apr":  time.April,
	"May":  time.May,
	"Jun":  time.June,
	"Jul":  time.July,
	"Aug":  time.August,
	"Sept": time.September,
	"Oct":  time.October,
	"Nov":  time.November,
	"Dec":  time.December,
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
}
