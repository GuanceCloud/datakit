package parser

import (
	"testing"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
)

type parseCase struct {
	in       string
	expected interface{}
	err      string
	fail     bool
}

var queryCases = []*parseCase{

	// 时间间隔为10秒，点数为5，计算所得interval为2s
	{
		in: `M::mem [1615883598:1615883608:auto(5)]`,
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"mem"},
				TimeRange: &TimeRange{
					Start: &TimeExpr{
						Time: time.Unix(0, 1615883598000000000),
					},
					End: &TimeExpr{
						Time: time.Unix(0, 1615883608000000000),
					},
					Resolution: &TimeResolution{PointNum: &NumberLiteral{IsInt: true, Int: 5}, Duration: time.Second * 2},
				},
			},
		},
	},

	{
		in: `M::mem [1615883598:1615883608:auto]`,
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"mem"},
				TimeRange: &TimeRange{
					Start: &TimeExpr{
						Time: time.Unix(0, 1615883598000000000),
					},
					End: &TimeExpr{
						Time: time.Unix(0, 1615883608000000000),
					},
					Resolution: &TimeResolution{Auto: true /*, Duration: time.Duration(27777777)*/},
				},
			},
		},
	},

	{
		in: `M::mem [::point(0)]`,
		// parse error: point() only accept integer param and should larger than 0
		fail: true,
	},

	{
		in: `M::mem [::point(10)]`,
		// parse error: use point() should have start time
		fail: true,
	},

	// order by list
	{
		in: `cpu ORDER BY avg_b`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			OrderBy: &OrderBy{
				List: NodeList{
					&OrderByElem{Column: "avg_b", Opt: OrderAsc},
				},
			},
		}},
	},

	{
		in: `cpu ORDER BY avg_b desc`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			OrderBy: &OrderBy{
				List: NodeList{
					&OrderByElem{Column: "avg_b", Opt: OrderDesc},
				},
			},
		}},
	},

	{
		in: `cpu ORDER BY avg_b desc, avg_c`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			OrderBy: &OrderBy{
				List: NodeList{
					&OrderByElem{Column: "avg_b", Opt: OrderDesc},
					&OrderByElem{Column: "avg_c", Opt: OrderAsc},
				},
			},
		}},
	},

	{
		in: `cpu ORDER BY avg_b, avg_c`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			OrderBy: &OrderBy{
				List: NodeList{
					&OrderByElem{Column: "avg_b", Opt: OrderAsc},
					&OrderByElem{Column: "avg_c", Opt: OrderAsc},
				},
			},
		}},
	},

	{
		in: `cpu ORDER BY "avg_b", "avg_c"`,
		// unexpected: string "\"avg_b\""
		fail: true,
	},

	{
		in: `cpu ORDER BY 1212, "avg_c"`,
		// parse error: unexpected: number "1212"
		fail: true,
	},

	{
		in: `cpu ORDER BY ok desc, "avg_c"`,
		// unexpected: string "\"avg_c\""
		fail: true,
	},

	// in statement
	{
		in: `M::cpu {x > 0, f in [1,2,3,4]}`,
		expected: Stmts{&DFQuery{
			Namespace: "M",
			Names:     []string{"cpu"},
			WhereCondition: []Node{
				&BinaryExpr{
					Op:  GT,
					LHS: &Identifier{Name: "x"},
					RHS: &NumberLiteral{IsInt: true, Int: 0},
				},
				&BinaryExpr{
					Op:  IN,
					LHS: &Identifier{Name: "f"},
					RHS: &NodeList{
						&NumberLiteral{IsInt: true, Int: 1},
						&NumberLiteral{IsInt: true, Int: 2},
						&NumberLiteral{IsInt: true, Int: 3},
						&NumberLiteral{IsInt: true, Int: 4},
					},
				},
			},
		},
		},
	},

	// comments
	{
		in: `###
		M::cpu
		###`,

		expected: Stmts{&DFQuery{
			Namespace: "M",
			Names:     []string{"cpu"},
		},
		},
	},

	{
		in: `# comments header
		metrics::           # this is namespace
				NAME, mem, re('*def') # this is metric list
				:(host)               # this is target-list
				# end of ql`,

		expected: Stmts{&DFQuery{
			Namespace:  "metrics",
			Names:      []string{"NAME", "mem"},
			RegexNames: []*Regex{{Regex: "*def"}},
			Targets: []*Target{
				{
					Col: &Identifier{Name: "host"},
				},
			},
		},
		},
	},

	// static_cast to int
	{
		in: "M::mem:(int(usage))",
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"mem"},
				Targets:   []*Target{{Col: &StaticCast{IsInt: true, Val: &Identifier{Name: "usage"}}}},
			},
		},
	},

	{
		in: "M::mem:(avg(int(usage)))",
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"mem"},
				Targets: []*Target{
					{
						Col: &FuncExpr{
							Name: "avg",
							Param: []Node{
								&StaticCast{IsInt: true, Val: &Identifier{Name: "usage"}},
							},
						},
					},
				},
			},
		},
	},

	// static_cast to float
	{
		in: "M::mem:(float(usage))",
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"mem"},
				Targets:   []*Target{{Col: &StaticCast{IsFloat: true, Val: &Identifier{Name: "usage"}}}},
			},
		},
	},

	{
		in:   "M::mem { float(usage) > 10.11 }",
		fail: true,
	},

	// link-with syntax
	{
		in: `cpu:(f1, cpu_host) {host="abc"} FILTER O::ecs:(ecs_1, host) { host = "ubuntu" } FILTER O::aliyun:(aliyun_1, hostname as host) LIMIT 10 WITH { cpu_host = host }`,
		expected: Stmts{
			&Lambda{
				Left: &DFQuery{
					Names: []string{"cpu"},
					Targets: []*Target{
						{Col: &Identifier{Name: "f1"}},
						{Col: &Identifier{Name: "cpu_host"}},
					},
					WhereCondition: []Node{&BinaryExpr{Op: EQ, LHS: &Identifier{Name: "host"}, RHS: &StringLiteral{Val: "abc"}}},
				},
				Opt: LambdaFilter,
				Right: []*DFQuery{
					{
						Namespace: "O",
						Names:     []string{"ecs"},
						Targets: []*Target{
							{Col: &Identifier{Name: "ecs_1"}},
							{Col: &Identifier{Name: "host"}},
						},
						WhereCondition: []Node{
							&BinaryExpr{
								Op:  EQ,
								LHS: &Identifier{Name: "host"},
								RHS: &StringLiteral{Val: "ubuntu"},
							},
						},
					},
					{
						Namespace: "O",
						Names:     []string{"aliyun"},
						Targets: []*Target{
							{Col: &Identifier{Name: "aliyun_1"}},
							{Col: &Identifier{Name: "hostname"}, Alias: "host"},
						},
						Limit: &Limit{Limit: 10},
					},
				},
				WhereCondition: []Node{
					&BinaryExpr{
						Op:  EQ,
						LHS: &Identifier{Name: "cpu_host"},
						RHS: &Identifier{Name: "host"},
					},
				},
			},
		},
	},

	// link-with syntax
	{
		in: `cpu:(f1, f2) {host="abc"} LINK O::ecs:(a1,a2) {a1 = "xxx"} WITH {f1=a1}`,
		expected: Stmts{
			&Lambda{
				Left: &DFQuery{
					Names: []string{"cpu"},
					Targets: []*Target{
						{Col: &Identifier{Name: `f1`}},
						{Col: &Identifier{Name: `f2`}},
					},
					WhereCondition: []Node{&BinaryExpr{Op: EQ, LHS: &Identifier{Name: "host"}, RHS: &StringLiteral{Val: "abc"}}},
				},
				Opt: LambdaLink,
				Right: []*DFQuery{{
					Namespace: "O",
					Names:     []string{"ecs"},
					Targets: []*Target{
						{Col: &Identifier{Name: `a1`}},
						{Col: &Identifier{Name: `a2`}},
					},
					WhereCondition: []Node{&BinaryExpr{Op: EQ, LHS: &Identifier{Name: "a1"}, RHS: &StringLiteral{Val: "xxx"}}},
				}},
				WhereCondition: []Node{&BinaryExpr{Op: EQ, LHS: &Identifier{Name: "f1"}, RHS: &Identifier{Name: "a1"}}},
			},
		},
	},

	{
		in: "M::mem:(identifier('abc```def'))",
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"mem"},
				Targets:   []*Target{{Col: &Identifier{Name: "abc```def"}}},
			},
		},
	},

	{
		in: `M::mem [10m:5m:1m]`,
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"mem"},
				TimeRange: &TimeRange{
					Start:      &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
					End:        &TimeExpr{IsDuration: true, Duration: time.Minute * 5},
					Resolution: &TimeResolution{Duration: time.Minute * 1},
				},
			},
		},
	},
	{
		in: `F::dataflux_dev.dql:(sort())`,
		expected: Stmts{
			&DFQuery{
				Namespace: "F",
				Names:     []string{`dataflux_dev__dql`},
				Targets:   []*Target{{Col: &FuncExpr{Name: `sort`}}},
			},
		},
	},

	// cascade functions
	{
		in: `F::DQL:(
		sort(dql("cpu:(usage) {host='a'}").nextf(f1=123).nextf(f2="abc").nextf(f3=456.7))
		)`,

		expected: Stmts{
			&DFQuery{
				Namespace: `F`,
				Names:     []string{"DQL"},
				Targets: []*Target{{
					Col: &FuncExpr{
						Name: "sort",
						Param: []Node{&CascadeFunctions{
							Funcs: []*FuncExpr{
								{
									Name:  "dql",
									Param: []Node{&StringLiteral{Val: "cpu:(usage) {host='a'}"}},
								},
								{
									Name:  "nextf",
									Param: []Node{&FuncArg{ArgName: "f1", ArgVal: &NumberLiteral{IsInt: true, Int: 123}}},
								},
								{
									Name:  "nextf",
									Param: []Node{&FuncArg{ArgName: "f2", ArgVal: &StringLiteral{Val: "abc"}}},
								},
								{
									Name:  "nextf",
									Param: []Node{&FuncArg{ArgName: "f3", ArgVal: &NumberLiteral{Float: 456.7}}},
								},
							}},
						},
					}},
				},
			},
		},
	},
	{

		in: `F::DQL:(
			sort(
			identifier("in")=dql("cpu:(usage) {host='a'} [10m]")).next_func1().next_func2(),
			sort2(identifier("in")="cpu:(usage) {host='a'} [10m]")
			)`,
		expected: Stmts{
			&DFQuery{
				Namespace: "F",
				Names:     []string{"DQL"},
				Targets: []*Target{
					{
						Col: &CascadeFunctions{
							Funcs: []*FuncExpr{
								{
									Name: "sort",
									Param: []Node{
										&FuncArg{
											ArgName: "in",
											ArgVal: &FuncExpr{
												Name: "dql",
												Param: []Node{&StringLiteral{
													Val: "cpu:(usage) {host='a'} [10m]",
												}}}},
									},
								},
								{
									Name: "next_func1",
								},
								{
									Name: "next_func2",
								},
							},
						},
					},

					{
						Col: &FuncExpr{
							Name: "sort2",
							Param: []Node{
								&FuncArg{ArgName: "in", ArgVal: &StringLiteral{Val: "cpu:(usage) {host='a'} [10m]"}}}},
					},
				},
			},
		},
	},

	{

		in: `F::DQL_expr__module:(SORT(data=dql("cpu:(usage))) {host='a'} [5m]"), key=nil, reverse=FALSE))`,

		expected: Stmts{
			&DFQuery{
				//Anonymous: true,
				Namespace: "F",
				Names:     []string{"DQL_expr__module"},
				Targets: []*Target{
					{
						Col: &FuncExpr{
							Name: "SORT",
							Param: []Node{
								&FuncArg{ArgName: "data", ArgVal: &FuncExpr{Name: "dql", Param: []Node{&StringLiteral{Val: "cpu:(usage))) {host='a'} [5m]"}}}},
								&FuncArg{ArgName: "key", ArgVal: &NilLiteral{}},
								&FuncArg{ArgName: "reverse", ArgVal: &BoolLiteral{Val: false}},
							},
						},
					},
				},
			},
		},
	},

	{
		in: `_:(ratio(cpu.f1, m.mem.f3))`,
		expected: Stmts{
			&DFQuery{
				Anonymous: true,
				Names:     []string{"_"},
				Targets: []*Target{
					{
						Col: &FuncExpr{
							Name: `ratio`,
							Param: []Node{
								&AttrExpr{Obj: &Identifier{Name: "cpu"}, Attr: &Identifier{Name: "f1"}},
								&AttrExpr{Obj: &Identifier{Name: "m"},
									Attr: &AttrExpr{
										Obj: &Identifier{Name: "mem"}, Attr: &Identifier{Name: "f3"},
									},
								},
							},
						},
					},
				},
			},
		},
	},

	// semantic checking
	// {
	// 	in:   `cpu:(avg(b) AS avg_b, "name", 2077)`,
	// 	fail: true,
	// },

	// FIXME:
	// type assert
	{
		in:   `cpu BY "string1", "string2"`,
		fail: true,
	},
	{
		in:   `cpu BY 2077, 2066`,
		fail: true,
	},
	{
		in:   `cpu BY nil`,
		fail: true,
	},
	{
		in:   `cpu ORDER BY "string"`,
		fail: true,
	},

	{
		in: "M::cpu {x>1.23E3}", // scientific notation
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{"cpu"},
				WhereCondition: []Node{
					&BinaryExpr{Op: GT, LHS: &Identifier{Name: "x"}, RHS: &NumberLiteral{Float: 1230.00}, ReturnBool: true},
				},
			},
		},
	},

	{
		in: `(
				M::cpu {x>1.23E3}
			):(user_usage) {y>1.23}`,
		expected: Stmts{

			&DFQuery{
				Targets:        []*Target{{Col: &Identifier{Name: "user_usage"}}},
				WhereCondition: []Node{&BinaryExpr{Op: GT, LHS: &Identifier{Name: "y"}, RHS: &NumberLiteral{Float: 1.23}}},
				Subquery: &DFQuery{
					Namespace: "M",
					Names:     []string{"cpu"},
					WhereCondition: []Node{
						&BinaryExpr{Op: GT, LHS: &Identifier{Name: "x"}, RHS: &NumberLiteral{Float: 1230.00}, ReturnBool: true},
					},
				},
			},
		},
	},

	{
		in: `x{( a!=true || b=false && c="abc" )}`,
		expected: Stmts{
			&DFQuery{
				//Namespace: "M",
				Names: []string{`x`},
				WhereCondition: []Node{
					&ParenExpr{
						Param: &BinaryExpr{
							ReturnBool: true,
							Op:         OR,
							LHS:        &BinaryExpr{ReturnBool: true, Op: NEQ, LHS: &Identifier{Name: "a"}, RHS: &BoolLiteral{Val: true}},
							RHS: &BinaryExpr{
								ReturnBool: true,
								Op:         AND,
								LHS:        &BinaryExpr{ReturnBool: true, Op: EQ, LHS: &Identifier{Name: "b"}, RHS: &BoolLiteral{Val: false}},
								RHS: &BinaryExpr{
									ReturnBool: true,
									Op:         EQ,
									LHS:        &Identifier{Name: "c"},
									RHS:        &StringLiteral{Val: "abc"},
								}}},
					}},
			}},
	},

	// multi-stmt
	{
		in: "M::cpu; M::mem",
		expected: Stmts{
			&DFQuery{
				Namespace: "M",
				Names:     []string{`cpu`},
			},
			&DFQuery{
				Namespace: "M",
				Names:     []string{`mem`},
			},
		},
	},

	// paren expr
	{
		in: `cpu { (a!=true || b=false), c=("abc") }`,
		expected: Stmts{&DFQuery{
			Names: []string{`cpu`},
			WhereCondition: []Node{
				&ParenExpr{
					Param: &BinaryExpr{
						Op:         OR,
						LHS:        &BinaryExpr{ReturnBool: true, Op: NEQ, LHS: &Identifier{Name: "a"}, RHS: &BoolLiteral{Val: true}},
						RHS:        &BinaryExpr{ReturnBool: true, Op: EQ, LHS: &Identifier{Name: "b"}, RHS: &BoolLiteral{Val: false}},
						ReturnBool: true,
					}},

				&BinaryExpr{
					Op:  EQ,
					LHS: &Identifier{Name: "c"},
					RHS: &ParenExpr{Param: &StringLiteral{Val: "abc"}},
				}},
		}},
	},

	{
		in: `show_tag_value(cpu, keyin = ["region", "host"]) {service="redis"} [10m] LIMIT 3 OFFSET 5`,
		expected: Stmts{&Show{
			Func: &FuncExpr{
				Name: `show_tag_value`,
				Param: []Node{
					&Identifier{Name: "cpu"},
					&FuncArg{
						ArgName: "keyin",
						ArgVal: FuncArgList{
							&StringLiteral{Val: "region"},
							&StringLiteral{Val: "host"},
						}}}},
			WhereCondition: []Node{
				&BinaryExpr{
					Op:  EQ,
					LHS: &Identifier{Name: "service"},
					RHS: &StringLiteral{Val: "redis"},
				},
			},
			TimeRange: &TimeRange{
				Start: &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			},
			Limit:  &Limit{Limit: 3},
			Offset: &Offset{Offset: 5},
		}},
	},

	{
		in:   `show_tag_value(cpuuuuu, keyin = ["region", "host"]) {service="redis"} [10m::1m] SLIMIT 3 SOFFSET 5`,
		fail: true,
	},

	{
		in: `show_tag_value(cpu, keyin = ["region", "host"]) {service="redis"}`,
		expected: Stmts{&Show{
			Func: &FuncExpr{
				Name: `show_tag_value`,
				Param: []Node{
					&Identifier{Name: "cpu"},
					&FuncArg{
						ArgName: "keyin",
						ArgVal: FuncArgList{
							&StringLiteral{Val: "region"},
							&StringLiteral{Val: "host"},
						},
					},
				},
			},
			WhereCondition: []Node{
				&BinaryExpr{
					Op:  EQ,
					LHS: &Identifier{Name: "service"},
					RHS: &StringLiteral{Val: "redis"},
				},
			},
		}},
	},

	{
		in: ` show_tag_value(cpu) `,
		expected: Stmts{&Show{
			Func: &FuncExpr{
				Name: `show_tag_value`,
				Param: []Node{
					&Identifier{Name: "cpu"},
				},
			},
		}},
	},

	// ----------
	// show end

	// TRUE/ FALSE
	{
		in: `cpu {a!=true, b=false}`,
		expected: Stmts{&DFQuery{
			Names: []string{`cpu`},
			WhereCondition: []Node{
				&BinaryExpr{
					Op:  NEQ,
					LHS: &Identifier{Name: "a"},
					RHS: &BoolLiteral{Val: true},
				},

				&BinaryExpr{
					Op:  EQ,
					LHS: &Identifier{Name: "b"},
					RHS: &BoolLiteral{Val: false},
				}}}}},

	// NIL
	{
		in: `cpu {a!=NIL}`,
		expected: Stmts{&DFQuery{
			Names: []string{`cpu`},
			WhereCondition: []Node{
				&BinaryExpr{
					Op:  NEQ,
					LHS: &Identifier{Name: "a"},
					RHS: &NilLiteral{},
				},
			},
		},
		},
	},

	// AND OR test
	{
		in: `cpu {a>123.45 and b!=re("*abc")}`,
		expected: Stmts{&DFQuery{
			Names: []string{`cpu`},
			WhereCondition: []Node{
				&BinaryExpr{
					Op: AND,
					LHS: &BinaryExpr{
						Op:  GT,
						LHS: &Identifier{Name: "a"},
						RHS: &NumberLiteral{Float: 123.45},
					},
					RHS: &BinaryExpr{
						Op:  NEQ,
						LHS: &Identifier{Name: "b"},
						RHS: &Regex{Regex: `*abc`},
					},
				}}}},
	},

	// timezone
	{
		in:   `metrics:: cpu tz("Asia/Shangdong")`,
		fail: true,
	},
	{
		in:   `metrics:: cpu tz("+22")`,
		fail: true,
	},
	{
		in: `metrics:: cpu tz("+8")`,
		expected: Stmts{&DFQuery{
			Namespace: "metrics",
			Names:     []string{"cpu"},
			TimeZone:  &TimeZone{Input: "+8", TimeZone: "Asia/Shanghai"},
		}}},

	// multiple embed metric query
	{
		in: `E::(
					O::(
							M::a
						):(d)
				):(e)
			`,
		expected: Stmts{&DFQuery{
			Namespace: "E",
			Subquery: &DFQuery{
				Namespace: "O",
				Subquery: &DFQuery{
					Namespace: "M",
					Names:     []string{"a"},
				},
				Targets: []*Target{
					{Col: &Identifier{Name: "d"}},
				},
			},
			Targets: []*Target{
				{Col: &Identifier{Name: "e"}},
			},
		},
		},
	},

	// embed metric query
	{
		in: `O::(
					M::a:(b,c)
				):(d,e)`,
		expected: Stmts{&DFQuery{
			Namespace: "O",
			Subquery: &DFQuery{
				Namespace: "M",
				Names:     []string{"a"},
				Targets: []*Target{
					{Col: &Identifier{Name: "b"}},
					{Col: &Identifier{Name: "c"}},
				},
			},
			Targets: []*Target{
				{Col: &Identifier{Name: "d"}},
				{Col: &Identifier{Name: "e"}},
			},
		},
		},
	},

	// from list with trailing `,`
	{
		in: `metrics:: NAME, mem, :(host)`,

		expected: Stmts{&DFQuery{
			Namespace: "metrics",
			Names:     []string{"NAME", "mem"},
			Targets: []*Target{
				{
					Col: &Identifier{Name: "host"},
				},
			},
		},
		},
	},

	// special characters
	{
		in: "metrics:: `NAME👍`, mem, re(`*¶¶¶¶¶¶¶¶`):(host) ",
		expected: Stmts{&DFQuery{
			Namespace: "metrics",

			Names:      []string{"NAME👍", "mem"},
			RegexNames: []*Regex{{Regex: "*¶¶¶¶¶¶¶¶"}},

			Targets: []*Target{
				{
					Col: &Identifier{Name: "host"},
				},
			},
		},
		},
	},

	{
		in: "metrics:: `中文`, mem, re(`*¶¶¶¶¶¶¶¶`):(host) ",
		expected: Stmts{&DFQuery{
			Namespace: "metrics",

			Names:      []string{"中文", "mem"},
			RegexNames: []*Regex{{Regex: "*¶¶¶¶¶¶¶¶"}},

			Targets: []*Target{
				{
					Col: &Identifier{Name: "host"},
				},
			},
		},
		},
	},

	{
		in: "metrics:: `中文`, mem, re(`*¶¶¶¶¶¶¶¶`):(host, `'xxx'`) ",
		expected: Stmts{&DFQuery{
			Namespace: "metrics",

			Names:      []string{"中文", "mem"},
			RegexNames: []*Regex{{Regex: "*¶¶¶¶¶¶¶¶"}},

			Targets: []*Target{
				{
					Col: &Identifier{Name: "host"},
				},
				{
					Col: &Identifier{Name: `'xxx'`},
				},
			},
		},
		},
	},

	// with keyword as metric-name
	{
		in: "metrics:: `and`, mem, re('*disk'):(host) ",
		expected: Stmts{&DFQuery{
			Namespace: "metrics",

			Names:      []string{"and", "mem"},
			RegexNames: []*Regex{{Regex: "*disk"}},

			Targets: []*Target{
				{
					Col: &Identifier{Name: "host"},
				},
			},
		},
		},
	},
	{
		in: `metrics::
				cpu, mem, re("*disk"):(host)`,
		expected: Stmts{&DFQuery{
			Namespace: "metrics",

			Names:      []string{"cpu", "mem"},
			RegexNames: []*Regex{{Regex: "*disk"}},

			Targets: []*Target{
				{
					Col: &Identifier{Name: "host"},
				},
			},
		},
		},
	},

	// multi-metric, no namespace
	{
		in: `cpu,mem,re("*disk"):()`,

		expected: Stmts{&DFQuery{
			Names:      []string{"cpu", "mem"},
			RegexNames: []*Regex{{Regex: "*disk"}},
		},
		},
	},

	// alias
	{
		in: "mem:(active AS `激活`)",
		expected: Stmts{
			&DFQuery{
				Names: []string{"mem"},
				Targets: []*Target{
					{Col: &Identifier{Name: "active"}, Alias: "激活"},
				},
			},
		},
	},

	{
		in: ` cpu:(avg(b) AS avg_b) ORDER BY avg_b`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{
					Col: &FuncExpr{
						Name: `avg`,
						Param: []Node{
							&Identifier{Name: "b"},
						},
					},
					Alias: "avg_b",
				},
			},
			OrderBy: &OrderBy{
				List: NodeList{
					&OrderByElem{Column: "avg_b", Opt: OrderAsc},
				},
			},
		}},
	},

	// limit clause
	{
		in:   `cpu:() {} [] LIMIT -3`,
		fail: true,
	},

	{
		in:   `cpu:() {} [] SLIMIT 3.6`,
		fail: true,
	},
	{
		in: `cpu:() {} [] LIMIT 3`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Limit: &Limit{Limit: 3},
		},
		},
	},
	{
		in: `cpu:() SLIMIT 10`,
		expected: Stmts{&DFQuery{
			Names:  []string{"cpu"},
			SLimit: &SLimit{SLimit: 10},
		}}},

	// offset
	{
		in:   `cpu:() OFFSET -1.3`,
		fail: true,
	},
	{
		in: `cpu:() SOFFSET 10`,
		expected: Stmts{&DFQuery{
			Names:   []string{"cpu"},
			SOffset: &SOffset{SOffset: 10},
		},
		}},
	{
		in:   `cpu:() OFFSET 10 LIMIT 4`,
		fail: true,
	},

	// trailing `,`
	{
		in: `cpu:(a,)`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &Identifier{Name: "a"}},
			}}}},
	{
		in: `cpu:(avg(a,))`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{
					Col: &FuncExpr{
						Name:  `avg`,
						Param: []Node{&Identifier{Name: "a"}},
					}}}}}},

	{
		in: `cpu {a!="xyz",}`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			WhereCondition: []Node{
				&BinaryExpr{Op: NEQ, LHS: &Identifier{Name: "a"}, RHS: &StringLiteral{Val: "xyz"}, ReturnBool: true},
			}},
		},
	},

	{
		in: `cpu:(f1, f2, avg(max(b))) ORDER BY f1, f2, f3 DESC`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &Identifier{Name: "f1"}},
				{Col: &Identifier{Name: "f2"}},
				{
					Col: &FuncExpr{
						Name: `avg`,
						Param: []Node{
							&FuncExpr{Name: `max`, Param: []Node{&Identifier{Name: "b"}}},
						},
					},
				},
			},
			OrderBy: &OrderBy{
				List: NodeList{
					&OrderByElem{Column: "f1", Opt: OrderAsc},
					&OrderByElem{Column: "f2", Opt: OrderAsc},
					&OrderByElem{Column: "f3", Opt: OrderDesc},
				},
			},
		}},
	},

	// group by
	{
		in: `cpu:() BY t1`,
		expected: Stmts{&DFQuery{
			Names:   []string{"cpu"},
			GroupBy: &GroupBy{List: NodeList{&Identifier{Name: "t1"}}},
		}}},
	{
		in: `cpu:() BY t1`,
		expected: Stmts{&DFQuery{
			Names:   []string{"cpu"},
			GroupBy: &GroupBy{List: NodeList{&Identifier{Name: "t1"}}},
		}}},
	{
		in: `cpu:() [] BY t1`,
		expected: Stmts{&DFQuery{
			Names:   []string{"cpu"},
			GroupBy: &GroupBy{List: NodeList{&Identifier{Name: "t1"}}},
		}}},

	{
		in:   `cpu:() [] group by`,
		fail: true,
	},

	// empty target/filter/time-options
	{
		in: `cpu:() {} []`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
		}}},

	// metric cases with fileters
	{
		in: `M::cpu:(f1) {a="123", b != 3, c!= re("some_regex")} [3m:2m:10s]`,
		expected: Stmts{&DFQuery{
			Names:     []string{"cpu"},
			Namespace: "M",
			Targets:   []*Target{{Col: &Identifier{Name: "f1"}}},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{IsDuration: true, Duration: time.Minute * 3},
				End:        &TimeExpr{IsDuration: true, Duration: time.Minute * 2},
				Resolution: &TimeResolution{Duration: time.Second * 10}},

			WhereCondition: []Node{
				&BinaryExpr{Op: EQ, LHS: &Identifier{Name: "a"}, RHS: &StringLiteral{Val: "123"}, ReturnBool: true},
				&BinaryExpr{Op: NEQ, LHS: &Identifier{Name: "b"}, RHS: &NumberLiteral{IsInt: true, Int: 3}, ReturnBool: true},
				&BinaryExpr{Op: NEQ, LHS: &Identifier{Name: "c"}, RHS: &Regex{Regex: "some_regex"}, ReturnBool: true},
			}},
		}},

	// nested functions
	{
		in: `cpu:(a, avg(max(b)))`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &Identifier{Name: "a"}},
				{
					Col: &FuncExpr{
						Name: `AVG`,
						Param: []Node{
							&FuncExpr{Name: `max`, Param: []Node{&Identifier{Name: "b"}}},
						}}}}}}},

	// function without args
	{
		in: `cpu:(a, avg())`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &Identifier{Name: "a"}},
				{Col: &FuncExpr{Name: `avg`}},
			},
		},
		}},

	// function position arg
	{
		in: `cpu:(MAX(arg1=[123, 456], arg2="xyz"), f2)`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &FuncExpr{Name: `max`, Param: []Node{
					&FuncArg{ArgName: "arg1", ArgVal: FuncArgList{
						&NumberLiteral{IsInt: true, Int: 123},
						&NumberLiteral{IsInt: true, Int: 456},
					}},
					&FuncArg{ArgName: "arg2", ArgVal: &StringLiteral{Val: "xyz"}},
				}}},
				{Col: &Identifier{Name: "f2"}},
			},
		},
		},
	},
	{
		in: `cpu:(MAX(arg1=123, arg2=456, arg3="raw-str"))`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &FuncExpr{Name: `max`, Param: []Node{
					&FuncArg{ArgName: "arg1", ArgVal: &NumberLiteral{IsInt: true, Int: 123}},
					&FuncArg{ArgName: "arg2", ArgVal: &NumberLiteral{IsInt: true, Int: 456}},
					&FuncArg{ArgName: "arg3", ArgVal: &StringLiteral{Val: "raw-str"}},
				}}},
			},
		}},
	},

	{
		in: `cpu { b!=re("abc")}`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			WhereCondition: []Node{
				&BinaryExpr{ReturnBool: true, Op: NEQ, LHS: &Identifier{Name: "b"}, RHS: &Regex{Regex: "abc"}},
			},
		},
		}},

	// func call with value fill
	{
		// various fill
		in: `cpu:(fill(max(a), linear), fill(min(b), nil), fill(avg(c), previous), fill(avg(d), 123.4), fill(avg(e), 123), fill(avg(f), "str"))`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &FuncExpr{Name: `max`, Param: []Node{&Identifier{Name: "a"}}}, Fill: &Fill{FillType: FillLinear}},
				{Col: &FuncExpr{Name: `min`, Param: []Node{&Identifier{Name: "b"}}}, Fill: &Fill{FillType: FillNil}},
				{Col: &FuncExpr{Name: `avg`, Param: []Node{&Identifier{Name: "c"}}}, Fill: &Fill{FillType: FillPrevious}},
				{Col: &FuncExpr{Name: `avg`, Param: []Node{&Identifier{Name: "d"}}}, Fill: &Fill{FillType: FillFloat, Float: 123.4}},
				{Col: &FuncExpr{Name: `avg`, Param: []Node{&Identifier{Name: "e"}}}, Fill: &Fill{FillType: FillInt, Int: 123}},
				{Col: &FuncExpr{Name: `avg`, Param: []Node{&Identifier{Name: "f"}}}, Fill: &Fill{FillType: FillStr, Str: "str"}},
			},
		}},
	},

	{
		in: `cpu:(fill(a, 123))`, // fill must work with function
		expected: Stmts{
			&DFQuery{
				Names: []string{"cpu"},
				Targets: []*Target{
					{Col: &Identifier{Name: "a"}, Fill: &Fill{FillType: FillInt, Int: 123}},
				},
			},
		},
	},
	{
		in:   `{cpu:(fill(min(a), unknown_fill))}`, // fill must work with function
		fail: true,
	},

	// aggr func call
	{
		in: `cpu:(avg(a, b) AS avg_ab) [10m:5m:1m]`,
		expected: Stmts{&DFQuery{
			Names:   []string{"cpu"},
			Targets: []*Target{{Col: &FuncExpr{Name: `avg`, Param: []Node{&Identifier{Name: "a"}, &Identifier{Name: "b"}}}, Alias: "avg_ab"}},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
				End:        &TimeExpr{IsDuration: true, Duration: time.Minute * 5},
				Resolution: &TimeResolution{Duration: time.Minute * 1}},
		},
		}},

	{
		in: `cpu:(unknown_func(a, b) AS xx_ab) [10m:5m:1m]`,
		expected: Stmts{&DFQuery{
			Names:   []string{"cpu"},
			Targets: []*Target{{Col: &FuncExpr{Name: `unknown_func`, Param: []Node{&Identifier{Name: "a"}, &Identifier{Name: "b"}}}, Alias: "xx_ab"}},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
				End:        &TimeExpr{IsDuration: true, Duration: time.Minute * 5},
				Resolution: &TimeResolution{Duration: time.Minute * 1}},
		},
		}},

	// () expr in filters
	{
		in: `cpu {(a>123.45 && b!=re("*abc")) || (z!="abc"), c=re("xyz*")}`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			WhereCondition: []Node{
				&BinaryExpr{
					ReturnBool: true,
					Op:         OR,
					LHS: &ParenExpr{
						Param: &BinaryExpr{
							ReturnBool: true,
							Op:         AND,
							LHS:        &BinaryExpr{ReturnBool: true, Op: GT, LHS: &Identifier{Name: "a"}, RHS: &NumberLiteral{Float: 123.45}},
							RHS:        &BinaryExpr{ReturnBool: true, Op: NEQ, LHS: &Identifier{Name: "b"}, RHS: &Regex{Regex: "*abc"}},
						},
					},
					RHS: &ParenExpr{&BinaryExpr{ReturnBool: true, Op: NEQ, LHS: &Identifier{Name: "z"}, RHS: &StringLiteral{Val: "abc"}}},
				},
				&BinaryExpr{ReturnBool: true, Op: EQ, LHS: &Identifier{Name: "c"}, RHS: &Regex{Regex: "xyz*"}},
			},
		},
		}},

	// &&, || in filters
	{
		in: `cpu {a>123 || b!=re("*abc"), c=re("xyz*")}`,
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			WhereCondition: []Node{
				&BinaryExpr{
					Op:         OR,
					LHS:        &BinaryExpr{ReturnBool: true, Op: GT, LHS: &Identifier{Name: "a"}, RHS: &NumberLiteral{Int: 123, IsInt: true}},
					RHS:        &BinaryExpr{ReturnBool: true, Op: NEQ, LHS: &Identifier{Name: "b"}, RHS: &Regex{Regex: "*abc"}},
					ReturnBool: true,
				},

				&BinaryExpr{
					Op:         EQ,
					LHS:        &Identifier{Name: "c"},
					RHS:        &Regex{Regex: "xyz*"},
					ReturnBool: true,
				},
			},
		},
		}},
	{
		in: `cpu {a && b} `,
		expected: Stmts{&DFQuery{
			Names:          []string{"cpu"},
			WhereCondition: []Node{&BinaryExpr{Op: AND, LHS: &Identifier{Name: "a"}, RHS: &Identifier{Name: "b"}, ReturnBool: true}},
		},
		}},

	// metric cases with time range
	{
		in:   `c [:1603881643]`, //
		fail: true,
	},
	{
		in: `c [::1h]`, //
		expected: Stmts{&DFQuery{
			Names: []string{"c"},
			TimeRange: &TimeRange{
				Resolution: &TimeResolution{Duration: time.Hour},
			},
		}},
	},
	{
		in: `c [::1h,13m]`, //
		expected: Stmts{&DFQuery{
			Names: []string{"c"},
			TimeRange: &TimeRange{
				Resolution:       &TimeResolution{Duration: time.Hour},
				ResolutionOffset: time.Minute * 13},
		}},
	},
	{
		in: `c [1603881643:1h:13m]`, //
		expected: Stmts{&DFQuery{
			Names: []string{"c"},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{Time: time.Unix(1603881643, 0)},
				End:        &TimeExpr{Duration: time.Hour, IsDuration: true},
				Resolution: &TimeResolution{Duration: time.Minute * 13}},
		},
		}},
	{
		in: `c [2020-10-28 10:46:02:1h:13m]`, //
		expected: Stmts{&DFQuery{
			Names: []string{"c"},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{Time: time.Unix(1603881962, 0).UTC()},
				End:        &TimeExpr{Duration: time.Hour, IsDuration: true},
				Resolution: &TimeResolution{Duration: time.Minute * 13}}},
		}},
	{
		in: `c [2020-10-28:1h:13m]`, //
		expected: Stmts{&DFQuery{
			Names: []string{"c"},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{Time: time.Unix(1603843200, 0).UTC()},
				End:        &TimeExpr{Duration: time.Hour, IsDuration: true},
				Resolution: &TimeResolution{Duration: time.Minute * 13}}},
		}},
	{
		in: `c [2020-1-2:1h:13m]`, // ok
		expected: Stmts{&DFQuery{
			Names: []string{"c"},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{Time: time.Unix(1577923200, 0).UTC()},
				End:        &TimeExpr{Duration: time.Hour, IsDuration: true},
				Resolution: &TimeResolution{Duration: time.Minute * 13}}},
		}},
	{
		in: "cpu [10m::1m,31s]",
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:            &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			Resolution:       &TimeResolution{Duration: time.Minute},
			ResolutionOffset: 31 * time.Second}},
		}},
	{
		in: "cpu [10m::1m,-31s]", // minus offset
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:            &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			Resolution:       &TimeResolution{Duration: time.Minute},
			ResolutionOffset: -31 * time.Second}},
		}},
	{
		in: "cpu [10m::1m,+31s]", // minus offset
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:            &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			Resolution:       &TimeResolution{Duration: time.Minute},
			ResolutionOffset: 31 * time.Second}},
		}},
	{
		in: "cpu [10m:5m:1m,31s]",
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:            &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			End:              &TimeExpr{IsDuration: true, Duration: time.Minute * 5},
			Resolution:       &TimeResolution{Duration: time.Minute},
			ResolutionOffset: 31 * time.Second}},
		}},
	{
		in: "cpu [10m::1m]",
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:      &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			Resolution: &TimeResolution{Duration: time.Minute}}},
		}},
	{
		in: "cpu [10m::1m]",
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:      &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			Resolution: &TimeResolution{Duration: time.Minute}}},
		}},
	{
		in: "cpu [10m:5m:5s]",
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:      &TimeExpr{IsDuration: true, Duration: time.Minute * 10},
			End:        &TimeExpr{IsDuration: true, Duration: time.Minute * 5},
			Resolution: &TimeResolution{Duration: time.Second * 5}}},
		}},
	{in: "yyy [10ym]", fail: true},
	{
		in: "cpu [10us:10ns:5ns]",
		expected: Stmts{&DFQuery{Names: []string{"cpu"}, TimeRange: &TimeRange{
			Start:      &TimeExpr{IsDuration: true, Duration: time.Microsecond * 10},
			End:        &TimeExpr{IsDuration: true, Duration: time.Nanosecond * 10},
			Resolution: &TimeResolution{Duration: time.Nanosecond * 5}}},
		}},
	{
		in: "cpu:(user_idle) [1s:10s:5s]",
		expected: Stmts{&DFQuery{Names: []string{"cpu"},
			Targets: []*Target{{Col: &Identifier{Name: "user_idle"}}},
			TimeRange: &TimeRange{
				Start:      &TimeExpr{IsDuration: true, Duration: time.Second * 1},
				End:        &TimeExpr{IsDuration: true, Duration: time.Second * 10},
				Resolution: &TimeResolution{Duration: time.Second * 5}}},
		}},

	// simple cases
	{in: "cpu", expected: Stmts{&DFQuery{Names: []string{"cpu"}}}},
	{in: `re("cpu")`, expected: Stmts{&DFQuery{RegexNames: []*Regex{{Regex: "cpu"}}}}},

	{
		in: "M::re(`cpu|mem|disk{a-z_1-9}`)",
		expected: Stmts{&DFQuery{
			Namespace:  "M",
			RegexNames: []*Regex{{Regex: "cpu|mem|disk{a-z_1-9}"}},
		},
		}},

	{
		in: "M::re(`cpu|mem|disk{a-z_1-9}`):(f1, f2)",
		expected: Stmts{&DFQuery{
			Namespace: "M", RegexNames: []*Regex{{Regex: "cpu|mem|disk{a-z_1-9}"}}, Targets: []*Target{{Col: &Identifier{Name: "f1"}}, {Col: &Identifier{Name: "f2"}}},
		},
		}},

	/* prefix/suffix space on :: and : */
	{in: "metric::cpu: (user_idle)", expected: Stmts{&DFQuery{Namespace: "metric", Names: []string{"cpu"}, Targets: []*Target{{Col: &Identifier{Name: "user_idle"}}}}}},
	{in: "metric::cpu : (user_idle)", expected: Stmts{&DFQuery{Namespace: "metric", Names: []string{"cpu"}, Targets: []*Target{{Col: &Identifier{Name: "user_idle"}}}}}},
	{in: "metric :: cpu : (user_idle)", expected: Stmts{&DFQuery{Namespace: "metric", Names: []string{"cpu"}, Targets: []*Target{{Col: &Identifier{Name: "user_idle"}}}}}},
	{in: "metric :: re('cpu') : (user_idle)", expected: Stmts{&DFQuery{Namespace: "metric", RegexNames: []*Regex{{Regex: "cpu"}}, Targets: []*Target{{Col: &Identifier{Name: "user_idle"}}}}}},

	{in: "metric::cpu", expected: Stmts{&DFQuery{Namespace: "metric", Names: []string{"cpu"}}}},
	{in: "metric::cpu:(user_idle)", expected: Stmts{&DFQuery{Namespace: "metric", Names: []string{"cpu"}, Targets: []*Target{{Col: &Identifier{Name: "user_idle"}}}}}},
	{in: "cpu:(user_idle)", expected: Stmts{&DFQuery{Namespace: "", Names: []string{"cpu"}, Targets: []*Target{{Col: &Identifier{Name: "user_idle"}}}}}},
	{in: "cpu:(user_idle as UI)", expected: Stmts{&DFQuery{Namespace: "", Names: []string{"cpu"}, Targets: []*Target{{Alias: "UI", Col: &Identifier{Name: "user_idle"}}}}}},

	{
		in: "cpu:(user_idle/sys_idle)",
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Col: &BinaryExpr{Op: DIV, LHS: &Identifier{Name: "user_idle"}, RHS: &Identifier{Name: "sys_idle"}, ReturnBool: false}},
			}},
		}},

	{
		in: "cpu:(user_idle+1 AS UI, sys_idle)",
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			Targets: []*Target{
				{Alias: "UI", Col: &BinaryExpr{Op: ADD, LHS: &Identifier{Name: "user_idle"}, RHS: &NumberLiteral{IsInt: true, Int: 1}, ReturnBool: false}},
				{Col: &Identifier{Name: "sys_idle"}},
			},
		}},
	},
	{
		in: "cpu {x>1.23E3}", // scientific notation
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			WhereCondition: []Node{
				&BinaryExpr{Op: GT, LHS: &Identifier{Name: "x"}, RHS: &NumberLiteral{Float: 1230.00}, ReturnBool: true},
			},
		}},
	},
	{
		in: "cpu {x<1.23E3}", // scientific notation
		expected: Stmts{&DFQuery{
			Names: []string{"cpu"},
			WhereCondition: []Node{
				&BinaryExpr{Op: LT, LHS: &Identifier{Name: "x"}, RHS: &NumberLiteral{Float: 1230.00}, ReturnBool: true},
			},
		}},
	},
}

func TestParseFuncExpr(t *testing.T) {
	var funcExprCases = []*parseCase{
		{
			in: "a.b__.c()",
			expected: &FuncExpr{
				Name: `a.b__.c`,
			},
		},
		{
			in: "`中文函数`()",
			expected: &FuncExpr{
				Name: `中文函数`,
			},
		},
		{
			in: "count(*)",
			expected: &FuncExpr{
				Name:  "count",
				Param: []Node{&Star{}},
			},
		},
		{
			in: "count(avg(2,5))",
			expected: &FuncExpr{
				Name: "count",
				Param: []Node{
					&FuncExpr{
						Name: "avg",
						Param: []Node{
							&NumberLiteral{IsInt: true, Int: 2},
							&NumberLiteral{IsInt: true, Int: 5},
						},
					},
				},
			},
		},
		{
			in: "count(`中国` = 1000)",
			expected: &FuncExpr{
				Name: "count",
				Param: []Node{
					&FuncArg{
						ArgName: "中国",
						ArgVal:  &NumberLiteral{IsInt: true, Int: 1000},
					},
				},
			},
		},
	}

	for idx, tc := range funcExprCases {
		_ = idx

		fexpr, err := ParseFuncExpr(tc.in)

		var x, y string

		if err != nil {
			t.Logf("parse error: %s", err.Error())
		} else {
			if tc.expected != nil {
				t.Logf("tc.expected: %+#v", tc.expected)
				expected, ok := tc.expected.(*FuncExpr)
				if ok {
					x = expected.String()
				}
				y = fexpr.String()
			}
		}

		if !tc.fail {
			testutil.Ok(t, err)
			testutil.Equals(t, x, y)
			t.Logf("[%d] ok %s -> %s", idx, tc.in, y)
		} else {
			t.Logf("[%d] %s -> expect fail: %v", idx, tc.in, err)
			testutil.NotOk(t, err, "")
		}
	}
}

func TestParseBinaryExpr(t *testing.T) {
	var binaryExprCases = []*parseCase{

		// division by zero
		{
			in:   "10/0",
			fail: true,
		},
		// modulo by zero
		{
			in:   "10%0.0",
			fail: true,
		},
		// single binary-expr
		{
			in: "a||b&&c",
			expected: &BinaryExpr{
				LHS: &Identifier{Name: "a"},
				Op:  OR,
				RHS: &BinaryExpr{
					Op:  AND,
					LHS: &Identifier{Name: "b"},
					RHS: &Identifier{Name: "c"},
				},
			},
		},

		{
			in: "a&&(b||c)",
			expected: &BinaryExpr{
				LHS: &Identifier{Name: "a"},
				Op:  AND,
				RHS: &ParenExpr{Param: &BinaryExpr{
					Op:  OR,
					LHS: &Identifier{Name: "b"},
					RHS: &Identifier{Name: "c"},
				},
				},
			},
		},

		{
			in: `a<5`,
			expected: &BinaryExpr{
				Op:  LT,
				RHS: &NumberLiteral{IsInt: true, Int: 5},
				LHS: &Identifier{Name: "a"},
			},
		},

		// IN [...]
		{
			in: "a in [1,2,3]",
			expected: &BinaryExpr{
				LHS: &Identifier{Name: "a"},
				Op:  IN,
				RHS: NodeList{
					&NumberLiteral{IsInt: true, Int: 1},
					&NumberLiteral{IsInt: true, Int: 2},
					&NumberLiteral{IsInt: true, Int: 3},
				},
			},
		},
	}

	for idx, tc := range binaryExprCases {
		_ = idx

		bexpr, err := ParseBinaryExpr(tc.in)

		var x, y string

		if err != nil {
			t.Logf("parse error: %s", err.Error())
		} else {
			if tc.expected != nil {
				t.Logf("tc.expected: %+#v", tc.expected)
				expected, ok := tc.expected.(*BinaryExpr)
				if ok {
					x = expected.String()
				}
				y = bexpr.String()
			}
		}

		if !tc.fail {
			testutil.Ok(t, err)
			testutil.Equals(t, x, y)
			t.Logf("[%d] ok %s -> %s", idx, tc.in, y)
		} else {
			t.Logf("[%d] %s -> expect fail: %v", idx, tc.in, err)
			testutil.NotOk(t, err, "")
		}
	}
}

func TestParseQuery(t *testing.T) {
	runCases(t, queryCases)
}

func runCases(t *testing.T, cases []*parseCase) {
	for idx := len(cases) - 1; idx >= 0; idx-- {
		tc := cases[idx]

		nodes, err := ParseDQL(tc.in)
		if err != nil {
			t.Log(err)
		}

		var x, y string

		_ = x
		_ = y

		// for debug
		switch v := nodes.(type) {
		case Stmts:
			if err != nil {
				t.Logf("parse error: %s", err.Error())
			} else {
				if tc.expected != nil {
					// t.Logf("tc.expected: %+#v", tc.expected)
					expected, ok := tc.expected.(Stmts)
					if ok {
						x = expected.String()
					}
					y = v.String()
				}
			}

		default:
			if !tc.fail {
				t.Errorf("[%d] in: %s -> unknown parse result: %v, err: %v", idx, tc.in, nodes, err)
			}
		}

		if !tc.fail {
			testutil.Ok(t, err)
			testutil.Equals(t, x, y)
			t.Logf("[%d] ok %s -> %s", idx, tc.in, y)
		} else {
			t.Logf("[%d] %s -> expect fail: %v", idx, tc.in, err)
			testutil.NotOk(t, err, "")
		}
	}
}

func BenchmarkParser(b *testing.B) {
	logger.SetStdoutRootLogger(logger.INFO, logger.OPT_DEFAULT)
	log = logger.DefaultSLogger("parser")

	b.ReportAllocs()

	nbytes := 0
	for _, c := range queryCases {
		nbytes += len(c.in)
	}
	b.SetBytes(int64(nbytes))

	nparse := 0
	for x := 0; x < b.N; x++ {
		for _, c := range queryCases {
			_, _ = ParseDQL(c.in)
			nparse++
		}
	}
}

func TestDQLToJson(t *testing.T) {
	for idx := len(queryCases) - 1; idx >= 0; idx-- {
		tc := queryCases[idx]

		nodes, err := ParseDQL(tc.in)
		if err != nil {
			continue
		}

		// for debug
		switch v := nodes.(type) {
		case Stmts:
			for _, stmt := range v {
				switch vv := stmt.(type) {
				case *DFQuery:
					b, err := vv.JSON()
					if err != nil {
						t.Fatalf("[%d] %s -> expect fail: %v", idx, tc.in, err)
					}

					t.Logf("[%d] ok %s -> %s", idx, tc.in, b)
				}
			}

		default:
			t.Fatal("panic")
		}
	}
}

func TestOuterFuncParse(t *testing.T) {
	var testCases = []map[string]string{

		{
			// 1, 测试逻辑not
			"index": "1",

			// "input": "difference(2, dql=`L::logging_b:(filesize) limit 3`).moving_average(2)",
			// "input": "difference(dql=`L::logging_b:(filesize) limit 3`).moving_average(size=2)",
			// "input": "difference(difference(dql=`L::logging_b:(filesize) limit 10`))",
			// "input": "moving_average(dql=`L::logging_b:(filesize) limit 10`, size=2)",
			"input":    "difference(`L::logging_b:(filesize) limit 3`)",
			"expected": ``,
		},
	}

	for _, item := range testCases {
		asts, perr := ParseDQL(item["input"])
		if perr != nil {
			t.Errorf(
				"parse error: the input is:\n\n %s \n\n err is:\n\n %s \n",
				item["input"],
				perr,
			)
		}
		log.Info(asts)
	}
}
