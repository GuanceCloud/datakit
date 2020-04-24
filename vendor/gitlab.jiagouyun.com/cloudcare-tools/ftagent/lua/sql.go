package lua

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/junhsieh/goexamples/fieldbinding/fieldbinding"
	lua "github.com/yuin/gopher-lua"
)

const (
	_SQL_CLIENT_TYPENAME = "sql{client}"
	_SQL_MAX_IDEL_CONN   = 16
	_SQL_MAX_OPEN_CONN   = 32
)

type sqlClient struct {
	DB             *sql.DB
	driverName     string
	dataSourceName string
	maxIdelConn    int
	maxOpenConn    int
	// atomic uint64
	querySuccessCount uint64
	queryFailedCount  uint64
	// start time
	connectTime time.Time
}

var sqlMethods = map[string]lua.LGFunction{
	"query": sqlQuery,
	"close": sqlClose,
}

func sqlConnect(L *lua.LState) int {

	var ud = L.NewUserData()
	var dn, dsn string // driverName, dataSourceName
	var err error

	if dn = L.CheckString(1); dn == "" {
		err = errors.New("invaild sql driver_name")
		goto conn_err
	}

	if dsn = L.CheckString(2); dsn == "" {
		err = errors.New("invaild sql data_source_name")
		goto conn_err
	}

	if client, ok := connPool.Load(connKey(dn, dsn)); ok {
		ud.Value = client
	} else {
		var client = sqlClient{}
		var idle, open int

		if L.GetTop()-2 >= 2 {
			idle, open = L.CheckInt(3), L.CheckInt(4)
		} else {
			idle, open = _SQL_MAX_IDEL_CONN, _SQL_MAX_OPEN_CONN
		}

		if client.DB, err = sql.Open(dn, dsn); err != nil {
			log.Printf("[error] Lua-SQL create connection failed to '%s-%s', err: %s", dn, dsn, err.Error())
			goto conn_err
		}

		client.DB.SetMaxIdleConns(idle)
		client.DB.SetMaxOpenConns(open)
		client.connectTime = time.Now()
		client.driverName = dn
		client.dataSourceName = dsn

		connPool.Store(connKey(dn, dsn), &client)
		ud.Value = &client
		log.Printf("[info] Lua-SQL create connection success to '%s-%s'", dn, dsn)
	}

	L.SetMetatable(ud, L.GetTypeMetatable(_SQL_CLIENT_TYPENAME))
	L.Push(ud)
	return 1

conn_err:
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2
}

func sqlQuery(L *lua.LState) int {
	// simplify code..

	ud := L.CheckUserData(1)
	client, ok := ud.Value.(*sqlClient)
	if !ok {
		L.Push(lua.LNil)
		L.Push(lua.LString("client expected"))
		return 2
	}

	query := L.ToString(2)
	if query == "" {
		L.Push(lua.LNil)
		L.Push(lua.LString("query string required"))
		return 2
	}

	// 取栈深度
	top := L.GetTop()
	// 可变参数数量 = 栈的深度 - （栈顶形参数量 + this）
	args := make([]interface{}, 0, top-2)

	// 取可变参数
	for i := 3; i <= top; i++ {
		args = append(args, sqlGetValue(L, i))
	}

	rows, err := client.DB.Query(query, args...)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		atomic.AddUint64(&client.queryFailedCount, 1)
		log.Printf("[error] Lua-SQL query failed to '%s-%s-%v', err: %s", client.driverName, client.dataSourceName, query, err.Error())
		return 2
	}
	defer rows.Close()

	fb := fieldbinding.NewFieldBinding()
	cols, err := rows.Columns()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		// ignore rows errors..
		return 2
	}

	fb.PutFields(cols)
	tb := L.NewTable()
	for rows.Next() {
		if err := rows.Scan(fb.GetFieldPtrArr()...); err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}

		tbRow := sqlToTableFromMap(L, reflect.ValueOf(fb.GetFieldArr()))
		tb.Append(tbRow)
	}

	atomic.AddUint64(&client.querySuccessCount, 1)
	L.Push(tb)
	return 1
}

func sqlClose(L *lua.LState) int {
	// FIXME: remove the function of close of connection pool
	// there is resource leak!

	// err := client.DB.Close()
	// // always clean
	// client.DB = nil
	// if err != nil {
	// 	L.Push(lua.LBool(false))
	// 	L.Push(lua.LString(err.Error()))
	// 	return 2
	// }

	// L.Push(lua.LBool(true))
	return 1
}

func sqlGetValue(l *lua.LState, n int) interface{} {
	return sqlGetArbitraryValue(l, l.Get(n))
}

func sqlGetArbitraryValue(l *lua.LState, v lua.LValue) interface{} {
	switch t := v.Type(); t {
	case lua.LTNil:
		return nil
	case lua.LTBool:
		return lua.LVAsBool(v)
	case lua.LTNumber:
		f := lua.LVAsNumber(v)
		if float64(f) == float64(int(f)) {
			return int(f)
		}
		return float64(f)
	case lua.LTString:
		return lua.LVAsString(v)
	case lua.LTTable:
		m := map[string]interface{}{}
		tb := v.(*lua.LTable)
		arrSize := 0
		tb.ForEach(func(k, val lua.LValue) {
			key := sqlGetArbitraryValue(l, k)
			if keyi, ok := key.(int); ok {
				if arrSize >= 0 && arrSize < keyi {
					arrSize = keyi
				}
				key = strconv.Itoa(keyi)
			} else {
				arrSize = -1
			}
			m[key.(string)] = sqlGetArbitraryValue(l, val)
		})

		if arrSize >= 0 {
			ms := make([]interface{}, arrSize)
			for i := 0; i < arrSize; i++ {
				ms[i] = m[strconv.Itoa(i+1)]
			}
			return ms
		}

		return m
	default:
		panic(fmt.Sprintf("unknown lua type: %s", t))
	}
}

func sqlToArbitraryValue(l *lua.LState, i interface{}) lua.LValue {
	if i == nil {
		return lua.LNil
	}

	switch ii := i.(type) {
	case bool:
		return lua.LBool(ii)
	case int:
		return lua.LNumber(ii)
	case int8:
		return lua.LNumber(ii)
	case int16:
		return lua.LNumber(ii)
	case int32:
		return lua.LNumber(ii)
	case int64:
		return lua.LNumber(ii)
	case uint:
		return lua.LNumber(ii)
	case uint8:
		return lua.LNumber(ii)
	case uint16:
		return lua.LNumber(ii)
	case uint32:
		return lua.LNumber(ii)
	case uint64:
		return lua.LNumber(ii)
	case float64:
		return lua.LNumber(ii)
	case float32:
		return lua.LNumber(ii)
	case string:
		return lua.LString(ii)
	case []byte:
		return lua.LString(ii)
	default:
		v := reflect.ValueOf(i)
		switch v.Kind() {
		case reflect.Ptr:
			return sqlToArbitraryValue(l, v.Elem().Interface())

		case reflect.Struct:
			return sqlToTableFromStruct(l, v)

		case reflect.Map:
			return sqlToTableFromMap(l, v)

		case reflect.Slice:
			return sqlToTableFromSlice(l, v)

		default:
			panic(fmt.Sprintf("unknown type being pushed onto lua stack: %T %+v", i, i))
		}
	}
}

func sqlToTableFromStruct(l *lua.LState, v reflect.Value) lua.LValue {
	tb := l.NewTable()
	return sqlToTableFromStructInner(l, tb, v)
}

func sqlToTableFromStructInner(l *lua.LState, tb *lua.LTable, v reflect.Value) lua.LValue {
	t := v.Type()
	for j := 0; j < v.NumField(); j++ {
		var inline bool
		name := t.Field(j).Name
		if tag := t.Field(j).Tag.Get("luautil"); tag != "" {
			tagParts := strings.Split(tag, ",")
			if tagParts[0] == "-" {
				continue
			} else if tagParts[0] != "" {
				name = tagParts[0]
			}
			if len(tagParts) > 1 && tagParts[1] == "inline" {
				inline = true
			}
		}
		if inline {
			sqlToTableFromStructInner(l, tb, v.Field(j))
		} else {
			tb.RawSetString(name, sqlToArbitraryValue(l, v.Field(j).Interface()))
		}
	}
	return tb
}

func sqlToTableFromMap(l *lua.LState, v reflect.Value) lua.LValue {
	tb := &lua.LTable{}
	for _, k := range v.MapKeys() {
		tb.RawSet(sqlToArbitraryValue(l, k.Interface()),
			sqlToArbitraryValue(l, v.MapIndex(k).Interface()))
	}
	return tb
}

func sqlToTableFromSlice(l *lua.LState, v reflect.Value) lua.LValue {
	tb := &lua.LTable{}
	for j := 0; j < v.Len(); j++ {
		tb.RawSet(sqlToArbitraryValue(l, j+1), // because lua is 1-indexed
			sqlToArbitraryValue(l, v.Index(j).Interface()))
	}
	return tb
}
