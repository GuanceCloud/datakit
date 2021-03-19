package rabbitmqlog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "rabbitmqlog"

	sampleCfg = `
[[inputs.tailf]]
    # required, glob logfiles
    logfiles = ["/var/log/rabbitmq/*.log"]

    # glob filteer
    ignore = [""]

    source = "rabbitmqlog"

    # add service tag, if it's empty, use $source.
    service = "rabbitmqlog"

    # grok pipeline script path
    pipeline = "rabbitmqlog.p"

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
grok(_, "%{LOGLEVEL:status}%{DATA}====%{SPACE}%{DATA:time}%{SPACE}===%{SPACE}%{GREEDYDATA:msg}")

grok(_, "%{DATA:time} \\[%{LOGLEVEL:status}\\] %{GREEDYDATA:msg}")

default_time(time)
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"rabbitmq": pipelineCfg},
		)
		t.Source = inputName
		return t
	})
}
