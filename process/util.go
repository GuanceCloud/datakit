package process

import (
	"strings"
	"github.com/tidwall/gjson"
	"github.com/mssola/user_agent"
	"github.com/GuilhermeCaruso/kair"
)

func UrldecodeParse(url string) (interface{}, error) {
	params, err := url.ParseQuery(url)
	if err != nil {
		return nil, err
	}

	return params, nil
}

func UserAgentParse(str string) interface{} {
	ua := user_agent.Parse(str)

	return ua
}

func DateFormat(pattern string, data string) (interface{}, error) {
	t, err := time.Parse(data, "2006-01-02T15:04:05Z")

	if err !- nil {
		return nil, err
	}

	day := t.Day()
    year := t.Year()
    mounth := int(t.Month())
    hour := t.Hour()
    minute := t.Minute()
    sec := t.Second()

	datetime := kair.DateTime(day, mounth, year, hour, minute, sec)

	return date.CustomFormat(pattern), err
}