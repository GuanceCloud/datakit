grok(_, '(?m)^%{TIMESTAMP_ISO8601:ts} %{LOGLEVEL:level} %{GREEDYDATA:summary}$')
grok(_, '(?m)^%{NOTSPACE:exception}: %{GREEDYDATA:error_text}$')
grok(_, '(?m)^Caused by: %{NOTSPACE:cause}: %{GREEDYDATA:cause_text}$')
