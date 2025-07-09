// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package funcs

import (
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

type FnCall func(ctx *runtime.Task, funcExpr *ast.CallExpr, vals ...any) *errchain.PlError

func VarPDefaultVal() (any, ast.DType) {
	return []any(nil), ast.List
}

func NewFunc(name string, params []*Param, ret []ast.DType, doc [2]*PLDoc, run FnCall) *Function {
	return &Function{
		Name: name,
		Args: params,
		// Return: ret,
		Doc:   doc,
		Call:  WrapFnCall(run, params),
		Check: WrapFnCheck(params),
	}
}

func WrapFnCall(fn FnCall, paramDesc []*Param) runtime.FuncCall {
	return func(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
		// The parameters of the function call expression need to be normalized in advance.

		// Note that some functions do not take the value of the variable
		// corresponding to the parameter, but its name.

		vals := make([]any, len(funcExpr.Param))

		lenP := len(paramDesc)
		varP := false
		if lenP > 0 {
			if paramDesc[lenP-1].VariableP {
				lenP -= 1
				varP = true
			}

			for i := 0; i < lenP; i++ {
				if val, err := getParam(ctx, paramDesc[i], funcExpr.Param[i]); err != nil {
					return err
				} else {
					vals[i] = val
				}
			}

			if varP {
				if v, err := getVarParam(ctx, paramDesc[lenP], funcExpr.Param[lenP:]); err != nil {
					return err
				} else {
					vals[lenP] = v
				}
			}
		}
		return fn(ctx, funcExpr, vals...)
	}
}

func WrapFnCheck(paramDesc []*Param) runtime.FuncCheck {
	kIndex := map[string]int{}

	prvOptP := false
	for i, p := range paramDesc {
		if _, ok := kIndex[p.Name]; ok {
			panic(fmt.Sprintf("duplicate parameter name: %s", p.Name))
		} else {
			kIndex[p.Name] = i
		}

		if p.VariableP {
			if i != len(paramDesc)-1 {
				panic(fmt.Sprintf("parameter %s: variable parameter should be the last one",
					p.Name))
			}
			if p.DefaultVal == nil {
				panic(fmt.Sprintf("parameter %s: variable parameter should have default value", p.Name))
			}

			val, _ := p.DefaultVal()
			switch val := val.(type) {
			case []any:
				for _, v := range val {
					dtyp := getValDtype(v)
					if !typInclude(dtyp, p.Type) {
						panic(fmt.Sprintf("parameter %s: default value data type not match", p.Name))
					}
				}
			case nil:
			default:
				panic(fmt.Sprintf("parameter %s: default value type not match", p.Name))
			}
		} else {
			if p.Optional {
				if p.DefaultVal == nil {
					panic(fmt.Sprintf("parameter %s: optional parameter should have default value",
						p.Name))
				}
				val, dt := p.DefaultVal()
				if getValDtype(val) != dt {
					panic(fmt.Sprintf("parameter %s: value type not match", p.Name))
				}
				if !typInclude(dt, p.Type) {
					panic(fmt.Sprintf("parameter %s: default value data type not match", p.Name))
				}
			}
		}

		if !p.Optional && !p.VariableP {
			if prvOptP {
				panic(fmt.Sprintf("parameter %s: required parameter should not follow optional parameter",
					p.Name))
			}
		} else {
			prvOptP = true
		}
	}

	return func(ctx *runtime.Task, funcExpr *ast.CallExpr) *errchain.PlError {
		if err := normalizeFuncParams(ctx, kIndex, paramDesc, funcExpr); err != nil {
			return err
		}
		if err := checkParams(ctx, funcExpr, paramDesc); err != nil {
			return err
		}
		return nil
	}
}

func getValDtype(v any) ast.DType {
	var dtyp ast.DType
	switch v.(type) {
	case string:
		dtyp = ast.String
	case int64:
		dtyp = ast.Int
	case float64:
		dtyp = ast.Float
	case bool:
		dtyp = ast.Bool
	case []any:
		dtyp = ast.List
	case map[string]any:
		dtyp = ast.Map
	case nil:
		dtyp = ast.Nil
	default:
	}
	return dtyp
}

func checkParams(ctx *runtime.Task, funcExpr *ast.CallExpr, paramDesc []*Param) *errchain.PlError {
	lenP := len(paramDesc)
	varP := false
	if paramDesc[lenP-1].VariableP {
		lenP -= 1
		varP = true
	}

	for i := 0; i < lenP; i++ {
		if p := funcExpr.Param[i]; p != nil {
			if ok := checkLiteralType(p.NodeType, paramDesc[i].Type); !ok {
				return runtime.NewRunError(ctx, "unexpected data type", p.StartPos())
			}
		}
	}
	if varP {
		if p := funcExpr.Param[lenP]; p != nil {
			if p.NodeType == ast.TypeAssignmentExpr {
				expr := p.AssignmentExpr()
				switch expr.RHS[0].NodeType {
				case ast.TypeListLiteral:
					for _, n := range expr.RHS[0].ListLiteral().List {
						if ok := checkLiteralType(n.NodeType, paramDesc[lenP].Type); !ok {
							return runtime.NewRunError(ctx, "unexpected data type", p.StartPos())
						}
					}
				case ast.TypeNilLiteral:
				default:
					return runtime.NewRunError(ctx,
						"unexpected data type, for named variable parameter only list and nil are supported",
						p.StartPos())
				}
			} else {
				for i := lenP; i < len(funcExpr.Param); i++ {
					if ok := checkLiteralType(funcExpr.Param[i].NodeType, paramDesc[lenP].Type); !ok {
						return runtime.NewRunError(ctx, "unexpected data type", p.StartPos())
					}
				}
			}
		}
	}
	return nil
}

func checkLiteralType(typ ast.NodeType, typs []ast.DType) bool {
	var dtype ast.DType
	switch typ {
	case ast.TypeStringLiteral:
		dtype = ast.String
	case ast.TypeIntegerLiteral:
		dtype = ast.Int
	case ast.TypeFloatLiteral:
		dtype = ast.Float
	case ast.TypeBoolLiteral:
		dtype = ast.Bool
	case ast.TypeNilLiteral:
		dtype = ast.Nil
	case ast.TypeListLiteral:
		dtype = ast.List
	case ast.TypeMapLiteral:
		dtype = ast.Map
	}

	if dtype != ast.Invalid {
		return typInclude(dtype, typs)
	}

	return true
}

func typInclude(typ ast.DType, typs []ast.DType) bool {
	for _, dt := range typs {
		if dt == typ {
			return true
		}
	}
	return false
}

func normalizeFuncParams(ctx *runtime.Task, keyMapp map[string]int,
	paramDesc []*Param, funcExpr *ast.CallExpr) *errchain.PlError {
	includeVarP := false
	if len(paramDesc) > 0 {
		if paramDesc[len(paramDesc)-1].VariableP {
			includeVarP = true
		}
	}

	if !includeVarP && len(funcExpr.Param) > len(keyMapp) {
		return runtime.NewRunError(ctx, "too many parameters", funcExpr.NamePos)
	}

	var dstArgs []*ast.Node
	if len(funcExpr.Param) < len(keyMapp) {
		dstArgs = make([]*ast.Node, len(keyMapp))
	} else {
		dstArgs = make([]*ast.Node, len(funcExpr.Param))
	}

	includeNamedP := false
	prvIsPosVar := true
	for idx, arg := range funcExpr.Param {
		if arg.NodeType == ast.TypeAssignmentExpr {
			includeNamedP = true
			if prvIsPosVar {
				prvIsPosVar = false
			}
			if arg.AssignmentExpr().LHS[0].NodeType != ast.TypeIdentifier {
				return runtime.NewRunError(ctx, "named parameter must be an identifier", arg.StartPos())
			}

			kname := arg.AssignmentExpr().LHS[0].Identifier().Name

			kIndex, ok := keyMapp[kname]
			if !ok {
				return runtime.NewRunError(ctx, "unknown named parameter", arg.StartPos())
			}

			dstArgs[kIndex] = arg
		} else {
			if !prvIsPosVar {
				return runtime.NewRunError(ctx, "positional parameter should not follow keyword parameter", arg.StartPos())
			}
			dstArgs[idx] = arg
		}
	}

	for i, v := range paramDesc {
		if !v.Optional && !v.VariableP {
			if dstArgs[i] == nil {
				return runtime.NewRunError(ctx, fmt.Sprintf("missing required parameter `%s`", v.Name), funcExpr.NamePos)
			}
		}
	}

	if includeNamedP && len(funcExpr.Param) > len(keyMapp) {
		return runtime.NewRunError(ctx, "too many parameters", funcExpr.NamePos)
	}

	funcExpr.Param = dstArgs

	return nil
}

func normalizeFuncArgsDeprecated(fnStmt *ast.CallExpr, keyList []string, reqParm int) error {
	// reqParm >= 1, if < 0, no optional args
	args := fnStmt.Param

	if reqParm < 0 || reqParm > len(keyList) {
		reqParm = len(keyList)
	}

	if len(args) > len(keyList) {
		return fmt.Errorf("the number of parameters does not match")
	}

	beforPosArg := true

	kMap := map[string]int{}
	for k, v := range keyList {
		kMap[v] = k
	}

	ret := make([]*ast.Node, len(keyList))

	for idx, arg := range args {
		if arg.NodeType == ast.TypeAssignmentExpr {
			if beforPosArg {
				beforPosArg = false
			}
			kname, err := getKeyName(arg.AssignmentExpr().LHS[0])
			if err != nil {
				return err
			}
			kIndex, ok := kMap[kname]
			if !ok {
				return fmt.Errorf("argument %s does not exist", kname)
			}
			ret[kIndex] = arg.AssignmentExpr().RHS[0]
		} else {
			if !beforPosArg {
				return fmt.Errorf("positional arguments cannot follow keyword arguments")
			}
			ret[idx] = arg
		}
	}

	for i := 0; i < reqParm; i++ {
		if v := ret[i]; v == nil {
			return fmt.Errorf("parameter %s is required", keyList[i])
		}
	}

	fnStmt.Param = ret
	return nil
}

func getVarParam(ctx *runtime.Task, pDesc *Param, p []*ast.Node) ([]any, *errchain.PlError) {
	if p[0] == nil {
		if pDesc.DefaultVal != nil {
			val, _ := pDesc.DefaultVal()
			switch val := val.(type) {
			case []any:
				return val, nil
			}
		}
		return nil, nil
	}

	if p[0].NodeType == ast.TypeAssignmentExpr {
		val, _, errR := runtime.RunStmt(ctx, p[0])
		if errR != nil {
			return nil, errR
		}

		switch val := val.(type) {
		case []any:
			return val, nil
		case nil:
			return nil, nil
		default:
			return nil, runtime.NewRunError(ctx, fmt.Sprintf("parameter %s: type not match",
				pDesc.Name), p[0].StartPos())
		}
	} else {
		vals := make([]any, len(p))
		for i, v := range p {
			val, dtype, errR := runtime.RunStmt(ctx, v)
			if errR != nil {
				return nil, errR
			}

			typNotMatch := true
			for _, d := range pDesc.Type {
				if dtype == d {
					typNotMatch = false
					break
				}
			}

			if typNotMatch {
				return nil, runtime.NewRunError(ctx, fmt.Sprintf("parameter %s element: type not match",
					pDesc.Name), v.StartPos())
			} else {
				vals[i] = val
			}
		}
		return vals, nil
	}
}

func getParam(ctx *runtime.Task, pDesc *Param, p *ast.Node) (any, *errchain.PlError) {
	var val any
	var dtype ast.DType
	if p == nil {
		// not need to check type here
		if pDesc.DefaultVal != nil {
			val, _ = pDesc.DefaultVal()
			return val, nil
		} else {
			return nil, nil
		}
	} else {
		var errR *errchain.PlError
		val, dtype, errR = runtime.RunStmt(ctx, p)
		if errR != nil {
			return nil, errR
		}
	}

	for _, d := range pDesc.Type {
		if dtype == d {
			return val, nil
		}
	}

	return nil, runtime.NewRunError(ctx, fmt.Sprintf("parameter %s: type not match",
		pDesc.Name), p.StartPos())
}
