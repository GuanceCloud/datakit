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
add_pattern("date2", "%{monthday} %{month} %{time}")

grok(_, "%{int:pid}:%{word:role} %{date2:date_access} %{notspace:serverity} %{greedydata:msg}")

group_in(serverity, ["."], "debug", level)
group_in(serverity, ["-"], "verbose", level)
group_in(serverity, ["*"], "notice", level)
group_in(serverity, ["#"], "warnning", level)
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{inputName: pipelineCfg},
		)
		t.Source = inputName
		t.Pipeline = inputName + ".p"
		return t
	})
}
