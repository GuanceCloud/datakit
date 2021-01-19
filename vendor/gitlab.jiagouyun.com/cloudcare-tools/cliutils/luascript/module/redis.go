package module

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-redis/redis"
	lua "github.com/yuin/gopher-lua"
)

const redisClientName = "redis{client}"

var redisMethods = map[string]lua.LGFunction{
	"docmd": redisDoCmd,
	"close": redisClose,
}

type redisClient struct {
	*redis.Client
}

func newRedisClient(host, port, passwd string, dbindex int) *redisClient {
	return &redisClient{redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: passwd,
		DB:       dbindex,
	})}
}

func (r *redisClient) close() {
	r.Close()
}

func redisConnect(L *lua.LState) int {
	var ud = L.NewUserData()
	opt := L.CheckTable(1)

	host := "127.0.0.1"
	port := "6579"
	passwd := ""
	dbindex := 0

	opt.ForEach(func(k, v lua.LValue) {
		k1, ok := k.(lua.LString)
		if !ok {
			L.ArgError(1, "only string allowed in table index")
			return
		}

		switch string(k1) {
		case "host":
			v1, ok := v.(lua.LString)
			if !ok {
				L.ArgError(1, "string required for host")
				return
			}
			host = string(v1)
		case "port":
			switch v1 := v.(type) {
			case lua.LString:
				port = string(v1)
			case lua.LNumber:
				port = fmt.Sprintf("%d", int64(v1))
			default:
				L.ArgError(1, "string or number required for port")
				return
			}
		case "passwd":
			v1, ok := v.(lua.LString)
			if !ok {
				L.ArgError(1, "string required for passwd")
				return
			}
			passwd = string(v1)
		case "index":
			switch v1 := v.(type) {
			case lua.LString:
				db1, err := strconv.Atoi(string(v1))
				if err == nil {
					dbindex = db1
				}
			case lua.LNumber:
				dbindex = int(v1)
			default:
				L.ArgError(1, "string or number required for index")
				return
			}
		}
	})

	if client, ok := connPool.load(joinKey(host, port, passwd)); ok {
		ud.Value = client
	} else {
		var client = newRedisClient(host, port, passwd, dbindex)
		ud.Value = client
		connPool.store(joinKey(host, port, passwd), client)
	}

	L.SetMetatable(ud, L.GetTypeMetatable(redisClientName))
	L.Push(ud)
	return 1
}

func redisClose(L *lua.LState) int {
	// pass
	return 0
}

func redisDoCmd(L *lua.LState) int {
	ud := L.CheckUserData(1)
	client, ok := ud.Value.(*redisClient)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("client expected"))
		return 2
	}
	args := []interface{}{}
	for i := 2; i <= L.GetTop(); i++ {
		a := L.Get(i)
		switch a1 := a.(type) {
		case lua.LString:
			args = append(args, string(a1))
		case lua.LNumber:
			args = append(args, int64(a1))
		case *lua.LTable:
			a1.ForEach(func(k, v lua.LValue) {
				switch k1 := k.(type) {
				case lua.LString:
					args = append(args, string(k1))
				case lua.LNumber:
				default:
					L.ArgError(i, "only string or number index allowed in table")
					return
				}

				switch v1 := v.(type) {
				case lua.LString:
					args = append(args, string(v1))
				case lua.LNumber:
					args = append(args, int64(v1))
				default:
					L.ArgError(i, "only string or number value allowed in table")
					return
				}
			})
		default:
			L.ArgError(i, "need string, number or table")
			return 0
		}
	}

	res, err := client.Do(args...).Result()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	v := reidsToLValue(reflect.ValueOf(res), L)
	L.Push(v)
	return 1
}

func reidsToLValue(a reflect.Value, L *lua.LState) lua.LValue {
	switch a.Kind() {
	case reflect.String:
		return lua.LString(a.String())
	case reflect.Slice:
		t := L.NewTable()
		for i := 0; i < a.Len(); i++ {
			v := a.Index(i)
			v1 := reidsToLValue(reflect.ValueOf(v.Interface()), L)
			t.RawSetInt(i+1, v1)
		}
		return t
	case reflect.Int, reflect.Int64:
		return lua.LNumber(a.Int())
	default:
		// nil
	}
	return lua.LNil
}
