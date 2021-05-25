package module

import (
	"io"
	"net/http"

	lua "github.com/yuin/gopher-lua"
)

func Clean() {
	connPool.close()
}

func RegisterAllFuncs(l *lua.LState, cache *LuaCache, logOutput io.Writer) {
	RegisterHTTPFuncs(l)
	RegisterSQLFuncs(l)
	RegisterRedisFuncs(l)
	RegisterMongoFuncs(l)
	RegisterJsonFuncs(l)
	RegisterCsvFuncs(l)
	RegisterXmlFuncs(l)
	RegisterCryptoFuncs(l)
	RegisterRegexFuncs(l)

	if cache != nil {
		RegisterCacheFuncs(l, cache)
	}

	if logOutput != nil {
		RegisterLogFuncs(l, logOutput)
	}
}

func RegisterHTTPFuncs(l *lua.LState) {
	var hc = NewHttpModule(&http.Client{})
	l.SetGlobal("http_request", l.NewFunction(hc.request))

	mt := l.NewTypeMetatable(luaHttpResponseTypeName)
	l.SetField(mt, "__index", l.NewFunction(httpResponseIndex))
}

func RegisterSQLFuncs(l *lua.LState) {
	l.SetGlobal("sql_connect", l.NewFunction(sqlConnect))

	mt := l.NewTypeMetatable(sqlClientName)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), sqlMethods))
}

func RegisterRedisFuncs(l *lua.LState) {
	l.SetGlobal("redis_connect", l.NewFunction(redisConnect))

	mt := l.NewTypeMetatable(redisClientName)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), redisMethods))
}

func RegisterMongoFuncs(l *lua.LState) {
	l.SetGlobal("mongo_connect", l.NewFunction(mongoConnect))

	mt := l.NewTypeMetatable(mongoClientName)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), mongoMethods))
}

func RegisterJsonFuncs(l *lua.LState) {
	l.SetGlobal("json_decode", l.NewFunction(jsonDecode))
	l.SetGlobal("json_encode", l.NewFunction(jsonEncode))
}

func RegisterCsvFuncs(l *lua.LState) {
	l.SetGlobal("csv_decode", l.NewFunction(csvDecode))
}

func RegisterXmlFuncs(l *lua.LState) {
	l.SetGlobal("xml_decode", l.NewFunction(xmlDecode))
}

func RegisterRegexFuncs(l *lua.LState) {
	l.SetGlobal("re_quote", l.NewFunction(reQuote))
	l.SetGlobal("re_find", l.NewFunction(reFind))
	l.SetGlobal("re_gsub", l.NewFunction(reGsub))
	l.SetGlobal("re_match", l.NewFunction(reMatch))
}

func RegisterCryptoFuncs(l *lua.LState) {
	l.SetGlobal("base64_encode", l.NewFunction(base64EncodeFn))
	l.SetGlobal("base64_decode", l.NewFunction(base64DecodeFn))
	l.SetGlobal("hex_encode", l.NewFunction(hexEncodeToStringFn))
	l.SetGlobal("crc32", l.NewFunction(crc32Fn))
	l.SetGlobal("hmac", l.NewFunction(hmacFn))
	l.SetGlobal("encrypt", l.NewFunction(encryptFn))
	l.SetGlobal("decrypt", l.NewFunction(decryptFn))
}

func RegisterCacheFuncs(l *lua.LState, c *LuaCache) {
	l.SetGlobal("cache_get", l.NewFunction(c.get))
	l.SetGlobal("cache_set", l.NewFunction(c.set))
	l.SetGlobal("cache_list", l.NewFunction(c.list))
}

func RegisterLogFuncs(l *lua.LState, ouput io.Writer) {
	log := lualog{ouput}
	l.SetGlobal("log_info", l.NewFunction(log.logInfo))
	l.SetGlobal("log_debug", l.NewFunction(log.logDebug))
	l.SetGlobal("log_warn", l.NewFunction(log.logWarn))
	l.SetGlobal("log_error", l.NewFunction(log.logError))
}
