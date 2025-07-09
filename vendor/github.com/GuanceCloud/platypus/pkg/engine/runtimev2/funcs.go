package runtimev2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/errchain"
)

type Fn struct {
	CallCheck FnCall
	Call      FnCall
	Cat       []string

	Desc FnDesc
}

type Param struct {
	Name string
	Desc string
	Typs []ast.DType

	Val func() any

	Variable bool
}

func (p *Param) TypStr() string {
	typs := make([]string, 0, len(p.Typs))
	for _, dt := range p.Typs {
		typs = append(typs, dt.String())
	}
	return strings.Join(typs, "|")
}

type FnDesc struct {
	Name    string
	Params  []*Param
	Returns []*Param

	Desc string
}

func (desc FnDesc) String() string {
	return desc.Signature()
}

type Desc struct {
	Name   string            `json:"name"`
	Desc   string            `json:"desc"`
	Params map[string]string `json:"params"`
	Return map[string]string `json:"return"`
}

func (desc FnDesc) OStruct() Desc {
	d := Desc{
		Name:   desc.Name,
		Desc:   desc.Desc,
		Params: map[string]string{},
		Return: map[string]string{},
	}

	for _, p := range desc.Params {
		d.Params[p.Name] = p.Desc
	}
	for _, r := range desc.Returns {
		d.Return[r.TypStr()] = r.Desc
	}
	return d
}

func (desc FnDesc) OMarkdown(tab string) string {
	fn := desc.Signature() + "\n"
	fn += fmt.Sprintf("%s- desc: %s\n", tab, desc.Desc)
	if len(desc.Params) > 0 {
		fn += fmt.Sprintf("%s- params:\n", tab)
		for _, p := range desc.Params {
			fn += fmt.Sprintf("%s%s- %s: %s\n", tab, tab, p.Name, p.Desc)
		}
	}
	if len(desc.Returns) > 0 {
		fn += fmt.Sprintf("%s- return:\n", tab)
		for _, p := range desc.Returns {
			fn += fmt.Sprintf("%s%s- %s: %s\n", tab, tab, p.TypStr(), p.Desc)
		}
	}

	return fn
}

func (desc FnDesc) Signature() string {
	var fn string
	args := make([]string, 0, len(desc.Params))
	p := desc.Params
	for i := range p {
		argStr := p[i].Name
		if typ := p[i].TypStr(); typ != "" {
			if p[i].Variable {
				argStr += ": " + "..." + typ
			} else {
				argStr += ": " + typ
			}
		} else if p[i].Variable {
			argStr += ": " + "..."
		}

		if p[i].Val != nil {
			v := p[i].Val()
			if v == nil {
				argStr += " = nil"
			} else {
				buf := bytes.NewBuffer([]byte{})
				enc := json.NewEncoder(buf)
				if err := enc.Encode(v); err == nil {
					argStr += " = " + strings.TrimRight(buf.String(), "\n")
				}
			}
		}
		args = append(args, argStr)
	}

	fn += fmt.Sprintf("fn %s(%s)", desc.Name, strings.Join(args, ", "))
	if len(desc.Returns) > 0 {
		rets := make([]string, 0, len(desc.Returns))
		for i := range desc.Returns {
			rets = append(rets, desc.Returns[i].TypStr())
		}
		if len(rets) > 1 {
			fn += fmt.Sprintf(" -> (%s)", strings.Join(rets, ", "))
		} else if len(rets) > 0 {
			fn += fmt.Sprintf(" -> %s", strings.Join(rets, ", "))
		}
	}
	return fn
}

// isValidParamName checks if a string is a valid param name
func isValidParamName(name string) error {
	// check if the string is empty
	if len(name) == 0 {
		return fmt.Errorf("the variable name cannot be empty")
	}

	// check if the first character is a letter or underscore
	r := []rune(name)[0]
	if !unicode.IsLetter(r) && r != '_' {
		return fmt.Errorf(
			"the parameter name must start with a letter or underscore. The first character is %c", r)
	}

	// check if the remaining characters are letters, digits, or underscores
	for _, r := range name[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf(
				"the variable name can only contain letters, digits, and underscores, invalid character %c found", r)
		}
	}

	// TODO: check keyword
	return nil
}

func CheckFnParamDef(params []*Param) error {
	optional := false
	variable := false
	names := map[string]struct{}{}
	for i, p := range params {
		// valid param
		if err := isValidParamName(p.Name); err != nil {
			return err
		}
		if _, ok := names[p.Name]; ok {
			return fmt.Errorf("repeated parameters %s", p.Name)
		}
		if p.Val != nil {
			if !optional {
				optional = true
			}
			// TODO check value data type
		} else {
			if optional {
				return fmt.Errorf(
					"comparison parameter must be before optional parameter: %s",
					p.Name)
			}
		}
		if p.Variable {
			if optional {
				return fmt.Errorf("variable parameters and optional parameters cannot be used at the same time")
			}
			if !variable {
				variable = true
			} else {
				return fmt.Errorf("only one variable parameter is allowed: %s", p.Name)
			}

			if i != len(params)-1 {
				return fmt.Errorf("variable parameters can only be the last parameter")
			}
		}
	}
	return nil
}

func CheckPassParam(ctx *Task, expr *ast.CallExpr, params []*Param) *errchain.PlError {
	var newArgs []*ast.Node
	if len(expr.Param) > len(params) {
		newArgs = make([]*ast.Node, len(expr.Param))
	} else {
		newArgs = make([]*ast.Node, len(params))
	}

	namedParam := false
	varbParam := false
	if len(params) > 0 {
		if params[len(params)-1].Variable {
			varbParam = true
		}
	}
	for ePIndex, p := range expr.Param {
		if p.NodeType == ast.TypeAssignmentExpr { // named param
			if varbParam {
				return NewRunError(ctx,
					"when there are variable parameters, named parameters cannot be used to pass parameters", p.StartPos(),
				)
			}
			if !namedParam {
				namedParam = true
			}
			ass := p.AssignmentExpr()
			if len(ass.LHS) != 1 || len(ass.RHS) != 1 {
				return NewRunError(ctx, "not named parameter node", ass.OpPos)
			}
			if ass.LHS[0].NodeType != ast.TypeIdentifier {
				return NewRunError(ctx,
					"the name of a named parameter needs to be an identifier", p.StartPos())
			}
			for pi := range params {
				if params[pi].Name == ass.LHS[0].String() {
					if newArgs[pi] != nil {
						return NewRunError(ctx, fmt.Sprintf(
							"duplicate parameters %s", params[pi].Name), p.StartPos())
					}
					newArgs[pi] = ass.RHS[0]
				}
			}
		} else { // pos param
			if namedParam {
				return NewRunError(ctx,
					"positional parameters should come before named parameters", p.StartPos(),
				)
			}
			newArgs[ePIndex] = p
		}
	}

	for i := range params {
		if params[i].Val == nil && !params[i].Variable {
			if newArgs[i] == nil {
				return NewRunError(ctx,
					fmt.Sprintf("missing parameter `%s`", params[i].Name),
					expr.NamePos,
				)
			}
		} else {
			break
		}
	}

	expr.ParamNormalized = newArgs

	return nil
}

// GetParam get the value by specifying the parameter position, index starts from 1
func GetParam(ctx *Task, expr *ast.CallExpr, params []*Param, i int) (any, *errchain.PlError) {
	if i >= len(params) || i >= len(expr.ParamNormalized) {
		return nil, NewRunError(ctx, fmt.Sprintf(
			"no such parameter, index %d", i), expr.NamePos)
	}

	if params[i].Variable {
		var ret []any
		if expr.ParamNormalized[i] == nil {
			return ret, nil
		}
		for ePIndex, p := range expr.ParamNormalized[i:] {
			if p == nil {
				return nil, NewRunError(ctx, fmt.Sprintf(
					"variable parameter value %d not passed", ePIndex), expr.ParamNormalized[i].StartPos())
			}
			err := RunExpr(ctx, p)
			if err != nil {
				return nil, err
			}
			v, errReg := ctx.Regs.GetRet()
			if errReg != nil {
				return nil, NewRunError(ctx, errReg.Error(), p.StartPos())
			}
			ret = append(ret, v.V)
		}
		return ret, nil
	} else {
		if expr.ParamNormalized[i] == nil {
			if params[i].Val != nil {
				return params[i].Val(), nil
			} else {
				return nil, NewRunError(ctx, fmt.Sprintf(
					"parameter %s was not passed", params[i].Name), expr.NamePos)
			}
		}
		err := RunExpr(ctx, expr.ParamNormalized[i])
		if err != nil {
			return nil, err
		}
		v, errReg := ctx.Regs.GetRet()
		if errReg != nil {
			return nil, NewRunError(ctx, errReg.Error(), expr.ParamNormalized[i].StartPos())
		}
		return v.V, nil
	}
}

func GetParamInt(ctx *Task, expr *ast.CallExpr, params []*Param, i int) (int64, *errchain.PlError) {
	p, err := GetParam(ctx, expr, params, i)
	if err != nil {
		return 0, err
	}

	switch p := p.(type) {
	case int:
		return int64(p), nil
	case int64:
		return p, nil
	case uint32:
		return int64(p), nil
	default:
		return 0, NewRunError(ctx, fmt.Sprintf(
			"param %s with unsupported value type", params[i].Name), expr.NamePos)
	}
}

func GetParamFloat(ctx *Task, expr *ast.CallExpr, params []*Param, i int) (float64, *errchain.PlError) {
	p, err := GetParam(ctx, expr, params, i)
	if err != nil {
		return 0, err
	}

	switch p := p.(type) {
	case float32:
		return float64(p), nil
	case float64:
		return p, nil
	default:
		return 0, NewRunError(ctx, fmt.Sprintf(
			"param %s with unsupported value type", params[i].Name), expr.NamePos)
	}
}

func GetParamBool(ctx *Task, expr *ast.CallExpr, params []*Param, i int) (bool, *errchain.PlError) {
	p, err := GetParam(ctx, expr, params, i)
	if err != nil {
		return false, err
	}

	if v, ok := p.(bool); ok {
		return v, nil
	}
	return false, NewRunError(ctx, fmt.Sprintf(
		"param %s with unsupported value type", params[i].Name), expr.NamePos)
}

func GetParamString(ctx *Task, expr *ast.CallExpr, params []*Param, i int) (string, *errchain.PlError) {
	p, err := GetParam(ctx, expr, params, i)
	if err != nil {
		return "", err
	}

	if v, ok := p.(string); ok {
		return v, nil
	}
	return "", NewRunError(ctx, fmt.Sprintf(
		"param %s with unsupported value type", params[i].Name), expr.NamePos)
}

func GetParamList(ctx *Task, expr *ast.CallExpr, params []*Param, i int) ([]any, *errchain.PlError) {
	p, err := GetParam(ctx, expr, params, i)
	if err != nil {
		return nil, err
	}

	if v, ok := p.([]any); ok {
		return v, nil
	}

	return nil, NewRunError(ctx, fmt.Sprintf(
		"param %s with unsupported value type", params[i].Name), expr.NamePos)
}

func GetParamMap(ctx *Task, expr *ast.CallExpr, params []*Param, i int) (map[string]any, *errchain.PlError) {
	p, err := GetParam(ctx, expr, params, i)
	if err != nil {
		return nil, err
	}

	if v, ok := p.(map[string]any); ok {
		return v, nil
	}
	return nil, NewRunError(ctx, fmt.Sprintf(
		"param %s with unsupported value type", params[i].Name), expr.NamePos)
}
