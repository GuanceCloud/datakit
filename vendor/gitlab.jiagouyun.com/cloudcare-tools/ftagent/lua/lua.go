package lua

import (
	"io/ioutil"
	"net/http"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/ftagent/utils"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

type LMode struct {
	*lua.LState
}

func NewLuaMode() LMode {
	return LMode{lua.NewState()}
}

func (l *LMode) RegisterFuncs() {
	l.RegisterCryptoFuncs()
	l.RegisterCsvFuncs()
	l.RegisterHTTPFuncs()
	l.RegisterJsonFuncs()
	l.RegisterMongoFuncs()
	l.RegisterRedisFuncs()
	l.RegisterRegexFuncs()
	l.RegisterSQLFuncs()
	l.RegisterXmlFuncs()
}

func (l *LMode) PointsOnHandle(pts []*influxdb.Point) ([]*influxdb.Point, error) {
	tb, err := sendMetatable(l.LState, pts)
	if err != nil {
		return nil, err
	}

	return table2Points(tb)
}

func (l *LMode) RegisterHTTPFuncs() {
	var hc = NewHttpModule(&http.Client{})
	l.SetGlobal("http_request", l.NewFunction(hc.request))

	mt := l.NewTypeMetatable(luaHttpResponseTypeName)
	l.SetField(mt, "__index", l.NewFunction(httpResponseIndex))
}

func (l *LMode) RegisterSQLFuncs() {
	l.SetGlobal("sql_connect", l.NewFunction(sqlConnect))

	mt := l.NewTypeMetatable(_SQL_CLIENT_TYPENAME)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), sqlMethods))
}

func (l *LMode) RegisterRedisFuncs() {
	l.SetGlobal("redis_connect", l.NewFunction(redisConnect))

	mt := l.NewTypeMetatable(_REDIS_CLIENT_TYPENAME)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), redisMethods))
}

func (l *LMode) RegisterMongoFuncs() {
	l.SetGlobal("mongo_connect", l.NewFunction(mongoConnect))

	mt := l.NewTypeMetatable(_MONGO_CLIENT_TYPENAME)
	l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), mongoMethods))
}

func (l *LMode) RegisterJsonFuncs() {
	l.SetGlobal("json_decode", l.NewFunction(jsonDecode))
	l.SetGlobal("json_encode", l.NewFunction(jsonEncode))
}

func (l *LMode) RegisterCsvFuncs() {
	l.SetGlobal("csv_decode", l.NewFunction(csvDecode))
}

func (l *LMode) RegisterXmlFuncs() {
	l.SetGlobal("xml_decode", l.NewFunction(xmlDecode))
}

func (l *LMode) RegisterCacheFuncs(c *Cache) {
	l.SetGlobal("cache_get", l.NewFunction(c.get))
	l.SetGlobal("cache_set", l.NewFunction(c.set))
	l.SetGlobal("cache_list", l.NewFunction(c.list))
}

func (l *LMode) RegisterRegexFuncs() {
	l.SetGlobal("re_quote", l.NewFunction(reQuote))
	l.SetGlobal("re_find", l.NewFunction(reFind))
	l.SetGlobal("re_gsub", l.NewFunction(reGsub))
	l.SetGlobal("re_match", l.NewFunction(reMatch))

	// l.SetGlobal("re_gmatch", l.NewFunction(reGmatch))
}

func (l *LMode) RegisterCryptoFuncs() {
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

func CheckSyntaxToFile(path string) error {
	code, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return CheckSyntaxToString(utils.Bytes2String(code))
}

func CheckSyntaxToBytes(code []byte) error {
	return CheckSyntaxToString(utils.Bytes2String(code))
}

func CheckSyntaxToString(code string) error {
	reader := strings.NewReader(code)
	_, err := parse.Parse(reader, "<string>")
	return err
}
