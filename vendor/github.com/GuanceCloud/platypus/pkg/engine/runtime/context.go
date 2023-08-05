// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/spf13/cast"
)

const (
	ploriginkey    = "message"
	PlRunInfoField = "pl_msg"
)

type Script struct {
	CallRef []*ast.CallExpr

	FuncCall map[string]FuncCall

	Name      string
	Namespace string
	Category  string
	FilePath  string

	Content string

	Ast ast.Stmts
}

type Signal interface {
	ExitSignal() bool
}

type Context struct {
	Regs PlReg

	stackHeader *PlProcStack
	stackCur    *PlProcStack

	funcCall  map[string]FuncCall
	funcCheck map[string]FuncCheck

	inType       InType
	inRMap       InputWithRMap
	inWithoutMap InputWithoutMap

	// for 循环结束后需要清理此标志
	loopBreak    bool
	loopContinue bool

	signal Signal

	procExit bool

	callRCef []*ast.CallExpr

	name    string
	content string
	// namespace string
	// filepath  string
}

func (ctx *Context) Name() string {
	return ctx.name
}

func (ctx *Context) InData() any {
	switch ctx.inType {
	case InRMap:
		return ctx.inRMap
	default:
		return ctx.inWithoutMap
	}
}

func (ctx *Context) Signal() Signal {
	return ctx.signal
}

func InitCtxWithoutMap(ctx *Context, inWithoutMap InputWithoutMap, funcs map[string]FuncCall,
	callRef []*ast.CallExpr, signal Signal, name, content string,
) *Context {
	ctx.Regs.Reset()

	ctx.inType = InWithoutMap
	ctx.inWithoutMap = inWithoutMap

	ctx.funcCall = funcs
	ctx.funcCheck = nil

	ctx.callRCef = callRef
	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.signal = signal
	ctx.procExit = false

	ctx.name = name
	ctx.content = content

	return ctx
}

func InitCtxWithRMap(ctx *Context, inWithRMap InputWithRMap, funcs map[string]FuncCall,
	callRef []*ast.CallExpr, signal Signal, name, content string,
) *Context {
	ctx.Regs.Reset()

	ctx.inType = InRMap
	ctx.inRMap = inWithRMap

	ctx.funcCall = funcs
	ctx.funcCheck = nil

	ctx.callRCef = callRef
	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.signal = signal
	ctx.procExit = false

	ctx.name = name
	ctx.content = content

	return ctx
}

func InitCtxForCheck(ctx *Context, funcs map[string]FuncCall, funcsCheck map[string]FuncCheck,
	name, content string,
) *Context {
	ctx.stackHeader = &PlProcStack{
		data: map[string]*Varb{},
	}
	ctx.stackCur = ctx.stackHeader

	ctx.Regs.Reset()

	ctx.funcCall = funcs
	ctx.funcCheck = funcsCheck

	ctx.callRCef = []*ast.CallExpr{}
	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.procExit = false

	ctx.name = name
	ctx.content = content

	return ctx
}

func (ctx *Context) SetVarb(key string, value any, dtype ast.DType) error {
	if key == "_" {
		key = ploriginkey
	}

	ctx.stackCur.Set(key, value, dtype)
	return nil
}

func (ctx *Context) SetExit() {
	ctx.procExit = true
}

func (ctx *Context) SetCallRef(expr *ast.CallExpr) {
	if ctx.callRCef == nil {
		ctx.callRCef = []*ast.CallExpr{}
	}
	ctx.callRCef = append(ctx.callRCef, expr)
}

func (ctx *Context) GetKey(key string) (*Varb, error) {
	if key == "_" {
		key = ploriginkey
	}
	if v, err := ctx.stackCur.Get(key); err == nil {
		return v, nil
	}

	switch ctx.inType {
	case InRMap:
		if v, t, err := ctx.inRMap.Get(key); err == nil {
			return &Varb{
				Value: v,
				DType: t,
			}, nil
		}
	}

	return nil, fmt.Errorf("key not found")
}

func (ctx *Context) GetKeyConv2Str(key string) (string, error) {
	if key == "_" {
		key = ploriginkey
	}

	if v, err := ctx.stackCur.Get(key); err == nil {
		return Conv2String(v.Value, v.DType)
	}

	switch ctx.inType {
	case InRMap:
		if v, t, err := ctx.inRMap.Get(key); err == nil {
			return Conv2String(v, t)
		}
	}

	return "", fmt.Errorf("nil")
}

func (ctx *Context) GetFuncCall(key string) (FuncCall, bool) {
	if ctx.funcCall == nil {
		return nil, false
	}
	v, ok := ctx.funcCall[key]
	return v, ok
}

func (ctx *Context) GetFuncCheck(key string) (FuncCheck, bool) {
	if ctx.funcCheck == nil {
		return nil, false
	}
	v, ok := ctx.funcCheck[key]
	return v, ok
}

func (ctx *Context) StackEnterNew() {
	next := &PlProcStack{
		data:   map[string]*Varb{},
		before: ctx.stackCur,
	}

	ctx.stackCur = next
}

func (ctx *Context) StackExitCur() {
	ctx.stackCur.data = nil
	ctx.stackCur.checkPattern = nil

	ctx.stackCur = ctx.stackCur.before
}

func (ctx *Context) StackClear() {
	ctx.stackCur.Clear()
}

func (ctx *Context) GetPattern(pattern string) (*grok.GrokPattern, bool) {
	v, ok := ctx.stackCur.GetPattern(pattern)
	if ok {
		return v, ok
	}

	v, ok = DenormalizedGlobalPatterns[pattern]
	if ok {
		return v, ok
	}

	return nil, false
}

func (ctx *Context) SetPattern(patternAlias string, gPattern *grok.GrokPattern) {
	ctx.stackCur.SetPattern(patternAlias, gPattern)
}

func (ctx *Context) StmtRetrun() bool {
	if ctx.ProcExit() || ctx.loopBreak || ctx.loopContinue {
		return true
	}
	return false
}

func (ctx *Context) ProcExit() bool {
	if !ctx.procExit && ctx.signal != nil {
		if ctx.signal.ExitSignal() {
			ctx.procExit = true
		}
	}
	return ctx.procExit
}

var ctxPool sync.Pool = sync.Pool{
	New: func() any {
		return &Context{}
	},
}

func GetContext() *Context {
	ctx, _ := ctxPool.Get().(*Context)

	ctx.stackHeader = &PlProcStack{
		data: map[string]*Varb{},
	}
	ctx.stackCur = ctx.stackHeader
	return ctx
}

func PutContext(ctx *Context) {
	ctx.stackHeader = nil
	ctx.stackCur = nil

	ctx.funcCall = nil
	ctx.funcCheck = nil

	ctx.inRMap = nil
	ctx.inRMap = nil
	ctx.inWithoutMap = nil
	ctx.inType = InNoSet

	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.procExit = false

	ctx.callRCef = nil

	ctxPool.Put(ctx)
}

func Conv2String(v any, dtype ast.DType) (string, error) {
	switch dtype { //nolint:exhaustive
	case ast.Int, ast.Float, ast.Bool, ast.String:
		return cast.ToString(v), nil
	case ast.List, ast.Map:
		res, err := json.Marshal(v)
		return string(res), err
	case ast.Nil:
		return "", nil
	default:
		return "", fmt.Errorf("unsupported data type %d", dtype)
	}
}
