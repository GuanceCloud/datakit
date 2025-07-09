package runtimev2

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
	"github.com/spf13/cast"
)

func RunStmts(ctx *Task, nodes ast.Stmts) *errchain.PlError {
	for _, node := range nodes {
		if err := RunExpr(ctx, node); err != nil {
			ctx.procExit = true
			return err
		}

		if ctx.StmtRetrun() {
			return nil
		}
	}
	return nil
}

func RunIfElseStmt(ctx *Task, stmt *ast.IfelseStmt) *errchain.PlError {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	// check if or elif condition
	for _, ifstmt := range stmt.IfList {
		// check condition
		err := RunExpr(ctx, ifstmt.Condition)
		if err != nil {
			return err
		}
		val, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), ifstmt.Condition.StartPos())
		}
		if !condTrue(val) {
			continue
		}

		if ifstmt.Block != nil {
			// run if or elif stmt
			ctx.StackEnterNew()

			if err := RunStmts(ctx, ifstmt.Block.Stmts); err != nil {
				return err
			}

			ctx.StackExitCur()
		}

		return nil
	}

	if stmt.Else != nil {
		// run else stmt
		ctx.StackEnterNew()
		if err := RunStmts(ctx, stmt.Else.Stmts); err != nil {
			return err
		}
		ctx.StackExitCur()
	}

	return nil
}

func condTrue(val V) bool {
	switch val.T { //nolint:exhaustive
	case ast.String:
		if cast.ToString(val.V) == "" {
			return false
		}
	case ast.Bool:
		return cast.ToBool(val.V)
	case ast.Int:
		if cast.ToInt64(val.V) == 0 {
			return false
		}
	case ast.Float:
		if cast.ToFloat64(val.V) == 0 {
			return false
		}
	case ast.List:
		if len(cast.ToSlice(val.V)) == 0 {
			return false
		}
	case ast.Map:
		switch v := val.V.(type) {
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

func RunForStmt(ctx *Task, stmt *ast.ForStmt) *errchain.PlError {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	// for init
	if stmt.Init != nil {
		err := RunExpr(ctx, stmt.Init)
		if err != nil {
			return err
		}
	}

	for {
		if stmt.Cond != nil {
			err := RunExpr(ctx, stmt.Cond)
			if err != nil {
				return err
			}
			val, errReg := ctx.Regs.GetRet()
			if errReg != nil {
				return NewRunError(ctx, errReg.Error(), stmt.Cond.StartPos())
			}
			if !condTrue(val) {
				break
			}
		}

		if stmt.Body != nil {
			// for body
			ctx.StackEnterNew()
			if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
				return err
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
			err := RunExpr(ctx, stmt.Loop)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func RunForInStmt(ctx *Task, stmt *ast.ForInStmt) *errchain.PlError {
	ctx.StackEnterNew()
	defer ctx.StackExitCur()

	err := RunExpr(ctx, stmt.Iter)
	if err != nil {
		return err
	}
	iter, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), stmt.Iter.StartPos())
	}

	ctx.StackEnterNew()
	defer ctx.StackExitCur()
	switch iter.T { //nolint:exhaustive
	case ast.String:
		iter, ok := iter.V.(string)
		if !ok {
			return NewRunError(ctx,
				"inner type error", stmt.Iter.StartPos())
		}
		for _, x := range iter {
			char := string(x)
			if stmt.Varb.NodeType != ast.TypeIdentifier {
				return err
			}
			ctx.SetVarb(stmt.Varb.Identifier().Name, V{char, ast.String})
			if stmt.Body != nil {
				if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
					return err
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
		iter, ok := iter.V.(map[string]any)
		if !ok {
			return NewRunError(ctx,
				"inner type error", stmt.Iter.StartPos())
		}
		for x := range iter {
			ctx.stackCur.Clear()
			ctx.SetVarb(stmt.Varb.Identifier().Name, V{x, ast.String})
			if stmt.Body != nil {
				if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
					return err
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
		iter, ok := iter.V.([]any)
		if !ok {
			return NewRunError(ctx,
				"inner type error", stmt.Iter.StartPos())
		}
		for _, x := range iter {
			ctx.stackCur.Clear()
			x, dtype := ast.DectDataType(x)
			if dtype == ast.Invalid {
				return NewRunError(ctx,
					"inner type error", stmt.Iter.StartPos())
			}
			ctx.SetVarb(stmt.Varb.Identifier().Name, V{x, dtype})
			if stmt.Body != nil {
				if err := RunStmts(ctx, stmt.Body.Stmts); err != nil {
					return err
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
		return NewRunError(ctx, fmt.Sprintf(
			"unsupported type: %s, not iter value", iter.T), stmt.Iter.StartPos())
	}

	return nil
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

func RunBreakStmt(ctx *Task, stmt *ast.BreakStmt) *errchain.PlError {
	ctx.loopBreak = true
	return nil
}

func RunContinueStmt(ctx *Task, stmt *ast.ContinueStmt) *errchain.PlError {
	ctx.loopContinue = true
	return nil
}

// RunExpr for all expr.
func RunExpr(ctx *Task, node *ast.Node) *errchain.PlError {
	// TODO
	// 存在个别 node 为 nil 的情况
	if node == nil {
		return nil
	}
	switch node.NodeType {
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
	case ast.TypeSliceExpr:
		return RunSliceExpr(ctx, node.SliceExpr())
	case ast.TypeInExpr:
		return RunInExpr(ctx, node.InExpr())
	case ast.TypeListLiteral:
		return RunListInitExpr(ctx, node.ListLiteral())
	case ast.TypeIdentifier:
		if v, err := ctx.GetKey(node.Identifier().Name); err != nil {
			return NewRunError(ctx, fmt.Sprintf("name `%s` is not defined",
				node.Identifier().Name), node.StartPos())
		} else {
			ctx.Regs.ReturnAppend(V{v.Value, v.DType})
			return nil
		}
	case ast.TypeMapLiteral:
		return RunMapInitExpr(ctx, node.MapLiteral())
	// use for map, slice and array
	case ast.TypeIndexExpr:
		return RunIndexExprGet(ctx, node.IndexExpr())

	// TODO
	case ast.TypeAttrExpr:
		return nil

	case ast.TypeBoolLiteral:
		ctx.Regs.ReturnAppend(V{node.BoolLiteral().Val, ast.Bool})
		return nil

	case ast.TypeIntegerLiteral:
		ctx.Regs.ReturnAppend(V{node.IntegerLiteral().Val, ast.Int})
		return nil

	case ast.TypeFloatLiteral:
		ctx.Regs.ReturnAppend(V{node.FloatLiteral().Val, ast.Float})
		return nil

	case ast.TypeStringLiteral:
		ctx.Regs.ReturnAppend(V{node.StringLiteral().Val, ast.String})
		return nil

	case ast.TypeNilLiteral:
		ctx.Regs.ReturnAppend(V{nil, ast.Nil})
		return nil

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
		return NewRunError(ctx, fmt.Sprintf(
			"unsupported ast node: %s", reflect.TypeOf(node).String()), node.StartPos())
	}
}

func RunUnaryExpr(ctx *Task, expr *ast.UnaryExpr) *errchain.PlError {
	switch expr.Op {
	case ast.SUB, ast.ADD:
		err := RunExpr(ctx, expr.RHS)
		if err != nil {
			return err
		}
		v, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), expr.RHS.StartPos())
		}
		switch v.T {
		case ast.Bool:
			val, _ := v.V.(bool)
			if expr.Op == ast.SUB {
				if val {
					ctx.Regs.ReturnAppend(V{int64(-1), ast.Int})
					return nil
				} else {
					ctx.Regs.ReturnAppend(V{int64(0), ast.Int})
					return nil
				}
			} else {
				if val {
					ctx.Regs.ReturnAppend(V{int64(1), ast.Int})
					return nil
				} else {
					ctx.Regs.ReturnAppend(V{int64(0), ast.Int})
					return nil
				}
			}
		case ast.Float:
			val, _ := v.V.(float64)
			if expr.Op == ast.SUB {
				ctx.Regs.ReturnAppend(V{-val, ast.Float})
				return nil
			} else {
				ctx.Regs.ReturnAppend(V{val, ast.Float})
				return nil
			}
		case ast.Int:
			val, _ := v.V.(int64)
			if expr.Op == ast.SUB {
				ctx.Regs.ReturnAppend(V{-val, ast.Int})
				return nil
			} else {
				ctx.Regs.ReturnAppend(V{val, ast.Int})
				return nil
			}
		default:
			return NewRunError(ctx,
				fmt.Sprintf("unsuppored operand type for unary op %s: %s",
					expr.Op, reflect.TypeOf(expr).String()), expr.OpPos)
		}

	case ast.NOT:
		err := RunExpr(ctx, expr.RHS)
		if err != nil {
			return err
		}

		v, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), expr.RHS.StartPos())
		}

		if v.V == nil {
			ctx.Regs.ReturnAppend(V{true, ast.Bool})
			return nil
		}

		switch val := v.V.(type) {
		case bool:
			ctx.Regs.ReturnAppend(V{!val, ast.Bool})
			return nil
		case float64:
			if val == 0 {
				ctx.Regs.ReturnAppend(V{true, ast.Bool})
				return nil
			} else {
				ctx.Regs.ReturnAppend(V{false, ast.Bool})
				return nil
			}
		case int64:
			if val == 0 {
				ctx.Regs.ReturnAppend(V{true, ast.Bool})
				return nil
			} else {
				ctx.Regs.ReturnAppend(V{false, ast.Bool})
				return nil
			}
		case string:
			if len(val) == 0 {
				ctx.Regs.ReturnAppend(V{true, ast.Bool})
				return nil
			} else {
				ctx.Regs.ReturnAppend(V{false, ast.Bool})
				return nil
			}
		case map[string]any:
			if len(val) == 0 {
				ctx.Regs.ReturnAppend(V{true, ast.Bool})
				return nil
			} else {
				ctx.Regs.ReturnAppend(V{false, ast.Bool})
				return nil
			}
		case []any:
			if len(val) == 0 {
				ctx.Regs.ReturnAppend(V{true, ast.Bool})
				return nil
			} else {
				ctx.Regs.ReturnAppend(V{false, ast.Bool})
				return nil
			}

		default:
			return NewRunError(ctx,
				fmt.Sprintf("unsuppored operand type for unary op %s: %s",
					expr.Op, reflect.TypeOf(expr).String()), expr.OpPos)
		}
	default:
		return NewRunError(ctx,
			fmt.Sprintf("unsupported op for unary expr: %s", expr.Op), expr.OpPos)
	}
}

func RunListInitExpr(ctx *Task, expr *ast.ListLiteral) *errchain.PlError {
	ret := []any{}
	for _, v := range expr.List {
		err := RunExpr(ctx, v)
		if err != nil {
			return err
		}
		val, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), v.StartPos())
		}
		ret = append(ret, val.V)
	}
	ctx.Regs.ReturnAppend(V{ret, ast.List})
	return nil
}

func RunMapInitExpr(ctx *Task, expr *ast.MapLiteral) *errchain.PlError {
	ret := map[string]any{}

	for _, v := range expr.KeyValeList {
		err := RunExpr(ctx, v[0])
		if err != nil {
			return err
		}
		k, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), v[0].StartPos())
		}

		key, ok := k.V.(string)
		if !ok {
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported key data type: %s", k.T), v[0].StartPos())
		}
		err = RunExpr(ctx, v[1])
		if err != nil {
			return err
		}
		value, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), v[1].StartPos())
		}
		switch value.T { //nolint:exhaustive
		case ast.String, ast.Bool, ast.Float, ast.Int,
			ast.Nil, ast.List, ast.Map:
		default:
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported value data type: %s", value.T), v[1].StartPos())
		}
		ret[key] = value.V
	}

	ctx.Regs.ReturnAppend(V{ret, ast.Map})
	return nil
}

// func indexKeyType(dtype ast.DType) bool {
// 	switch dtype { //nolint:exhaustive
// 	case ast.Int, ast.String:
// 		return true
// 	default:
// 		return false
// 	}
// }

func RunIndexExprGet(ctx *Task, expr *ast.IndexExpr) *errchain.PlError {
	key := expr.Obj.Name

	varb, err := ctx.GetKey(key)
	if err != nil {
		return NewRunError(ctx, err.Error(), expr.Obj.Start)
	}
	switch varb.DType { //nolint:exhaustive
	case ast.List:
		switch varb.Value.(type) {
		case []any:
		default:
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported type: %v", reflect.TypeOf(varb.Value)), expr.Obj.Start)
		}
	case ast.Map:
		switch varb.Value.(type) {
		case map[string]any: // for json map
		default:
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported type: %v", reflect.TypeOf(varb.Value)), expr.Obj.Start)
		}
	default:
		return NewRunError(ctx, fmt.Sprintf(
			"unindexable type: %s", varb.DType), expr.Obj.Start)
	}

	return searchListAndMap(ctx, varb.Value, expr.Index)
}

func searchListAndMap(ctx *Task, obj any, index []*ast.Node) *errchain.PlError {
	cur := obj

	for _, i := range index {
		err := RunExpr(ctx, i)
		if err != nil {
			return err
		}
		key, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), i.StartPos())
		}
		switch curVal := cur.(type) {
		case map[string]any:
			if key.T != ast.String {
				return NewRunError(ctx,
					"key type is not string", i.StartPos())
			}
			var ok bool
			cur, ok = curVal[key.V.(string)]
			if !ok {
				ctx.Regs.ReturnAppend(V{nil, ast.Nil})
				return nil
			}
		case []any:
			if key.T != ast.Int {
				return NewRunError(ctx,
					"key type is not int", i.StartPos())
			}
			keyInt := cast.ToInt(key.V)

			// 反转负数
			if keyInt < 0 {
				keyInt = len(curVal) + keyInt
			}

			if keyInt < 0 || keyInt >= len(curVal) {
				return NewRunError(ctx,
					"list index out of range", i.StartPos())
			}
			cur = curVal[keyInt]
		default:
			return NewRunError(ctx,
				"not found", i.StartPos())
		}
	}
	cur, dtype := ast.DectDataType(cur)
	ctx.Regs.ReturnAppend(V{cur, dtype})
	return nil
}

func RunParenExpr(ctx *Task, expr *ast.ParenExpr) *errchain.PlError {
	return RunExpr(ctx, expr.Param)
}

// BinarayExpr

func RunInExpr(ctx *Task, expr *ast.InExpr) *errchain.PlError {
	err := RunExpr(ctx, expr.LHS)
	if err != nil {
		return err
	}
	lhs, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), expr.LHS.StartPos())
	}

	err = RunExpr(ctx, expr.RHS)
	if err != nil {
		return err
	}
	rhs, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), expr.LHS.StartPos())
	}

	switch rhs.T {
	case ast.String:
		if lhs.T != ast.String {
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported lhs data type: %s", lhs.T), expr.OpPos)
		}
		if s, ok := lhs.V.(string); ok {
			if v, ok := rhs.V.(string); ok {
				ctx.Regs.ReturnAppend(V{strings.Contains(v, s), ast.Bool})
				return nil
			}
		}
		ctx.Regs.ReturnAppend(V{false, ast.Bool})
		return nil
	case ast.Map:
		if lhs.T != ast.String {
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported lhs data type: %s", lhs.T), expr.OpPos)
		}
		if s, ok := lhs.V.(string); ok {
			if v, ok := rhs.V.(map[string]any); ok {
				if _, ok := v[s]; ok {
					ctx.Regs.ReturnAppend(V{true, ast.Bool})
					return nil
				}
			}
		}
		ctx.Regs.ReturnAppend(V{false, ast.Bool})
		return nil
	case ast.List:
		if v, ok := rhs.V.([]any); ok {
			for _, elem := range v {
				if reflect.DeepEqual(lhs.V, elem) {
					ctx.Regs.ReturnAppend(V{true, ast.Bool})
					return nil
				}
			}
		}
		ctx.Regs.ReturnAppend(V{false, ast.Bool})
		return nil

	default:
		return NewRunError(ctx, fmt.Sprintf(
			"unsupported rhs data type: %s", rhs.T), expr.OpPos)
	}
}

func RunConditionExpr(ctx *Task, expr *ast.ConditionalExpr) *errchain.PlError {
	err := RunExpr(ctx, expr.LHS)
	if err != nil {
		return err
	}
	lhs, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), expr.LHS.StartPos())
	}

	if lhs.T == ast.Bool {
		switch expr.Op { //nolint:exhaustive
		case ast.OR:
			if cast.ToBool(lhs.V) {
				ctx.Regs.ReturnAppend(V{true, ast.Bool})
				return nil
			}
		case ast.AND:
			if !cast.ToBool(lhs.V) {
				ctx.Regs.ReturnAppend(V{false, ast.Bool})
				return nil
			}
		}
	}

	err = RunExpr(ctx, expr.RHS)
	if err != nil {
		return err
	}
	rhs, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), expr.RHS.StartPos())
	}
	if val, dtype, err := condOp(lhs, rhs, expr.Op); err != nil {
		return NewRunError(ctx, err.Error(), expr.OpPos)
	} else {
		ctx.Regs.ReturnAppend(V{val, dtype})
		return nil
	}
}

func RunArithmeticExpr(ctx *Task, expr *ast.ArithmeticExpr) *errchain.PlError {
	// 允许字符串通过操作符 '+' 进行拼接

	errOpInt := RunExpr(ctx, expr.LHS)
	if errOpInt != nil {
		return errOpInt
	}

	lhsVal, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), expr.LHS.StartPos())
	}

	errOpInt = RunExpr(ctx, expr.RHS)
	if errOpInt != nil {
		return errOpInt
	}
	rhsVal, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), expr.RHS.StartPos())
	}

	if !arithType(lhsVal.T) {
		return NewRunError(ctx, fmt.Sprintf(
			"unsupported lhs data type: %s", lhsVal.T), expr.OpPos)
	}

	if !arithType(rhsVal.T) {
		return NewRunError(ctx, fmt.Sprintf(
			"unsupported rhs data type: %s", rhsVal.T), expr.OpPos)
	}

	// string
	if lhsVal.T == ast.String || rhsVal.T == ast.String {
		if expr.Op != ast.ADD {
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				expr.Op, lhsVal.T, rhsVal.T), expr.OpPos)
		}
		if lhsVal.T == ast.String && rhsVal.T == ast.String {
			ctx.Regs.ReturnAppend(V{cast.ToString(lhsVal.V) + cast.ToString(rhsVal.V), ast.String})
			return nil
		} else {
			return NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				expr.Op, lhsVal.T, rhsVal.T), expr.OpPos)
		}
	}

	// float
	if lhsVal.T == ast.Float || rhsVal.T == ast.Float {
		v, dtype, err := arithOpFloat(cast.ToFloat64(lhsVal.V), cast.ToFloat64(rhsVal.V), expr.Op)
		if err != nil {
			return NewRunError(ctx, err.Error(), expr.OpPos)
		}
		ctx.Regs.ReturnAppend(V{v, dtype})
		return nil
	}

	// bool or int

	v, dtype, errOp := arithOpInt(cast.ToInt64(lhsVal.V), cast.ToInt64(rhsVal.V), expr.Op)

	if errOp != nil {
		return NewRunError(ctx, errOp.Error(), expr.OpPos)
	}
	ctx.Regs.ReturnAppend(V{v, dtype})
	return nil
}

func runAssignArith(ctx *Task, l, r V, op ast.Op, pos token.LnColPos) (
	V, *errchain.PlError) {
	arithOp, ok := assign2arithOp(op)
	if !ok {
		return V{}, NewRunError(ctx,
			"unsupported op", pos)
	}

	if !arithType(l.T) {
		return V{}, NewRunError(ctx, fmt.Sprintf(
			"unsupported lhs data type: %s", l.T), pos)
	}

	if !arithType(r.T) {
		return V{}, NewRunError(ctx, fmt.Sprintf(
			"unsupported rhs data type: %s", r.T), pos)
	}

	// string
	if l.T == ast.String || r.T == ast.String {
		if arithOp != ast.ADD {
			return V{}, NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				op, l.T, r.T), pos)
		}
		if l.T == ast.String && r.T == ast.String {
			return V{cast.ToString(l.V) + cast.ToString(r.V), ast.String}, nil
		} else {
			return V{}, NewRunError(ctx, fmt.Sprintf(
				"unsupported operand type(s) for %s: %s and %s",
				op, l.T, r.T), pos)
		}
	}

	// float
	if l.T == ast.Float || r.T == ast.Float {
		v, dtype, err := arithOpFloat(cast.ToFloat64(l.V), cast.ToFloat64(r.V), arithOp)
		if err != nil {
			return r, NewRunError(ctx, err.Error(), pos)
		}
		return V{v, dtype}, nil
	}

	// bool or int

	v, dtype, errOp := arithOpInt(cast.ToInt64(l.V), cast.ToInt64(r.V), arithOp)

	if errOp != nil {
		return r, NewRunError(ctx, errOp.Error(), pos)
	}
	return V{v, dtype}, nil
}

// RunAssignmentExpr runs assignment expression, but actually it is a stmt
func RunAssignmentExpr(ctx *Task, expr *ast.AssignmentExpr) *errchain.PlError {
	lhsCount := len(expr.LHS)
	var vals []V

	for _, e := range expr.RHS {
		if err := RunExpr(ctx, e); err != nil {
			return err
		}
		switch ctx.Regs.Count() {
		case 1:
			if v, errReg := ctx.Regs.GetRet(); errReg != nil {
				return NewRunError(ctx, errReg.Error(), expr.RHS[0].StartPos())
			} else {
				vals = append(vals, v)
			}
		default:
			if v, errReg := ctx.Regs.GetMultiRet(); errReg != nil {
				return NewRunError(ctx, errReg.Error(), expr.RHS[0].StartPos())
			} else {
				if lhsCount == 1 {
					return NewRunError(ctx, "multiple return values", e.StartPos())
				}
				vals = append(vals, v...)
			}
		}
	}

	valsCount := len(vals)

	if lhsCount != valsCount {
		return NewRunError(ctx, "the number of left and right operands is not equal", expr.OpPos)
	}

	for i, e := range expr.LHS {
		switch expr.Op {
		case ast.SUBEQ,
			ast.ADDEQ,
			ast.MULEQ,
			ast.DIVEQ,
			ast.MODEQ:
			if valsCount != 1 {
				return NewRunError(ctx, "can be only one right value", expr.OpPos)
			}
			if err := RunExpr(ctx, e); err != nil {
				return err
			}
			lval, errReg := ctx.Regs.GetRet()
			if errReg != nil {
				return NewRunError(ctx, errReg.Error(), e.StartPos())
			}
			r, err := runAssignArith(ctx, lval, vals[i], expr.Op, expr.OpPos)
			if err != nil {
				return err
			}
			switch e.NodeType {
			case ast.TypeIdentifier:
				ctx.SetVarb(e.Identifier().Name, r)
			case ast.TypeIndexExpr:
				if varb, err := ctx.GetKey(e.IndexExpr().Obj.Name); err != nil {
					return NewRunError(ctx, err.Error(), e.IndexExpr().Obj.Start)
				} else {
					if err := changeListOrMapValue(ctx, varb.Value, e.IndexExpr().Index, r); err != nil {
						return err
					}
				}
			default:
				return NewRunError(ctx, fmt.Sprintf(
					"unsupported lhs type: %s", e.NodeType), e.StartPos())
			}
		case ast.EQ:
			switch e.NodeType {
			case ast.TypeIdentifier:
				ctx.SetVarb(e.Identifier().Name, vals[i])
			case ast.TypeIndexExpr:
				if varb, err := ctx.GetKey(e.IndexExpr().Obj.Name); err != nil {
					return NewRunError(ctx, err.Error(), e.IndexExpr().Obj.Start)
				} else {
					if err := changeListOrMapValue(ctx, varb.Value, e.IndexExpr().Index, vals[i]); err != nil {
						return err
					}
				}
			default:
				return NewRunError(ctx, fmt.Sprintf(
					"unsupported lhs type: %s", e.NodeType), e.StartPos())
			}
		}

	}

	return nil
}

func changeListOrMapValue(ctx *Task, obj any, index []*ast.Node, val V) *errchain.PlError {
	cur := obj
	lenIdx := len(index)

	for idx, node := range index {
		if err := RunExpr(ctx, node); err != nil {
			return err
		}
		key, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), node.StartPos())
		}
		switch curVal := cur.(type) {
		case map[string]any:
			if key.T != ast.String {
				return NewRunError(ctx,
					"key type is not string", node.StartPos())
			}
			if idx+1 == lenIdx {
				curVal[key.V.(string)] = val.V
				return nil
			}

			var ok bool
			cur, ok = curVal[key.V.(string)]
			if !ok {
				return NewRunError(ctx,
					"key not found", node.StartPos())
			}
		case []any:
			if key.T != ast.Int {
				return NewRunError(ctx,
					"key type is not int", node.StartPos())
			}
			keyInt := cast.ToInt(key.V)

			// 反转负数
			if keyInt < 0 {
				keyInt = len(curVal) + keyInt
			}

			if keyInt < 0 || keyInt >= len(curVal) {
				return NewRunError(ctx,
					"list index out of range", node.StartPos())
			}

			if idx+1 == lenIdx {
				curVal[keyInt] = val.V
				return nil
			}

			cur = curVal[keyInt]
		default:
			return NewRunError(ctx,
				"obj not map or list", node.StartPos())
		}
	}
	return nil
}

func RunCallExpr(ctx *Task, expr *ast.CallExpr) *errchain.PlError {
	if funcCall, ok := ctx.GetFn(expr.Name); ok {
		if err := funcCall(ctx, expr); err != nil {
			return err
		}
	}
	return nil
}

func RunSliceExpr(ctx *Task, expr *ast.SliceExpr) *errchain.PlError {
	err := RunExpr(ctx, expr.Obj)
	if err != nil {
		return err
	}
	obj, errReg := ctx.Regs.GetRet()
	if errReg != nil {
		return NewRunError(ctx, errReg.Error(), expr.Obj.StartPos())
	}
	var start, end, step V
	if expr.Start != nil {
		err = RunExpr(ctx, expr.Start)
		if err != nil {
			return err
		}
		start, errReg = ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), expr.Start.StartPos())
		}
	}
	if expr.End != nil {
		err = RunExpr(ctx, expr.End)
		if err != nil {
			return err
		}
		end, errReg = ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), expr.End.StartPos())
		}
	}
	if expr.Step != nil {
		err = RunExpr(ctx, expr.Step)
		if err != nil {
			return err
		}
		step, errReg = ctx.Regs.GetRet()
		if errReg != nil {
			return NewRunError(ctx, errReg.Error(), expr.Step.StartPos())
		}
	}
	var startInt, endInt, stepInt int
	var length int
	switch obj.T { //nolint:exhaustive
	case ast.String:
		length = len(obj.V.(string))
	case ast.List, ast.DType(ast.TypeSliceExpr):
		length = len(obj.V.([]any))
	default:
		return NewRunError(ctx, "invalid obj type", expr.Obj.StartPos())
	}

	switch step.T {
	case ast.Invalid:
		stepInt = 1
	case ast.Int:
		stepInt = cast.ToInt(step.V)
		if stepInt == 0 {
			return NewRunError(ctx, "step must be non-zero", expr.Step.StartPos())
		}
	default:
		return NewRunError(ctx, "step type must be integer", expr.Step.StartPos())

	}

	switch start.T {
	case ast.Invalid:
		if stepInt > 0 {
			startInt = 0
		} else {
			startInt = length - 1
		}
	case ast.Int:
		startInt = cast.ToInt(start.V)
		if startInt < 0 {
			startInt = length + startInt
		}
	default:
		return NewRunError(ctx, "start type must be integer", expr.Start.StartPos())
	}

	switch end.T {
	case ast.Invalid:
		if stepInt > 0 {
			endInt = length
		} else {
			endInt = -1
		}
	case ast.Int:
		endInt = cast.ToInt(end.V)
		if endInt < 0 {
			endInt = length + endInt
		}
	default:
		return NewRunError(ctx, "end type must be integer", expr.End.StartPos())

	}

	switch obj.T {
	case ast.String:
		str := obj.V.(string)
		if stepInt > 0 {
			result := ""
			if startInt < 0 {
				startInt = 0
			}
			for i := startInt; i < endInt && i < length; i += stepInt {
				result += string(str[i])
			}
			ctx.Regs.ReturnAppend(V{result, ast.String})
			return nil
		} else {
			result := ""
			if startInt > length-1 {
				startInt = length - 1
			}
			for i := startInt; i > endInt && i >= 0; i += stepInt {
				result += string(str[i])
			}
			ctx.Regs.ReturnAppend(V{result, ast.String})
			return nil
		}
	default:
		list := obj.V.([]any)
		if stepInt > 0 {
			if startInt < 0 {
				startInt = 0
			}
			if endInt > length {
				endInt = length
			}
			result := make([]any, 0, (endInt-startInt+stepInt-1)/stepInt)
			for i := startInt; i < endInt; i += stepInt {
				result = append(result, list[i])
			}
			ctx.Regs.ReturnAppend(V{result, ast.List})
			return nil
		} else {
			if startInt > length-1 {
				startInt = length - 1
			}
			if endInt < 0 {
				endInt = -1
			}
			result := make([]any, 0, (startInt-endInt-stepInt-1)/(-stepInt))
			for i := startInt; i > endInt; i += stepInt {
				result = append(result, list[i])
			}
			ctx.Regs.ReturnAppend(V{result, ast.List})
			return nil
		}
	}
}
func typePromotion(l ast.DType, r ast.DType) ast.DType {
	if l == ast.Float || r == ast.Float {
		return ast.Float
	}

	return ast.Int
}

func condOp(lhs, rhs V, op ast.Op) (any, ast.DType, error) {
	switch op { //nolint:exhaustive
	case ast.EQEQ:
		switch lhs.T { //nolint:exhaustive
		case ast.Int, ast.Bool, ast.Float:
			switch rhs.T { //nolint:exhaustive
			case ast.Int, ast.Bool, ast.Float:
			default:
				return false, ast.Bool, nil
			}
			dtype := typePromotion(lhs.T, rhs.T)
			if dtype == ast.Float {
				return cast.ToFloat64(lhs.V) == cast.ToFloat64(rhs.V), ast.Bool, nil
			}
			return cast.ToFloat64(lhs.V) == cast.ToFloat64(rhs.V), ast.Bool, nil
		case ast.String:
			if rhs.T != ast.String {
				return false, ast.Bool, nil
			}
			return cast.ToString(lhs.V) == cast.ToString(rhs.V), ast.Bool, nil
		case ast.Nil:
			if rhs.T != ast.Nil {
				return false, ast.Bool, nil
			}
			return true, ast.Bool, nil
		default:
			return reflect.DeepEqual(lhs, rhs), ast.Bool, nil
		}

	case ast.NEQ:
		switch lhs.T { //nolint:exhaustive
		case ast.Int, ast.Bool, ast.Float:
			switch rhs.T { //nolint:exhaustive
			case ast.Int, ast.Bool, ast.Float:
			default:
				return true, ast.Bool, nil
			}
			dtype := typePromotion(lhs.T, rhs.T)
			if dtype == ast.Float {
				return cast.ToFloat64(lhs.V) != cast.ToFloat64(rhs.V), ast.Bool, nil
			}
			return cast.ToFloat64(lhs.V) != cast.ToFloat64(rhs.V), ast.Bool, nil
		case ast.String:
			if rhs.T != ast.String {
				return true, ast.Bool, nil
			}
			return cast.ToString(lhs.V) != cast.ToString(rhs.V), ast.Bool, nil
		case ast.Nil:
			if rhs.T != ast.Nil {
				return true, ast.Bool, nil
			}
			return false, ast.Bool, nil
		default:
			return !reflect.DeepEqual(lhs, rhs), ast.Bool, nil
		}
	}

	if !cmpType(lhs.T) {
		return nil, ast.Invalid, fmt.Errorf("not compareable")
	}
	if !cmpType(rhs.T) {
		return nil, ast.Invalid, fmt.Errorf("not compareable")
	}

	switch op { //nolint:exhaustive
	case ast.AND, ast.OR:
		if lhs.T != ast.Bool || rhs.T != ast.Bool {
			return nil, ast.Invalid, fmt.Errorf("unsupported operand type(s) for %s: %s and %s",
				op, lhs.T, rhs.T)
		}
		if op == ast.AND {
			return cast.ToBool(lhs.V) && cast.ToBool(rhs.V), ast.Bool, nil
		} else {
			return cast.ToBool(lhs.V) || cast.ToBool(rhs.V), ast.Bool, nil
		}

	case ast.LT:
		dtype := typePromotion(lhs.T, rhs.T)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs.V) < cast.ToFloat64(rhs.V), ast.Bool, nil
		}
		return cast.ToInt(lhs.V) < cast.ToInt(rhs.V), ast.Bool, nil
	case ast.LTE:
		dtype := typePromotion(lhs.T, rhs.T)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs.V) <= cast.ToFloat64(rhs.V), ast.Bool, nil
		}
		return cast.ToInt(lhs.V) <= cast.ToInt(rhs.V), ast.Bool, nil
	case ast.GT:
		dtype := typePromotion(lhs.T, rhs.T)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs.V) > cast.ToFloat64(rhs.V), ast.Bool, nil
		}
		return cast.ToInt(lhs.V) > cast.ToInt(rhs.V), ast.Bool, nil
	case ast.GTE:
		dtype := typePromotion(lhs.T, rhs.T)
		if dtype == ast.Float {
			return cast.ToFloat64(lhs.V) >= cast.ToFloat64(rhs.V), ast.Bool, nil
		}
		return cast.ToInt(lhs.V) >= cast.ToInt(rhs.V), ast.Bool, nil
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
