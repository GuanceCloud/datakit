package lua

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	lua "github.com/yuin/gopher-lua"
)

const _REDIS_CLIENT_TYPENAME = "redis{client}"

type redisClient struct {
	DB       *redis.Client
	host     string
	port     string
	password string
	dbindex  int
	// atomic
	querySuccessCount uint64
	queryFailedCount  uint64
	connectTime       time.Time
}

var redisMethods = map[string]lua.LGFunction{
	"docmd": redisDoCmd,
	"close": redisClose,
}

func redisConnect(L *lua.LState) int {
	var ud = L.NewUserData()
	opt := L.CheckTable(1)
	// host := "127.0.0.1"; port := "6379"; passwd := ""
	host := "host-null"
	port := "port-null"
	passwd := "password-null"
	dbindex := 0
	otFlag := false

	defer func() {
		if otFlag {
			log.Printf("[error] Lua-Redis create connection failed to '%s:%s/%s', invaild param", host, port, passwd)
		} else {
			log.Printf("[info] Lua-Redis create connection success to '%s:%s/%s'", host, port, passwd)
		}
	}()

	opt.ForEach(func(k, v lua.LValue) {
		k1, ok := k.(lua.LString)
		if !ok {
			L.ArgError(1, "only string allowed in table index")
			otFlag = true
			return
		}

		switch string(k1) {
		case "host":
			v1, ok := v.(lua.LString)
			if !ok {
				L.ArgError(1, "string required for host")
				otFlag = true
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
				otFlag = true
				return
			}
		case "passwd":
			v1, ok := v.(lua.LString)
			if !ok {
				L.ArgError(1, "string required for passwd")
				otFlag = true
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
				otFlag = true
				return
			}
		}
	})

	if client, ok := connPool.Load(connKey(host, port, passwd)); ok {
		ud.Value = client
	} else {
		var client = redisClient{}
		r := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: passwd,
			DB:       dbindex,
		})
		client.DB = r
		client.host = host
		client.port = port
		client.password = passwd
		client.dbindex = dbindex
		client.connectTime = time.Now()

		ud.Value = &client
		connPool.Store(connKey(host, port, passwd), &client)
	}

	L.SetMetatable(ud, L.GetTypeMetatable(_REDIS_CLIENT_TYPENAME))
	L.Push(ud)
	return 1
}

func redisClose(L *lua.LState) int {
	// ud := L.CheckUserData(1)
	// r := ud.Value.(*redis.Client)
	// r.Close()
	return 0
}

func redisDoCmd(L *lua.LState) int {
	ud := L.CheckUserData(1)
	rc, ok := ud.Value.(*redisClient)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("client expected"))
		return 2
	}
	r := rc.DB
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

	res, err := r.Do(args...).Result()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		atomic.AddUint64(&rc.queryFailedCount, 1)
		log.Printf("[error] Lua-Redis query failed to '%s:%s', err: %s", rc.host, rc.port, err.Error())
		return 2
	}
	atomic.AddUint64(&rc.querySuccessCount, 1)
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
		// log.Errorf("unhandled type %v", a.Kind())
	}
	return lua.LNil
}
