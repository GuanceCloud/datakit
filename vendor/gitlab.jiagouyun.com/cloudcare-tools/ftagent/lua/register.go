package lua

import (
	"net/http"

	lua "github.com/yuin/gopher-lua"
)

func RegisterHTTPFuncs(l *lua.LState) {
	var hc = NewHttpModule(&http.Client{})
	l.SetGlobal("http_request", l.NewFunction(hc.request))

	mt := l.NewTypeMetatable(luaHttpResponseTypeName)
	l.SetField(mt, "__index", l.NewFunction(httpResponseIndex))
}

func RegisterSQLFuncs(l *lua.LState) {
	l.SetGlobal("sql_connect", l.NewFunction(sqlConnect))

	mt := l.NewTypeMetatable(_SQL_CLIENT_TYPENAME)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), sqlMethods))
}

func RegisterRedisFuncs(l *lua.LState) {
	l.SetGlobal("redis_connect", l.NewFunction(redisConnect))

	mt := l.NewTypeMetatable(_REDIS_CLIENT_TYPENAME)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), redisMethods))
}

func RegisterMongoFuncs(l *lua.LState) {
	l.SetGlobal("mongo_connect", l.NewFunction(mongoConnect))

	mt := l.NewTypeMetatable(_MONGO_CLIENT_TYPENAME)
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

func RegisterCacheFuncs(l *lua.LState, c *Cache) {
	l.SetGlobal("cache_get", l.NewFunction(c.get))
	l.SetGlobal("cache_set", l.NewFunction(c.set))
	l.SetGlobal("cache_list", l.NewFunction(c.list))
}

func RegisterRegexFuncs(l *lua.LState) {
	l.SetGlobal("re_quote", l.NewFunction(reQuote))
	l.SetGlobal("re_find", l.NewFunction(reFind))
	l.SetGlobal("re_gsub", l.NewFunction(reGsub))
	l.SetGlobal("re_match", l.NewFunction(reMatch))

	// l.SetGlobal("re_gmatch", l.NewFunction(reGmatch))
}

func RegisterCryptoFuncs(l *lua.LState) {
	l.SetGlobal("base64_encode", l.NewFunction(base64EncodeFn))
	l.SetGlobal("base64_decode", l.NewFunction(base64DecodeFn))
	l.SetGlobal("hex_encode", l.NewFunction(hexEncodeToStringFn))
	l.SetGlobal("crc32", l.NewFunction(crc32Fn))
	l.SetGlobal("hmac", l.NewFunction(hmacFn))
	l.SetGlobal("encrypt", l.NewFunction(encryptFn))
	l.SetGlobal("decrypt", l.NewFunction(decryptFn))

	// l.SetGlobal("md5", l.NewFunction(md5Fn))
	// l.SetGlobal("sha1", l.NewFunction(sha1Fn))
	// l.SetGlobal("sha256", l.NewFunction(sha256Fn))
	// l.SetGlobal("sha512", l.NewFunction(sha512Fn))
}
