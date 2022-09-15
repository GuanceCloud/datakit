// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parser

import (
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/testutil"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
)

func TestExprSeparation(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		expected ast.Stmts
		err      string
		fail     bool
	}{
		{
			name: "test_if_nl",
			in: `

			# 0 
			[
				
			
			1, 3,
			"3",
			true,
			1.]
		
			[
				1, 3,
				"3",
				true,
				1.
			]

			# 只支持基础数据类型
			# [1, 3, "3", 1. ,
			# 	[1, 3, "3", 1. 
			# 		]
			# ]

			a = 1; func1()
			func2(1, 3, 
				"3", true, 1. 
			
			) 
			if true {
				;
			}

			#2
			if true {

			} elif false {
			} elif x == y {	 a =	
				a + 1	}  
				a = b; a=c
					a =d

			#3
			if true {
			} elif false {
			} elif x == y {	 
				a =	
				a + 
				1
			} else {

			}
			`,
			expected: ast.Stmts{
				ast.WrapListInitExpr(&ast.ListInitExpr{
					List: []*ast.Node{
						ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
						ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 3, IsInt: true}),
						ast.WrapStringLiteral(&ast.StringLiteral{Val: "3"}),
						ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
						ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 1.}),
					},
				}),
				ast.WrapListInitExpr(
					&ast.ListInitExpr{
						List: []*ast.Node{
							ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
							ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 3, IsInt: true}),
							ast.WrapStringLiteral(&ast.StringLiteral{Val: "3"}),
							ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
							ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 1.}),
						},
					}),
				ast.WrapAssignmentExpr(&ast.AssignmentExpr{
					LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
					RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
				}),
				ast.WrapCallExpr(&ast.CallExpr{Name: "func1"}),
				ast.WrapCallExpr(
					&ast.CallExpr{
						Name: "func2",
						Param: ast.FuncArgList{
							ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
							ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 3, IsInt: true}),
							ast.WrapStringLiteral(&ast.StringLiteral{Val: "3"}),
							ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
							ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 1.}),
						},
					},
				),
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: []*ast.IfStmtElem{
						{Condition: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true})},
					},
				}),

				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: []*ast.IfStmtElem{
						{Condition: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true})},
						{Condition: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false})},
						{
							Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
								Op:  AstOp(EQEQ),
								RHS: ast.WrapIdentifier(&ast.Identifier{Name: "y"}),
							}),
							Stmts: []*ast.Node{
								ast.WrapAssignmentExpr(&ast.AssignmentExpr{
									LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
									RHS: ast.WrapArithmeticExpr(&ast.ArithmeticExpr{
										LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
										RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
										Op:  AstOp(ADD),
									}),
								}),
							},
						},
					},
				}),

				ast.WrapAssignmentExpr(&ast.AssignmentExpr{
					LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
					RHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
				}),
				ast.WrapAssignmentExpr(&ast.AssignmentExpr{
					LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
					RHS: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
				}),
				ast.WrapAssignmentExpr(&ast.AssignmentExpr{
					LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
					RHS: ast.WrapIdentifier(&ast.Identifier{Name: "d"}),
				}),

				ast.WrapIfelseStmt(
					&ast.IfelseStmt{
						IfList: []*ast.IfStmtElem{
							{Condition: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true})},
							{Condition: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false})},
							{
								Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
									LHS: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
									Op:  AstOp(EQEQ),
									RHS: ast.WrapIdentifier(&ast.Identifier{Name: "y"}),
								}),
								Stmts: []*ast.Node{
									ast.WrapAssignmentExpr(&ast.AssignmentExpr{
										LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
										RHS: ast.WrapArithmeticExpr(&ast.ArithmeticExpr{
											LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
											RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
											Op:  AstOp(ADD),
										}),
									}),
								},
							},
						},
						Else: nil,
					},
				),
			},
		},
		{
			name: "+||",
			in:   "a * 1 + b || x + 1 && 1 / 2 == 1",
			expected: ast.Stmts{
				ast.WrapConditionExpr(
					&ast.ConditionalExpr{
						Op: AstOp(EQEQ),
						LHS: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op: AstOp(OR),
							LHS: ast.WrapArithmeticExpr(&ast.ArithmeticExpr{
								Op: AstOp(ADD),
								LHS: ast.WrapArithmeticExpr(&ast.ArithmeticExpr{
									Op:  AstOp(MUL),
									LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
									RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
								}),
								RHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
							}),
							RHS: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op: AstOp(AND),
								LHS: ast.WrapArithmeticExpr(&ast.ArithmeticExpr{
									Op:  AstOp(ADD),
									LHS: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
									RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
								}),
								RHS: ast.WrapArithmeticExpr(&ast.ArithmeticExpr{
									Op:  AstOp(DIV),
									LHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1, IsInt: true}),
									RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 2, IsInt: true}),
								}),
							}),
						}),
						RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
					},
				),
			},
		},
	}

	// for idx := len(cases) - 1; idx >= 0; idx-- {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Stmts, err := ParsePipeline(tc.in)

			if !tc.fail {
				testutil.Ok(t, err)
			} else {
				t.Logf("expected error: %s", err)
				testutil.NotOk(t, err, "")
				return
			}

			if !tc.fail {
				var x, y string
				x = tc.expected.String()
				y = Stmts.String()
				testutil.Ok(t, err)
				testutil.Equals(t, x, y)
				t.Logf("ok %s -> %s", tc.in, y)
			} else {
				t.Logf("%s -> expect fail: %v", tc.in, err)
				testutil.NotOk(t, err, "")
			}
		})
	}
}

func TestParserFor(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		expected ast.Stmts
		err      string
		fail     bool
	}{
		{
			name: "for_x_in_y",
			in: `for x in y {
				b = 1
			}`,
			expected: ast.Stmts{
				ast.WrapForInStmt(&ast.ForInStmt{
					Varb: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
					Iter: ast.WrapIdentifier(&ast.Identifier{Name: "y"}),
					Body: []*ast.Node{
						ast.WrapAssignmentExpr(&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
							RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1}),
						}),
					},
				}),
			},
		},
		{
			name: "for_x_in_y",
			in: `for x in [1,2.,"2"]  {
				if x != "2" {
					continue				} else {
					break
				}
				a = a + 1
			}`,
			expected: ast.Stmts{
				ast.WrapForInStmt(&ast.ForInStmt{
					Varb: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
					Iter: ast.WrapListInitExpr(&ast.ListInitExpr{
						List: []*ast.Node{
							ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
							ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 1}),
							ast.WrapStringLiteral(&ast.StringLiteral{Val: "2"}),
						},
					}),
					Body: []*ast.Node{
						ast.WrapIfelseStmt(&ast.IfelseStmt{
							IfList: []*ast.IfStmtElem{
								{
									Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
										LHS: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
										Op:  AstOp(NEQ),
										RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "2"}),
									}),
									Stmts: ast.Stmts{
										ast.WrapContinueStmt(&ast.ContinueStmt{}),
									},
								},
							},
							Else: ast.Stmts{
								ast.WrapBreakStmt(&ast.BreakStmt{}),
							},
						}),
						ast.WrapAssignmentExpr(&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
							RHS: ast.WrapArithmeticExpr(&ast.ArithmeticExpr{
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
								Op:  AstOp(ADD),
								RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
							}),
						}),
					},
				}),
			},
		},
		{
			name: "for cond ",
			in: `for ; a == func() ; {
				b = 1
			}`,
			expected: ast.Stmts{
				ast.WrapForStmt(&ast.ForStmt{
					Cond: ast.WrapConditionExpr(&ast.ConditionalExpr{
						Op:  AstOp(EQEQ),
						LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
						RHS: ast.WrapCallExpr(&ast.CallExpr{
							Name: "func",
						}),
					}),
					Body: []*ast.Node{
						ast.WrapAssignmentExpr(&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
							RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1}),
						}),
					},
				}),
			},
		},
		{
			name: "for_; cond ; loop ",
			in: `for ; a == func() ; x = 2. {
				b=2
			}`,
			expected: ast.Stmts{ast.WrapForStmt(&ast.ForStmt{
				Cond: ast.WrapConditionExpr(&ast.ConditionalExpr{
					Op:  AstOp(EQEQ),
					LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
					RHS: ast.WrapCallExpr(&ast.CallExpr{
						Name: "func",
					}),
				}),
				Loop: ast.WrapAssignmentExpr(&ast.AssignmentExpr{
					LHS: ast.WrapIdentifier(&ast.Identifier{
						Name: "x",
					}),
					RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{
						Float: 2.,
					}),
				}),
				Body: []*ast.Node{
					ast.WrapAssignmentExpr(&ast.AssignmentExpr{
						LHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
						RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{Int: 1}),
					}),
				},
			})},
		},
	}

	// for idx := len(cases) - 1; idx >= 0; idx-- {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Stmts, err := ParsePipeline(tc.in)

			if !tc.fail {
				testutil.Ok(t, err)
			} else {
				t.Logf("expected error: %s", err)
				testutil.NotOk(t, err, "")
				return
			}

			if !tc.fail {
				var x, y string
				x = tc.expected.String()
				y = Stmts.String()
				testutil.Ok(t, err)
				testutil.Equals(t, x, y)
				t.Logf("ok %s -> %s", tc.in, y)
			} else {
				t.Logf("%s -> expect fail: %v", tc.in, err)
				testutil.NotOk(t, err, "")
			}
		})
	}
}

func TestParser(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		expected ast.Stmts
		err      string
		fail     bool
	}{
		{
			name: "if-condition-list-paren2",
			in:   `if ((a==b) && (a==c)) || a==d { }`,
			expected: ast.Stmts{ast.WrapIfelseStmt(&ast.IfelseStmt{
				IfList: ast.IfList{
					&ast.IfStmtElem{Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
						Op: AstOp(OR),
						LHS: ast.WrapParenExpr(&ast.ParenExpr{
							Param: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op: AstOp(AND),
								LHS: ast.WrapParenExpr(&ast.ParenExpr{
									Param: ast.WrapConditionExpr(&ast.ConditionalExpr{
										Op:  AstOp(EQEQ),
										LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
										RHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
									}),
								}),
								RHS: ast.WrapParenExpr(&ast.ParenExpr{
									Param: ast.WrapConditionExpr(&ast.ConditionalExpr{
										Op:  AstOp(EQEQ),
										LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
										RHS: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
									}),
								}),
							}),
						}),
						RHS: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
							RHS: ast.WrapIdentifier(&ast.Identifier{Name: "d"}),
						}),
					})},
				},
			})},
		},

		{
			name: "if-condition-list-paren",
			in:   `if (a==b) && (a==c) { }`,
			expected: ast.Stmts{ast.WrapIfelseStmt(&ast.IfelseStmt{
				IfList: ast.IfList{&ast.IfStmtElem{
					Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
						Op: AstOp(AND),
						LHS: ast.WrapParenExpr(&ast.ParenExpr{
							Param: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op:  AstOp(EQEQ),
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
								RHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
							}),
						}),
						RHS: ast.WrapParenExpr(&ast.ParenExpr{
							Param: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op:  AstOp(EQEQ),
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
								RHS: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
							}),
						}),
					}),
				}},
			})},
		},

		{
			name: "if-condition-list",
			in:   `if a==b && a==c { }`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{
						&ast.IfStmtElem{
							Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op: AstOp(AND),
								LHS: ast.WrapConditionExpr(&ast.ConditionalExpr{
									Op:  AstOp(EQEQ),
									LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
									RHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
								}),
								RHS: ast.WrapConditionExpr(&ast.ConditionalExpr{
									Op:  AstOp(EQEQ),
									LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
									RHS: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
								}),
							}),
						},
					},
				}),
			},
		},

		{
			name: "if-error-non-condition",
			in:   `if { }`,
			fail: true,
		},

		{
			name: "if-elif-error-non-condition",
			in: `
			if key=="11" {

			} elif {

			}`,
			fail: true,
		},

		{
			name: "if-elif-elif-error-non-condition",
			in: `
			if key=="11" {

			} elif key=="22" {

			} elif {

			}`,
			fail: true,
		},

		{
			name: "if-elif-else-expr",
			in: `
			if key=="11" {
				g1(arg)
			} elif key=="22" {
				g2(arg)
			} else {
				g3(arg)
			}`,
			expected: ast.Stmts{ast.WrapIfelseStmt(&ast.IfelseStmt{
				IfList: ast.IfList{
					&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
						}),
						Stmts: ast.Stmts{
							ast.WrapCallExpr(&ast.CallExpr{
								Name: "g1",
								Param: []*ast.Node{ast.WrapIdentifier(
									&ast.Identifier{Name: "arg"})},
							}),
						},
					},
					&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "22"}),
						}),
						Stmts: ast.Stmts{ast.WrapCallExpr(&ast.CallExpr{
							Name: "g2",
							Param: []*ast.Node{ast.WrapIdentifier(
								&ast.Identifier{Name: "arg"})},
						})},
					},
				},
				Else: ast.Stmts{
					ast.WrapCallExpr(&ast.CallExpr{
						Name: "g3",
						Param: []*ast.Node{
							ast.WrapIdentifier(&ast.Identifier{Name: "arg"}),
						},
					}),
				},
			})},
		},

		{
			name: "if-elif-expr",
			in: `
			if key=="11" {
				g1(arg)
			} elif key=="22" {
				g2(arg)
			}`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{
						&ast.IfStmtElem{
							Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op:  AstOp(EQEQ),
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
								RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
							}),
							Stmts: ast.Stmts{ast.WrapCallExpr(&ast.CallExpr{
								Name: "g1",
								Param: []*ast.Node{
									ast.WrapIdentifier(&ast.Identifier{Name: "arg"}),
								},
							})},
						},
						&ast.IfStmtElem{
							Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op:  AstOp(EQEQ),
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
								RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "22"}),
							}),
							Stmts: ast.Stmts{
								ast.WrapCallExpr(&ast.CallExpr{
									Name: "g2",
									Param: []*ast.Node{ast.WrapIdentifier(
										&ast.Identifier{Name: "arg"})},
								}),
							},
						},
					},
				}),
			},
		},

		{
			name: "if-elif-else-non-ast.Stmts",
			in: `
			if key=="11"{

			} elif key=="22" {

			} else {

			}`,
			expected: ast.Stmts{ast.WrapIfelseStmt(&ast.IfelseStmt{
				IfList: ast.IfList{
					&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
						}),
					},
					&ast.IfStmtElem{Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
						Op:  AstOp(EQEQ),
						LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
						RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "22"}),
					})},
				},
				Else: ast.Stmts{},
			})},
		},

		{
			name: "if-elif-non-ast.Stmts",
			in: `
			if key=="11" {

			} elif key=="22" {

			}`,
			expected: ast.Stmts{ast.WrapIfelseStmt(&ast.IfelseStmt{
				IfList: ast.IfList{
					&ast.IfStmtElem{Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
						Op:  AstOp(EQEQ),
						LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
						RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
					})},
					&ast.IfStmtElem{Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
						Op:  AstOp(EQEQ),
						LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
						RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "22"}),
					})},
				},
			})},
		},

		{
			name: "if-expr-non-ast.Stmts",
			in:   `if key=="11" { }`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
						}),
					}},
				}),
			},
		},
		{
			name: "if-else-expr-non-ast.Stmts",
			in: `
			if key=="11" {

			} else {

			}`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{
						&ast.IfStmtElem{
							Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op:  AstOp(EQEQ),
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
								RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
							}),
						},
					},
				}),
			},
		},

		{
			name: "if-expr-non-ast.Stmts",
			in:   `if key=="11" { }`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
						}),
					}},
				}),
			},
		},

		{
			name: "if-else-expr",
			in:   `if key=="11" { g1(arg);   g2(arg) } else { h(arg) } #ddwwdw `,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{
						&ast.IfStmtElem{
							Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op:  AstOp(EQEQ),
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
								RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
							}),
							Stmts: ast.Stmts{
								ast.WrapCallExpr(&ast.CallExpr{
									Name: "g1",
									Param: []*ast.Node{
										ast.WrapIdentifier(
											&ast.Identifier{Name: "arg"}),
									},
								}),
								ast.WrapCallExpr(&ast.CallExpr{
									Name: "g2",
									Param: []*ast.Node{
										ast.WrapIdentifier(
											&ast.Identifier{Name: "arg"}),
									},
								}),
							},
						},
					},
					Else: ast.Stmts{
						ast.WrapCallExpr(&ast.CallExpr{
							Name: "h",
							Param: []*ast.Node{
								ast.WrapIdentifier(
									&ast.Identifier{Name: "arg"}),
							},
						}),
					},
				},
				),
			},
		},
		// {
		// 	name: "if-else-expr",
		// 	in:   `if match(_,"./*")=="11" { g1(arg) g2(arg) } else { h(arg) }`,
		// 	expected: ast.Stmts{
		// 		&ast.IfelseStmt{
		// 			ast.IfList: ast.IfList{
		// 				&IfExpr{
		// 					Condition: &ast.ConditionalExpr{
		// 						Op:  AstOp(EQEQ),
		// 						LHS: &ast.Identifier{Name: "match(_, './*')"},
		// 						RHS: &ast.StringLiteral{Val: "11"},
		// 					},
		// 					ast.Stmts: ast.Stmts{
		// 						&FuncStmt{
		// 							Name:  "g1",
		// 							Param: []ast.Node{&ast.Identifier{Name: "arg"}},
		// 						},
		// 						&FuncStmt{
		// 							Name:  "g2",
		// 							Param: []ast.Node{&ast.Identifier{Name: "arg"}},
		// 						},
		// 					},
		// 				},
		// 			},
		// 			Else: ast.Stmts{
		// 				&FuncStmt{
		// 					Name:  "h",
		// 					Param: []ast.Node{&ast.Identifier{Name: "arg"}},
		// 				},
		// 			},
		// 		},
		// 	},
		// },

		{
			name: "if-else-expr-newline",
			in: `
			if key=="11" { #11aa
				g(arg) } else {


				h(arg)
			}`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
						}),
						Stmts: ast.Stmts{
							ast.WrapCallExpr(&ast.CallExpr{
								Name: "g",
								Param: []*ast.Node{
									ast.WrapIdentifier(&ast.Identifier{Name: "arg"}),
								},
							}),
						},
					}},
					Else: ast.Stmts{ast.WrapCallExpr(&ast.CallExpr{
						Name: "h",
						Param: []*ast.Node{
							ast.WrapIdentifier(&ast.Identifier{Name: "arg"}),
						},
					})},
				}),
			},
		},

		{
			name: "if-nil",
			in:   `if abc == nil {}`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{
						&ast.IfStmtElem{
							Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
								Op:  AstOp(EQEQ),
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "abc"}),
								RHS: ast.WrapNilLiteral(&ast.NilLiteral{}),
							}),
						},
					},
				}),
			},
		},

		{
			name: "if-expr",
			in:   `if key=="11" { g(arg) }`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{IfList: ast.IfList{
					&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
						}),
						Stmts: ast.Stmts{
							ast.WrapCallExpr(&ast.CallExpr{
								Name: "g",
								Param: []*ast.Node{
									ast.WrapIdentifier(&ast.Identifier{Name: "arg"}),
								},
							}),
						},
					},
				}}),
			},
		},

		{
			name: "if-expr-newline",
			in: `
			if key=="11" {
				g(arg)
			}`,
			expected: ast.Stmts{
				ast.WrapIfelseStmt(&ast.IfelseStmt{
					IfList: ast.IfList{&ast.IfStmtElem{
						Condition: ast.WrapConditionExpr(&ast.ConditionalExpr{
							Op:  AstOp(EQEQ),
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "key"}),
							RHS: ast.WrapStringLiteral(&ast.StringLiteral{Val: "11"}),
						}),
						Stmts: ast.Stmts{ast.WrapCallExpr(&ast.CallExpr{
							Name: "g",
							Param: []*ast.Node{
								ast.WrapIdentifier(&ast.Identifier{Name: "arg"}),
							},
						})},
					}},
				}),
			},
		},

		{
			name: "array-list in function arg value",
			in:   `f([1, 2.0, "3", true, false, nil, null, id_123])`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "f",
					Param: []*ast.Node{ast.WrapListInitExpr(&ast.ListInitExpr{
						List: []*ast.Node{
							ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
							ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 2.0}),
							ast.WrapStringLiteral(&ast.StringLiteral{Val: "3"}),
							ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
							ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
							ast.WrapNilLiteral(&ast.NilLiteral{}),
							ast.WrapNilLiteral(&ast.NilLiteral{}),
							ast.WrapIdentifier(&ast.Identifier{Name: "id_123"}),
						},
					})},
				}),
			},
		},

		{
			name: "attr-expr in function arg value",
			in:   `f(arg=a.b.c)`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "f",
					Param: []*ast.Node{ast.WrapAssignmentExpr(&ast.AssignmentExpr{
						LHS: ast.WrapIdentifier(&ast.Identifier{Name: "arg"}),
						RHS: ast.WrapAttrExpr(&ast.AttrExpr{
							Obj: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
							Attr: ast.WrapAttrExpr(&ast.AttrExpr{
								Obj: ast.WrapAttrExpr(&ast.AttrExpr{
									Obj:  ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
									Attr: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
								}),
							}),
						}),
					})},
				}),
			},
		},

		{
			name: "func_call_in_assignement_right",
			in:   `a = fn("a", true, a1=["b", 1.1])`,
			expected: ast.Stmts{
				ast.WrapAssignmentExpr(&ast.AssignmentExpr{
					LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
					RHS: ast.WrapCallExpr(&ast.CallExpr{
						Name: "fn",
						Param: []*ast.Node{
							ast.WrapStringLiteral(&ast.StringLiteral{Val: "a"}),
							ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
							ast.WrapAssignmentExpr(&ast.AssignmentExpr{
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a1"}),
								RHS: ast.WrapListInitExpr(&ast.ListInitExpr{
									List: []*ast.Node{
										ast.WrapStringLiteral(&ast.StringLiteral{Val: "b"}),
										ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 1.1}),
									},
								}),
							}),
						},
					}),
				}),
			},
		},

		{
			name: "naming args",
			in:   `f(arg1=1, arg2=2)`,
			expected: ast.Stmts{
				ast.WrapCallExpr(
					&ast.CallExpr{
						Name: "f",
						Param: []*ast.Node{
							ast.WrapAssignmentExpr(&ast.AssignmentExpr{
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "arg1"}),
								RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
							}),
							ast.WrapAssignmentExpr(&ast.AssignmentExpr{
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "arg2"}),
								RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 2}),
							}),
						},
					},
				),
			},
		},

		{
			name: "func-cond-param-and-kwarg",
			in:   `f(arg1>1, arg2=2)`,
			expected: ast.Stmts{
				ast.WrapCallExpr(
					&ast.CallExpr{
						Name: "f",
						Param: []*ast.Node{
							ast.WrapConditionExpr(&ast.ConditionalExpr{
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "arg1"}),
								Op:  ast.GT,
								RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
							}),
							ast.WrapAssignmentExpr(&ast.AssignmentExpr{
								LHS: ast.WrapIdentifier(&ast.Identifier{Name: "arg2"}),
								RHS: ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 2}),
							}),
						},
					},
				),
			},
		},

		{
			name: "json attr",
			in:   `f(.[2].x[3])`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: `f`,
					Param: []*ast.Node{ast.WrapAttrExpr(&ast.AttrExpr{
						Obj: ast.WrapIndexExpr(
							&ast.IndexExpr{Index: []*ast.Node{
								ast.WrapNumberLiteral(&ast.NumberLiteral{
									IsInt: true,
									Int:   2,
								}),
							}},
						),
						Attr: ast.WrapIndexExpr(
							&ast.IndexExpr{
								Obj: &ast.Identifier{Name: "x"},
								Index: []*ast.Node{
									ast.WrapNumberLiteral(&ast.NumberLiteral{
										IsInt: true,
										Int:   3,
									}),
								},
							},
						),
					})},
				}),
			},
		},

		{ // 如果 y 是 map 则支持
			name: "multi-dim arr",
			in: `f(x.y[2.5])
			{}
			a = {"1": 2, 3.1: 1}
			`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "f",
					Param: []*ast.Node{
						ast.WrapAttrExpr(&ast.AttrExpr{
							Obj: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
							Attr: ast.WrapIndexExpr(&ast.IndexExpr{
								Obj:   &ast.Identifier{Name: "y"},
								Index: []*ast.Node{ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 2.5})},
							}),
						}),
					},
				}),
				ast.WrapMapInitExpr(&ast.MapInitExpr{}),
				ast.WrapAssignmentExpr(&ast.AssignmentExpr{
					LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
					RHS: ast.WrapMapInitExpr(&ast.MapInitExpr{
						KeyValeList: [][2]*ast.Node{
							{
								ast.WrapStringLiteral(&ast.StringLiteral{Val: "1"}),
								ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 2}),
							},
							{
								ast.WrapNumberLiteral(&ast.NumberLiteral{Float: 3.1}),
								ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
							},
						},
					}),
				}),
			},
		},

		{
			in: `f(x.y[1][2].z)`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: `f`,
					Param: []*ast.Node{
						ast.WrapAttrExpr(&ast.AttrExpr{
							Obj: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
							Attr: ast.WrapAttrExpr(&ast.AttrExpr{
								Obj: ast.WrapIndexExpr(&ast.IndexExpr{
									Obj: &ast.Identifier{Name: "y"},
									Index: []*ast.Node{
										ast.WrapNumberLiteral(
											&ast.NumberLiteral{
												IsInt: true,
												Int:   1,
											},
										),
										ast.WrapNumberLiteral(
											&ast.NumberLiteral{
												IsInt: true,
												Int:   2,
											},
										),
									},
								}),
								Attr: ast.WrapIdentifier(&ast.Identifier{Name: "z"}),
							}),
						},
						),
					},
				}),
			},
		},

		{
			name: "case:-multiple-functions",
			in: `f1()
		f2()
		f3()`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: `f1`,
				}),
				ast.WrapCallExpr(&ast.CallExpr{
					Name: `f2`,
				}),
				ast.WrapCallExpr(&ast.CallExpr{
					Name: `f3`,
				}),
			},
		},

		{
			name: "embedded functions",
			in:   `f1(g(f2("abc"), 123.0), 1, 2)`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "f1",
					Param: []*ast.Node{
						ast.WrapCallExpr(&ast.CallExpr{
							Name: "g",
							Param: []*ast.Node{
								ast.WrapCallExpr(
									&ast.CallExpr{
										Name: "f2",
										Param: []*ast.Node{
											ast.WrapStringLiteral(&ast.StringLiteral{Val: "abc"}),
										},
									},
								),
								ast.WrapNumberLiteral(&ast.NumberLiteral{
									Float: 123.0,
								}),
							},
						}),
						ast.WrapNumberLiteral(&ast.NumberLiteral{
							IsInt: true,
							Int:   1,
						}), ast.WrapNumberLiteral(&ast.NumberLiteral{
							IsInt: true,
							Int:   2,
						}),
					},
				},
				),
			},
		},

		{
			name: "case: attr syntax in function arg",
			in:   `avg(x.y.z, 1,2,3, p68, re("cd"), pqa)`,
			// 函数参数不可以 re()
			fail: true,
		},

		{
			name: "attr syntax with index syntax in function arg",
			in:   `json(_, x.y[1].z)`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "json",
					Param: []*ast.Node{
						ast.WrapIdentifier(&ast.Identifier{Name: "_"}),
						ast.WrapAttrExpr(&ast.AttrExpr{
							Obj: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
							Attr: ast.WrapAttrExpr(&ast.AttrExpr{
								Obj: ast.WrapIndexExpr(&ast.IndexExpr{
									Obj: &ast.Identifier{Name: "y"},
									Index: []*ast.Node{ast.WrapNumberLiteral(&ast.NumberLiteral{
										IsInt: true,
										Int:   1,
									})},
								}),
								Attr: ast.WrapIdentifier(&ast.Identifier{Name: "z"}),
							}),
						}),
					},
				}),
			},
		},

		{
			name: "simple attr syntax",
			in:   `json(_, x.y.z)`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "json",
					Param: []*ast.Node{
						ast.WrapIdentifier(&ast.Identifier{Name: "_"}),
						ast.WrapAttrExpr(&ast.AttrExpr{
							Obj: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
							Attr: ast.WrapAttrExpr(&ast.AttrExpr{
								Obj:  ast.WrapIdentifier(&ast.Identifier{Name: "y"}),
								Attr: ast.WrapIdentifier(&ast.Identifier{Name: "z"}),
							}),
						}),
					},
				}),
			},
		},

		{
			name: "simple attr syntax",
			in:   `match(_,"p([a-z]+)ch")`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "match",
					Param: []*ast.Node{
						ast.WrapIdentifier(&ast.Identifier{Name: "_"}),
						ast.WrapStringLiteral(&ast.StringLiteral{Val: "p([a-z]+)ch"}),
					},
				}),
			},
		},

		{
			name: "many param",
			in:   `f(a, b, 1, 2, )`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name: "f",
					Param: []*ast.Node{
						ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
						ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
						ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 1}),
						ast.WrapNumberLiteral(&ast.NumberLiteral{IsInt: true, Int: 2}),
					},
				}),
			},
		},

		{
			name: `func-arg-with-multi-line-string`,
			in: `abc(x, """
this
is
multiline-string
""")`,
			expected: ast.Stmts{
				ast.WrapCallExpr(
					&ast.CallExpr{
						Name: "abc",
						Param: []*ast.Node{
							ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
							ast.WrapStringLiteral(&ast.StringLiteral{Val: `
this
is
multiline-string
`}),
						},
					},
				),
			},
		},
		{
			name: `func-func`,
			in: `f1()
			f2()
			`,
			expected: ast.Stmts{
				ast.WrapCallExpr(&ast.CallExpr{
					Name:  "f1",
					Param: nil,
				}),
				ast.WrapCallExpr(&ast.CallExpr{
					Name:  "f2",
					Param: nil,
				}),
			},
		},
	}

	// for idx := len(cases) - 1; idx >= 0; idx-- {
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			Stmts, err := ParsePipeline(tc.in)

			if !tc.fail {
				testutil.Ok(t, err)
			} else {
				t.Logf("expected error: %s", err)
				if err == nil {
					t.Error(tc)
				}
				testutil.NotOk(t, err, "")
				return
			}

			if !tc.fail {
				var x, y string
				x = tc.expected.String()
				y = Stmts.String()
				testutil.Ok(t, err)
				testutil.Equals(t, x, y)
				t.Logf("ok %s -> %s", tc.in, y)
			} else {
				t.Logf("%s -> expect fail: %v", tc.in, err)
				testutil.NotOk(t, err, "")
			}
		})
	}
}
