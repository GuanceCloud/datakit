package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

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
	output  *Output
	content string

	stopRunPP bool // stop run()

	debugMode bool
	ts        time.Time

	patterns map[string]string
	grok     *vgrok.Grok

	callbacks     map[string]FuncCallback
	callbackCheck map[string]FuncCallbackCheck
	stmts         Stmts

	lastErr error
}

//nolint:structcheck,unused
type Output struct {
	Dropped bool
	Error   string
	Cost    map[string]string
	Tags    map[string]string
	Data    map[string]interface{}
}

func NewEngine(script string, callbacks map[string]FuncCallback, check map[string]FuncCallbackCheck, debug bool) (*Engine, error) {
	node, err := ParsePipeline(script)
	if err != nil {
		return nil, err
	}

	stmts, ok := node.(Stmts)
	if !ok {
		return nil, fmt.Errorf("invalid AST, should not been here")
	}
	ng := &Engine{
		debugMode: debug,
		output: &Output{
			Tags: make(map[string]string),
			Data: make(map[string]interface{}),
			Cost: make(map[string]string),
		},
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
	ng.ts = time.Now()
	ng.content = input
	ng.output.Data["message"] = input
	ng.stopRunPP = false
	ng.stmts.Run(ng)
	if ng.debugMode {
		ng.output.Cost["script-total"] = time.Since(ng.ts).String()
	}
	return ng.lastErr
}

func (ng *Engine) Result() *Output {
	for k, v := range ng.output.Data {
		switch v.(type) {
		case int, uint64, uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
		default:
			str, err := json.Marshal(v)
			if err != nil {
				log.Errorf("object type marshal error %v", err)
			}
			ng.output.Data[k] = string(str)
		}
	}
	return ng.output
}

func (ng *Engine) LastErr() error {
	return ng.lastErr
}

func (ng *Engine) reset() {
	ng.output = &Output{
		Tags: make(map[string]string),
		Data: make(map[string]interface{}),
		Cost: make(map[string]string),
	}
	ng.ts = time.Now()
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

	if v, ok := ng.output.Tags[k]; ok {
		return v, nil
	}
	v, ok := ng.output.Data[k]
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

	checkOutPutNilPtr(&ng.output)

	ng.output.Data[k] = v
}

func (ng *Engine) MarkDrop() {
	ng.output.Dropped = true
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

func checkOutPutNilPtr(outptr **Output) {
	if *outptr == nil {
		*outptr = &Output{
			Tags: make(map[string]string),
			Data: make(map[string]interface{}),
		}
	}
	if (*outptr).Data == nil {
		(*outptr).Data = make(map[string]interface{})
	}
	if (*outptr).Tags == nil {
		(*outptr).Tags = make(map[string]string)
	}
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

	checkOutPutNilPtr(&ng.output)

	if v == nil {
		return nil
	}

	if _, ok := ng.output.Tags[key]; ok {
		var value string
		switch v := v.(type) {
		case int, uint64, uint32, uint16, uint8, int64, int32, int16, int8, bool, float32, float64:
			value = conv.ToString(v)
		case string:
			value = v
		}
		ng.output.Tags[key] = value
	} else {
		ng.output.Data[key] = v
	}
	return nil
}

func (ng *Engine) SetTag(k interface{}, v string) error {
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
	checkOutPutNilPtr(&ng.output)

	delete(ng.output.Data, key)

	ng.output.Tags[key] = v

	return nil
}

func (ng *Engine) IsTag(k interface{}) bool {
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
		return false
	}
	if _, ok := ng.output.Tags[key]; ok {
		return true
	}
	return false
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

	if _, ok := ng.output.Tags[key]; ok {
		delete(ng.output.Tags, key)
	} else {
		delete(ng.output.Data, key)
	}
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
	if ng.lastErr != nil || ng.stopRunPP {
		return
	}
	for _, stmt := range e {
		switch v := stmt.(type) {
		case *IfelseStmt:
			v.Run(ng)
		case *FuncStmt:
			v.Run(ng)
			if v.Name == "exit" {
				ng.stopRunPP = true
			}

		case *AssignmentStmt:
			v.Run(ng)
		case Stmts:
			v.Run(ng)
		default:
			ng.lastErr = fmt.Errorf("unsupported Stmts type %s, from: %s", reflect.TypeOf(v), stmt)
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
		return false
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
		return false
	}

	switch v := e.Condition.(type) {
	case *ParenExpr:
		pass = v.Run(ng)
	case *ConditionalExpr:
		pass = v.Run(ng)
	case *BoolLiteral:
		pass = v.Val
	default:
		ng.lastErr = fmt.Errorf("unsupported IfExpr type %s, from: %s", reflect.TypeOf(v), e.Condition)
		return false
	}

	if pass {
		e.Stmts.Run(ng)
	}

	return
}

func (e *ConditionalExpr) Run(ng *Engine) (pass bool) {
	if ng.lastErr != nil {
		return false
	}

	// TODO
	// add 'Lazy Evaluation' to ConditionalExpr contrast

	var left, right interface{}

	switch v := e.LHS.(type) {
	case *Identifier:
		left = ng.output.Data[v.Name] // left maybe nil
	case *ParenExpr:
		left = v.Run(ng)
	case *ConditionalExpr:
		left = v.Run(ng)
	case *StringLiteral:
		left = v.Value()
	case *NumberLiteral:
		left = v.Value()
	case *BoolLiteral:
		left = v.Value()
	case *NilLiteral:
		left = v.Value()
	default:
		ng.lastErr = fmt.Errorf("unsupported ConditionalExpr type %s, from: %s", reflect.TypeOf(v), e.LHS)
		return false
	}

	switch v := e.RHS.(type) {
	case *Identifier:
		right = ng.output.Data[v.Name] // right maybe nil
	case *ParenExpr:
		right = v.Run(ng)
	case *ConditionalExpr:
		right = v.Run(ng)
	case *StringLiteral:
		right = v.Value()
	case *NumberLiteral:
		right = v.Value()
	case *BoolLiteral:
		right = v.Value()
	case *NilLiteral:
		right = v.Value()
	default:
		ng.lastErr = fmt.Errorf("unsupported ConditionalExpr type %s, from: %s", reflect.TypeOf(v), e.RHS)
		return false
	}

	if ng.lastErr != nil {
		return false
	}

	p, err := contrast(left, e.Op.String(), right)
	if err != nil {
		ng.lastErr = fmt.Errorf("failed to contrast, err: %w", err)
		return false
	}
	return p
}

func (e *ParenExpr) Run(ng *Engine) (pass bool) {
	if ng.lastErr != nil {
		return false
	}

	switch v := e.Param.(type) {
	case *ParenExpr:
		pass = v.Run(ng)
	case *ConditionalExpr:
		pass = v.Run(ng)
	case *BoolLiteral:
		pass = v.Val
	default:
		ng.lastErr = fmt.Errorf("unsupported ParenExpr type %s, from: %s", reflect.TypeOf(v), e.Param)
		return
	}
	return
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
			ng.output.Data[v.Name] = vv.Value()
		case *NumberLiteral:
			ng.output.Data[v.Name] = vv.Value()
		case *BoolLiteral:
			ng.output.Data[v.Name] = vv.Value()
		default:
			ng.lastErr = fmt.Errorf("unsupported AssignmentStmt type %s, from: %s", reflect.TypeOf(vv), e.RHS)
		}
	default:
		ng.lastErr = fmt.Errorf("unsupported AssignmentStmt type %s, from: %s", reflect.TypeOf(v), e.LHS)
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

func (e *NilLiteral) Value() interface{} { return nil }

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
		return fmt.Errorf(`unsupported AssignmentStmt type %s, from: %s`,
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
// contrast 数值比较
// 支持类型 int64, float64, json.Number, booler, string, nil  支持符号 < <= == != >= >
// 如果类型不一致，一定是 false，比如 int64 和 float64 比较
// 如果是 json.Number 类型，会先取其 float64 值，再进行 < <= > >= 比较
func contrast(left interface{}, op string, right interface{}) (b bool, err error) {
	var (
		float   []float64
		integer []int64
		booler  []bool
		typeErr = fmt.Errorf(`invalid operation: %s %s %s (mismatched types untyped %s and untyped %s)`,
			left, op, right, reflect.TypeOf(left), reflect.TypeOf(right))
	)

	// all value compared to nil is acceptable:
	//   if 10 == nil
	//   if "abc" == nil
	//   ...
	if right != nil && left != nil {
		if reflect.TypeOf(left) != reflect.TypeOf(right) {
			err = typeErr
			return
		}
	}

	switch op {
	case "==":
		b = reflect.DeepEqual(left, right)
		return
	case "!=":
		b = !reflect.DeepEqual(left, right)
		return
	}

	switch x := left.(type) {
	case json.Number:
		xnum, _err := x.Float64()
		if err != nil {
			err = fmt.Errorf("trans json.Number(%s) err, %w", x, _err)
			return
		}
		float = append(float, xnum)

		switch y := right.(type) {
		case json.Number:
			ynum, _err := y.Float64()
			if err != nil {
				err = fmt.Errorf("trans json.Number(%s) err, %w", y, _err)
				return
			}
			float = append(float, ynum)
		case float64:
			float = append(float, y)
		case nil:
			return
		default:
			err = typeErr
			return
		}

	case int64:
		switch y := right.(type) {
		case int64:
			integer = append(integer, x)
			integer = append(integer, y)
		case nil:
			return
		default:
			err = typeErr
			return
		}

	case float64:
		switch y := right.(type) {
		case float64:
			float = append(float, x)
			float = append(float, y)
		case nil:
			return
		default:
			err = typeErr
			return
		}

	case bool:
		switch y := right.(type) {
		case bool:
			booler = append(booler, x)
			booler = append(booler, y)
		case nil:
			return
		default:
			err = typeErr
			return
		}

	case string, nil:
		return

	default:
		err = typeErr
		return
	}

	switch op {
	case "&&":
		if len(booler) == 2 {
			b = booler[0] && booler[1]
			return
		}
	case "||":
		if len(booler) == 2 {
			b = booler[0] || booler[1]
			return
		}
	case "<=":
		if len(float) == 2 {
			b = float[0] <= float[1]
			return
		}
		if len(integer) == 2 {
			b = integer[0] <= integer[1]
			return
		}
	case "<":
		if len(float) == 2 {
			b = float[0] < float[1]
			return
		}
		if len(integer) == 2 {
			b = integer[0] < integer[1]
			return
		}
	case ">=":
		if len(float) == 2 {
			b = float[0] >= float[1]
			return
		}
		if len(integer) == 2 {
			b = integer[0] >= integer[1]
			return
		}
	case ">":
		if len(float) == 2 {
			b = float[0] > float[1]
			return
		}
		if len(integer) == 2 {
			b = integer[0] > integer[1]
			return
		}
	default:
		err = fmt.Errorf("unexpected operator %s", op)
		return
	}

	err = fmt.Errorf("the operator is not available for this type, %s", typeErr.Error())
	return
}
