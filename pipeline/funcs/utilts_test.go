// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/parser"
)

func TestReIndexFuncArgs(t *testing.T) {
	cases := []struct {
		name    string
		keyList []string
		reqParm int
		fnArgs  *parser.FuncStmt
		exp     *parser.FuncStmt
		fail    bool
	}{
		{
			name:    "t1",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.Identifier{Name: "p1"},
					&parser.BoolLiteral{Val: false},
				},
			},
			exp: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.Identifier{Name: "p1"},
					&parser.BoolLiteral{Val: false},
					nil,
				},
			},
		},
		{
			name:    "t1-fail-pos-after-kw",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "a"},
						RHS: &parser.BoolLiteral{Val: true},
					},
					&parser.BoolLiteral{Val: false},
				},
			},
			fail: true,
		},
		{
			name:    "t1-fail-key-not-exist",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.BoolLiteral{Val: false},
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "x"},
						RHS: &parser.BoolLiteral{Val: true},
					},
				},
			},
			fail: true,
		},
		{
			name:    "t1-nil-param",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "b"},
						RHS: &parser.BoolLiteral{Val: true},
					},
				},
			},
			fail: true,
		},
		{
			name:    "t1-1",
			keyList: []string{"a", "b", "c"},
			reqParm: 2,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "b"},
						RHS: &parser.BoolLiteral{Val: true},
					},
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "a"},
						RHS: &parser.BoolLiteral{Val: false},
					},
				},
			},
			exp: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.BoolLiteral{Val: false},
					&parser.BoolLiteral{Val: true},
					nil,
				},
			},
		},
		{
			name:    "t2",
			keyList: []string{"a", "b", "c"},
			reqParm: -1,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.Identifier{Name: "p1"},
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "c"},
						RHS: &parser.BoolLiteral{Val: true},
					},
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "b"},
						RHS: &parser.BoolLiteral{Val: false},
					},
				},
			},
			exp: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.Identifier{Name: "p1"},
					&parser.BoolLiteral{Val: false},
					&parser.BoolLiteral{Val: true},
				},
			},
		},
		{
			name:    "t3",
			keyList: []string{"a", "b", "c"},
			reqParm: 1,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.Identifier{Name: "p1"},
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "c"},
						RHS: &parser.BoolLiteral{Val: true},
					},
				},
			},
			exp: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.Identifier{Name: "p1"},
					nil,
					&parser.BoolLiteral{Val: true},
				},
			},
		},
		{
			name:    "t3-fail",
			keyList: []string{"a"},
			reqParm: 1,
			fnArgs: &parser.FuncStmt{
				Param: []parser.Node{
					&parser.Identifier{Name: "p1"},
					&parser.AssignmentStmt{
						LHS: &parser.Identifier{Name: "c"},
						RHS: &parser.BoolLiteral{Val: true},
					},
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
