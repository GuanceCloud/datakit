package dataclean

var (
	DWLuaPath string

	sampleConfig = `
#bind_addr = '0.0.0.0:9528'
#gin_log = 'gin.log'
#lua_worker = 4

#[[global_lua]]
#path = 'a.lua'
#circle = '*/10 * * * *'

#[[routes_config]]
#name = 'demo'
#disable_type_check = false
#disable_lua = false

# [[routes_config.lua]]
#  path = 'demo.lua'
`
)

type LuaConfig struct {
	Path   string `toml:"path"`
	Circle string `toml:"circle,omitempty"`
}

type RouteConfig struct {
	Name             string      `toml:"name"`
	Lua              []LuaConfig `toml:"lua,omitempty"`
	DisableTypeCheck bool        `toml:"disable_type_check"`
	DisableLua       bool        `toml:"disable_lua"`
}
