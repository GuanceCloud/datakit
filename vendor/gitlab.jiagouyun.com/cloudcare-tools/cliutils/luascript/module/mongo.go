package module

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	lua "github.com/yuin/gopher-lua"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoClientName = "mongo{client}"

var mongoMethods = map[string]lua.LGFunction{
	"query": mongoQuery,
	"close": mongoClose,
}

type mongoClient struct {
	*mongo.Client
}

func newMongoClient(uri string) (*mongoClient, error) {
	op := options.Client().ApplyURI(uri)
	conn, err := mongo.Connect(context.TODO(), op)
	return &mongoClient{conn}, err
}

func (m *mongoClient) close() {
	m.Disconnect(context.Background())
}

func mongoConnect(L *lua.LState) int {
	var ud = L.NewUserData()
	var uri string
	var err error

	if uri = L.CheckString(1); uri == "" {
		err = errors.New("invaild mongodb uri")
		goto conn_err
	}

	if client, ok := connPool.load(joinKey(uri)); ok {
		ud.Value = client
	} else {
		client, _err := newMongoClient(uri)
		if _err != nil {
			err = _err
			goto conn_err
		}
		ud.Value = client
		connPool.store(joinKey(uri), client)
	}

	L.SetMetatable(ud, L.GetTypeMetatable(mongoClientName))
	L.Push(ud)
	return 1

conn_err:
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2
}

func mongoClose(L *lua.LState) int {
	// pass
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
	collec := client.Database(db).Collection(collection)

	if err := collec.FindOne(context.Background(), mongoGetTable(L, query)).Decode(&result); err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(mongoToTable(L, result))
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
