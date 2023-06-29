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
	"github.com/spf13/cast"
)

type (
	FuncCheck func(*Context, *ast.CallExpr) *errchain.PlError
	FuncCall  func(*Context, *ast.CallExpr) *errchain.PlError
)

func RunScriptWithoutMapIn(proc *Script, data InputWithoutMap, signal Signal) *errchain.PlError {
	if proc == nil {
		return nil
	}

	ctx := GetContext()
	defer PutContext(ctx)

	ctx = InitCtxWithoutMap(ctx, data, proc.FuncCall, proc.CallRef, signal, proc.Name, proc.Content)
	return RunStmts(ctx, proc.Ast)
}

func RunScriptWithRMapIn(proc *Script, data InputWithRMap, signal Signal) *errchain.PlError {
	if proc == nil {
		return nil
	}

	ctx := GetContext()
	defer PutContext(ctx)

	ctx = InitCtxWithRMap(ctx, data, proc.FuncCall, proc.CallRef, signal, proc.Name, proc.Content)
	return RunStmts(ctx, proc.Ast)
}

func RefRunScript(ctx *Context, proc *Script) *errchain.PlError {
	if proc == nil {
		return nil
	}

	newctx := GetContext()
	defer PutContext(newctx)

	switch ctx.inType {
	case InRMap:
		InitCtxWithRMap(newctx, ctx.inRMap, proc.FuncCall, proc.CallRef, ctx.signal, proc.Name, proc.Content)
	case InWithoutMap:
		InitCtxWithoutMap(newctx, ctx.inWithoutMap, proc.FuncCall, proc.CallRef, ctx.signal, proc.Name, proc.Content)
	default:
		// TODO
		return nil
	}

	return RunStmts(newctx, proc.Ast)
}

func CheckScript(proc *Script, funcsCheck map[string]FuncCheck) *errchain.PlError {
	ctx := GetContext()
	defer PutContext(ctx)
	InitCtxForCheck(ctx, proc.FuncCall, funcsCheck, proc.Name, proc.Content)
	if err := RunStmtsCheck(ctx, &ContextCheck{}, proc.Ast); err != nil {
		return err
	}

	proc.CallRef = ctx.callRCef
	return nil
}

func RunStmts(ctx *Context, nodes ast.Stmts) *errchain.PlError {
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

func RunIfElseStmt(ctx *Context, stmt *ast.IfelseStmt) (any, ast.DType, *errchain.PlError) {
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

func RunForStmt(ctx *Context, stmt *ast.ForStmt) (any, ast.DType, *errchain.PlError) {
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

func RunForInStmt(ctx *Context, stmt *ast.ForInStmt) (any, ast.DType, *errchain.PlError) {
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
			_ = ctx.SetVarb(stmt.Varb.Identifier.Name, char, ast.String)
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
			_ = ctx.SetVarb(stmt.Varb.Identifier.Name, x, ast.String)
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
			_ = ctx.SetVarb(stmt.Varb.Identifier.Name, x, dtype)
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

func forbreak(ctx *Context) bool {
	if ctx.loopBreak {
		ctx.loopBreak = false
		return true
	}
	return false
}

func forcontinue(ctx *Context) {
	if ctx.loopContinue {
		ctx.loopContinue = false
	}
}

func RunBreakStmt(ctx *Context, stmt *ast.BreakStmt) (any, ast.DType, *errchain.PlError) {
	ctx.loopBreak = true
	return nil, ast.Void, nil
}

func RunContinueStmt(ctx *Context, stmt *ast.ContinueStmt) (any, ast.DType, *errchain.PlError) {
	ctx.loopContinue = true
	return nil, ast.Void, nil
}

// RunStmt for all expr.
func RunStmt(ctx *Context, node *ast.Node) (any, ast.DType, *errchain.PlError) {
	// TODO
	// 存在个别 node 为 nil 的情况
	if node == nil {
		return nil, ast.Void, nil
	}
	switch node.NodeType { //nolint:exhaustive
	case ast.TypeParenExpr:
		return RunParenExpr(ctx, node.ParenExpr)
	case ast.TypeArithmeticExpr:
		return RunArithmeticExpr(ctx, node.ArithmeticExpr)
	case ast.TypeConditionalExpr:
		return RunConditionExpr(ctx, node.ConditionalExpr)
	case ast.TypeAssignmentExpr:
		return RunAssignmentExpr(ctx, node.AssignmentExpr)
	case ast.TypeCallExpr:
		return RunCallExpr(ctx, node.CallExpr)
	case ast.TypeInExpr:
		return RunInExpr(ctx, node.InExpr)
	case ast.TypeListInitExpr:
		return RunListInitExpr(ctx, node.ListInitExpr)
	case ast.TypeIdentifier:
		if v, err := ctx.GetKey(node.Identifier.Name); err != nil {
			return nil, ast.Nil, nil
		} else {
			return v.Value, v.DType, nil
		}
	case ast.TypeMapInitExpr:
		return RunMapInitExpr(ctx, node.MapInitExpr)
	// use for map, slice and array
	case ast.TypeIndexExpr:
		return RunIndexExprGet(ctx, node.IndexExpr)

	// TODO
	case ast.TypeAttrExpr:
		return nil, ast.Void, nil

	case ast.TypeBoolLiteral:
		return node.BoolLiteral.Val, ast.Bool, nil

	case ast.TypeIntegerLiteral:
		return node.IntegerLiteral.Val, ast.Int, nil

	case ast.TypeFloatLiteral:
		return node.FloatLiteral.Val, ast.Float, nil

	case ast.TypeStringLiteral:
		return node.StringLiteral.Val, ast.String, nil

	case ast.TypeNilLiteral:
		return nil, ast.Nil, nil

	case ast.TypeIfelseStmt:
		return RunIfElseStmt(ctx, node.IfelseStmt)
	case ast.TypeForStmt:
		return RunForStmt(ctx, node.ForStmt)
	case ast.TypeForInStmt:
		return RunForInStmt(ctx, node.ForInStmt)
	case ast.TypeBreakStmt:
		return RunBreakStmt(ctx, node.BreakStmt)
	case ast.TypeContinueStmt:
		return RunContinueStmt(ctx, node.ContinueStmt)
	default:
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported ast node: %s", reflect.TypeOf(node).String()), node.StartPos())
	}
}

func RunListInitExpr(ctx *Context, expr *ast.ListInitExpr) (any, ast.DType, *errchain.PlError) {
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

func RunMapInitExpr(ctx *Context, expr *ast.MapInitExpr) (any, ast.DType, *errchain.PlError) {
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

func RunIndexExprGet(ctx *Context, expr *ast.IndexExpr) (any, ast.DType, *errchain.PlError) {
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

func searchListAndMap(ctx *Context, obj any, index []*ast.Node) (any, ast.DType, *errchain.PlError) {
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

func RunParenExpr(ctx *Context, expr *ast.ParenExpr) (any, ast.DType, *errchain.PlError) {
	return RunStmt(ctx, expr.Param)
}

// BinarayExpr

func RunInExpr(ctx *Context, expr *ast.InExpr) (any, ast.DType, *errchain.PlError) {
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

func RunConditionExpr(ctx *Context, expr *ast.ConditionalExpr) (any, ast.DType, *errchain.PlError) {
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

func RunArithmeticExpr(ctx *Context, expr *ast.ArithmeticExpr) (any, ast.DType, *errchain.PlError) {
	// 允许字符串通过操作符 '+' 进行拼接

	lhsVal, lhsValType, errOpInt := RunStmt(ctx, expr.LHS)
	if errOpInt != nil {
		return nil, ast.Invalid, errOpInt
	}
	if !arithType(lhsValType) {
		return nil, ast.Invalid, NewRunError(ctx, fmt.Sprintf(
			"unsupported lhs data type: %s", lhsValType), expr.OpPos)
	}

	rhsVal, rhsValType, errOpInt := RunStmt(ctx, expr.RHS)
	if errOpInt != nil {
		return nil, ast.Invalid, errOpInt
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

func RunAssignmentExpr(ctx *Context, expr *ast.AssignmentExpr) (any, ast.DType, *errchain.PlError) {
	v, dtype, err := RunStmt(ctx, expr.RHS)
	if err != nil {
		return nil, ast.Invalid, err
	}

	switch expr.LHS.NodeType { //nolint:exhaustive
	case ast.TypeIdentifier:
		_ = ctx.SetVarb(expr.LHS.Identifier.Name, v, dtype)
		return v, dtype, nil
	case ast.TypeIndexExpr:
		varb, err := ctx.GetKey(expr.LHS.IndexExpr.Obj.Name)
		if err != nil {
			return nil, ast.Invalid, NewRunError(ctx, err.Error(), expr.LHS.IndexExpr.Obj.Start)
		}
		return changeListOrMapValue(ctx, varb.Value, expr.LHS.IndexExpr.Index,
			v, dtype)
	default:
		return nil, ast.Void, nil
	}
}

func changeListOrMapValue(ctx *Context, obj any, index []*ast.Node, val any, dtype ast.DType) (any, ast.DType, *errchain.PlError) {
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

func RunCallExpr(ctx *Context, expr *ast.CallExpr) (any, ast.DType, *errchain.PlError) {
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
		return 0, ast.Invalid, fmt.Errorf("unsupported op: %s", op)
	}
}
