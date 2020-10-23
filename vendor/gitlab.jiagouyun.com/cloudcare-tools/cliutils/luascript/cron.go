package luascript

import (
	"io/ioutil"
	"strings"

	"github.com/robfig/cron/v3"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/luascript/module"
)

var (
	specParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month)

	// 进程级别的cache
	globalLuaCache = &module.LuaCache{}
)

type LuaCron struct {
	*cron.Cron
}

func NewLuaCron() *LuaCron {
	return &LuaCron{
		cron.New(cron.WithParser(specParser)),
	}
}

func (c *LuaCron) AddLua(code string, schedule string) (err error) {
	if err = CheckLuaCode(code); err != nil {
		return
	}

	luastate := lua.NewState()
	module.RegisterAllFuncs(luastate, globalLuaCache, nil)

	_, err = c.AddFunc(schedule, func() {
		luastate.DoString(code)
	})
	return err
}

func (c *LuaCron) AddLuaFromFile(file string, schedule string) (err error) {
	content, _err := ioutil.ReadFile(file)
	if _err != nil {
		return _err
	}

	code := string(content)

	if err = CheckLuaCode(code); err != nil {
		return err
	}

	luastate := lua.NewState()
	module.RegisterAllFuncs(luastate, globalLuaCache, nil)

	_, err = c.AddFunc(schedule, func() {
		luastate.DoString(code)
	})
	return err
}

func (c *LuaCron) Run() {
	c.Start()
}

func (c *LuaCron) Stoping() {
	c.Stop()
}

func CheckLuaCode(code string) error {
	reader := strings.NewReader(code)
	_, err := parse.Parse(reader, "<string>")
	return err
}
