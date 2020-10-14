package luascript

import (
	"strings"

	"github.com/robfig/cron"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/luascript/module"
)

var (
	specParser = cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month)

	globalLuaCache = &module.LuaCache{}
)

type LuaCron struct {
	*cron.Cron
}

func NewLuaCron() *LuaCron {
	return &LuaCron{
		cron.New(),
	}
}

func (c *LuaCron) AddHandle(code string, intervalSpec string) error {
	if err := CheckLuaCode(code); err != nil {
		return err
	}

	luastate := lua.NewState()
	module.RegisterAllFuncs(luastate, globalLuaCache, nil)

	return c.AddFunc(intervalSpec, func() {
		luastate.DoString(code)
	})
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
