package nginxlog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "nginxlog"

	sampleCfg = `
[[inputs.tailf]]
    # glob logfiles
    # required
    logfiles = ["/var/log/nginx/*.log"]

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
add_pattern("date2", "%{year}[./]%{monthnum}[./]%{monthday} %{time}")

# access log
grok(_, "%{iporhost:client_ip} %{notspace:http_ident} %{notspace:http_auth} \\[%{httpdate:date_access}\\] \"%{data:http_method} %{greedydata:http_url} HTTP/%{number:http_version}\" %{int:status_code} %{int:bytes}")

# access log
add_pattern("access_common", "%{iporhost:client_ip} %{notspace:http_ident} %{notspace:http_auth} \\[%{httpdate:date_access}\\] \"%{data:http_method} %{greedydata:http_url} HTTP/%{number:http_version}\" %{int:status_code} %{int:bytes}")
grok(_, '%{access_common} "%{notspace:referrer}" "%{greedydata:agent}')
user_agent(agent)

# error log
grok(_, "%{date2:date_access} \\[%{loglevel:level}\\] %{greedydata:msg}, client: %{iporhost:client}, server: %{iporhost:server}, request: \"%{data:http_method} %{greedydata:http_url} HTTP/%{number:http_version}\", host: \"%{iporhost:host}\"")

cast(status_code, "int")
cast(bytes, "int")

group_between(status_code, [200,299], "OK", status)
group_between(status_code, [300,399], "notice", status)
group_between(status_code, [400,499], "warnning", status)
group_between(status_code, [500,599], "error", status)

nullif(http_ident, "-")
nullif(http_auth, "-")
default_time(date_access)
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
