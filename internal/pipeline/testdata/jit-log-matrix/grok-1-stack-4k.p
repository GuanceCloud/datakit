grok(_, '(?s)^%{TIMESTAMP_ISO8601:ts} %{LOGLEVEL:level} %{GREEDYDATA:stack}$')
