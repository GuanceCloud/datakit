// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package runtime provide a runtime for the pipeline
package runtime

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
	"github.com/spf13/cast"
)

type (
	FuncCheck func(*Task, *ast.CallExpr) *errchain.PlError
	FuncCall  func(*Task, *ast.CallExpr) *errchain.PlError
)

type Script struct {
	CallRef []*ast.CallExpr

	FuncCall map[string]FuncCall

	Name      string
	Namespace string
	Category  string
	FilePath  string

	Content string // deprecated

	Ast ast.Stmts
}

type Signal interface {
	ExitSignal() bool
}

func WithVal(key string, val any) TaskFn {
	return func(ctx *Task) {
		_ = ctx.WithVal(key, val, false)
	}
}

type TaskFn func(ctx *Task)

func (s *Script) Run(data Input, signal Signal, fn ...TaskFn) *errchain.PlError {
	if s == nil {
		return nil
	}

	ctx := GetContext()
	defer PutContext(ctx)

	for _, fn := range fn {
		fn(ctx)
	}

	ctx = InitCtx(ctx, data, s, signal)
	return RunStmts(ctx, s.Ast)
}

func (s *Script) RefRun(ctx *Task) *errchain.PlError {
	if s == nil {
		return nil
	}

	newctx := GetContext()
	defer PutContext(newctx)

	InitCtx(newctx, ctx.input, s, ctx.signal)

	return RunStmts(newctx, s.Ast)
}

func (s *Script) Check(funcsCheck map[string]FuncCheck) *errchain.PlError {
	if s == nil {
		return nil
	}

	ctx := GetContext()
	defer PutContext(ctx)
	InitCtxForCheck(ctx, s, funcsCheck)
	if err := RunStmtsCheck(ctx, &ContextCheck{}, s.Ast); err != nil {
		return err
	}

	s.CallRef = ctx.callRef
	return nil
}

func RunStmts(ctx *Task, nodes ast.Stmts) *errchain.PlError {
	for _, node := range nodes {
		if _, _, err := RunStmt(ctx, node); err != nil {
			ctx.procExit = true
			return err
		}

		if ctx.StmtRetrun() {
			return nil
		}
	}
	return nil
}

func RunIfElseStmt(ctx *Task, stmt *ast.IfelseStmt) (any, ast.DType, *errchain.PlError) {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	// check if or elif condition
	for _, ifstmt := range stmt.IfList {
		// check condition
		val, dtype, err := RunStmt(ctx, ifstmt.Condition)
		if err != nil {
			return nil, ast.Invalid, err
		}
		if !condTrue(val, dtype) {
			continue
		}

		if ifstmt.Block != nil {
			// run if or elif stmt
			ctx.StackEnterNew()

			if err := RunStmts(ctx, ifstmt.Block.Stmts); err != nil {
				return nil, ast.Void, err
			}

			ctx.StackExitCur()
		}

		return nil, ast.Void, nil
	}

	if stmt.Else != nil {
		// run else stmt
		ctx.StackEnterNew()
		if err := RunStmts(ctx, stmt.Else.Stmts); err != nil {
			return nil, ast.Void, err
		}
		ctx.StackExitCur()
	}

	return nil, ast.Void, nil
}

func condTrue(val any, dtype ast.DType) bool {
	switch dtype { //nolint:exhaustive
	case ast.String:
		if cast.ToString(val) == "" {
			return false
		}
	case ast.Bool:
		return cast.ToBool(val)
	case ast.Int:
		if cast.ToInt64(val) == 0 {
			return false
		}
	case ast.Float:
		if cast.ToFloat64(val) == 0 {
			return false
		}
	case ast.List:
		if len(cast.ToSlice(val)) == 0 {
			return false
		}
	case ast.Map:
		switch v := val.(type) {
		case map[string]any:
			if len(v) == 0 {
				return false
			}
		default:
			return false
		}
	default:
		return false
	}
	return true
}

func RunForStmt(ctx *Task, stmt *ast.ForStmt) (any, ast.DType, *errchain.PlError) {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	// for init
	if stmt.Init != nil {
		_, _, err := RunStmt(ctx, stmt.Init)
		if err != nil {
			return nil, ast.Invalid, err
		}
	}

	for {
		if stmt.Cond != nil {
			val, dtype, err := RunStmt(ctx, stmt.Cond)
			if err != nil {
				return nil, ast.Invalid, err
			}
			if !condTrue(val, dtype) {
				break
			}
		}

		if stmt.Body != nil {
			// for body
			ctx.StackEnterNew()
			if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
				return nil, ast.Invalid, err
			}
			ctx.StackExitCur()
		}

		if ctx.loopBreak {
			ctx.loopBreak = false
			break
		}

		if ctx.loopContinue {
			ctx.loopContinue = false
		}

		if ctx.StmtRetrun() {
			break
		}

		// loop stmt
		if stmt.Loop != nil {
			_, _, err := RunStmt(ctx, stmt.Loop)
			if err != nil {
				return nil, ast.Invalid, err
			}
		}
	}

	return nil, ast.Void, nil
}

func RunForInStmt(ctx *Task, stmt *ast.ForInStmt) (any, ast.DType, *errchain.PlError) {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	iter, dtype, err := RunStmt(ctx, stmt.Iter)
	if err != nil {
		return nil, ast.Invalid, err
	}

	ctx.StackEnterNew()
	defer ctx.StackExitCur()
	switch dtype { //nolint:exhaustive
	case ast.String:
		iter, ok := iter.(string)
		if !ok {
			return nil, ast.Invalid, NewRunError(ctx,
				"inner type error", stmt.Iter.StartPos())
		}
		for _, x := range iter {
			char := string(x)
			if stmt.Varb.NodeType != ast.TypeIdentifier {
				return nil, ast.Invalid, err
			}
			_ = ctx.SetVarb(stmt.Varb.Identifier().Name, char, ast.String)
			if stmt.Body != nil {
				if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
					return nil, ast.Invalid, err
				}
			}
			ctx.stackCur.Clear()

			if forbreak(ctx) {
				break
			}
			forcontinue(ctx)
			if ctx.StmtRetrun() {
				break
			}
		}
	case ast.Map:
		iter, ok := iter.(map[string]any)
		if !ok {
			return nil, ast.Invalid, NewRunError(ctx,
				"inner type error", stmt.Iter.StartPos())
		}
		for x := range iter {
			ctx.stackCur.Clear()
			_ = ctx.SetVarb(stmt.Varb.Identifier().Name, x, ast.String)
			if stmt.Body != nil {
				if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
					return nil, ast.Invalid, err
				}
			}
			if forbreak(ctx) {
				break
			}
			forcontinue(ctx)
			if ctx.StmtRetrun() {
				break
			}
		}
	case ast.List:
		iter, ok := iter.([]any)
		if !ok {
			return nil, ast.Invalid, NewRunError(ctx,
				"inner type error", stmt.Iter.StartPos())
		}
		for _, x := range iter {
			ctx.stackCur.Clear()
			x, dtype := ast.DectDataType(x)
			if dtype == ast.Invalid {
				return nil, ast.Invalid, NewRunError(ctx,
					"inner type error", stmt.Iter.StartPos())
			}
			_ = ctx.SetVarb(stmt.Varb.Identifier().Name, x, dtype)
			if stmt.Body != nil {
				if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
					return nil, ast.Invalid, err
				}
			}
			if forbreak(ctx) {
				break
			}
			forcontinue(ctx)
			if ctx.StmtRetrun() {
				break
			}
		}
	default:
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported type: %s, not iter value", dtype), stmt.Iter.StartPos())
	}

	return nil, ast.Void, nil
}

func forbreak(ctx *Task) bool {
	if ctx.loopBreak {
		ctx.loopBreak = false
		return true
	}
	return false
}

func forcontinue(ctx *Task) {
	if ctx.loopContinue {
		ctx.loopContinue = false
	}
}

func RunBreakStmt(ctx *Task, stmt *ast.BreakStmt) (any, ast.DType, *errchain.PlError) {
	ctx.loopBreak = true
	return nil, ast.Void, nil
}

func RunContinueStmt(ctx *Task, stmt *ast.ContinueStmt) (any, ast.DType, *errchain.PlError) {
	ctx.loopContinue = true
	return nil, ast.Void, nil
}

// RunStmt for all expr.
func RunStmt(ctx *Task, node *ast.Node) (any, ast.DType, *errchain.PlError) {
	// TODO
	// 存在个别 node 为 nil 的情况
	if node == nil {
		return nil, ast.Void, nil
	}
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeParenExpr:
		return RunParenExpr(ctx, node.ParenExpr())
	case ast.TypeArithmeticExpr:
		return RunArithmeticExpr(ctx, node.ArithmeticExpr())
	case ast.TypeConditionalExpr:
		return RunConditionExpr(ctx, node.ConditionalExpr())
	case ast.TypeUnaryExpr:
		return RunUnaryExpr(ctx, node.UnaryExpr())
	case ast.TypeAssignmentExpr:
		return RunAssignmentExpr(ctx, node.AssignmentExpr())
	case ast.TypeCallExpr:
		return RunCallExpr(ctx, node.CallExpr())
	case ast.TypeInExpr:
		return RunInExpr(ctx, node.InExpr())
	case ast.TypeListLiteral:
		return RunListInitExpr(ctx, node.ListLiteral())
	case ast.TypeIdentifier:
		if v, err := ctx.GetKey(node.Identifier().Name); err != nil {
			return nil, ast.Nil, nil
		} else {
			return v.Value, v.DType, nil
		}
	case ast.TypeMapLiteral:
		return RunMapInitExpr(ctx, node.MapLiteral())
	// use for map, slice and array
	case ast.TypeIndexExpr:
		return RunIndexExprGet(ctx, node.IndexExpr())

	// TODO
	case ast.TypeAttrExpr:
		return nil, ast.Void, nil

	case ast.TypeBoolLiteral:
		return node.BoolLiteral().Val, ast.Bool, nil

	case ast.TypeIntegerLiteral:
		return node.IntegerLiteral().Val, ast.Int, nil

	case ast.TypeFloatLiteral:
		return node.FloatLiteral().Val, ast.Float, nil

	case ast.TypeStringLiteral:
		return node.StringLiteral().Val, ast.String, nil

	case ast.TypeNilLiteral:
		return nil, ast.Nil, nil

	case ast.TypeIfelseStmt:
		return RunIfElseStmt(ctx, node.IfelseStmt())
	case ast.TypeForStmt:
		return RunForStmt(ctx, node.ForStmt())
	case ast.TypeForInStmt:
		return RunForInStmt(ctx, node.ForInStmt())
	case ast.TypeBreakStmt:
		return RunBreakStmt(ctx, node.BreakStmt())
	case ast.TypeContinueStmt:
		return RunContinueStmt(ctx, node.ContinueStmt())
	default:
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported ast node: %s", reflect.TypeOf(node).String()), node.StartPos())
	}
}

func RunUnaryExpr(ctx *Task, expr *ast.UnaryExpr) (any, ast.DType, *errchain.PlError) {
	switch expr.Op {
	case ast.SUB, ast.ADD:
		v, dtype, err := RunStmt(ctx, expr.RHS)
		if err != nil {
			return nil, ast.Invalid, err
		}
		switch dtype {
		case ast.Bool:
			val, _ := v.(bool)
			if expr.Op == ast.SUB {
				if val {
					return int64(-1), ast.Int, nil
				} else {
					return 0, ast.Int, nil
				}
			} else {
				if val {
					return int64(1), ast.Int, nil
				} else {
					return 0, ast.Int, nil
				}
			}
		case ast.Float:
			val, _ := v.(float64)
			if expr.Op == ast.SUB {
				return -val, ast.Float, nil
			} else {
				return val, ast.Float, nil
			}
		case ast.Int:
			val, _ := v.(int64)
			if expr.Op == ast.SUB {
				return -val, ast.Int, nil
			} else {
				return val, ast.Int, nil
			}
		default:
			return nil, ast.Invalid, NewRunError(ctx,
				fmt.Sprintf("unsuppored operand type for unary op %s: %s",
					expr.Op, reflect.TypeOf(expr).String()), expr.OpPos)
		}

	case ast.NOT:
		v, _, err := RunStmt(ctx, expr.RHS)
		if err != nil {
			return nil, ast.Invalid, err
		}

		if v == nil {
			return true, ast.Bool, nil
		}

		switch v := v.(type) {
		case bool:
			return !v, ast.Bool, nil
		case float64:
			if v == 0 {
				return true, ast.Bool, nil
			} else {
				return false, ast.Bool, nil
			}
		case int64:
			if v == 0 {
				return true, ast.Bool, nil
			} else {
				return false, ast.Bool, nil
			}
		case string:
			if len(v) == 0 {
				return true, ast.Bool, nil
			} else {
				return false, ast.Bool, nil
			}
		case map[string]any:
			if len(v) == 0 {
				return true, ast.Bool, nil
			} else {
				return false, ast.Bool, nil
			}
		case []any:
			if len(v) == 0 {
				return true, ast.Bool, nil
			} else {
				return false, ast.Bool, nil
			}

		default:
			return nil, ast.Invalid, NewRunError(ctx,
				fmt.Sprintf("unsuppored operand type for unary op %s: %s",
					expr.Op, reflect.TypeOf(expr).String()), expr.OpPos)
		}
	default:
		return nil, ast.Invalid, NewRunError(ctx,
			fmt.Sprintf("unsupported op for unary expr: %s", expr.Op), expr.OpPos)
	}
}

func RunListInitExpr(ctx *Task, expr *ast.ListLiteral) (any, ast.DType, *errchain.PlError) {
	ret := []any{}
	for _, v := range expr.List {
		v, _, err := RunStmt(ctx, v)
		if err != nil {
			return nil, ast.Invalid, err
		}
		ret = append(ret, v)
	}
	return ret, ast.List, nil
}

func RunMapInitExpr(ctx *Task, expr *ast.MapLiteral) (any, ast.DType, *errchain.PlError) {
	ret := map[string]any{}

	for _, v := range expr.KeyValeList {
		k, keyType, err := RunStmt(ctx, v[0])
		if err != nil {
			return nil, ast.Invalid, err
		}

		key, ok := k.(string)
		if !ok {
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported key data type: %s", keyType), v[0].StartPos())
		}
		value, valueType, err := RunStmt(ctx, v[1])
		if err != nil {
			return nil, ast.Invalid, err
		}
		switch valueType { //nolint:exhaustive
		case ast.String, ast.Bool, ast.Float, ast.Int,
			ast.Nil, ast.List, ast.Map:
		default:
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported value data type: %s", keyType), v[1].StartPos())
		}
		ret[key] = value
	}

	return ret, ast.Map, nil
}

// func indexKeyType(dtype ast.DType) bool {
// 	switch dtype { //nolint:exhaustive
// 	case ast.Int, ast.String:
// 		return true
// 	default:
// 		return false
// 	}
// }

func RunIndexExprGet(ctx *Task, expr *ast.IndexExpr) (any, ast.DType, *errchain.PlError) {
	key := expr.Obj.Name

	varb, err := ctx.GetKey(key)
	if err != nil {
		return nil, ast.Invalid, NewRunError(ctx, err.Error(), expr.Obj.Start)
	}
	switch varb.DType { //nolint:exhaustive
	case ast.List:
		switch varb.Value.(type) {
		case []any:
		default:
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported type: %v", reflect.TypeOf(varb.Value)), expr.Obj.Start)
		}
	case ast.Map:
		switch varb.Value.(type) {
		case map[string]any: // for json map
		default:
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported type: %v", reflect.TypeOf(varb.Value)), expr.Obj.Start)
		}
	default:
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unindexable type: %s", varb.DType), expr.Obj.Start)
	}

	return searchListAndMap(ctx, varb.Value, expr.Index)
}

func searchListAndMap(ctx *Task, obj any, index []*ast.Node) (any, ast.DType, *errchain.PlError) {
	cur := obj

	for _, i := range index {
		key, keyType, err := RunStmt(ctx, i)
		if err != nil {
			return nil, ast.Invalid, err
		}
		switch curVal := cur.(type) {
		case map[string]any:
			if keyType != ast.String {
				return nil, ast.Invalid, NewRunError(ctx,
					"key type is not string", i.StartPos())
			}
			var ok bool
			cur, ok = curVal[key.(string)]
			if !ok {
				return nil, ast.Nil, nil
			}
		case []any:
			if keyType != ast.Int {
				return nil, ast.Invalid, NewRunError(ctx,
					"key type is not int", i.StartPos())
			}
			keyInt := cast.ToInt(key)

			// 反转负数
			if keyInt < 0 {
				keyInt = len(curVal) + keyInt
			}

			if keyInt < 0 || keyInt >= len(curVal) {
				return nil, ast.Invalid, NewRunError(ctx,
					"list index out of range", i.StartPos())
			}
			cur = curVal[keyInt]
		default:
			return nil, ast.Invalid, NewRunError(ctx,
				"not found", i.StartPos())
		}
	}
	var dtype ast.DType
	cur, dtype = ast.DectDataType(cur)
	return cur, dtype, nil
}

func RunParenExpr(ctx *Task, expr *ast.ParenExpr) (any, ast.DType, *errchain.PlError) {
	return RunStmt(ctx, expr.Param)
}

// BinarayExpr

func RunInExpr(ctx *Task, expr *ast.InExpr) (any, ast.DType, *errchain.PlError) {
	lhs, lhsT, err := RunStmt(ctx, expr.LHS)
	if err != nil {
		return nil, ast.Invalid, err
	}

	rhs, rhsT, err := RunStmt(ctx, expr.RHS)
	if err != nil {
		return nil, ast.Invalid, err
	}

	switch rhsT {
	case ast.String:
		if lhsT != ast.String {
			return false, ast.Bool, NewRunError(ctx, fmt.Sprintf(
				"unsupported lhs data type: %s", lhsT), expr.OpPos)
		}
		if s, ok := lhs.(string); ok {
			if v, ok := rhs.(string); ok {
				return strings.Contains(v, s), ast.Bool, nil
			}
		}

		return false, ast.Bool, nil
	case ast.Map:
		if lhsT != ast.String {
			return false, ast.Bool, NewRunError(ctx, fmt.Sprintf(
				"unsupported lhs data type: %s", lhsT), expr.OpPos)
		}
		if s, ok := lhs.(string); ok {
			if v, ok := rhs.(map[string]any); ok {
				if _, ok := v[s]; ok {
					return true, ast.Bool, nil
				}
			}
		}
		return false, ast.Bool, nil
	case ast.List:
		if v, ok := rhs.([]any); ok {
			for _, elem := range v {
				if reflect.DeepEqual(lhs, elem) {
					return true, ast.Bool, nil
				}
			}
		}
		return false, ast.Bool, nil

	default:
		return false, ast.Bool, NewRunError(ctx, fmt.Sprintf(
			"unsupported rhs data type: %s", rhsT), expr.OpPos)
	}
}

func RunConditionExpr(ctx *Task, expr *ast.ConditionalExpr) (any, ast.DType, *errchain.PlError) {
	lhs, lhsT, err := RunStmt(ctx, expr.LHS)
	if err != nil {
		return nil, ast.Invalid, err
	}

	if lhsT == ast.Bool {
		switch expr.Op { //nolint:exhaustive
		case ast.OR:
			if cast.ToBool(lhs) {
				return true, ast.Bool, nil
			}
		case ast.AND:
			if !cast.ToBool(lhs) {
				return false, ast.Bool, nil
			}
		}
	}

	rhs, rhsT, err := RunStmt(ctx, expr.RHS)
	if err != nil {
		return nil, ast.Invalid, err
	}

	if val, dtype, err := condOp(lhs, rhs, lhsT, rhsT, expr.Op); err != nil {
		return nil, ast.Invalid, NewRunError(ctx, err.Error(), expr.OpPos)
	} else {
		return val, dtype, nil
	}
}

func RunArithmeticExpr(ctx *Task, expr *ast.ArithmeticExpr) (any, ast.DType, *errchain.PlError) {
	// 允许字符串通过操作符 '+' 进行拼接

	lhsVal, lhsValType, errOpInt := RunStmt(ctx, expr.LHS)
	if errOpInt != nil {
		return nil, ast.Invalid, errOpInt
	}

	rhsVal, rhsValType, errOpInt := RunStmt(ctx, expr.RHS)
	if errOpInt != nil {
		return nil, ast.Invalid, errOpInt
	}

	if !arithType(lhsValType) {
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported lhs data type: %s", lhsValType), expr.OpPos)
	}

	if !arithType(rhsValType) {
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported rhs data type: %s", rhsValType), expr.OpPos)
	}

	// string
	if lhsValType == ast.String || rhsValType == ast.String {
		if expr.Op != ast.ADD {
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				expr.Op, lhsValType, rhsValType), expr.OpPos)
		}
		if lhsValType == ast.String && rhsValType == ast.String {
			return cast.ToString(lhsVal) + cast.ToString(rhsVal), ast.String, nil
		} else {
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				expr.Op, lhsValType, rhsValType), expr.OpPos)
		}
	}

	// float
	if lhsValType == ast.Float || rhsValType == ast.Float {
		v, dtype, err := arithOpFloat(cast.ToFloat64(lhsVal), cast.ToFloat64(rhsVal), expr.Op)
		if err != nil {
			return nil, ast.Invalid, NewRunError(ctx, err.Error(), expr.OpPos)
		}
		return v, dtype, nil
	}

	// bool or int

	v, dtype, errOp := arithOpInt(cast.ToInt64(lhsVal), cast.ToInt64(rhsVal), expr.Op)

	if errOp != nil {
		return nil, ast.Invalid, NewRunError(ctx, errOp.Error(), expr.OpPos)
	}
	return v, dtype, nil
}

func runAssignArith(ctx *Task, l, r *Varb, op ast.Op, pos token.LnColPos) (
	any, ast.DType, *errchain.PlError) {

	arithOp, ok := assign2arithOp(op)
	if !ok {
		return nil, ast.Invalid, NewRunError(ctx,
			"unsupported op", pos)
	}

	if !arithType(l.DType) {
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported lhs data type: %s", l.DType), pos)
	}

	if !arithType(r.DType) {
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported rhs data type: %s", r.DType), pos)
	}

	// string
	if l.DType == ast.String || r.DType == ast.String {
		if arithOp != ast.ADD {
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				op, l.DType, r.DType), pos)
		}
		if l.DType == ast.String && r.DType == ast.String {
			return cast.ToString(l.Value) + cast.ToString(r.Value), ast.String, nil
		} else {
			return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				op, l.DType, r.DType), pos)
		}
	}

	// float
	if l.DType == ast.Float || r.DType == ast.Float {
		v, dtype, err := arithOpFloat(cast.ToFloat64(l.Value), cast.ToFloat64(r.Value), arithOp)
		if err != nil {
			return nil, ast.Invalid, NewRunError(ctx, err.Error(), pos)
		}
		return v, dtype, nil
	}

	// bool or int

	v, dtype, errOp := arithOpInt(cast.ToInt64(l.Value), cast.ToInt64(r.Value), arithOp)

	if errOp != nil {
		return nil, ast.Invalid, NewRunError(ctx, errOp.Error(), pos)
	}
	return v, dtype, nil
}

// RunAssignmentExpr runs assignment expression, but actually it is a stmt
func RunAssignmentExpr(ctx *Task, expr *ast.AssignmentExpr) (any, ast.DType, *errchain.PlError) {
	v, dtype, err := RunStmt(ctx, expr.RHS)
	if err != nil {
		return nil, ast.Invalid, err
	}
	rVarb := &Varb{Value: v, DType: dtype}

	switch expr.LHS.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
		switch expr.Op {
		case ast.EQ:
			_ = ctx.SetVarb(expr.LHS.Identifier().Name, v, dtype)
			return v, dtype, nil

		case ast.SUBEQ,
			ast.ADDEQ,
			ast.MULEQ,
			ast.DIVEQ,
			ast.MODEQ:
			lVarb, err := ctx.GetKey(expr.LHS.Identifier().Name)
			if err != nil {
				return nil, ast.Nil, nil
			}
			if v, dt, errR := runAssignArith(ctx, lVarb, rVarb, expr.Op, expr.OpPos); errR != nil {
				return nil, ast.Void, errR
			} else {
				_ = ctx.SetVarb(expr.LHS.Identifier().Name, v, dt)
				return v, dt, nil
			}

		default:
			return nil, ast.Invalid, NewRunError(ctx,
				"unsupported op", expr.OpPos)
		}
	case ast.TypeIndexExpr:
		switch expr.Op {
		case ast.EQ:
			varb, err := ctx.GetKey(expr.LHS.IndexExpr().Obj.Name)
			if err != nil {
				return nil, ast.Invalid, NewRunError(ctx, err.Error(), expr.LHS.IndexExpr().Obj.Start)
			}
			return changeListOrMapValue(ctx, varb.Value, expr.LHS.IndexExpr().Index,
				v, dtype)
		case ast.ADDEQ,
			ast.SUBEQ,
			ast.MULEQ,
			ast.DIVEQ,
			ast.MODEQ:
			varb, err := ctx.GetKey(expr.LHS.IndexExpr().Obj.Name)
			if err != nil {
				return nil, ast.Invalid, NewRunError(ctx, err.Error(), expr.LHS.IndexExpr().Obj.Start)
			}
			if v, dt, errR := searchListAndMap(ctx, varb.Value, expr.LHS.IndexExpr().Index); errR != nil {
				return nil, ast.Invalid, errR
			} else {
				v, dt, err := runAssignArith(ctx, &Varb{Value: v, DType: dt}, rVarb, expr.Op, expr.OpPos)
				if err != nil {
					return nil, ast.Invalid, err
				}
				return changeListOrMapValue(ctx, varb.Value, expr.LHS.IndexExpr().Index,
					v, dt)
			}
		default:
			return nil, ast.Invalid, NewRunError(ctx,
				"unsupported op", expr.OpPos)
		}

	default:
		return nil, ast.Void, nil
	}
}

func changeListOrMapValue(ctx *Task, obj any, index []*ast.Node, val any, dtype ast.DType) (any, ast.DType, *errchain.PlError) {
	cur := obj
	lenIdx := len(index)

	for idx, node := range index {
		key, keyType, err := RunStmt(ctx, node)
		if err != nil {
			return nil, ast.Invalid, err
		}
		switch curVal := cur.(type) {
		case map[string]any:
			if keyType != ast.String {
				return nil, ast.Invalid, NewRunError(ctx,
					"key type is not string", node.StartPos())
			}
			if idx+1 == lenIdx {
				curVal[key.(string)] = val
				return val, dtype, nil
			}

			var ok bool
			cur, ok = curVal[key.(string)]
			if !ok {
				return nil, ast.Invalid, NewRunError(ctx,
					"key not found", node.StartPos())
			}
		case []any:
			if keyType != ast.Int {
				return nil, ast.Invalid, NewRunError(ctx,
					"key type is not int", node.StartPos())
			}
			keyInt := cast.ToInt(key)

			// 反转负数
			if keyInt < 0 {
				keyInt = len(curVal) + keyInt
			}

			if keyInt < 0 || keyInt >= len(curVal) {
				return nil, ast.Invalid, NewRunError(ctx,
					"list index out of range", node.StartPos())
			}

			if idx+1 == lenIdx {
				curVal[keyInt] = val
				return val, dtype, nil
			}

			cur = curVal[keyInt]
		default:
			return nil, ast.Invalid, NewRunError(ctx,
				"obj not map or list", node.StartPos())
		}
	}
	return nil, ast.Nil, nil
}

func RunCallExpr(ctx *Task, expr *ast.CallExpr) (any, ast.DType, *errchain.PlError) {
	defer ctx.Regs.Reset()
	if funcCall, ok := ctx.GetFuncCall(expr.Name); ok {
		if err := funcCall(ctx, expr); err != nil {
			return nil, ast.Invalid, err
		}
		if ctx.Regs.Count() > 0 {
			if v, dtype, err := ctx.Regs.Get(RegR0); err != nil {
				return v, dtype, NewRunError(ctx, err.Error(), expr.NamePos)
			} else {
				return v, dtype, nil
			}
		}
	}
	return nil, ast.Void, nil
}

func typePromotion(l ast.DType, r ast.DType) ast.DType {
	if l == ast.Float || r == ast.Float {
		return ast.Float
	}

	return ast.Int
}

func condOp(lhs, rhs any, lhsT, rhsT ast.DType, op ast.Op) (any, ast.DType, error) {
	switch op { //nolint:exhaustive
	case ast.EQEQ:
		switch lhsT { //nolint:exhaustive
		case ast.Int, ast.Bool, ast.Float:
			switch rhsT { //nolint:exhaustive
			case ast.Int, ast.Bool, ast.Float:
			default:
				return false, ast.Bool, nil
			}
			dtype := typePromotion(lhsT, rhsT)
			if dtype == ast.Float {
				return cast.ToFloat64(lhs) == cast.ToFloat64(rhs), ast.Bool, nil
			}
			return cast.ToFloat64(lhs) == cast.ToFloat64(rhs), ast.Bool, nil
		case ast.String:
			if rhsT != ast.String {
				return false, ast.Bool, nil
			}
			return cast.ToString(lhs) == cast.ToString(rhs), ast.Bool, nil
		case ast.Nil:
			if rhsT != ast.Nil {
				return false, ast.Bool, nil
			}
			return true, ast.Bool, nil
		default:
			return reflect.DeepEqual(lhs, rhs), ast.Bool, nil
		}

	case ast.NEQ:
		switch lhsT { //nolint:exhaustive
		case ast.Int, ast.Bool, ast.Float:
			switch rhsT { //nolint:exhaustive
			case ast.Int, ast.Bool, ast.Float:
			default:
				return true, ast.Bool, nil
			}
			dtype := typePromotion(lhsT, rhsT)
			if dtype == ast.Float {
				return cast.ToFloat64(lhs) != cast.ToFloat64(rhs), ast.Bool, nil
			}
			return cast.ToFloat64(lhs) != cast.ToFloat64(rhs), ast.Bool, nil
		case ast.String:
			if rhsT != ast.String {
				return true, ast.Bool, nil
			}
			return cast.ToString(lhs) != cast.ToString(rhs), ast.Bool, nil
		case ast.Nil:
			if rhsT != ast.Nil {
				return true, ast.Bool, nil
			}
			return false, ast.Bool, nil
		default:
			return !reflect.DeepEqual(lhs, rhs), ast.Bool, nil
		}
	}

	if !cmpType(lhsT) {
		return nil, ast.Invalid, fmt.Errorf("not compareable")
	}
	if !cmpType(rhsT) {
		return nil, ast.Invalid, fmt.Errorf("not compareable")
	}

	switch op { //nolint:exhaustive
	case ast.AND, ast.OR:
		if lhsT != ast.Bool || rhsT != ast.Bool {
			return nil, ast.Invalid, fmt.Errorf("unsupported operand type(s) for %s: %s and %s",
				op, lhsT, rhsT)
		}
		if op == ast.AND {
			return cast.ToBool(lhs) && cast.ToBool(rhs), ast.Bool, nil
		} else {
			return cast.ToBool(lhs) || cast.ToBool(rhs), ast.Bool, nil
		}

	case ast.LT:
		dtype := typePromotion(lhsT, rhsT)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs) < cast.ToFloat64(rhs), ast.Bool, nil
		}
		return cast.ToInt(lhs) < cast.ToInt(rhs), ast.Bool, nil
	case ast.LTE:
		dtype := typePromotion(lhsT, rhsT)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs) <= cast.ToFloat64(rhs), ast.Bool, nil
		}
		return cast.ToInt(lhs) <= cast.ToInt(rhs), ast.Bool, nil
	case ast.GT:
		dtype := typePromotion(lhsT, rhsT)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs) > cast.ToFloat64(rhs), ast.Bool, nil
		}
		return cast.ToInt(lhs) > cast.ToInt(rhs), ast.Bool, nil
	case ast.GTE:
		dtype := typePromotion(lhsT, rhsT)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs) >= cast.ToFloat64(rhs), ast.Bool, nil
		}
		return cast.ToInt(lhs) >= cast.ToInt(rhs), ast.Bool, nil
	default:
		return nil, ast.Invalid, fmt.Errorf("op error")
	}
}

func cmpType(dtype ast.DType) bool {
	switch dtype { //nolint:exhaustive
	case ast.Int, ast.Float, ast.Bool:
		return true
	}
	return false
}

func assign2arithOp(op ast.Op) (ast.Op, bool) {
	switch op {
	case ast.ADDEQ:
		return ast.ADD, true
	case ast.SUBEQ:
		return ast.SUB, true
	case ast.MULEQ:
		return ast.MUL, true
	case ast.DIVEQ:
		return ast.DIV, true
	case ast.MODEQ:
		return ast.MOD, true
	default:
		return "", false
	}
}

func arithType(dtype ast.DType) bool {
	switch dtype { //nolint:exhaustive
	case ast.Int, ast.Float, ast.Bool, ast.String:
		return true
	default:
		return false
	}
}

func arithOpInt(l int64, r int64, op ast.Op) (int64, ast.DType, error) {
	switch op { //nolint:exhaustive
	case ast.ADD:
		return l + r, ast.Int, nil
	case ast.SUB:
		return l - r, ast.Int, nil
	case ast.MUL:
		return l * r, ast.Int, nil
	case ast.DIV:
		if r == 0 {
			return 0, ast.Invalid, fmt.Errorf("integer division by zero")
		}
		return l / r, ast.Int, nil
	case ast.MOD:
		if r == 0 {
			return 0, ast.Invalid, fmt.Errorf("integer modulo by zero")
		}
		return l % r, ast.Int, nil
	default:
		return 0, ast.Invalid, fmt.Errorf("unsupported op: %s", op)
	}
}

func arithOpFloat(l float64, r float64, op ast.Op) (float64, ast.DType, error) {
	switch op { //nolint:exhaustive
	case ast.ADD:
		return l + r, ast.Float, nil
	case ast.SUB:
		return l - r, ast.Float, nil
	case ast.MUL:
		return l * r, ast.Float, nil
	case ast.DIV:
		if r == 0 {
			return 0, ast.Invalid, fmt.Errorf("float division by zero")
		}
		return l / r, ast.Float, nil
	default:
		return 0, ast.Invalid, fmt.Errorf("float does not support modulo operations")
	}
}
