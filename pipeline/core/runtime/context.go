// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"fmt"
	"sync"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/core/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/pipeline/grok"
)

type (
	PtFlag  uint
	KeyKind uint
)

const (
	PtMeasurement PtFlag = iota
	PtTag
	PtField
	PtTFDefaulutOrKeep
	PtTime
)

const (
	KindPtDefault KeyKind = iota
	KindPtTag
)

const (
	ploriginkey    = "message"
	PlRunInfoField = "pl_msg"
)

type Script struct {
	CallRef map[string]*Script

	FuncCall map[string]FuncCall

	Name      string
	Namespace string
	Category  string
	FilePath  string

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

	pt *Point

	// for 循环结束后需要清理此标志
	loopBreak    bool
	loopContinue bool

	signal Signal

	procExit bool

	callRCef map[string]*Script
	// name      string
	// namespace string
	// filepath  string
}

func InitCtx(ctx *Context, pt *Point, funcs map[string]FuncCall,
	callRef map[string]*Script, signal Signal,
) *Context {
	ctx.Regs.Reset()
	ctx.pt = pt
	ctx.funcCall = funcs
	ctx.funcCheck = nil

	ctx.callRCef = callRef
	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.signal = signal
	ctx.procExit = false

	return ctx
}

func InitCtxForCheck(ctx *Context, funcs map[string]FuncCall, funcsCheck map[string]FuncCheck) *Context {
	ctx.stackHeader = &PlProcStack{
		data: map[string]*Varb{},
	}
	ctx.stackCur = ctx.stackHeader

	ctx.Regs.Reset()

	ctx.funcCall = funcs
	ctx.funcCheck = funcsCheck

	ctx.callRCef = map[string]*Script{}
	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.procExit = false

	return ctx
}

func (ctx *Context) Point() (*Point, bool) {
	if ctx.pt != nil {
		return ctx.pt, true
	}

	return nil, false
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

func (ctx *Context) GetCallRef(name string) (*Script, bool) {
	if ctx.callRCef == nil {
		return nil, false
	}
	if v, ok := ctx.callRCef[name]; ok && v != nil {
		return v, ok
	}
	return nil, false
}

func (ctx *Context) SetCallRef(name string) {
	if ctx.callRCef == nil {
		ctx.callRCef = map[string]*Script{}
	}
	ctx.callRCef[name] = nil
}

func (ctx *Context) MarkPtDrop() {
	ctx.pt.Drop = true
}

func (ctx *Context) PtDropped() bool {
	return ctx.pt.Drop
}

func (ctx *Context) SetMeasurement(name string) {
	ctx.pt.Measurement = name
}

func (ctx *Context) RenamePtKey(to string, from string) {
	if to == "_" {
		to = ploriginkey
	}

	if from == "_" {
		from = ploriginkey
	}

	v, ok := ctx.pt.Meta[from]
	if !ok {
		return
	}

	delete(ctx.pt.Meta, from)
	ctx.pt.Meta[to] = v

	switch v.PtFlag { //nolint:exhaustive
	case PtField:
		if v, ok := ctx.pt.Fields[from]; ok {
			ctx.pt.Fields[to] = v
		}
		delete(ctx.pt.Fields, from)
	case PtTag:
		if v, ok := ctx.pt.Tags[from]; ok {
			ctx.pt.Tags[to] = v
		}
		delete(ctx.pt.Tags, from)
	}
}

func (ctx *Context) AddKey2PtWithVal(key string, value any, dtype ast.DType, kind KeyKind) error {
	if key == "_" {
		key = ploriginkey
	}

	switch kind { //nolint:exhaustive
	case KindPtDefault:
		return ctx.pt.Set(key, value, dtype)
	default:
		return ctx.pt.SetTag(key, value, dtype)
	}
}

func (ctx *Context) AddKey2Pt(key string, kind KeyKind) error {
	if key == "_" {
		key = ploriginkey
	}

	switch kind { //nolint:exhaustive
	case KindPtDefault:
		if v, err := ctx.GetKey(key); err == nil {
			return ctx.pt.Set(key, v.Value, v.DType)
		} else {
			return ctx.pt.Set(key, nil, ast.Nil)
		}
	default:
		if v, err := ctx.GetKey(key); err == nil {
			return ctx.pt.SetTag(key, v.Value, v.DType)
		} else {
			return ctx.pt.SetTag(key, nil, ast.Nil)
		}
	}
}

func (ctx *Context) PointTime() int64 {
	if ctx.pt.Time.IsZero() {
		return time.Now().UnixNano()
	} else {
		return ctx.pt.Time.UnixNano()
	}
}

func (ctx *Context) DeleteKeyPt(key string) {
	if key == "_" {
		key = ploriginkey
	}
	ctx.pt.Delete(key)
}

func (ctx *Context) GetKeyFromPt(key string) (any, ast.DType, error) {
	return ctx.pt.Get(key)
}

func (ctx *Context) GetKey(key string) (*Varb, error) {
	if key == "_" {
		key = ploriginkey
	}
	if v, err := ctx.stackCur.Get(key); err == nil {
		return v, nil
	}

	if v, t, err := ctx.pt.Get(key); err == nil {
		return &Varb{
			Name:  key,
			Value: v,
			DType: t,
		}, nil
	}

	return nil, fmt.Errorf("key not found")
}

func (ctx *Context) GetKeyConv2Str(key string) (string, error) {
	if key == "_" {
		key = ploriginkey
	}

	if v, err := ctx.stackCur.Get(key); err == nil {
		return conv2String(v.Value, v.DType)
	}

	if v, t, err := ctx.pt.Get(key); err == nil {
		return conv2String(v, t)
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

	ctx.pt = nil

	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.procExit = false

	ctx.callRCef = nil

	ctxPool.Put(ctx)
}
