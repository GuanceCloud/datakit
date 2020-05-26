package lua

import (
	"errors"
	"log"
	"reflect"
	"strconv"
	"time"

	influxdb "github.com/influxdata/influxdb1-client/v2"
	lua "github.com/yuin/gopher-lua"
)

const (
	luaCallbackFnName = "handle"
	luaPointsTypeName = "points"
)

var (
	ErrInvalidTagKeyType   = errors.New(`invalid tag key type, string expected`)
	ErrInvalidTagValueType = errors.New(`invalid tag value type, string expected`)

	ErrInvalidFieldKeyType   = errors.New(`invalid field key type, string expected`)
	ErrInvalidFieldValueType = errors.New(`invalid field value type, string/number/bool expected`)
)

func tagsToTable(l *lua.LState, t map[string]string) *lua.LTable {
	tb := l.NewTable()

	for k, v := range t {
		tb.RawSetString(k, lua.LString(v))
	}

	return tb
}

func fieldsToTable(l *lua.LState, f map[string]interface{}) *lua.LTable {
	tb := l.NewTable()

	for k, v := range f {
		switch v.(type) {
		case uint64:
			tb.RawSetString(k, lua.LNumber(v.(uint64)))
		case int64:
			tb.RawSetString(k, lua.LNumber(v.(int64)))
		case float64:
			tb.RawSetString(k, lua.LNumber(v.(float64)))
		case bool:
			tb.RawSetString(k, lua.LBool(v.(bool)))
		case string:
			tb.RawSetString(k, lua.LString(v.(string)))
		}
	}

	return tb
}

func getName(tb *lua.LTable) (string, error) {
	if name, _ := tb.RawGetString("name").(lua.LString); name != "" {
		return name.String(), nil
	}
	return "", errors.New("invalid point name")
}

func getTime(tb *lua.LTable) time.Time {
	var tim time.Time
	if t, ok := tb.RawGetString("time").(lua.LNumber); !ok {
		tim = time.Now()
	} else {
		if ts, err := strconv.Atoi(t.String()); err != nil {
			tim = time.Now()
		} else {
			tim = time.Unix(0, int64(ts))
		}
	}
	return tim
}

func getTags(tb *lua.LTable) (map[string]string, error) {

	tagtb, ok := tb.RawGetString("tags").(*lua.LTable)
	if !ok {
		return nil, errors.New("invalid tags table")
	}

	var res = make(map[string]string)
	var err error

	tagtb.ForEach(func(k, v lua.LValue) {

		// For influxdb, tag keys & values only support string type
		switch k.(type) {
		case lua.LString:
			switch v.(type) {
			case lua.LString:
				res[k.String()] = v.String()
			default:
				err = ErrInvalidTagValueType
				return
			}
		default:
			err = ErrInvalidTagKeyType
			return
		}
	})

	if err != nil {
		return nil, err
	}
	return res, nil
}

func getFields(tb *lua.LTable) (map[string]interface{}, error) {

	fdtb, ok := tb.RawGetString("fields").(*lua.LTable)
	if !ok {
		return nil, errors.New("invalid fields table")
	}

	var res = make(map[string]interface{})
	var err error

	fdtb.ForEach(func(k, v lua.LValue) {

		// For influxdb, fields keys only support string type
		switch k.(type) {
		case lua.LString:
			switch v.(type) {
			case lua.LString:
				res[k.String()] = v.(lua.LString).String()

			case lua.LNumber:
				res[k.String()] = float64(v.(lua.LNumber))

			case lua.LBool:
				bs := v.(lua.LBool).String()
				if b, err := strconv.ParseBool(bs); err == nil {
					res[k.String()] = b
				} else {
					res[k.String()] = bs
				}

			default:
				err = ErrInvalidFieldValueType
				return
			}

		default:
			err = ErrInvalidFieldKeyType
			return
		}
	})

	if err != nil {
		return nil, err
	}
	return res, nil
}

func table2Points(tb *lua.LTable) ([]*influxdb.Point, error) {

	res := []*influxdb.Point{}

	for i := 1; i <= tb.MaxN(); i++ {
		t, ok := tb.RawGetInt(i).(*lua.LTable)
		if !ok {
			continue
		}

		name, err := getName(t)
		if err != nil {
			continue
		}

		tags, err := getTags(t)
		if err != nil {
			continue
		}

		fields, err := getFields(t)
		if err != nil {
			continue
		}

		pt, err := influxdb.NewPoint(
			name,
			tags,
			fields,
			getTime(t),
		)

		if err != nil {
			log.Printf("[warn] %s", err.Error())
			return nil, err
		}

		res = append(res, pt)
	}

	return res, nil
}

func sendMetatable(l *lua.LState, pts []*influxdb.Point) (*lua.LTable, error) {

	tb := l.NewTable()

	for _, pt := range pts {

		tbPoint := l.NewTable()

		tbPoint.RawSetString("name", lua.LString(pt.Name()))
		tbPoint.RawSetString("tags", tagsToTable(l, pt.Tags()))

		if f, err := pt.Fields(); err == nil {
			tbPoint.RawSetString("fields", fieldsToTable(l, f))
		}

		tbPoint.RawSetString("time", lua.LNumber(pt.UnixNano()))

		tb.Append(tbPoint)
	}

	l.SetMetatable(tb, l.GetTypeMetatable(luaPointsTypeName))

	lv := l.GetGlobal(luaCallbackFnName)

	switch lv.(type) {
	case *lua.LFunction:
	default:
		log.Printf("[error] invalid lua value type: %s", reflect.TypeOf(lv))
		return nil, errors.New("invalid lua function: " + luaCallbackFnName)
	}

	gf, _ := lv.(*lua.LFunction)

	if err := l.CallByParam(lua.P{
		Fn:      gf,
		NRet:    1,
		Protect: true,
	}, tb); err != nil {
		return nil, err
	}

	var ret *lua.LTable
	lt := l.Get(-1)

	ret, ok := lt.(*lua.LTable)
	if !ok {
		return nil, errors.New("get lua LTable failed")
	}

	l.Pop(1)
	return ret, nil
}
