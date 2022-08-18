
#------------------------------------   警告   -------------------------------------
# 不要修改本文件，如果要更新，请拷贝至其它文件，最好以某种前缀区分，避免重启后被覆盖
#-----------------------------------------------------------------------------------

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
