// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package runtime

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/GuanceCloud/grok"
	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/spf13/cast"
)

const (
	ploriginkey    = "message"
	PlRunInfoField = "pl_msg"
)

type Task struct {
	Regs PlReg

	stackHeader *PlProcStack
	stackCur    *PlProcStack

	funcCall  map[string]FuncCall
	funcCheck map[string]FuncCheck

	input Input

	// for 循环结束后需要清理此标志
	loopBreak    bool
	loopContinue bool

	signal Signal

	procExit bool

	callRef []*ast.CallExpr

	name string

	withValue map[any]any
}

func (ctx *Task) Name() string {
	return ctx.name
}

var (
	ErrNilKey           = errors.New("key is nil")
	ErrKeyNotComparable = errors.New("key is not comparable")
	ErrKeyExists        = errors.New("key exists")
)

func (ctx *Task) WithVal(key, val any, force bool) error {
	if key == nil {
		return ErrNilKey
	}
	if !reflect.TypeOf(key).Comparable() {
		return ErrKeyNotComparable
	}
	if ctx.withValue == nil {
		ctx.withValue = map[any]any{
			key: val,
		}
		return nil
	}

	if _, ok := ctx.withValue[key]; !ok {
		ctx.withValue[key] = val
	} else {
		if force {
			ctx.withValue[key] = val
		} else {
			return ErrKeyExists
		}
	}

	return nil
}

func (ctx *Task) Val(key any) any {
	if ctx.withValue == nil {
		return nil
	}
	return ctx.withValue[key]
}

func (ctx *Task) InData() any {
	return ctx.input
}

func (ctx *Task) Signal() Signal {
	return ctx.signal
}

func InitCtx(ctx *Task, input Input, script *Script, signal Signal) *Task {
	ctx.Regs.Reset()

	ctx.input = input

	ctx.funcCall = script.FuncCall
	ctx.funcCheck = nil

	ctx.callRef = script.CallRef
	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.signal = signal
	ctx.procExit = false

	ctx.name = script.Name

	return ctx
}

func InitCtxForCheck(ctx *Task, script *Script, checkFn map[string]FuncCheck) *Task {
	ctx.stackHeader = &PlProcStack{
		data: map[string]*Varb{},
	}
	ctx.stackCur = ctx.stackHeader

	ctx.Regs.Reset()

	ctx.funcCall = script.FuncCall
	ctx.funcCheck = checkFn

	ctx.callRef = []*ast.CallExpr{}
	ctx.loopBreak = false
	ctx.loopContinue = false

	ctx.procExit = false

	ctx.name = script.Name
	return ctx
}

func (ctx *Task) SetVarb(key string, value any, dtype ast.DType) error {
	if key == "_" {
		key = ploriginkey
	}

	ctx.stackCur.Set(key, value, dtype)
	return nil
}

func (ctx *Task) SetExit() {
	ctx.procExit = true
}

func (ctx *Task) SetCallRef(expr *ast.CallExpr) {
	if ctx.callRef == nil {
		ctx.callRef = []*ast.CallExpr{}
	}
	ctx.callRef = append(ctx.callRef, expr)
}

func (ctx *Task) GetKey(key string) (*Varb, error) {
	if key == "_" {
		key = ploriginkey
	}
	if v, err := ctx.stackCur.Get(key); err == nil {
		return v, nil
	}

	if v, t, err := ctx.input.Get(key); err == nil {
		return &Varb{
			Value: v,
			DType: t,
		}, nil
	}

	return nil, fmt.Errorf("key not found")
}

func (ctx *Task) GetKeyConv2Str(key string) (string, error) {
	if key == "_" {
		key = ploriginkey
	}

	if v, err := ctx.stackCur.Get(key); err == nil {
		return Conv2String(v.Value, v.DType)
	}

	if v, t, err := ctx.input.Get(key); err == nil {
		return Conv2String(v, t)
	}

	return "", fmt.Errorf("nil")
}

func (ctx *Task) GetFuncCall(key string) (FuncCall, bool) {
	if ctx.funcCall == nil {
		return nil, false
	}
	v, ok := ctx.funcCall[key]
	return v, ok
}

func (ctx *Task) GetFuncCheck(key string) (FuncCheck, bool) {
	if ctx.funcCheck == nil {
		return nil, false
	}
	v, ok := ctx.funcCheck[key]
	return v, ok
}

func (ctx *Task) StackEnterNew() {
	next := &PlProcStack{
		data:   map[string]*Varb{},
		before: ctx.stackCur,
	}

	ctx.stackCur = next
}

func (ctx *Task) StackExitCur() {
	ctx.stackCur.data = nil
	ctx.stackCur.checkPattern = nil

	ctx.stackCur = ctx.stackCur.before
}

func (ctx *Task) StackClear() {
	ctx.stackCur.Clear()
}

func (ctx *Task) GetPattern(pattern string) (*grok.GrokPattern, bool) {
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

func (ctx *Task) SetPattern(patternAlias string, gPattern *grok.GrokPattern) {
	ctx.stackCur.SetPattern(patternAlias, gPattern)
}

func (ctx *Task) StmtRetrun() bool {
	if ctx.ProcExit() || ctx.loopBreak || ctx.loopContinue {
		return true
	}
	return false
}

func (ctx *Task) ProcExit() bool {
	if !ctx.procExit && ctx.signal != nil {
		if ctx.signal.ExitSignal() {
			ctx.procExit = true
		}
	}
	return ctx.procExit
}

var ctxPool sync.Pool = sync.Pool{
	New: func() any {
		return &Task{}
	},
}

func GetContext() *Task {
	ctx, _ := ctxPool.Get().(*Task)

	ctx.stackHeader = &PlProcStack{
		data: map[string]*Varb{},
	}
	ctx.stackCur = ctx.stackHeader
	return ctx
}

func PutContext(ctx *Task) {
	*ctx = Task{}
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
