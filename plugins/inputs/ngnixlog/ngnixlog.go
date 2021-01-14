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
    pipeline_path = ""

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
grok(_, "%{_iporhost:clientip} - %{_username:remote_user} \\[%{_httpdate:date_timestamp}\\] \"(?:%{_word:method} %{_notspace:request_uri}(?: HTTP/%{_number:httpversion})?|%{_data:raw_request})\" %{_int:status} %{_number:body_bytes_sent} \"%{_data:http_referer}\" \"%{_data:http_user_agent}\"")

grok(_, "(?:%{_ipv4:clientip}|-)(?:,\s[\d.]+)* (?:%{_data:remote_user}|-) (?:%{_data:ident}|-) \\[%{_httpdate:date_timestamp}\\] \"(?:%{_word:method} %{_notspace:request_uri}(?: HTTP/%{_number:httpversion})?|%{_data:raw_request})\" %{_number:status} (?:%{_number:body_bytes_sent}|-) \"%{_data:http_referer}\" \"%{_data:http_user_agent}\" (?:%{_number:request_length}|-) (?:%{_number:bytes_sent}|-) (?:%{_number:request_time}|-")

grok(_, "(?<timestamp>%{_year}[./]%{_monthnum}[./]%{_monthday} %{time}) \[%{_loglevel:level}\] %{_posint:pid}#%{_number:threadid}\: \*%{_number:connectionid} %{_greedydata:message}, client: %{_ip:clientip}, server: %{_greedydata:server}, request: "(?<request>%{_word:method} %{_unixpath:path} http/(?<httpversion>[0-9.]*))"(, )?(upstream: "(?<upstream>[^,]*)")?(, )?(host: "(?<host>[^,]*)")?")

cast(body_bytes_sent, "int")
cast(status, "int")
`
)

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"ngnixlog.p": pipelineCfg},
		)
	})
}
