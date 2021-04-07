package kafkalog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "kafkalog"

	sampleCfg = `
[[inputs.tailf]]
    # required, glob logfiles
    logfiles = ["/var/log/kafka/*.log"]

    # glob filteer
    ignore = [""]

    source = "kafkalog"

    # add service tag, if it's empty, use $source.
    service = "kafkalog"

    # grok pipeline script path
    pipeline = "kafkalog.p"

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
grok(_, "%{DATA:time} \\[%{WORD:thread_name}\\] %{WORD:status}  %{WORD:name} - %{GREEDYDATA:msg}")


grok(_, "^%{INT:duration} \\[%{WORD:thread_name}\\] %{LOGLEVEL:status} %{GREEDYDATA:name} - %{GREEDYDATA:msg}")

add_pattern("date", "%{INT}-%{INT}-%{INT} %{INT}:%{INT}:%{INT}")
grok(_, "^%{date:time} %{LOGLEVEL:status} %{DATA:name}:%{INT:line} - %{GREEDYDATA:msg}")


add_pattern("date1", "%{INT}-%{INT}-%{INT} %{INT}:%{INT}:%{INT},%{INT}")
grok(_, "^\\[%{date1:time}\\] %{WORD:status} %{DATA:msg} \\(%{DATA:name}\\)")

default_time(time)
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"kafka": pipelineCfg},
		)
		t.Source = inputName
		return t
	})
}
