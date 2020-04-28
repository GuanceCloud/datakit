package lua

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	lua "github.com/yuin/gopher-lua"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const _MONGO_CLIENT_TYPENAME = "mongo{client}"

type mongoClient struct {
	DB                *mongo.Client
	uri               string
	querySuccessCount uint64
	queryFailedCount  uint64
	connectTime       time.Time
}

var mongoMethods = map[string]lua.LGFunction{
	"query": mongoQuery,
	"close": mongoClose,
}

var mongoConnPool sync.Map

func mongoConnect(L *lua.LState) int {

	var ud = L.NewUserData()
	var uri string
	var err error

	if uri = L.CheckString(1); uri == "" {
		err = errors.New("invaild mongodb uri")
		goto conn_err
	}

	if client, ok := connPool.Load(connKey(uri)); ok {
		ud.Value = client
	} else {
		var client = mongoClient{}
		op := options.Client().ApplyURI(uri)
		client.DB, err = mongo.Connect(context.TODO(), op)
		fmt.Println(err)
		if err != nil {
			log.Printf("[error] Lua-MongoDB create connection failed to '%s', err: %s", uri, err.Error())
			goto conn_err
		}

		client.uri = uri
		client.connectTime = time.Now()

		ud.Value = &client
		connPool.Store(connKey(uri), &client)
		log.Printf("[info] Lua-MongoDB create connection success to '%s'", uri)
	}

	L.SetMetatable(ud, L.GetTypeMetatable(_MONGO_CLIENT_TYPENAME))
	L.Push(ud)
	return 1

conn_err:
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2
}

func mongoClose(L *lua.LState) int {
	// FIXME: remove the function of close of connection pool
	// there is resource leak!

	// ud := L.CheckUserData(1)
	// m := ud.Value.(*mongo.Client)
	// m.Disconnect(context.Background())
	return 0
}

func mongoQuery(L *lua.LState) int {

	ud := L.CheckUserData(1)
	client, ok := ud.Value.(*mongoClient)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("client expected"))
		return 2
	}

	db := L.CheckString(2)
	collection := L.CheckString(3)
	query := L.CheckTable(4)

	var result map[string]interface{}
	collec := client.DB.Database(db).Collection(collection)

	if err := collec.FindOne(context.Background(), mongoGetTable(L, query)).Decode(&result); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		log.Printf("[error] Lua-MongoDB query failed to '%s:%s:%s', err: %s", client.uri, db, collection, err.Error())
		atomic.AddUint64(&client.queryFailedCount, 1)
		return 2
	}

	L.Push(mongoToTable(L, result))
	atomic.AddUint64(&client.querySuccessCount, 1)
	return 1
}

func mongoGetTable(l *lua.LState, tb *lua.LTable) map[string]interface{} {

	var m = make(map[string]interface{}, tb.Len())
	tb.ForEach(func(k, v lua.LValue) {
		k1, ok := k.(lua.LString)
		if !ok {
			return
		}
		switch v.(type) {
		case lua.LString:
			m[k1.String()] = v.(lua.LString).String()
		case lua.LNumber:
			m[k1.String()] = float64(v.(lua.LNumber))
		case lua.LBool:
			bs := v.(lua.LBool).String()
			if b, err := strconv.ParseBool(bs); err == nil {
				m[k1.String()] = b
			} else {
				m[k1.String()] = bs
			}
		default:
			// TODO
		}

	})
	return m
}

func mongoToTable(l *lua.LState, m map[string]interface{}) *lua.LTable {

	tb := l.NewTable()
	for k, v := range m {
		switch v.(type) {
		case int64:
			tb.RawSetString(k, lua.LNumber(v.(int64)))
		case int32:
			tb.RawSetString(k, lua.LNumber(v.(int32)))
		case float64:
			tb.RawSetString(k, lua.LNumber(v.(float64)))
		case bool:
			tb.RawSetString(k, lua.LBool(v.(bool)))
		case string:
			tb.RawSetString(k, lua.LString(v.(string)))
		case []byte:
			tb.RawSetString(k, lua.LString(v.([]byte)))
		default:
			tb.RawSetString(k, lua.LString(fmt.Sprintf("%#v", v)))
		}
	}

	return tb
}
