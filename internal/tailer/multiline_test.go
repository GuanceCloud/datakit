package tailer

import (
	"testing"
)

func TestMultiline(t *testing.T) {
	m, err := NewMultiline("^(# Time|\\d{4}-\\d{2}-\\d{2}|\\d{6}\\s+\\d{2}:\\d{2}:\\d{2})", 10)
	if err != nil {
		panic(err)
	}

	slows := [][]string{
		{
			"# Time: 2021-05-31T11:15:26.043419Z",
			"# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228",
			"# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288",
			"SET timestamp=1622459726;",
			"SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb",
			"                FROM     information_schema.tables",
			"                GROUP BY table_schema;",
		},
		{
			"# Time: 2021-05-31T11:15:26.043419Z",
			"# Line: 2 ========================",
			"# Line: 3 ========================",
			"# Line: 4 ========================",
			"# Line: 5 ========================",
			"# Line: 6 ========================",
			"# Line: 7 ========================",
			"# Line: 8 ========================",
			"# Line: 9 ========================",
			"# Line: 10 =======================",
			"# Line: 11 =======================",
		},
		{
			"2021-05-31T11:15:26.043419Z",
		},
		{
			"2021-05-31T11:15:26.043419Z",
		},
	}

	for _, slow := range slows {
		for _, in := range slow {
			res := m.ProcessLine(in)
			if res != "" {
				t.Log(res)
			}
		}
	}
}
