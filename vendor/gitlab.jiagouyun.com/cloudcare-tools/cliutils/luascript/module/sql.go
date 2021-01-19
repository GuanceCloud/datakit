package module

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/junhsieh/goexamples/fieldbinding/fieldbinding"
	lua "github.com/yuin/gopher-lua"
)

const (
	sqlClientName  = "sql{client}"
	sqlMaxIdelConn = 16
	sqlMaxOpenConn = 32
)

var sqlMethods = map[string]lua.LGFunction{
	"query": sqlQuery,
	"close": sqlClose,
}

type sqlClient struct {
	*sql.DB
}

func newSQLClient(dn, dsn string, maxIdel, maxOpen int) (*sqlClient, error) {
	client, err := sql.Open(dn, dsn)
	if err != nil {
		return nil, err
	}

	client.SetMaxIdleConns(maxIdel)
	client.SetMaxOpenConns(maxOpen)
	return &sqlClient{client}, nil
}

func (s *sqlClient) close() {
	s.Close()
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

	if client, ok := connPool.load(joinKey(dn, dsn)); ok {
		ud.Value = client
	} else {
		var idle, open int
		if L.GetTop()-2 >= 2 {
			idle, open = L.CheckInt(3), L.CheckInt(4)
		} else {
			idle, open = sqlMaxIdelConn, sqlMaxOpenConn
		}

		client, _err := newSQLClient(dn, dsn, idle, open)
		if _err != nil {
			err = _err
			goto conn_err
		}

		ud.Value = client
		connPool.store(joinKey(dn, dsn), client)
	}

	L.SetMetatable(ud, L.GetTypeMetatable(sqlClientName))
	L.Push(ud)
	return 1

conn_err:
	L.Push(lua.LNil)
	L.Push(lua.LString(err.Error()))
	return 2
}

func sqlQuery(L *lua.LState) int {
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
		return 2
	}
	defer rows.Close()

	fb := fieldbinding.NewFieldBinding()
	cols, err := rows.Columns()
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
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

	L.Push(tb)
	return 1
}

func sqlClose(L *lua.LState) int {
	// pass
	return 0
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
		return fmt.Sprintf("ERROR: unknown lua type: %s", t)
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
