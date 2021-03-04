package ginlog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "ginlog"

	sampleCfg = `
[[inputs.tailf]]

    logfiles = [""]  # required
    source = "<your-source>" # required

    # glob filteer
    ignore = [""]

    # add service tag, if it's empty, use $source.
    service = "" # default same as $source

    # grok pipeline script path
    pipeline = "ginlog.p"

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # optional encodings:
    #    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''this regexp'''
    match = '''^\S.*'''

    [inputs.tailf.tags]
    # tags1 = "value1"
`
	pipelineCfg = `
add_pattern("gin_sep", "%{SPACE}\\|%{SPACE}")
add_pattern("gin_date", "%{YEAR}/%{MONTHNUM}/%{MONTHDAY}%{SPACE}-%{SPACE}%{HOUR}:%{MINUTE}:%{SECOND}")
add_pattern("gin_http_status_code", "%{INT}")
add_pattern("gin_resp_time", "%{NUMBER}(ns|µs|ms|s)")
add_pattern("gin_source_ip", "%{IP}")
add_pattern("gin_http_method", "%{WORD}")
add_pattern("gin_url", "%{GREEDYDATA}")

grok(_, '\\[GIN\\]%{SPACE}%{gin_date:time}%{gin_sep}%{gin_http_status_code:status}%{gin_sep}%{gin_resp_time:cost}%{gin_sep}%{gin_source_ip:client_ip}%{gin_sep}%{gin_http_method:method}%{SPACE}"%{gin_url:url}"')
default_time(time)

cast(status, "int")
group_between(status, [200,299], "OK", status)
group_between(status, [400,499], "warning", status)
group_between(status, [500,599], "error", status)
parse_duration(cost)
rename(cost_ns, cost)`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"ginlog": pipelineCfg},
		)
		return t
	})
}
