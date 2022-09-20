// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package runtime provide a runtime for the pipeline
package runtime

import (
	"fmt"
	"reflect"
	"time"

	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
)

type PlPanic error

type (
	FuncCheck func(*Context, *ast.CallExpr) error
	FuncCall  func(*Context, *ast.CallExpr) PlPanic
)

func RunScript(proc *Script, measurement string,
	tags map[string]string, fields map[string]any, tn time.Time) (
	string, map[string]string, map[string]any, time.Time, bool, error,
) {
	if proc == nil {
		return "", nil, nil, tn, false, fmt.Errorf("vaule of script is nil")
	}
	ctx := GetContext()
	defer PutContext(ctx)

	pt := GetPoint()
	defer PutPoint(pt)

	pt = InitPt(pt, measurement, tags, fields, tn)

	ctx = InitCtx(ctx, pt, proc.FuncCall, proc.CallRef)
	RunStmts(ctx, proc.Ast)

	pt.KeyTime2Time()
	return pt.Measurement, pt.Tags, pt.Fields, pt.Time, ctx.PtDropped(), nil
}

func RunScriptWithCtx(ctx *Context, proc *Script) error {
	pt, _ := ctx.Point()

	newctx := GetContext()
	defer PutContext(newctx)

	newctx = InitCtx(newctx, pt, proc.FuncCall, proc.CallRef)
	RunStmts(newctx, proc.Ast)
	return nil
}

func CheckScript(proc *Script, funcsCheck map[string]FuncCheck) error {
	ctx := GetContext()
	defer PutContext(ctx)
	InitCtxForCheck(ctx, proc.FuncCall, funcsCheck)
	if err := RunStmtsCheck(ctx, &ContextCheck{}, proc.Ast); err != nil {
		return err
	}

	proc.CallRef = ctx.callRCef
	return nil
}

func RunStmts(ctx *Context, nodes ast.Stmts) {
	for _, node := range nodes {
		_, _, _ = RunStmt(ctx, node)
		if ctx.StmtRetrun() {
			return
		}
	}
}

func RunIfElseStmt(ctx *Context, stmt *ast.IfelseStmt) (any, ast.DType, error) {
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

		// run if or elif stmt
		ctx.StackEnterNew()

		RunStmts(ctx, ifstmt.Stmts)

		ctx.StackExitCur()
		return nil, ast.Void, nil
	}

	// run else stmt
	ctx.StackEnterNew()
	RunStmts(ctx, stmt.Else)
	ctx.StackExitCur()

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

func RunForStmt(ctx *Context, stmt *ast.ForStmt) (any, ast.DType, error) {
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

		// for body
		ctx.StackEnterNew()
		RunStmts(ctx, stmt.Body)
		ctx.StackExitCur()

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

func RunForInStmt(ctx *Context, stmt *ast.ForInStmt) (any, ast.DType, error) {
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
			return nil, ast.Invalid, fmt.Errorf("inner type error")
		}
		for _, x := range iter {
			char := string(x)
			if stmt.Varb.NodeType != ast.TypeIdentifier {
				return nil, ast.Invalid, err
			}
			_ = ctx.SetVarb(stmt.Varb.Identifier.Name, char, ast.String)
			RunStmts(ctx, stmt.Body)
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
			return nil, ast.Invalid, fmt.Errorf("inner type error")
		}
		for x := range iter {
			ctx.stackCur.Clear()
			_ = ctx.SetVarb(stmt.Varb.Identifier.Name, x, ast.String)
			RunStmts(ctx, stmt.Body)
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
			return nil, ast.Invalid, fmt.Errorf("inner type error")
		}
		for _, x := range iter {
			ctx.stackCur.Clear()
			x, dtype := ast.DectDataType(x)
			if dtype == ast.Invalid {
				return nil, ast.Invalid, fmt.Errorf("inner type error")
			}
			_ = ctx.SetVarb(stmt.Varb.Identifier.Name, x, dtype)
			RunStmts(ctx, stmt.Body)
			if forbreak(ctx) {
				break
			}
			forcontinue(ctx)
			if ctx.StmtRetrun() {
				break
			}
		}
	default:
		return nil, ast.Invalid, fmt.Errorf("unsupported type: %s, not iter value", dtype)
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

func RunBreakStmt(ctx *Context, stmt *ast.BreakStmt) (any, ast.DType, error) {
	ctx.loopBreak = true
	return nil, ast.Void, nil
}

func RunContinueStmt(ctx *Context, stmt *ast.ContinueStmt) (any, ast.DType, error) {
	ctx.loopContinue = true
	return nil, ast.Void, nil
}

// RunStmt for all expr.
func RunStmt(ctx *Context, node *ast.Node) (any, ast.DType, error) {
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
	case ast.TypeNumberLiteral:
		if node.NumberLiteral.IsInt {
			return node.NumberLiteral.Int, ast.Int, nil
		} else {
			return node.NumberLiteral.Float, ast.Float, nil
		}
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
		return nil, ast.Invalid, fmt.Errorf("unsupported ast node: %s", reflect.TypeOf(node).String())
	}
}

func RunListInitExpr(ctx *Context, expr *ast.ListInitExpr) (any, ast.DType, error) {
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

func RunMapInitExpr(ctx *Context, expr *ast.MapInitExpr) (any, ast.DType, error) {
	ret := map[string]any{}

	for _, v := range expr.KeyValeList {
		k, keyType, err := RunStmt(ctx, v[0])
		if err != nil {
			return nil, ast.Invalid, err
		}

		key, ok := k.(string)
		if !ok {
			return nil, ast.Invalid, fmt.Errorf("unsupported data type: %s", keyType)
		}
		value, valueType, err := RunStmt(ctx, v[1])
		if err != nil {
			return nil, ast.Invalid, err
		}
		switch valueType { //nolint:exhaustive
		case ast.String, ast.Bool, ast.Float, ast.Int,
			ast.Nil, ast.List, ast.Map:
		default:
			return nil, ast.Invalid, fmt.Errorf("unsupported data type: %s", keyType)
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

func RunIndexExprGet(ctx *Context, expr *ast.IndexExpr) (any, ast.DType, error) {
	key := expr.Obj.Name

	varb, err := ctx.GetKey(key)
	if err != nil {
		return nil, ast.Invalid, err
	}
	switch varb.DType { //nolint:exhaustive
	case ast.List:
		switch varb.Value.(type) {
		case []any:
		default:
			return nil, ast.Invalid, fmt.Errorf("unsupported type: %v", reflect.TypeOf(varb.Value))
		}
	case ast.Map:
		switch varb.Value.(type) {
		case map[string]any: // for json map
		default:
			return nil, ast.Invalid, fmt.Errorf("unsupported type: %v", reflect.TypeOf(varb.Value))
		}
	default:
		return nil, ast.Invalid, fmt.Errorf("unindexable type: %s", varb.DType)
	}

	return searchListAndMap(ctx, varb.Value, expr.Index)
}

func searchListAndMap(ctx *Context, obj any, index []*ast.Node) (any, ast.DType, error) {
	cur := obj

	for _, i := range index {
		key, keyType, err := RunStmt(ctx, i)
		if err != nil {
			return nil, ast.Invalid, err
		}
		switch curVal := cur.(type) {
		case map[string]any:
			if keyType != ast.String {
				return nil, ast.Invalid, fmt.Errorf("key type is not string")
			}
			var ok bool
			cur, ok = curVal[key.(string)]
			if !ok {
				return nil, ast.Invalid, fmt.Errorf("key not found")
			}
		case []any:
			if keyType != ast.Int {
				return nil, ast.Invalid, fmt.Errorf("key type is not int")
			}
			keyInt := cast.ToInt(key)

			// 反转负数
			if keyInt < 0 {
				keyInt = len(curVal) + keyInt
			}

			if keyInt < 0 || keyInt >= len(curVal) {
				return nil, ast.Invalid, fmt.Errorf("list index out of range")
			}
			cur = curVal[keyInt]
		default:
			return nil, ast.Invalid, fmt.Errorf("not found")
		}
	}
	var dtype ast.DType
	cur, dtype = ast.DectDataType(cur)
	return cur, dtype, nil
}

func RunParenExpr(ctx *Context, expr *ast.ParenExpr) (any, ast.DType, error) {
	return RunStmt(ctx, expr.Param)
}

// BinarayExpr

func RunConditionExpr(ctx *Context, expr *ast.ConditionalExpr) (any, ast.DType, error) {
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
		return nil, ast.Invalid, fmt.Errorf("err %w", err)
	} else {
		return val, dtype, nil
	}
}

func RunArithmeticExpr(ctx *Context, expr *ast.ArithmeticExpr) (any, ast.DType, error) {
	// 允许字符串通过操作符 '+' 进行拼接

	lhsVal, lhsValType, err := RunStmt(ctx, expr.LHS)
	if err != nil {
		return nil, ast.Invalid, err
	}
	if !arithType(lhsValType) {
		return nil, ast.Invalid, fmt.Errorf("unsupported lhs data type: %s", lhsValType)
	}

	rhsVal, rhsValType, err := RunStmt(ctx, expr.RHS)
	if err != nil {
		return nil, ast.Invalid, err
	}

	if !arithType(lhsValType) {
		return nil, ast.Invalid, fmt.Errorf("unsupported rhs data type: %s", lhsValType)
	}

	// string
	if lhsValType == ast.String || rhsValType == ast.String {
		if expr.Op != ast.ADD {
			return nil, ast.Invalid, fmt.Errorf(
				"unsupported operand type(s) for %s: %s and %s",
				expr.Op, lhsValType, rhsValType)
		}
		if lhsValType == ast.String && rhsValType == ast.String {
			return cast.ToString(lhsVal) + cast.ToString(rhsVal), ast.String, nil
		} else {
			return nil, ast.Invalid, fmt.Errorf(
				"unsupported operand type(s) for %s: %s and %s",
				expr.Op, lhsValType, rhsValType)
		}
	}

	// float
	if lhsValType == ast.Float || rhsValType == ast.Float {
		return arithOpFloat(cast.ToFloat64(lhsVal), cast.ToFloat64(rhsVal), expr.Op)
	}

	// bool or int
	return arithOpInt(cast.ToInt64(lhsVal), cast.ToInt64(rhsVal), expr.Op)
}

func RunAssignmentExpr(ctx *Context, expr *ast.AssignmentExpr) (any, ast.DType, error) {
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
			return nil, ast.Invalid, err
		}
		return changeListOrMapValue(ctx, varb.Value, expr.LHS.IndexExpr.Index,
			v, dtype)
	default:
		return nil, ast.Void, nil
	}
}

func changeListOrMapValue(ctx *Context, obj any, index []*ast.Node, val any, dtype ast.DType) (any, ast.DType, error) {
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
				return nil, ast.Invalid, fmt.Errorf("key type is not string")
			}
			if idx+1 == lenIdx {
				curVal[key.(string)] = val
				return val, dtype, nil
			}

			var ok bool
			cur, ok = curVal[key.(string)]
			if !ok {
				return nil, ast.Invalid, fmt.Errorf("key not found")
			}
		case []any:
			if keyType != ast.Int {
				return nil, ast.Invalid, fmt.Errorf("key type is not int")
			}
			keyInt := cast.ToInt(key)

			// 反转负数
			if keyInt < 0 {
				keyInt = len(curVal) + keyInt
			}

			if keyInt < 0 || keyInt >= len(curVal) {
				return nil, ast.Invalid, fmt.Errorf("list index out of range")
			}

			if idx+1 == lenIdx {
				curVal[keyInt] = val
				return val, dtype, nil
			}

			cur = curVal[keyInt]
		default:
			return nil, ast.Invalid, fmt.Errorf("not found")
		}
	}
	return nil, ast.Nil, nil
}

func RunCallExpr(ctx *Context, expr *ast.CallExpr) (any, ast.DType, error) {
	defer ctx.Regs.Reset()
	if funcCall, ok := ctx.GetFuncCall(expr.Name); ok {
		if err := funcCall(ctx, expr); err != nil {
			return nil, ast.Invalid, err
		}
		if ctx.Regs.Count() > 0 {
			return ctx.Regs.Get(RegR0)
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
