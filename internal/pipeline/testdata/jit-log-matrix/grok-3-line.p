grok(_, 'service=%{WORD:service}')
grok(_, 'level=%{LOGLEVEL:level}')
grok(_, 'request_id=%{NOTSPACE:request_id}')
