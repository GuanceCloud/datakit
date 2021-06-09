package dql

import (
	"testing"

	"github.com/luci/go-render/render"
	// "github.com/prometheus/prometheus/util/testutil"

	"gitlab.jiagouyun.com/cloudcare-tools/kodo/dql/parser"
)

/*
type testin struct {
	dql     string
	rewrite *RewriteAST
}

var (
	tests = []struct {
		in       testin
		expected string
		err      string
		fail     bool
	}{
		// FIXME:rewrite error
		// {
		// 	in:   testin{`cpu LIMIT 6000`, nil}, // rewrite fail: Limit should less than 5000
		// 	fail: true,
		// },
		// {
		// 	in:       testin{`cpu [2020-11-25 06:27:34:2020-11-25 08:27:34]`, &Rewrite{MaxDuration: "1h"}},
		// 	fail: true, // rewrite fail: time range should less than 1h0m0s
		// },
		// {
		// 	in:   testin{`cpu [2018-11-22 11:01:07:2020-11-25 08:27:34]`, nil},
		// 	fail: true, // rewrite fail: time range should less than 8760h0m0s  -> 1year
		// },
		{
			in:       testin{`cpu`, nil},
			expected: `SELECT * FROM "cpu" LIMIT 1000 SLIMIT 30`,
		},

		{
			in:       testin{`cpu LIMIT 2000`, nil},
			expected: `SELECT * FROM "cpu" LIMIT 2000 SLIMIT 30`,
		},

		// 1606285654363, 1606292854363
		// 2020-11-25 06:27:34  2020-11-25 08:27:34'

		{
			in:       testin{`cpu`, &RewriteAST{TimeRange: []int64{1606285654363, 1606292854363}}},
			expected: `SELECT * FROM "cpu" WHERE "time" >= '2020-11-25 06:27:34' AND "time" < '2020-11-25 08:27:34' LIMIT 1000 SLIMIT 30`,
		},

		{
			in:       testin{`cpu [2020-11-20 06:27:34:2020-11-21 08:27:34]`, &RewriteAST{TimeRange: []int64{1606285654363, 1606292854363}}},
			expected: `SELECT * FROM "cpu" WHERE "time" >= '2020-11-25 06:27:34' AND "time" < '2020-11-25 08:27:34' LIMIT 1000 SLIMIT 30`,
		},

		{
			in:       testin{`cpu`, &RewriteAST{TimeRange: []int64{1606285654363, 1606292854363}, Conditions: "a && (b||c)"}},
			expected: `SELECT * FROM "cpu" WHERE ("a" and ("b" or "c")) AND "time" >= '2020-11-25 06:27:34' AND "time" < '2020-11-25 08:27:34' LIMIT 1000 SLIMIT 30`,
		},

		{
			in:       testin{`cpu:(count(*)) [2020-11-20 06:27:34:2020-11-21 08:27:34:240000ms]`, nil},
			expected: `SELECT COUNT(*) FROM "cpu" WHERE "time" >= '2020-11-20 06:27:34' AND "time" < '2020-11-21 08:27:34' GROUP BY time(240000ms) LIMIT 1000 SLIMIT 30`,
		},

		{
			in:       testin{`cpu:(count(*)) [::AUTO]`, &RewriteAST{TimeRange: []int64{1606285654363, 1606292854363}}},
			expected: `SELECT COUNT(*) FROM "cpu" WHERE "time" >= '2020-11-25 06:27:34' AND "time" < '2020-11-25 08:27:34' GROUP BY time(20000ms) LIMIT 1000 SLIMIT 30`,
		},
	}
)

func TestRewrite(t *testing.T) {
	for index, tt := range tests {
		_ = index

		res, err := parser.ParseDQL(tt.in.dql)
		var out string

		if err != nil {
			t.Fatalf("TestRewrite @in '%s' @rewrite '%+#v' should always parse ok: %v",
				tt.in.dql, tt.in.rewrite, err)
		}

		stmts, ok := res.(parser.Stmts)
		if !ok {
			t.Fatal("no stmts")
		}

		testutil.Ok(t, err)

		for _, stmt := range stmts {
			switch stmt.(type) {
			case *parser.DFQuery:
				m := stmt.(*parser.DFQuery)
				if err := RewriteDFQuery(m, tt.in.rewrite); err != nil {
					t.Fatalf("[%d] %v (%+#v)-> rewrite fail: %v", index, tt.in.dql, tt.in.rewrite, err)
				}

				out, err = stmt.(*parser.DFQuery).InfluxQL()
			case *parser.Show:
				out, err = stmt.(*parser.Show).InfluxQL()
			default:
				t.Fatal("not support")
			}
		}

		if !tt.fail {
			t.Logf("[%d] %s -> %s", index, tt.in.dql, out)

			testutil.Ok(t, err)
			testutil.Equals(t, tt.expected, out)
		} else {
			t.Logf("[%d] %v -> expect fail: %v", index, tt.in.dql, err)
			testutil.NotOk(t, err, "")
		}
	}
}
*/

func TestAST(t *testing.T) {
	query := `cpu {x<1.23E3}`

	if query == "" {
		return
	}

	var err error
	res, err := parser.ParseDQL(query)
	if err != nil {
		t.Fatal(err)
	}

	stmts, ok := res.(parser.Stmts)
	if !ok {
		t.Fatal("no stmts")
	}

	t.Logf("Query -> %s\n\n", query)

	var printAST = func(m *parser.DFQuery) {
		const base = "\t\t%s: %s\n\n"
		t.Logf("AST:\n")
		if len(m.Targets) > 0 {
			t.Logf(base, "Targets", render.Render(m.Targets))
		}
		if len(m.WhereCondition) > 0 {
			t.Logf(base, "Where", render.Render(m.WhereCondition))
		}
		if m.TimeRange != nil {
			t.Logf(base, "TimeRange", render.Render(m.TimeRange))
			t.Logf(base, "GroupBy", render.Render(m.GroupBy))
		}
		if m.OrderBy != nil {
			t.Logf(base, "OrderBy", render.Render(m.OrderBy))
		}
	}

	var out interface{}
	for _, stmt := range stmts {
		switch v := stmt.(type) {
		case *parser.DFQuery:
			printAST(v)

			switch v.Namespace {
			case NSMetric, NSMetricAbbr, "":
				out, err = v.InfluxQL()
			default: // ES
				out, err = v.ESQL()
			}

		case *parser.Show:
			switch v.Namespace {
			case NSMetric, NSMetricAbbr, "":
				out, err = v.InfluxQL()
			default: // ES
				out, err = v.ESQL()
			}

		default:
			t.Fatal("not support")
		}
	}

	if err != nil {
		t.Logf("Err -> %s\n", err)
	} else {
		t.Logf("OK -> %s\n", out)
	}

}
