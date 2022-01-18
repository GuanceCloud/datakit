package multiline

import (
	"testing"

	tu "gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

func TestMultiline(t *testing.T) {
	const maxLines = 10

	cases := []struct {
		pattern string
		in, out [][]byte
	}{
		{
			pattern: "^(# Time|\\d{4}-\\d{2}-\\d{2}|\\d{6}\\s+\\d{2}:\\d{2}:\\d{2})",
			in: [][]byte{
				[]byte("# Time: 2021-05-31T11:15:26.043419Z"),
				[]byte("# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228"),
				[]byte("# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288"),
				[]byte("SET timestamp=1622459726;"),
				[]byte("SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb"),
				[]byte("                FROM     information_schema.tables"),
				[]byte("                GROUP BY table_schema;"),
			},
			out: [][]byte{
				[]byte("# Time: 2021-05-31T11:15:26.043419Z\n# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228\n# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288\nSET timestamp=1622459726;\nSELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb\n                FROM     information_schema.tables\n                GROUP BY table_schema;"),
			},
		},
		{
			pattern: "^# Time",
			in: [][]byte{
				[]byte("# Time: 2021-05-31T11:15:26.043419Z"),
				[]byte("# Line: 2 ========================"),
				[]byte("# Line: 3 ========================"),
				[]byte("# Line: 4 ========================"),
				[]byte("# Line: 5 ========================"),
				[]byte("# Line: 6 ========================"),
				[]byte("# Line: 7 ========================"),
				[]byte("# Line: 8 ========================"),
				[]byte("# Line: 9 ========================"),
				[]byte("# Line: 10 ======================="),
				[]byte("# Line: 11 ======================="),
			},
			out: [][]byte{
				[]byte(
					"# Time: 2021-05-31T11:15:26.043419Z\n# Line: 2 ========================\n# Line: 3 ========================\n# Line: 4 ========================\n# Line: 5 ========================\n# Line: 6 ========================\n# Line: 7 ========================\n# Line: 8 ========================\n# Line: 9 ========================\n# Line: 10 ======================="),
				[]byte("# Line: 11 ======================="),
			},
		},
		{
			pattern: "^(# Time|\\d{4}-\\d{2}-\\d{2}|\\d{6}\\s+\\d{2}:\\d{2}:\\d{2})",
			in:      [][]byte{[]byte("2021-05-31T11:15:26.043419Z")},
			out:     [][]byte{[]byte("2021-05-31T11:15:26.043419Z")},
		},
		{
			pattern: "",
			in:      [][]byte{[]byte("2021-05-31T11:15:26.043419Z\n  Line:10================")},
			out:     [][]byte{[]byte("2021-05-31T11:15:26.043419Z\n  Line:10================")},
		},
	}

	for _, tc := range cases {
		m, err := New(tc.pattern, maxLines)
		tu.Equals(t, nil, err)

		outIdx := 0
		for _, line := range tc.in {
			res := m.ProcessLine(line)
			if len(res) != 0 {
				tu.Equals(t, tc.out[outIdx], res)
				outIdx++
			}
		}

		if m.CacheLines() != 0 {
			tu.Equals(t, tc.out[outIdx], m.Flush())
		}
	}
}

func TestMultilineString(t *testing.T) {
	const maxLines = 10

	cases := []struct {
		pattern string
		in, out []string
	}{
		{
			pattern: "^(# Time|\\d{4}-\\d{2}-\\d{2}|\\d{6}\\s+\\d{2}:\\d{2}:\\d{2})",
			in: []string{
				"# Time: 2021-05-31T11:15:26.043419Z",
				"# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228",
				"# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288",
				"SET timestamp=1622459726;",
				"SELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb",
				"                FROM     information_schema.tables",
				"                GROUP BY table_schema;",
			},
			out: []string{"# Time: 2021-05-31T11:15:26.043419Z\n# User@Host: datakitMonitor[datakitMonitor] @ localhost []  Id:  1228\n# Query_time: 0.015214  Lock_time: 0.000112 Rows_sent: 4  Rows_examined: 288\nSET timestamp=1622459726;\nSELECT   table_schema, IFNULL(SUM(data_length+index_length)/1024/1024,0) AS total_mb\n                FROM     information_schema.tables\n                GROUP BY table_schema;"},
		},
		{
			pattern: "^# Time",
			in: []string{
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
			out: []string{
				"# Time: 2021-05-31T11:15:26.043419Z\n# Line: 2 ========================\n# Line: 3 ========================\n# Line: 4 ========================\n# Line: 5 ========================\n# Line: 6 ========================\n# Line: 7 ========================\n# Line: 8 ========================\n# Line: 9 ========================\n# Line: 10 =======================",
				"# Line: 11 =======================",
			},
		},
		{
			pattern: "^(# Time|\\d{4}-\\d{2}-\\d{2}|\\d{6}\\s+\\d{2}:\\d{2}:\\d{2})",
			in:      []string{"2021-05-31T11:15:26.043419Z"},
			out:     []string{"2021-05-31T11:15:26.043419Z"},
		},
		{
			pattern: "",
			in:      []string{"2021-05-31T11:15:26.043419Z\n  Line:10================"},
			out:     []string{"2021-05-31T11:15:26.043419Z\n  Line:10================"},
		},
	}

	for _, tc := range cases {
		m, err := New(tc.pattern, maxLines)
		tu.Equals(t, nil, err)

		outIdx := 0
		for _, line := range tc.in {
			res := m.ProcessLineString(line)
			if res != "" {
				tu.Equals(t, tc.out[outIdx], res)
				outIdx++
			}
		}

		if m.CacheLines() != 0 {
			tu.Equals(t, tc.out[outIdx], m.FlushString())
		}
	}
}
