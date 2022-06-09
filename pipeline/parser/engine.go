// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"sync"
	"time"

	conv "github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
)

const (
	PipelineTimeField = "time"
)

var ngDataPool = sync.Pool{
	New: func() interface{} {
		return &EngineData{
			grokPatternStack: make([]map[string]*grok.GrokPattern, 0),
			grokPatternIndex: make([]int, 0),
			ts:               time.Now(),
		}
	},
}

func getNgData() *EngineData {
	data, _ := ngDataPool.Get().(*EngineData)
	return data
}

func putNgData(ngData *EngineData) {
	ngData.Reset()
	ngDataPool.Put(ngData)
}

type phase string

const (
	MakePhase phase = "make"
	OffPhase  phase = "off"
)

type (
	FuncCallback      func(*EngineData, Node) interface{}
	FuncCallbackCheck func(*EngineData, Node) error
)

type Engine struct {
	grok *grok.Grok

	callbacks     map[string]FuncCallback
	callbackCheck map[string]FuncCallbackCheck
	stmts         Stmts
}

type EngineData struct {
	output *Output

	contentKey string

	OnlyForCheckFunc **grok.Grok
	grokPatternStack []map[string]*grok.GrokPattern
	grokPatternIndex []int
	stackDeep        int

	stopRunPL bool // stop run()

	ts      time.Time
	lastErr error
}

func (ngData *EngineData) Reset() {
	ngData.output = nil
	ngData.contentKey = ""

	ngData.OnlyForCheckFunc = nil
	if len(ngData.grokPatternStack) > 0 {
		ngData.grokPatternStack = ngData.grokPatternStack[:0]
	} else if ngData.grokPatternStack == nil {
		ngData.grokPatternStack = make([]map[string]*grok.GrokPattern, 0)
	}
	if len(ngData.grokPatternIndex) > 0 {
		ngData.grokPatternIndex = ngData.grokPatternIndex[:0]
	} else if ngData.grokPatternIndex == nil {
		ngData.grokPatternIndex = make([]int, 0)
	}
	ngData.stackDeep = 0

	ngData.stopRunPL = false

	ngData.lastErr = nil
}

//nolint:structcheck,unused
type Output struct {
	Error error

	Drop bool

	Measurement string
	Time        time.Time

	Tags   map[string]string
	Fields map[string]interface{}
}

func NewOutput() *Output {
	return &Output{
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
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
		grok: &grok.Grok{
			GlobalDenormalizedPatterns: DenormalizedGlobalPatterns,
			DenormalizedPatterns:       make(map[string]*grok.GrokPattern),
			CompliedGrokRe:             make(map[string]map[string]*grok.GrokRegexp),
		},
		callbackCheck: check,
		callbacks:     callbacks,
		stmts:         stmts,
	}
	if err := ng.Check(); err != nil {
		return nil, err
	}
	return ng, nil
}

func (ng *Engine) Check() error {
	data := &EngineData{
		output:           NewOutput(),
		grokPatternStack: make([]map[string]*grok.GrokPattern, 0),
		grokPatternIndex: make([]int, 0),
		OnlyForCheckFunc: &ng.grok,
	}

	return ng.stmts.Check(ng, data)
}

func (ng *Engine) Run(measurement string, tags map[string]string, fields map[string]interface{},
	contentKey string, rTime time.Time) (*Output, error) {
	data := getNgData()
	defer putNgData(data)

	if rTime.IsZero() {
		rTime = time.Now()
	}

	if tags == nil {
		tags = map[string]string{}
	}
	if fields == nil {
		fields = map[string]interface{}{}
	}

	data.contentKey = contentKey

	data.output = &Output{
		Fields:      fields,
		Tags:        tags,
		Measurement: measurement,
		Time:        rTime,
	}

	data.stopRunPL = false
	ng.stmts.Run(ng, data)

	return result(data), nil
}

func result(data *EngineData) *Output {
	for k, v := range data.output.Fields {
		switch v.(type) {
		case int, uint64, uint32, uint16, uint8, int64, int32, int16, int8, bool, string, float32, float64:
		default:
			str, err := json.Marshal(v)
			if err != nil {
				log.Errorf("object type marshal error %v", err)
			}
			data.output.Fields[k] = string(str)
		}
	}
	if v, err := data.GetContent(PipelineTimeField); err == nil {
		if nanots, ok := v.(int64); ok {
			t := time.Unix(nanots/int64(time.Second),
				nanots%int64(time.Second))
			if !t.IsZero() {
				data.output.Time = t
			}
			_ = data.DeleteContent(PipelineTimeField)
		}
	}

	if data.lastErr != nil {
		data.output.Error = data.lastErr
	}
	return data.output
}

func (ngData *EngineData) GetContentStr(key interface{}) (string, error) {
	c, err := ngData.GetContent(key)
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

func (ngData *EngineData) GetContent(key interface{}) (interface{}, error) {
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
		k = ngData.contentKey
	}

	if v, ok := ngData.output.Tags[k]; ok {
		return v, nil
	}
	v, ok := ngData.output.Fields[k]
	if !ok {
		return nil, fmt.Errorf("%s no found", k)
	}

	return v, nil
}

func (ngData *EngineData) SetKey(k string, v interface{}) {
	if v == nil { // ignored
		return
	}

	checkOutputNilPtr(&ngData.output)

	if k == "_" {
		k = ngData.contentKey
	}
	delete(ngData.output.Tags, k)
	ngData.output.Fields[k] = v
}

func (ngData *EngineData) MarkDrop() {
	ngData.output.Drop = true
}

const (
	DefaultStr   = ""
	InvalidInt   = math.MinInt32 // error: MinInt64?
	DefaultInt   = int64(0xdeadbeef)
	InvalidStr   = "deadbeaf"
	InvalidFloat = math.SmallestNonzeroFloat64
)

func getStrArg(ngData *EngineData, node Node) (string, error) {
	switch v := node.(type) {
	case *StringLiteral:
		return v.Val, nil
	case *AttrExpr, *Identifier:
		return ngData.GetContentStr(v)
	default:
		return "", fmt.Errorf("invalid arg type %s(%s)",
			reflect.TypeOf(node).String(), node.String())
	}
}

func kwGetStrArg(ngData *EngineData, args map[string]Node, kw string) (string, error) {
	v, ok := args[kw]
	if !ok {
		return DefaultStr, nil
	}
	return getStrArg(ngData, v)
}

func getIntArg(ngData *EngineData, node Node) (int64, error) {
	str, err := getStrArg(ngData, node)
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

func kwGetIntArg(ngData *EngineData, args map[string]Node, kw string) (int64, error) {
	v, ok := args[kw]
	if !ok {
		return DefaultInt, nil
	}
	return getIntArg(ngData, v)
}

func GetFuncStrArg(ngData *EngineData, f *FuncStmt, idx int, kw string) (string, error) {
	if len(f.KwParam) > 0 {
		return kwGetStrArg(ngData, f.KwParam, kw)
	}

	if f.Param != nil {
		if idx >= len(f.Param) {
			return InvalidStr, fmt.Errorf("arg index out of range")
		}
		return getStrArg(ngData, f.Param[idx])
	}

	return InvalidStr, fmt.Errorf("no params available")
}

func GetFuncIntArg(ngData *EngineData, f *FuncStmt, idx int, kw string) (int64, error) {
	if len(f.KwParam) > 0 {
		return kwGetIntArg(ngData, f.KwParam, kw)
	}

	if f.Param != nil {
		if idx >= len(f.Param) {
			return InvalidInt, fmt.Errorf("arg index outof range")
		}
		return getIntArg(ngData, f.Param[idx])
	}

	return InvalidInt, fmt.Errorf("no params available")
}

func GetFuncFloatArg(ngData *EngineData, f *FuncStmt, idx int, kw string) (float64, error) {
	return InvalidFloat, fmt.Errorf("not implemented")
}

func (ngData *EngineData) GetEngineRGrok() (*grok.Grok, bool) {
	if ngData.OnlyForCheckFunc != nil && *ngData.OnlyForCheckFunc != nil {
		return *ngData.OnlyForCheckFunc, true
	}
	return nil, false
}

func (ngData *EngineData) SetEngineRGrok(grok *grok.Grok) bool {
	if ngData.OnlyForCheckFunc != nil {
		*ngData.OnlyForCheckFunc = grok
		return true
	}
	return false
}

func (ngData *EngineData) StackDeep() int {
	return ngData.stackDeep
}

func (ngData *EngineData) PatternStack() []map[string]*grok.GrokPattern {
	return ngData.grokPatternStack
}

func (ngData *EngineData) PatternIndex() string {
	idx := ""
	for _, v := range ngData.grokPatternIndex {
		idx = strconv.Itoa(v) + "," + idx
	}
	return idx
}

func checkOutputNilPtr(outptr **Output) {
	if *outptr == nil {
		*outptr = &Output{
			Tags:   make(map[string]string),
			Fields: make(map[string]interface{}),
		}
	}
	if (*outptr).Fields == nil {
		(*outptr).Fields = make(map[string]interface{})
	}
	if (*outptr).Tags == nil {
		(*outptr).Tags = make(map[string]string)
	}
}

func (ngData *EngineData) SetContent(k, v interface{}) error {
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

	checkOutputNilPtr(&ngData.output)

	if v == nil {
		return nil
	}

	if key == "_" {
		key = ngData.contentKey
	}

	if _, ok := ngData.output.Tags[key]; ok {
		var value string
		switch v := v.(type) {
		case int, uint64, uint32, uint16, uint8, int64, int32, int16, int8, bool, float32, float64:
			value = conv.ToString(v)
		case string:
			value = v
		}
		ngData.output.Tags[key] = value
	} else {
		ngData.output.Fields[key] = v
	}
	return nil
}

func (ngData *EngineData) SetMeasurement(v string) {
	ngData.output.Measurement = v
}

func (ngData *EngineData) SetTag(k interface{}, v string) error {
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
	checkOutputNilPtr(&ngData.output)

	if key == "_" {
		key = ngData.contentKey
	}
	delete(ngData.output.Fields, key)

	ngData.output.Tags[key] = v

	return nil
}

func (ngData *EngineData) IsTag(k interface{}) bool {
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

	if key == "_" {
		key = ngData.contentKey
	}
	if _, ok := ngData.output.Tags[key]; ok {
		return true
	}
	return false
}

func (ngData *EngineData) DeleteContent(k interface{}) error {
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

	if key == "_" {
		key = ngData.contentKey
	}

	if _, ok := ngData.output.Tags[key]; ok {
		delete(ngData.output.Tags, key)
	} else {
		delete(ngData.output.Fields, key)
	}
	return nil
}

///
// Runner
///

func (e Stmts) Run(ng *Engine, data *EngineData) {
	for _, stmt := range e {
		if data.lastErr != nil || data.stopRunPL {
			return
		}
		switch v := stmt.(type) {
		case *IfelseStmt:
			v.Run(ng, data)
		case *FuncStmt:
			v.Run(ng, data)
			if v.Name == "exit" {
				data.stopRunPL = true
			}

		case *AssignmentStmt:
			v.Run(ng, data)
		case Stmts:
			v.Run(ng, data)
		default:
			data.lastErr = fmt.Errorf("unsupported Stmts type %s, from: %s", reflect.TypeOf(v), stmt)
		}
	}
}

func (e *IfelseStmt) Run(ng *Engine, data *EngineData) {
	// data.stackDeep += 1
	// data.grokPatternIndex = append(data.grokPatternIndex, 0)
	// defer func() {
	// 	data.stackDeep -= 1
	// 	data.grokPatternIndex = data.grokPatternIndex[:data.stackDeep]
	// }()

	if data.lastErr != nil {
		return
	}

	if !e.IfList.Run(ng, data) {
		// data.grokPatternIndex[data.stackDeep-1] += 1
		e.Else.Run(ng, data)
	}
}

func (e IfList) Run(ng *Engine, data *EngineData) (end bool) {
	if data.lastErr != nil {
		return false
	}
	for _, ifexpr := range e {
		// data.grokPatternIndex[data.stackDeep-1] += 1
		end = ifexpr.Run(ng, data)
		if end {
			return
		}
	}
	return
}

func (e *IfExpr) Run(ng *Engine, data *EngineData) (pass bool) {
	if data.lastErr != nil {
		return false
	}

	switch v := e.Condition.(type) {
	case *ParenExpr:
		pass = v.Run(ng, data)
	case *ConditionalExpr:
		pass = v.Run(ng, data)
	case *BoolLiteral:
		pass = v.Val
	default:
		data.lastErr = fmt.Errorf("unsupported IfExpr type %s, from: %s", reflect.TypeOf(v), e.Condition)
		return false
	}

	if pass {
		e.Stmts.Run(ng, data)
	}

	return
}

func (e *ConditionalExpr) Run(ng *Engine, data *EngineData) (pass bool) {
	if data.lastErr != nil {
		return false
	}

	// TODO
	// add 'Lazy Evaluation' to ConditionalExpr contrast

	var left, right interface{}

	switch v := e.LHS.(type) {
	case *Identifier:
		left, _ = data.GetContent(v.Name) // left maybe nil
	case *ParenExpr:
		left = v.Run(ng, data)
	case *ConditionalExpr:
		left = v.Run(ng, data)
	case *StringLiteral:
		left = v.Value()
	case *NumberLiteral:
		left = v.Value()
	case *BoolLiteral:
		left = v.Value()
	case *NilLiteral:
		left = v.Value()
	// case *FuncStmt:
	// 	switch ret := v.Run(ng).(type) {
	// 	case error:
	// 		return false
	// 	default:
	// 		left = ret
	// 	}
	default:
		data.lastErr = fmt.Errorf("unsupported ConditionalExpr type %s, from: %s", reflect.TypeOf(v), e.LHS)
		return false
	}

	switch v := e.RHS.(type) {
	case *Identifier:
		right, _ = data.GetContent(v.Name) // right maybe nil
	case *ParenExpr:
		right = v.Run(ng, data)
	case *ConditionalExpr:
		right = v.Run(ng, data)
	case *StringLiteral:
		right = v.Value()
	case *NumberLiteral:
		right = v.Value()
	case *BoolLiteral:
		right = v.Value()
	case *NilLiteral:
		right = v.Value()
	default:
		data.lastErr = fmt.Errorf("unsupported ConditionalExpr type %s, from: %s", reflect.TypeOf(v), e.RHS)
		return false
	}

	if data.lastErr != nil {
		return false
	}

	p, err := contrast(left, e.Op.String(), right)
	if err != nil {
		data.lastErr = fmt.Errorf("failed to contrast, err: %w", err)
		return false
	}
	return p
}

func (e *ParenExpr) Run(ng *Engine, data *EngineData) (pass bool) {
	if data.lastErr != nil {
		return false
	}

	switch v := e.Param.(type) {
	case *ParenExpr:
		pass = v.Run(ng, data)
	case *ConditionalExpr:
		pass = v.Run(ng, data)
	case *BoolLiteral:
		pass = v.Val
	default:
		data.lastErr = fmt.Errorf("unsupported ParenExpr type %s, from: %s", reflect.TypeOf(v), e.Param)
		return
	}
	return
}

func (e *ComputationExpr) Run(ng *Engine) {
	// TODO
}

func (e *AssignmentStmt) Run(ng *Engine, data *EngineData) {
	if data.lastErr != nil {
		return
	}

	switch v := e.LHS.(type) {
	case *Identifier:
		switch vv := e.RHS.(type) {
		case *StringLiteral:
			_ = data.SetContent(v.Name, vv.Value())
		case *NumberLiteral:
			_ = data.SetContent(v.Name, vv.Value())
		case *BoolLiteral:
			_ = data.SetContent(v.Name, vv.Value())
		default:
			data.lastErr = fmt.Errorf("unsupported AssignmentStmt type %s, from: %s", reflect.TypeOf(vv), e.RHS)
		}
	default:
		data.lastErr = fmt.Errorf("unsupported AssignmentStmt type %s, from: %s", reflect.TypeOf(v), e.LHS)
	}
}

func (e *FuncStmt) Run(ng *Engine, data *EngineData) interface{} {
	if fn := ng.callbacks[e.Name]; fn == nil {
		data.lastErr = fmt.Errorf("unsupported func: `%v'", e.Name)
		return data.lastErr
	} else {
		switch ret := fn(data, e).(type) {
		case error:
			data.lastErr = fmt.Errorf("Run func %v: %w", e.Name, ret)
			return ret
		case nil:
			return nil
		default:
			return ret
		}
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
func (e Stmts) Check(ng *Engine, data *EngineData) error {
	for _, stmt := range e {
		switch v := stmt.(type) {
		case *IfelseStmt:
			if err := v.Check(ng, data); err != nil {
				return err
			}
		case *FuncStmt:
			if err := v.Check(ng, data); err != nil {
				return fmt.Errorf("func %s: %w", v.Name, err)
			}
		case *AssignmentStmt:
			if err := v.Check(); err != nil {
				return err
			}
		case Stmts:
			if err := v.Check(ng, data); err != nil {
				return err
			}
		default:
			return fmt.Errorf(`unsupported type %s, from: %s`,
				reflect.TypeOf(stmt), stmt)
		}
	}
	return nil
}

// Check IfExpr
//   Condition support BoolLiteral/ConditionalExpr
func (e *FuncStmt) Check(ng *Engine, data *EngineData) error {
	if _, ok := ng.callbacks[e.Name]; !ok {
		return fmt.Errorf("unsupported func: `%v'", e.Name)
	}

	checkFn, ok := ng.callbackCheck[e.Name]
	if !ok {
		return fmt.Errorf("not found check for func: `%v'", e.Name)
	}
	return checkFn(data, e)
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
	case *NumberLiteral, *StringLiteral, *BoolLiteral, *NilLiteral:
		// nil
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.RHS), e.RHS)
	}
	return nil
}

// Check IfelseStmt.
func (e *IfelseStmt) Check(ng *Engine, data *EngineData) error {
	data.stackDeep += 1
	data.grokPatternStack = append(data.grokPatternStack, map[string]*grok.GrokPattern{})
	data.grokPatternIndex = append(data.grokPatternIndex, 0)
	defer func() {
		data.stackDeep -= 1
		data.grokPatternStack = data.grokPatternStack[:data.stackDeep]
		data.grokPatternIndex = data.grokPatternIndex[:data.stackDeep]
	}()

	if err := e.IfList.Check(ng, data); err != nil {
		return err
	}

	data.grokPatternStack[data.stackDeep-1] = make(map[string]*grok.GrokPattern)
	data.grokPatternIndex[data.stackDeep-1] += 1
	return e.Else.Check(ng, data)
}

// Check IfList.
func (e IfList) Check(ng *Engine, data *EngineData) error {
	for _, i := range e {
		data.grokPatternStack[data.stackDeep-1] = make(map[string]*grok.GrokPattern)
		data.grokPatternIndex[data.stackDeep-1] += 1
		if err := i.Check(ng, data); err != nil {
			return err
		}
	}
	return nil
}

// Check IfExpr
//   Condition support BoolLiteral/ConditionalExpr
func (e *IfExpr) Check(ng *Engine, data *EngineData) error {
	switch v := e.Condition.(type) {
	case *ParenExpr:
		// nil
	case *BoolLiteral:
		// nil
	case *ConditionalExpr:
		if err := v.Check(ng); err != nil {
			return err
		}
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.Condition), e.Condition)
	}
	return e.Stmts.Check(ng, data)
}

// Check ConditionalExpr
//   left node only support Identifier
//   right node support NumberLiteral/StringLiteral/BoolLiteral
func (e *ConditionalExpr) Check(ng *Engine) error {
	switch e.LHS.(type) {
	case *Identifier:
	case *ParenExpr:
	case *ConditionalExpr:
	case *StringLiteral:
	case *NumberLiteral:
	case *BoolLiteral:
	case *NilLiteral:
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.LHS), e.LHS)
	}

	switch e.RHS.(type) {
	case *Identifier:
	case *ParenExpr:
	case *ConditionalExpr:
	case *StringLiteral:
	case *NumberLiteral:
	case *BoolLiteral:
	case *NilLiteral:
	default:
		return fmt.Errorf(`unsupported type %s, from: %s`,
			reflect.TypeOf(e.RHS), e.RHS)
	}
	return nil
}

// contrast 数值比较
// 支持类型 int64, float64, json.Number, booler, string, nil  支持符号 < <= == != >= >
// 如果类型不一致，一定是 false，比如 int64 和 float64 比较
// 如果是 json.Number 类型，会先取其 float64 值，再进行 < <= > >= 比较.
func contrast(left interface{}, op string, right interface{}) (b bool, err error) {
	var (
		float   []float64
		integer []int64
		booler  []bool
		typeErr = fmt.Errorf(`invalid operation: %s %s %s (mismatched types untyped %s and untyped %s)`,
			left, op, right, reflect.TypeOf(left), reflect.TypeOf(right))
	)

	// All value compared to nil is acceptable:
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

	return b, fmt.Errorf("the operator is not available for this type, %w", typeErr)
}
