package ngnixlog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "ngnixlog"

	sampleCfg = `
[[inputs.tailf]]
    # glob logfiles
    # required
    logfiles = ["/usr/local/cloudcare/dataflux/datakit/*.txt"]

    # glob filteer
    ignore = [""]

    # required
    source = "ngnixlog"

    # grok pipeline script path
    pipeline_path = "ngnixlog.p"

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

    ## multiline parser/codec
    #[inputs.tailf.multiline]
    ## The pattern should be a regexp which matches what you believe to be an indicator that the field is part of an event consisting of multiple lines of log data.
    ## Note the use of '''XXX'''
    #pattern = '''^\s'''

    ## The field's value must be previous or next and indicates the relation to the multi-line event.
    #match_which_line = "previous"

    ## The invert_match can be true or false (defaults to false).
    ## If true, a message not matching the pattern will constitute a match of the multiline filter and the what will be applied. (vice-versa is also true)
    #invert_match = false

    # [inputs.tailf.tags]
    # tags1 = "value1"
`
	pipelineCfg = `
add_pattern("httpversion", "\\d+\\.\\d+")
add_pattern("date2", "%{year}[./]%{monthnum}[./]%{monthday} %{time}")

grok(_, "%{iporhost:client_ip} %{notspace:http_ident} %{notspace:http_auth} \\[%{httpdate:date}\\] \"%{data:http_method} %{greedydata:http_url} HTTP/%{httpversion:http_version}\" %{int:status_code} %{int:bytes}")

add_pattern("access_common", "%{iporhost:client_ip} %{notspace:http_ident} %{notspace:http_auth} \\[%{httpdate:date}\\] \"%{data:http_method} %{greedydata:http_url} HTTP/%{httpversion:http_version}\" %{int:status_code} %{int:bytes}")
grok(_, '%{access_common} "%{notspace:referrer}" "%{greedydata:agent}')

user_agent(agent)

grok(_, "%{date2:date} \\[%{loglevel:status}\\] %{greedydata:msg}")
grok(msg, "request: \"%{data:http_method} %{greedydata:http_url} HTTP/%{httpversion:http_version}\"")

cast(status_code, "int")
group_between(status_code, [200,299], "ok")

add_key(status, "error")
nullif(http_ident, "-")
default_time(date)
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"ngnixlog": pipelineCfg},
		)
	})
}
