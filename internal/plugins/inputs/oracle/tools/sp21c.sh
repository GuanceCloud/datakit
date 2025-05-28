./sqlplus.t \
	-conn "oracle://sys:123456@8.153.108.66:1522/XE?timeout=30" \
	-f $1
