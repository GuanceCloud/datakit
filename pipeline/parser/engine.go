package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"

	conv "github.com/spf13/cast"
	vgrok "github.com/vjeantet/grok"
)

type phase string

const (
	MakePhase phase = "make"
	OffPhase  phase = "off"
)

type (
	FuncCallback      func(*Engine, Node) error
	FuncCallbackCheck func(Node) error
)

type Engine struct {
	output  map[string]interface{}
	content string

	patterns map[string]string
	grok     *vgrok.Grok

	callbacks     map[string]FuncCallback
	callbackCheck map[string]FuncCallbackCheck
	stmts         Stmts

	lastErr error
}

func NewEngine(script string, callbacks map[string]FuncCallback, check map[string]FuncCallbackCheck) (*Engine, error) {
	node, err := ParsePipeline(script)
	if err != nil {
		return nil, err
	}

	stmts, ok := node.(Stmts)
	if !ok {
		return nil, fmt.Errorf("invalid AST, should not been here")
	}
	ng := &Engine{
		output:    make(map[string]interface{}),
		patterns:  CopyGlobalPatterns(),
		callbacks: callbacks,
		stmts:     stmts,
	}
	ng.grok, _ = vgrok.NewWithConfig(&vgrok.Config{
		SkipDefaultPatterns: true,
		NamedCapturesOnly:   true,
		Patterns:            ng.patterns,
	})

	return ng, nil
}

func (ng *Engine) Check() error {
	return ng.stmts.Check(ng)
}

func (ng *Engine) Run(input string) error {
	ng.reset()
	ng.content = input
	ng.output["message"] = input
	ng.stmts.Run(ng)
	return ng.lastErr
}

func (ng *Engine) Result() map[string]interface{} {
	for k, v := range ng.output {
		switch v.(type) {
		case int, uint64, uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
		default:
			str, err := json.Marshal(v)
			if err != nil {
				log.Errorf("object type marshal error %v", err)
			}
			ng.output[k] = string(str)
		}
	}
	return ng.output
}

func (ng *Engine) LastErr() error {
	return ng.lastErr
}

func (ng *Engine) reset() {
	ng.output = make(map[string]interface{})
	ng.lastErr = nil
	ng.content = ""
}

func (ng *Engine) GetContentStr(key interface{}) (string, error) {
	c, err := ng.GetContent(key)
	if err != nil {
		return "", err
	}

	switch v := reflect.ValueOf(c); v.Kind() { //nolint:exhaustive
	case reflect.Map:
		res, err := json.Marshal(v.Interface())
		return string(res), err
	default:
		return conv.ToString(v.Interface()), err
	}
}

func (ng *Engine) GetContent(key interface{}) (interface{}, error) {
	var k string

	switch t := key.(type) {
	case *Identifier:
		k = t.String()
	case *AttrExpr:
		k = t.String()
	case *StringLiteral:
		k = t.Val
	case string:
		k = t
	default:
		return nil, fmt.Errorf("unsupported %v get", reflect.TypeOf(key).String())
	}

	if k == "_" {
		return ng.content, nil
	}

	v, ok := ng.output[k]
	if !ok {
		return nil, fmt.Errorf("%s no found", k)
	}

	return v, nil
}

func (ng *Engine) GetPatterns() map[string]string {
	return ng.patterns
}

func (ng *Engine) SetKey(k string, v interface{}) {
	if v == nil { // ignored
		return
	}

	if ng.output == nil {
		ng.output = map[string]interface{}{}
	}
	ng.output[k] = v
}

const (
	DefaultStr   = ""
	InvalidInt   = math.MinInt32 // error: MinInt64?
	DefaultInt   = int64(0xdeadbeef)
	InvalidStr   = "deadbeaf"
	InvalidFloat = math.SmallestNonzeroFloat64
)

func (ng *Engine) getStrArg(node Node) (string, error) {
	switch v := node.(type) {
	case *StringLiteral:
		return v.Val, nil
	case *AttrExpr, *Identifier:
		return ng.GetContentStr(v)
	default:
		return "", fmt.Errorf("invalid arg type %s(%s)",
			reflect.TypeOf(node).String(), node.String())
	}
}

func (ng *Engine) kwGetStrArg(args map[string]Node, kw string) (string, error) {
	v, ok := args[kw]
	if !ok {
		return DefaultStr, nil
	}
	return ng.getStrArg(v)
}

func (ng *Engine) getIntArg(node Node) (int64, error) {
	str, err := ng.getStrArg(node)
	if err != nil {
		return InvalidInt, err
	}
	if str == "" {
		return DefaultInt, nil
	}

	v, err := strconv.ParseInt(str, 10, 64) //nolint: gomnd
	if err != nil {
		return InvalidInt, err
	}
	return v, nil
}

func (ng *Engine) kwGetIntArg(args map[string]Node, kw string) (int64, error) {
	v, ok := args[kw]
	if !ok {
		return DefaultInt, nil
	}
	return ng.getIntArg(v)
}

func (ng *Engine) GetFuncStrArg(f *FuncStmt, idx int, kw string) (string, error) {
	if len(f.KwParam) > 0 {
		return ng.kwGetStrArg(f.KwParam, kw)
	}

	if f.Param != nil {
		if idx >= len(f.Param) {
			return InvalidStr, fmt.Errorf("arg index out of range")
		}
		return ng.getStrArg(f.Param[idx])
	}

	return InvalidStr, fmt.Errorf("no params available")
}

func (ng *Engine) GetFuncIntArg(f *FuncStmt, idx int, kw string) (int64, error) {
	if len(f.KwParam) > 0 {
		return ng.kwGetIntArg(f.KwParam, kw)
	}

	if f.Param != nil {
		if idx >= len(f.Param) {
			return InvalidInt, fmt.Errorf("arg index outof range")
		}
		return ng.getIntArg(f.Param[idx])
	}

	return InvalidInt, fmt.Errorf("no params available")
}

func (ng *Engine) GetFuncFloatArg(f *FuncStmt, idx int, kw string) (float64, error) {
	return InvalidFloat, fmt.Errorf("not implemented")
}

func (ng *Engine) GetGrok() *vgrok.Grok {
	return ng.grok
}

func (ng *Engine) SetContent(k, v interface{}) error {
	var key string

	switch t := k.(type) {
	case *Identifier:
		key = t.String()
	case *AttrExpr:
		key = t.String()
	case *StringLiteral:
		key = t.Val
	case string:
		key = t
	default:
		return fmt.Errorf("unsupported %v set", reflect.TypeOf(key).String())
	}

	if ng.output == nil {
		ng.output = make(map[string]interface{})
	}

	if v == nil {
		return nil
	}

	ng.output[key] = v
	return nil
}

func (ng *Engine) SetGrok(g *vgrok.Grok) {
	if g != nil {
		ng.grok = g
	}
}

func (ng *Engine) DeleteContent(k interface{}) error {
	var key string

	switch t := k.(type) {
	case *Identifier:
		key = t.String()
	case *AttrExpr:
		key = t.String()
	case *StringLiteral:
		key = t.Val
	case string:
		key = t
	default:
		return fmt.Errorf("unsupported %v set", reflect.TypeOf(key).String())
	}

	delete(ng.output, key)
	return nil
}

func (ng *Engine) SetPatterns(patterns map[string]string) error {
	var err error

	if ng.patterns == nil {
		ng.patterns = CopyGlobalPatterns()
	}

	if ng.grok == nil {
		ng.grok, err = vgrok.NewWithConfig(&vgrok.Config{
			SkipDefaultPatterns: true,
			NamedCapturesOnly:   true,
			Patterns:            ng.patterns,
		})
		if err != nil {
			return err
		}
	}

	for k, v := range patterns {
		if _, ok := ng.patterns[k]; !ok && v != "" {
			ng.patterns[k] = v
			if err = ng.grok.AddPattern(k, v); err != nil {
				return err
			}
		}
	}

	return nil
}

///
// Runner
///

func (e Stmts) Run(ng *Engine) {
	if ng.lastErr != nil {
		return
	}
	for _, stmt := range e {
		switch v := stmt.(type) {
		case *IfelseStmt:
			v.Run(ng)
		case *FuncStmt:
			v.Run(ng)
		case *AssignmentStmt:
			v.Run(ng)
		case Stmts:
			v.Run(ng)
		default:
			ng.lastErr = fmt.Errorf("unsupported type %s, from: %s", reflect.TypeOf(v), stmt)
		}
	}
}

func (e *IfelseStmt) Run(ng *Engine) {
	if ng.lastErr != nil {
		return
	}

	if !e.IfList.Run(ng) {
		e.Else.Run(ng)
	}
}

func (e IfList) Run(ng *Engine) (end bool) {
	if ng.lastErr != nil {
		return true
	}
	for _, ifexpr := range e {
		end = ifexpr.Run(ng)
		if end {
			return
		}
	}
	return
}

func (e *IfExpr) Run(ng *Engine) (pass bool) {
	if ng.lastErr != nil {
		return true
	}

	switch v := e.Condition.(type) {
	case *ConditionalExpr:
		pass = v.Run(ng)
	case *BoolLiteral:
		pass = v.Val
	default:
		ng.lastErr = fmt.Errorf("unsupported type %s, from: %s", reflect.TypeOf(v), e.Condition)
		return
	}

	if pass {
		e.Stmts.Run(ng)
	}

	return
}

func (e *ConditionalExpr) Run(ng *Engine) (pass bool) {
	if ng.lastErr != nil {
		return true
	}

	switch v := e.LHS.(type) {
	case *Identifier:
		left, ok := ng.output[v.Name]
		if !ok {
			return false
		}

		switch vv := e.RHS.(type) {
		case *StringLiteral:
			return contrast(left, e.Op.String(), vv.Value())
		case *NumberLiteral:
			return contrast(left, e.Op.String(), vv.Value())
		case *BoolLiteral:
			return contrast(left, e.Op.String(), vv.Value())
		default:
			ng.lastErr = fmt.Errorf("unsupported type %s, from: %s", reflect.TypeOf(vv), e.RHS)
		}
	default:
		ng.lastErr = fmt.Errorf("unsupported type %s, from: %s", reflect.TypeOf(v), e.LHS)
	}

	return false
}

func (e *ComputationExpr) Run(ng *Engine) {
	// TODO
}

func (e *AssignmentStmt) Run(ng *Engine) {
	if ng.lastErr != nil {
		return
	}

	switch v := e.LHS.(type) {
	case *Identifier:
		switch vv := e.RHS.(type) {
		case *StringLiteral:
			ng.output[v.Name] = vv.Value()
		case *NumberLiteral:
			ng.output[v.Name] = vv.Value()
		case *BoolLiteral:
			ng.output[v.Name] = vv.Value()
		default:
			ng.lastErr = fmt.Errorf("unsupported type %s, from: %s", reflect.TypeOf(vv), e.RHS)
		}
	default:
		ng.lastErr = fmt.Errorf("unsupported type %s, from: %s", reflect.TypeOf(v), e.LHS)
	}
}

func (e *FuncStmt) Run(ng *Engine) {
	if ng.lastErr != nil {
		return
	}

	fn := ng.callbacks[e.Name]
	if fn == nil {
		ng.lastErr = fmt.Errorf("unsupported func: `%v'", e.Name)
		return
	}
	if err := fn(ng, e); err != nil {
		ng.lastErr = fmt.Errorf("Run func %v: %w", e.Name, err)
	}
}

///
// Literal Value: StringLiteral, BoolLiteral, NumberLiteral

func (e *StringLiteral) Value() interface{} { return e.Val }
func (e *BoolLiteral) Value() interface{}   { return e.Val }
func (e *NumberLiteral) Value() interface{} {
	if e.IsInt {
		return e.Int
	}
	return e.Float
}

///
// Checking: Stmts, FuncStmt, AssignmentStmt, IfelseStmt,
///

// Check Stmts
//   stmt only support IfelseStmt/FuncStmt/AssignmentStmt
func (e Stmts) Check(ng *Engine) error {
	for _, stmt := range e {
		switch v := stmt.(type) {
		case *IfelseStmt:
			return v.Check(ng)
		case *FuncStmt:
			return v.Check(ng)
		case *AssignmentStmt:
			return v.Check()
		default:
			return fmt.Errorf(`unsupported type %s, from: %s`,
				reflect.TypeOf(stmt), stmt)
		}
	}
	return nil
}

// Check IfExpr
//   Condition support BoolLiteral/ConditionalExpr
func (e *FuncStmt) Check(ng *Engine) error {
	if _, ok := ng.callbacks[e.Name]; !ok {
		return fmt.Errorf("unsupported func: `%v'", e.Name)
	}

	checkFn, ok := ng.callbackCheck[e.Name]
	if !ok {
		return fmt.Errorf("not found check for func: `%v'", e.Name)
	}
	return checkFn(e)
}

// Check AssignmentStmt
//   left node only support Identifier
//   right node support NumberLiteral/StringLiteral/BoolLiteral
func (e *AssignmentStmt) Check() error {
	switch e.LHS.(type) {
	case *Identifier:
		// nil
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.LHS), e.LHS)
	}

	switch e.RHS.(type) {
	case *NumberLiteral, *StringLiteral, *BoolLiteral:
		// nil
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.RHS), e.RHS)
	}
	return nil
}

// Check IfelseStmt.
func (e *IfelseStmt) Check(ng *Engine) error {
	if err := e.IfList.Check(ng); err != nil {
		return err
	}
	return e.Else.Check(ng)
}

// Check IfList.
func (e IfList) Check(ng *Engine) error {
	for _, i := range e {
		if err := i.Check(ng); err != nil {
			return err
		}
	}
	return nil
}

// Check IfExpr
//   Condition support BoolLiteral/ConditionalExpr
func (e *IfExpr) Check(ng *Engine) error {
	switch v := e.Condition.(type) {
	case *BoolLiteral:
		// nil
	case *ConditionalExpr:
		return v.Check()
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.Condition), e.Condition)
	}
	return e.Stmts.Check(ng)
}

// Check ConditionalExpr
//   left node only support Identifier
//   right node support NumberLiteral/StringLiteral/BoolLiteral
func (e *ConditionalExpr) Check() error {
	switch e.LHS.(type) {
	case *Identifier:
		// nil
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.LHS), e.LHS)
	}

	switch e.RHS.(type) {
	case *NumberLiteral, *StringLiteral, *BoolLiteral:
		// nil
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.RHS), e.RHS)
	}
	return nil
}

// nolint
func contrast(x interface{}, op string, y interface{}) (b bool) {
	// It's looong! float==float is undefined.
	var (
		float  []float64
		str    []string
		booler []bool
	)
	var err error

	const typeErr = "mismatch of type: %s(%v) %s %s(%v)"

	switch vx := x.(type) {
	case json.Number:
		var xx float64
		xx, err = vx.Float64()
		if err != nil {
			log.Warn(err)
			return
		}

		float = append(float, xx)

		switch vy := y.(type) {
		case json.Number:
			var yy float64
			yy, err = vy.Float64()
			if err != nil {
				return
			}
			float = append(float, yy)
		case float64:
			float = append(float, vy)
		case int64:
			float = append(float, float64(vy))
		default:
			log.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}

	case int64:
		float = append(float, float64(vx))

		switch vy := y.(type) {
		case json.Number:
			var yy float64
			yy, err = vy.Float64()
			if err != nil {
				log.Warn(err)
				return
			}
			float = append(float, yy)
		case float64:
			float = append(float, vy)
		case int64:
			float = append(float, float64(vy))
		default:
			log.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}

	case float64:
		float = append(float, vx)

		switch vy := y.(type) {
		case json.Number:
			var yy float64
			yy, err = vy.Float64()
			if err != nil {
				log.Warn(err)
				return
			}
			float = append(float, yy)
		case float64:
			float = append(float, vy)
		case int64:
			float = append(float, float64(vy))
		default:
			log.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}

	case string:
		yy, ok := y.(string)
		if !ok {
			log.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}
		str = append(str, vx)
		str = append(str, yy)
	case bool:
		yy, ok := y.(bool)
		if !ok {
			log.Warnf(typeErr, reflect.TypeOf(x), x, op, reflect.TypeOf(y), y)
			return
		}
		booler = append(booler, x.(bool))
		booler = append(booler, yy)

	case nil:
		booler = append(booler, true)
		if y == nil {
			booler = append(booler, true)
		} else {
			booler = append(booler, false)
		}

	default:
		log.Warnf("mismatch of type: %s(%v)", reflect.TypeOf(x), x)
		return
	}

	switch op {
	case "==":
		b = reflect.DeepEqual(x, y)
		return
	case "!=":
		if len(float) != 0 {
			b = float[0] != float[1]
			return
		}
		if len(str) != 0 {
			b = str[0] != str[1]
			return
		}
		if len(booler) != 0 {
			b = booler[0] != booler[1]
			return
		}
	case "<=":
		if len(float) != 0 {
			b = float[0] <= float[1]
			return
		}
	case "<":
		if len(float) != 0 {
			b = float[0] < float[1]
			return
		}
	case ">=":
		if len(float) != 0 {
			b = float[0] >= float[1]
			return
		}
	case ">":
		if len(float) != 0 {
			b = float[0] > float[1]
			return
		}
	default:
		log.Warn("unexpected operator")
		return
	}

	log.Warn("the operator is not available for this type")
	return
}
