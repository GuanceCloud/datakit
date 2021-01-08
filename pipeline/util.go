package pipeline

import (
	"net/url"
	"time"

	"github.com/GuilhermeCaruso/kair"
	"github.com/mssola/user_agent"
)

func UrldecodeParse(path string) (interface{}, error) {
	params, err := url.ParseQuery(path)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func UserAgentParse(str string) interface{} {
	ua := user_agent.New(str)

	return ua
}

func DateFormat(pattern string, data string) (interface{}, error) {
	t, err := time.Parse(data, "2006-01-02T15:04:05Z")

	if err != nil {
		return nil, err
	}

	day := t.Day()
	year := t.Year()
	mounth := int(t.Month())
	hour := t.Hour()
	minute := t.Minute()
	sec := t.Second()

	datetime := kair.DateTime(day, mounth, year, hour, minute, sec)

	return datetime.CustomFormat(pattern), err
}

func GroupHandle(value interface{}, set []int, tag interface{}, with bool) interface{} {
	if with {
		for _, val := range set {
			if value == val {
				return tag
			}
		}
	} else {
		if value.(int) >= set[0] && value.(int) >= set[1] {
			return tag
		}
	}

	return nil
}
