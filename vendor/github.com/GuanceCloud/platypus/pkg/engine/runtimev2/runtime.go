package runtimev2

import (
	"fmt"

	"github.com/GuanceCloud/platypus/pkg/ast"
	"github.com/GuanceCloud/platypus/pkg/engine/runtime"
	"github.com/GuanceCloud/platypus/pkg/errchain"
	"github.com/GuanceCloud/platypus/pkg/token"
)

type Opt func(ctx *Task)

type FnCall func(ctx *Task, fn *ast.CallExpr) *errchain.PlError

type Script struct {
	Name  string
	Stmts ast.Stmts
	Fn    map[string]*Fn
}

func (s *Script) Run(signal Signal, opt ...Opt) *errchain.PlError {
	task := NewTask(s.Name, s.Fn)
	for _, o := range opt {
		if o != nil {
			o(task)
		}
	}
	if err := RunStmts(task, s.Stmts); err != nil {
		return err
	}
	return nil
}

func (s *Script) Check() *errchain.PlError {
	task := NewTask(s.Name, s.Fn)
	if err := RunStmtsCheck(task, &ContextCheck{}, s.Stmts); err != nil {
		return err
	}

	return nil
}

type Signal interface {
	ExitSignal() bool
}

func WithPrivate(v map[TaskP]any) Opt {
	return func(ctx *Task) {
		ctx.private = v
	}
}

type TaskP string

type Task struct {
	name    string
	private map[TaskP]any
	funcs   map[string]*Fn

	Regs        PlReg
	stackHeader *runtime.Stack
	stackCur    *runtime.Stack

	// for 循环结束后需要清理此标志
	loopBreak    bool
	loopContinue bool

	signal   Signal
	procExit bool
}

type PlReg struct {
	Val []V
}

func (reg *PlReg) Reset() {
	reg.Val = reg.Val[:0]
}

type V struct {
	V any
	T ast.DType
}

func (reg *PlReg) ReturnAppend(val ...V) {
	reg.Reset()
	reg.Val = append(reg.Val, val...)
}

func (reg *PlReg) Count() int {
	return len(reg.Val)
}

func (reg *PlReg) GetRet() (V, error) {
	switch {
	case len(reg.Val) == 1:
		return reg.Val[0], nil
	case len(reg.Val) == 0:
		return V{}, fmt.Errorf("no return value")
	default:
		return V{}, fmt.Errorf("there are multiple return values")
	}
}

func (reg *PlReg) GetMultiRet() ([]V, error) {
	switch {
	case len(reg.Val) > 1:
		return reg.Val, nil
	case len(reg.Val) == 0:
		return nil, fmt.Errorf("no return value")
	default:
		return nil, fmt.Errorf("only one return value")
	}
}

func (ctx *Task) SetExit() {
	ctx.procExit = true
}

func (ctx *Task) StackEnterNew() {
	next := &runtime.Stack{
		Data:   map[string]*runtime.Varb{},
		Before: ctx.stackCur,
	}

	ctx.stackCur = next
}

func (ctx *Task) PValue(k TaskP) (any, bool) {
	v, ok := ctx.private[k]
	return v, ok
}

func (ctx *Task) StackExitCur() {
	ctx.stackCur.Data = nil
	ctx.stackCur.CheckPattern = nil

	ctx.stackCur = ctx.stackCur.Before
}

func (ctx *Task) ProcExit() bool {
	if !ctx.procExit && ctx.signal != nil {
		if ctx.signal.ExitSignal() {
			ctx.procExit = true
		}
	}
	return ctx.procExit
}

func (ctx *Task) SetVarb(key string, v V) {
	ctx.stackCur.Set(key, v.V, v.T)
}

func (ctx *Task) GetKey(key string) (*runtime.Varb, error) {
	if v, err := ctx.stackCur.Get(key); err == nil {
		return v, nil
	}

	return nil, fmt.Errorf("key not found")
}

func (ctx *Task) GetFn(name string) (FnCall, bool) {
	if v, ok := ctx.funcs[name]; ok && v != nil && v.Call != nil {
		return v.Call, true
	}
	return nil, false
}

func (ctx *Task) GetFnCheck(name string) (FnCall, bool) {
	if v, ok := ctx.funcs[name]; ok && v != nil && v.CallCheck != nil {
		return v.CallCheck, true
	}
	return nil, false
}

func (ctx *Task) StmtRetrun() bool {
	if ctx.ProcExit() || ctx.loopBreak || ctx.loopContinue {
		return true
	}
	return false
}

func NewTask(name string, funcs map[string]*Fn) *Task {
	task := &Task{
		stackHeader: runtime.NewStack(),
		funcs:       funcs,
		name:        name,
	}
	task.StackEnterNew()
	return task
}

func NewRunError(ctx *Task, err string, pos token.LnColPos) *errchain.PlError {
	return errchain.NewErr(ctx.name, pos, err)
}
