package luascript

import (
	"fmt"
	"io/ioutil"
	"sync"

	lua "github.com/yuin/gopher-lua"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/luascript/module"
)

// LuLuaData LuaScript所接收的数据接口
type LuaData interface {

	// 主要执行函数，接收从lua中返回的数据的json字符串，和执行过程中出现的error
	Handle(value string, err error)

	// 要发送给lua的数据，建议是基础类型，如整型、浮点型、map和slice等
	DataToLua() interface{}

	// 数据要执行的lua分组名称
	Name() string

	// lua执行函数的函数名
	CallbackFnName() string

	// lua执行函数的唯一形参，即DataToLua数据在lua中的具体表现名称
	CallbackTypeName() string
}

type LuaScript struct {
	// 代码组，每个组名有N个代码块
	codes map[string][]string

	workerNum int

	dataChan chan LuaData

	// 注册lua模块所需，每个LuaScript共享同一份luaCache
	luaCache *module.LuaCache

	// 退出广播
	exit *cliutils.Sem

	// 是否处于运行状态
	runStatus bool

	wg sync.WaitGroup
}

func NewLuaScript(workerNum int) *LuaScript {
	return &LuaScript{
		codes:     make(map[string][]string),
		workerNum: workerNum,
		dataChan:  make(chan LuaData, workerNum*2),
		luaCache:  globalLuaCache,
		runStatus: false,
		wg:        sync.WaitGroup{},
	}
}

func (s *LuaScript) AddLuaCodes(name string, codes []string) error {
	if _, ok := s.codes[name]; ok {
		return fmt.Errorf("the %s runner line already exist", name)
	}

	for _, code := range codes {
		if err := CheckLuaCode(code); err != nil {
			return err
		}
	}
	s.codes[name] = codes
	return nil
}

func (s *LuaScript) AddLuaCodesFromFile(name string, filePath []string) error {
	if _, ok := s.codes[name]; ok {
		return fmt.Errorf("the %s runner line already exist", name)
	}

	codes := []string{}

	for _, file := range filePath {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}

		code := string(content)

		if err := CheckLuaCode(code); err != nil {
			return err
		}

		codes = append(codes, code)
	}
	s.codes[name] = codes
	return nil
}

func (s *LuaScript) Run() {
	if s.runStatus {
		return
	}

	s.exit = cliutils.NewSem()

	for i := 0; i < s.workerNum; i++ {
		s.wg.Add(1)
		go func() {
			wk := newWork(s, s.codes)
			wk.run()
			s.wg.Done()
		}()
	}

	s.runStatus = true
}

func (s *LuaScript) SendData(d LuaData) error {
	if _, ok := s.codes[d.Name()]; !ok {
		return fmt.Errorf("not found luaState of this name '%s'", d.Name())
	}

	s.dataChan <- d
	return nil
}

func (s *LuaScript) Stop() {
	if !s.runStatus {
		return
	}
	s.exit.Close()
	s.wg.Wait()
	s.runStatus = false
}

type worker struct {
	script *LuaScript
	ls     map[string][]*lua.LState
}

func newWork(script *LuaScript, lines map[string][]string) *worker {
	wk := &worker{
		script: script,
		ls:     make(map[string][]*lua.LState),
	}
	for name, codes := range lines {
		lst := []*lua.LState{}
		for _, code := range codes {
			luastate := lua.NewState()
			module.RegisterAllFuncs(luastate, wk.script.luaCache, nil)

			luastate.DoString(code)
			lst = append(lst, luastate)
		}
		wk.ls[name] = lst
	}
	return wk
}

func (wk *worker) run() {
	for {
	AGAIN:
		select {
		case data := <-wk.script.dataChan:
			var err error
			ls := wk.ls[data.Name()]
			val := lua.LNil

			for index, l := range ls {
				if index == 0 {
					val = ToLValue(l, data.DataToLua())
				}
				val, err = SendToLua(l, val, data.CallbackFnName(), data.CallbackTypeName())
				if err != nil {
					data.Handle("", fmt.Errorf("lua '%s' exec error: %v", data.Name(), err))
					goto AGAIN
				}
			}

			jsonStr, err := JsonEncode(val)
			if err != nil {
				data.Handle("", fmt.Errorf("lua '%s' exec error: %v", data.Name(), err))
				goto AGAIN
			}

			data.Handle(jsonStr, nil)

		case <-wk.script.exit.Wait():
			wk.clean()
			return
		}
	}
}

func (wk *worker) clean() {
	for _, ls := range wk.ls {
		for _, luastate := range ls {
			luastate.Close()
		}
	}
}

const defaultWorkerNum = 4

var defaultLuaScript = NewLuaScript(defaultWorkerNum)

func AddLuaCodes(name string, codes []string) error {
	return defaultLuaScript.AddLuaCodes(name, codes)
}

func AddLuaCodesFromFile(name string, filePath []string) error {
	return defaultLuaScript.AddLuaCodesFromFile(name, filePath)
}

func Run() {
	defaultLuaScript.Run()
}

func SendData(d LuaData) error {
	return defaultLuaScript.SendData(d)
}

func Stop() {
	defaultLuaScript.Stop()
}
