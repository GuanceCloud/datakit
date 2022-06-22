package parser

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	conv "github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
)

type EngineData struct {
	output *Output

	contentKey string

	OnlyForCheckFunc **grok.Grok
	grokPatternStack []map[string]*grok.GrokPattern
	grokPatternIndex []int
	stackDeep        int

	callRef map[string]*Engine

	stopRunPL bool // stop run()

	ts      time.Time
	lastErr error
}

func (ngData *EngineData) SetCallRef(name string, ng *Engine) bool {
	if name == "" {
		return false
	}
	if ngData.callRef != nil {
		ngData.callRef[name] = ng
		return true
	}
	return false
}

func (ngData *EngineData) GetCallRef(name string) (*Engine, bool) {
	if ng, ok := ngData.callRef[name]; ok {
		return ng, ok
	}
	return nil, false
}

func (ngData *EngineData) Reset() {
	ngData.callRef = nil
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

func (ngData *EngineData) GetEngineRGrok() (*grok.Grok, bool) {
	if ngData.OnlyForCheckFunc != nil && *ngData.OnlyForCheckFunc != nil {
		return *ngData.OnlyForCheckFunc, true
	}
	return nil, false
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
