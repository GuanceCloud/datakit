package parser

import (
	"testing"
	"time"

	"github.com/prometheus/prometheus/util/testutil"
)

const (
	layout    = "2006-01-02 15:04:05"
	startTime = "2020-11-09 11:20:46"
	endTime   = "2020-11-10 11:20:46"
)

var (
	stime, _ = time.Parse(layout, startTime)
	etime, _ = time.Parse(layout, endTime)
)

var (
	transferCases = []struct {
		in, expected string
		err          string
		fail         bool
	}{
		// rp test
		{
			in:       "df_metering:(usage)",
			expected: `SELECT "usage" FROM "autogen"."df_metering"`,
		},

		{
			in:       "df_bill:(usage)",
			expected: `SELECT "usage" FROM "autogen"."df_bill"`,
		},

		{
			in:       `show_tag_value(from=["df_metering", "mem"], keyin=["h2o"]) {name="dql"} LIMIT 10 OFFSET 3`,
			expected: `SHOW TAG VALUES FROM "autogen"."df_metering", "mem" WITH KEY IN ("h2o") WHERE "name" = 'dql' LIMIT 10 OFFSET 3`,
		},

		{
			in:       `show_tag_key(from=['df_bill', "h30"])`,
			expected: `SHOW TAG KEYS FROM "autogen"."df_bill", "h30"`,
		},

		{
			in:       `show_field_key(from=['df_bill', "h30"], from=["df_metering", "h50"])`,
			expected: `SHOW FIELD KEYS FROM "autogen"."df_bill", "h30", "autogen"."df_metering", "h50"`,
		},

		{
			in: `show_tag_key(from=[df_bill, "h30"])`,
			// unexpected error: from only accept string list values, got <nil>
			fail: true,
		},
		// rp test end

		{
			in:       "cpu:(count_distinct())",
			expected: `SELECT COUNT(DISTINCT(*)) FROM "cpu"`,
		},

		{
			in:       "cpu:(count_distinct(`人数`))",
			expected: `SELECT COUNT(DISTINCT("人数")) FROM "cpu"`,
		},

		{
			in:       "cpu:(count(int(`人数`)))",
			expected: `SELECT COUNT("人数"::integer) FROM "cpu"`,
		},

		{
			in:       "cpu:(int(`人数`))",
			expected: `SELECT "人数"::integer FROM "cpu"`,
		},

		{
			in:       `cpu:(float(usage))`,
			expected: `SELECT "usage"::float FROM "cpu"`,
		},

		{
			in:       `cpu { hostname = "RedHat", host in [red, true] }`,
			expected: `SELECT * FROM "cpu" WHERE "hostname" = 'RedHat' AND ("host" = "red" OR "host" = true)`,
		},

		{
			in:       `cpu { host in [mac, "ubuntu", 123] }`,
			expected: `SELECT * FROM "cpu" WHERE ("host" = "mac" OR "host" = 'ubuntu' OR "host" = 123)`,
		},

		{
			in: `cpu:(count(*), name)`,
			// expect fail: Metric mixing aggregate and non-aggregate queries is not supported
			fail: true,
		},

		{
			in: `cpu:(name) [::86400ms]`,
			// expect fail: Metric GROUP BY interval require at least a aggregate function in target list
			fail: true,
		},

		{
			in:       `cpu:(count(*)*3) [::86400ms] BY mem, syslog`,
			expected: `SELECT COUNT(*) * 3 FROM "cpu" GROUP BY time(86400ms), "mem", "syslog"`,
		},

		{
			in:       `cpu:(avg(count(*))) [::86400ms]`,
			expected: `SELECT MEAN(COUNT(*)) FROM "cpu" GROUP BY time(86400ms)`,
		},

		{
			in:       `cpu:(count(*)) [::86400ms]`,
			expected: `SELECT COUNT(*) FROM "cpu" GROUP BY time(86400ms)`,
		},

		{
			in:       "cpu {\"地🍺区\" = '中🍺国'}",
			expected: "SELECT * FROM \"cpu\" WHERE \"地🍺区\" = '中🍺国'",
		},

		{
			in:       `cpu {"地区" = "中🍺国"}`,
			expected: "SELECT * FROM \"cpu\" WHERE \"地区\" = '中🍺国'",
		},

		{
			in:       "cpu {`地区` = '中🍺国'}",
			expected: "SELECT * FROM \"cpu\" WHERE \"地区\" = '中🍺国'",
		},

		{
			in:       "cpu:(`中国`)",
			expected: `SELECT "中国" FROM "cpu"`,
		},

		{
			in:       `cpu:(f1)`,
			expected: `SELECT "f1" FROM "cpu"`,
		},

		// ORDER BY
		{
			in: "cpu:(COUNT(*) AS cnt) ORDER BY `notime`",
			// expect fail: Metric only ORDER BY time supported
			fail: true,
		},

		{
			in: `cpu:(f1, f2) ORDER BY f3`,
			// expect fail: Metric only ORDER BY time supported
			fail: true,
		},

		{
			in: `cpu:(f1, f2) ORDER BY f3, f4`,
			// expect fail: Metric only ORDER BY time supported
			fail: true,
		},

		{
			in:       "cpu:(COUNT(*) AS cnt) ORDER BY `time` DESC",
			expected: `SELECT COUNT(*) AS "cnt" FROM "cpu" ORDER BY "time" DESC`,
		},

		{
			in:       `cpu:(COUNT(*) AS cnt) ORDER BY time`,
			expected: `SELECT COUNT(*) AS "cnt" FROM "cpu" ORDER BY "time"`,
		},

		{
			in:       `cpu:(count(*)) [2020-11-22 11:01:07:2020-11-23 11:01:07]`,
			expected: `SELECT COUNT(*) FROM "cpu" WHERE "time" >= '2020-11-22 11:01:07' AND "time" < '2020-11-23 11:01:07'`,
		},

		{
			in:       `show_field_key(from=['h2o', "h30"], from=["h40", "h50"])`,
			expected: `SHOW FIELD KEYS FROM "h2o", "h30", "h40", "h50"`,
		},

		{
			in:       `show_field_key(from=['h2o', "h30"])`,
			expected: `SHOW FIELD KEYS FROM "h2o", "h30"`,
		},

		{
			in:       `show_tag_key(from=['h2o', "h30"])`,
			expected: `SHOW TAG KEYS FROM "h2o", "h30"`,
		},

		{
			in:   `show_field_key(from=["h2o"]) {name="dql"} LIMIT 10`,
			fail: true,
		},

		{
			in:       `show_tag_value(from=["cpu", "mem"], keyin=["h2o"]) {name="dql"} LIMIT 10 OFFSET 3`,
			expected: `SHOW TAG VALUES FROM "cpu", "mem" WITH KEY IN ("h2o") WHERE "name" = 'dql' LIMIT 10 OFFSET 3`,
		},

		{
			in:       `show_tag_value(keyin=["h2o"]) {name="dql"} LIMIT 10 OFFSET 3`,
			expected: `SHOW TAG VALUES WITH KEY IN ("h2o") WHERE "name" = 'dql' LIMIT 10 OFFSET 3`,
		},

		{
			in:   `show_tag_key(from=['h2o']) [10m::10s]`, // has time range interval
			fail: true,
		},

		{
			in:       `show_tag_key(from=['h2o']) [2006-01-02 15:04:05:2020-11-10 11:20:46] OFFSET 3`,
			expected: `SHOW TAG KEYS FROM "h2o" WHERE "time" >= '2006-01-02 15:04:05' AND "time" < '2020-11-10 11:20:46' OFFSET 3`,
		},

		{
			in:       `show_measurement(re("h2o.*")) {name="dql"} [2006-01-02 15:04:05:2020-11-10 11:20:46] LIMIT 10 OFFSET 3`,
			expected: `SHOW MEASUREMENTS WITH MEASUREMENT =~ /h2o.*/ WHERE "name" = 'dql' AND "time" >= '2006-01-02 15:04:05' AND "time" < '2020-11-10 11:20:46' LIMIT 10 OFFSET 3`,
		},

		{
			in:   `show_tag_value(keyin=["h2o", 123 ])`,
			fail: true,
		},

		{
			in:       `show_tag_value(keyin=["h2o"])`,
			expected: `SHOW TAG VALUES WITH KEY IN ("h2o")`,
		},

		{
			in:       `show_tag_value(keyin=["h2o", "cpu"])`,
			expected: `SHOW TAG VALUES WITH KEY IN ("h2o", "cpu")`,
		},

		{
			in:       `show_tag_key(from=['h2o'])`,
			expected: `SHOW TAG KEYS FROM "h2o"`,
		},

		{
			in:   `show_tag_key(a="b")`,
			fail: true,
		},

		{
			in:       `show_tag_key()`,
			expected: `SHOW TAG KEYS`,
		},

		{
			in:       `show_field_key(from=["h2o"])`,
			expected: `SHOW FIELD KEYS FROM "h2o"`,
		},

		{
			in:   `show_field_key(re("h2o"))`,
			fail: true,
		},

		{
			in:       `show_field_key()`,
			expected: `SHOW FIELD KEYS`,
		},

		{
			in:       `show_measurement(re("h2o.*"))`,
			expected: `SHOW MEASUREMENTS WITH MEASUREMENT =~ /h2o.*/`,
		},

		{
			in:       `show_measurement()`,
			expected: `SHOW MEASUREMENTS`,
		},

		{
			in:       `cpu { a>123.45 && (b!=re("*abc") || c!="123"), d=re("456*")}`,
			expected: `SELECT * FROM "cpu" WHERE "a" > 123.450000 and ("b" !~ /*abc/ or "c" != '123') AND "d" =~ /456*/`,
		},

		// subquery
		{

			in:       `((a):(d)):(e)`,
			expected: `SELECT "e" FROM (SELECT "d" FROM (SELECT * FROM "a"))`,
		},

		// metric
		{
			in: `metric::
				cpu:(fill(COUNT(), linear), COUNT(water_level))
					[2006-01-02 15:04:05:2020-11-10 11:20:46:10s,5s]
					BY water_level`,

			expected: `SELECT COUNT(*), COUNT("water_level") FROM "cpu" WHERE "time" >= '2006-01-02 15:04:05' AND "time" < '2020-11-10 11:20:46' GROUP BY time(10000ms, 5000ms), "water_level"`,
		},

		// timezone
		{
			in:       `cpu:(count()) tz("+8")`,
			expected: `SELECT COUNT(*) FROM "cpu" tz('Asia/Shanghai')`,
		},

		// metric
		{

			in: `cpu:(COUNT(*)) {name="萧峰"}
					[2006-01-02 15:04:05:2020-11-10 11:20:46:10s,5s]
					BY water_level`,
			expected: `SELECT COUNT(*) FROM "cpu" WHERE "name" = '萧峰' AND "time" >= '2006-01-02 15:04:05' AND "time" < '2020-11-10 11:20:46' GROUP BY time(10000ms, 5000ms), "water_level"`,
		},

		// case insensitive function name
		{
			in:       `cpu:(cOunt())`,
			expected: `SELECT COUNT(*) FROM "cpu"`,
		},
	}
)

func TestDQL2Influx(t *testing.T) {
	for idx, c := range transferCases {
		_ = idx

		res, err := ParseDQL(c.in)
		var out string

		if err != nil {
			t.Fatalf("TestDQL2Influx @in '%s' should always parse ok: %v", c.in, err)
		}

		stmts, ok := res.(Stmts)
		if !ok {
			t.Fatal("no stmts")
		}

		testutil.Ok(t, err)

		for _, stmt := range stmts {
			switch v := stmt.(type) {
			case *DFQuery:
				out, err = v.InfluxQL()
			case *Show:
				out, err = v.InfluxQL()
			default:
				t.Fatal("not support")
			}
		}

		if !c.fail {
			t.Logf("[%d] %s -> %s", idx, c.in, out)

			testutil.Ok(t, err)
			testutil.Equals(t, c.expected, out)
		} else {
			t.Logf("[%d] %v -> expect fail: %v", idx, c.in, err)
			testutil.NotOk(t, err, "")
		}
	}
}
