grok(_, '^apache %{IP:apache_ip}')
grok(_, '^mysql %{INT:mysql_id}')
grok(_, '^service=%{WORD:service} level=%{LOGLEVEL:level} request_id=%{NOTSPACE:request_id} method=%{WORD:method} path=%{NOTSPACE:path} status_code=%{INT:status_code} duration_ms=%{NUMBER:duration_ms} bytes=%{INT:bytes} region=%{NOTSPACE:region} user=%{NOTSPACE:user}$')
