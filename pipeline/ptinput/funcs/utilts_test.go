// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/stretchr/testify/assert"
)

func TestReIndexFuncArgs(t *testing.T) {
	cases := []struct {
		name    string
		keyList []string
		reqParm int
		fnArgs  *ast.CallExpr
		exp     *ast.CallExpr
		fail    bool
	}{
		{
			name:    "t1",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &ast.CallExpr{
				Param: ast.Stmts{
					ast.WrapIdentifier(&ast.Identifier{Name: "p1"}),
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
				},
			},
			exp: &ast.CallExpr{
				Param: ast.Stmts{
					ast.WrapIdentifier(&ast.Identifier{Name: "p1"}),
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
					nil,
				},
			},
		},
		{
			name:    "t1-fail-pos-after-kw",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &ast.CallExpr{
				Param: ast.Stmts{
					ast.WrapAssignmentExpr(&ast.AssignmentExpr{
						LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
						RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
					}),
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
				},
			},
			fail: true,
		},
		{
			name:    "t1-fail-key-not-exist",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "x"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
						},
					),
				},
			},
			fail: true,
		},
		{
			name:    "t1-nil-param",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
						},
					),
				},
			},
			fail: true,
		},
		{
			name:    "t1-1",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
						},
					),
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "a"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
						},
					),
				},
			},
			exp: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
					nil,
				},
			},
		},
		{
			name:    "t2",
			keyList: []string{"a", "b", "c"},
			reqParm: -1,
			fnArgs: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapIdentifier(&ast.Identifier{Name: "p1"}),
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
						},
					),
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "b"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
						},
					),
				},
			},
			exp: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapIdentifier(&ast.Identifier{Name: "p1"}),
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: false}),
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
				},
			},
		},
		{
			name:    "t3",
			keyList: []string{"a", "b", "c"},
			reqParm: 1,
			fnArgs: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapIdentifier(&ast.Identifier{Name: "p1"}),
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
						},
					),
				},
			},
			exp: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapIdentifier(&ast.Identifier{Name: "p1"}),
					nil,
					ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
				},
			},
		},
		{
			name:    "t3-fail",
			keyList: []string{"a"},
			reqParm: 1,
			fnArgs: &ast.CallExpr{
				Param: []*ast.Node{
					ast.WrapIdentifier(&ast.Identifier{Name: "p1"}),
					ast.WrapAssignmentExpr(
						&ast.AssignmentExpr{
							LHS: ast.WrapIdentifier(&ast.Identifier{Name: "c"}),
							RHS: ast.WrapBoolLiteral(&ast.BoolLiteral{Val: true}),
						},
					),
				},
			},
			fail: true,
		},
	}

	for _, v := range cases {
		t.Run(v.name, func(t *testing.T) {
			err := reIndexFuncArgs(v.fnArgs, v.keyList, v.reqParm)
			if err != nil {
				if !v.fail {
					t.Error(err)
				}
				return
			}
			for i, p := range v.fnArgs.Param {
				assert.Equal(t, v.exp.Param[i], p)
			}
		})
	}
}
