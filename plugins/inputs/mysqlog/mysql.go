package mysqlog

import (
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/plugins/inputs/tailf"
)

const (
	inputName = "mysqlog"

	sampleCfg = `
[[inputs.tailf]]
    # required, glob logfiles
    logfiles = ["/var/log/mysql/*.log"]

    # glob filteer
    ignore = [""]

    source = "mysqlog"

    # add service tag, if it's empty, use $source.
    service = "mysqlog"

    # grok pipeline script path
    pipeline = "mysql.p"

    # read file from beginning
    # if from_begin was false, off auto discovery file
    from_beginning = false

    # optional encodings:
    #    "utf-8", "utf-16le", "utf-16le", "gbk", "gb18030" or ""
    character_encoding = ""

    # The pattern should be a regexp. Note the use of '''this regexp'''
    # regexp link: https://golang.org/pkg/regexp/syntax/#hdr-Syntax
    match = '''^(# Time|\d{4}-\d{2}-\d{2}|\d{6}\s+\d{2}:\d{2}:\d{2}).*'''

    [inputs.tailf.tags]
    # tags1 = "value1"
`
	pipelineCfg = `
grok(_, "%{TIMESTAMP_ISO8601:time}\\s+%{INT:thread_id}\\s+%{WORD:operation}\\s+%{GREEDYDATA:raw_query}")
grok(_, "%{TIMESTAMP_ISO8601:time} %{INT:thread_id} \\[%{NOTSPACE:status}\\] %{GREEDYDATA:msg}")

add_pattern("date2", "%{YEAR}%{MONTHNUM2}%{MONTHDAY} %{TIME}")
grok(_, "%{date2:time} \\s+(InnoDB:|\\[%{NOTSPACE:status}\\])\\s+%{GREEDYDATA:msg}")

add_pattern("timeline", "# Time: %{TIMESTAMP_ISO8601:time}")
add_pattern("userline", "# User@Host: %{NOTSPACE:db_user}\\s+@\\s(%{NOTSPACE:db_host})?(\\s+)?\\[(%{NOTSPACE:db_ip})?\\](\\s+Id:\\s+%{INT:query_id})?")
add_pattern("kvline01", "# Query_time: %{NUMBER:query_time}\\s+Lock_time: %{NUMBER:lock_time}\\s+Rows_sent: %{INT:rows_sent}\\s+Rows_examined: %{INT:rows_examined}")

add_pattern("kvline02", "# Thread_id: %{INT:thread_id}\\s+Killed: %{INT:killed}\\s+Errno: %{INT:errno}")
add_pattern("kvline03", "# Bytes_sent: %{INT:bytes_sent}\\s+Bytes_received: %{INT:bytes_received}")

# multi-line SQLs
add_pattern("sqls", "(?s)(.*)")

grok(_, "%{timeline}\\n%{userline}\\n%{kvline01}(\\n)?(%{kvline02})?(\\n)?(%{kvline03})?\\n%{sqls:db_slow_statement}")

cast(thread_id, "int")
cast(query_id, "int")
cast(rows_sent, "int")
cast(rows_examined, "int")

cast(bytes_sent, "int")
cast(bytes_received, "int")
cast(killed, "int")
cast(errno, "int")
cast(query_timestamp, "int")
cast(query_time, "float")
cast(lock_time, "float")

nullif(thread_id, 0)
nullif(db_host, "")
nullif(killed, 0)
nullif(bytes_sent, 0)
nullif(bytes_received, 0)
nullif(errno, 0)

default_time(time)
`
)

// match sample:
//
// # Time: 2019-11-27T10:43:13.460744Z
// 2017-12-29T12:04:09.954078Z 0 [Warning] System table 'plugin' is expected to be transactional
// 171113 14:14:20  InnoDB: Shutdown completed; log sequence number 1595675
func init() {
	inputs.Add(inputName, func() inputs.Input {
		t := tailf.NewTailf(
			inputName,
			"log",
			sampleCfg,
			map[string]string{"mysql": pipelineCfg},
		)
		t.Source = inputName
		return t
	})
}
