package dataclean

var (
	DWLuaPath string

	sampleConfig = `
#bind_addr = '0.0.0.0:9528'
#gin_log = 'gin.log'
#lua_worker = 4
#enable_config_api = true
#cfg_api_pwd = 'xxx'

#[[global_lua]]
#path = 'a.lua'
#circle = '* */1 * * *'

#[[routes_config]]
#name = 'demo'
#disable_type_check = false
#disable_lua = false
#ak_open = false

# [[routes_config.lua]]
#  path = 'demo.lua'
#  circle = '* */5 * * *'	
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
	AkOpen           bool        `toml:"ak_open"`
}
