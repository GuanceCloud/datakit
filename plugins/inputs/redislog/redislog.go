package redislog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "redislog"

	sampleCfg = `
[[inputs.tailf]]
    # glob logfiles
    # required
    logfiles = ["/var/log/redis/*.log"]

    # glob filteer
    ignore = [""]

    # add service tag, if it's empty, use "redislog".
    service = ""

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    ## characters are replaced using the unicode replacement character
    ## When set to the empty string the data is not decoded to text.
    ## ex: character_encoding = "utf-8"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "utf-16le"
    ##     character_encoding = "gbk"
    ##     character_encoding = "gb18030"
    ##     character_encoding = ""
    #character_encoding = ""

    # [inputs.tailf.tags]
    # tags1 = "value1"
`
	pipelineCfg = `
add_pattern("date2", "%{MONTHDAY} %{MONTH} %{YEAR}?%{TIME}")

grok(_, "%{INT:pid}:%{WORD:role} %{date2:time} %{NOTSPACE:serverity} %{GREEDYDATA:msg}")

group_in(serverity, ["."], "debug", status)
group_in(serverity, ["-"], "verbose", status)
group_in(serverity, ["*"], "notice", status)
group_in(serverity, ["#"], "warnning", status)

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
		t.Pipeline = "redis.p"
		return t
	})
}
