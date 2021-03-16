package activemqlog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "activemqlog"

	sampleCfg = `
[[inputs.tailf]]
    # required, glob logfiles
    logfiles = ["/var/log/activemqlog/*.log"]

    # glob filteer
    ignore = [""]

    source = "activemqlog"

    # add service tag, if it's empty, use $source.
    service = "activemqlog"

    # grok pipeline script path
    pipeline = "activemqlog.p"

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
grok(_, "^%{TIMESTAMP_ISO8601:time}%{SPACE}\\|%{SPACE}%{WORD:status}%{SPACE}%{SPACE}\\|%{SPACE}%{USERNAME:user}%{SPACE}%{WORD}%{SPACE}%{URIPATH:action_url}%{SPACE}\\[%{GREEDYDATA:msg}\\] from %{IP:client_ip}")

grok(_, "^%{LOGLEVEL:status}%{SPACE}%{SPACE}\\|%{SPACE}%{WORD:user}%{SPACE}%{WORD}%{SPACE}%{URIPATH:action_url}%{SPACE}\\[%{GREEDYDATA}\\] from  %{IP:client_ip}")

grok(_, "^%{TIMESTAMP_ISO8601:time}%{SPACE}%{WORD:status}%{SPACE}\\[%{EMAILLOCALPART:name}\\]%{SPACE}%{WORD:message_id}:%{SPACE}%{GREEDYDATA:msg}")


grok(_, "^%{TIMESTAMP_ISO8601:time}%{SPACE}%{WORD:status}%{SPACE}\\[%{EMAILLOCALPART:name}\\]%{SPACE}%{GREEDYDATA:msg}")

grok(_, "^%{TIMESTAMP_ISO8601:time}%{SPACE}\\[%{WORD:type}\\]\\(%{USERNAME:thread_name}%{SPACE}\\(%{GREEDYDATA:name}\\)\\)%{SPACE}%{WORD:message_id}:%{SPACE}%{GREEDYDATA:msg}")

default_time(time)
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"activemq": pipelineCfg},
		)
		t.Source = inputName
		return t
	})
}
