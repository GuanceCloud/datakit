package redislog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "redislog"

	sampleCfg = `
[[inputs.tailf]]
    # required, glob logfiles
    logfiles = ["/var/log/redis/*.log"]

    # glob filteer
    ignore = [""]

    source = "redislog"

    # add service tag, if it's empty, use $source.
    service = "redislog"

    # grok pipeline script path
    pipeline = "redis.p"

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # optional encodings:
    #    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''this regexp'''
    # regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
    match = '''^\S.*'''

    [inputs.tailf.tags]
    # tags1 = "value1"
`
	pipelineCfg = `
add_pattern("date2", "%{MONTHDAY} %{MONTH} %{YEAR}?%{TIME}")

grok(_, "%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}")

group_in(serverity, ["."], "debug", status)
group_in(serverity, ["-"], "verbose", status)
group_in(serverity, ["*"], "notice", status)
group_in(serverity, ["#"], "warning", status)

cast(pid, "int")
default_time(time)
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"redis": pipelineCfg},
		)
		t.Source = inputName
		return t
	})
}
