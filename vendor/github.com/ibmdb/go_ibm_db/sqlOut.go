package go_ibm_db

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/ibmdb/go_ibm_db/api"
)

// Out struct is used to store the value of a OUT parameter in Stored Procedure
type Out struct {
	sqlOut          *sql.Out
	idx             int
	data            []byte
	ctype           api.SQLSMALLINT
	sqltype         api.SQLSMALLINT
	decimalDigits   api.SQLSMALLINT
	nullable        api.SQLSMALLINT
	inputOutputType api.SQLSMALLINT
	parameterSize   api.SQLULEN
	buflen          api.SQLLEN
	plen            *api.SQLLEN
}

func newOut(hstmt api.SQLHSTMT, sqlOut *sql.Out, idx int) (*Out, error) {
	var ctype, sqltype, decimalDigits, nullable, inputOutputType api.SQLSMALLINT
	var parameterSize api.SQLULEN
	var buflen api.SQLLEN
	var plen *api.SQLLEN
	var data []byte
	if sqlOut.In {
		inputOutputType = api.SQL_PARAM_INPUT_OUTPUT
		//convert sql.Out.Dest to a driver.Value so the number of possible type is limited.
		dv, err := driver.DefaultParameterConverter.ConvertValue(sqlOut.Dest)
		if err != nil {
			return nil, fmt.Errorf("%v : failed to convert Dest in sql.Out to driver.Value", err)
		}
		// use case with one type only. Otherwise d will turn into some other type and extract
		// will give incorrect result
		switch d := dv.(type) {
		case nil:
			var ind api.SQLLEN = api.SQL_NULL_DATA
			// nil has no type, so use SQLDescribeParam
			ret := api.SQLDescribeParam(hstmt, api.SQLUSMALLINT(idx+1),
				&sqltype, &parameterSize, &decimalDigits, &nullable)
			if IsError(ret) {
				return nil, NewError("SQLDescribeParam", hstmt)
			}
			// input value might be nil but the output value may not be so allocate buffer for output
			data = make([]byte, parameterSize)
			ctype = SqltoCtype(sqltype)
			buflen = api.SQLLEN(len(data))
			plen = &ind
		case string:
			var ind api.SQLLEN = api.SQL_NTS
			// string output buffer cannot be same as input, so use SQLDescribeParam
			ret := api.SQLDescribeParam(hstmt, api.SQLUSMALLINT(idx+1),
				&sqltype, &parameterSize, &decimalDigits, &nullable)
			if IsError(ret) {
				return nil, NewError("SQLDescribeParam", hstmt)
			}
			ctype = api.SQL_C_WCHAR
			sqltype = api.SQL_WCHAR
			s16 := api.StringToUTF16(d)
			b := api.ExtractUTF16Str(s16)
			data = make([]byte, (parameterSize*2)+2)
			if len(b) > len(data) {
				return nil,
					fmt.Errorf("At param. index %d INOUT string size is greater than the allocated OUT buffer size", idx+1)
			}
			copy(data, b)
			buflen = api.SQLLEN(len(data))
			// use SQL_NTS to indicate that the string null terminated
			plen = &ind
		case int64:
			ctype = api.SQL_C_SBIGINT
			sqltype = api.SQL_BIGINT
			data = api.Extract(unsafe.Pointer(&d), unsafe.Sizeof(d))
			parameterSize = 8
		case float64:
			ctype = api.SQL_C_DOUBLE
			sqltype = api.SQL_DOUBLE
			data = api.Extract(unsafe.Pointer(&d), unsafe.Sizeof(d))
			parameterSize = 8
		case bool:
			var b byte
			if d {
				b = 1
			}
			ctype = api.SQL_C_BIT
			sqltype = api.SQL_BIT
			data = api.Extract(unsafe.Pointer(&b), unsafe.Sizeof(b))
			parameterSize = 1
		case time.Time:
			ctype = api.SQL_C_TYPE_TIMESTAMP
			sqltype = api.SQL_TYPE_TIMESTAMP
			y, m, day := d.Date()
			t := api.SQL_TIMESTAMP_STRUCT{
				Year:     api.SQLSMALLINT(y),
				Month:    api.SQLUSMALLINT(m),
				Day:      api.SQLUSMALLINT(day),
				Hour:     api.SQLUSMALLINT(d.Hour()),
				Minute:   api.SQLUSMALLINT(d.Minute()),
				Second:   api.SQLUSMALLINT(d.Second()),
				Fraction: api.SQLUINTEGER(d.Nanosecond()),
			}
			data = api.Extract(unsafe.Pointer(&t), unsafe.Sizeof(t))
			decimalDigits = 3
			parameterSize = 20 + api.SQLULEN(decimalDigits)
		case []byte:
			ctype = api.SQL_C_BINARY
			sqltype = api.SQL_BINARY
			data = make([]byte, len(d))
			copy(data, d)
			buflen = api.SQLLEN(len(data))
			plen = &buflen
			parameterSize = api.SQLULEN(len(data))
		default:
			panic(fmt.Errorf("unsupported sql.Out.Dest type %T", d))
		}
	} else {
		inputOutputType = api.SQL_PARAM_OUTPUT
		ret := api.SQLDescribeParam(hstmt, api.SQLUSMALLINT(idx+1),
			&sqltype, &parameterSize, &decimalDigits, &nullable)
		if IsError(ret) {
			return nil, NewError("SQLDescribeParam", hstmt)
		}
		data = make([]byte, parameterSize + 1)
		ctype = SqltoCtype(sqltype)
		buflen = api.SQLLEN(len(data))
		plen = &buflen
	}

	return &Out{
		sqlOut:          sqlOut,
		idx:             idx + 1,
		ctype:           ctype,
		sqltype:         sqltype,
		decimalDigits:   decimalDigits,
		nullable:        nullable,
		inputOutputType: inputOutputType,
		parameterSize:   parameterSize,
		data:            data,
		buflen:          buflen,
		plen:            plen,
	}, nil
}

// Value function converts the database value to driver.value
func (o *Out) Value() (driver.Value, error) {
	var p unsafe.Pointer
	buf := o.data
	if len(buf) > 0 {
		p = unsafe.Pointer(&buf[0])
	}
	switch o.ctype {
	case api.SQL_C_BIT:
		return buf[0] != 0, nil
	case api.SQL_C_LONG:
		return *((*int32)(p)), nil
	case api.SQL_C_SBIGINT:
		return *((*int64)(p)), nil
	case api.SQL_C_DOUBLE:
		return *((*float64)(p)), nil
	case api.SQL_C_CHAR:
		buf = bytes.Trim(buf, "\x00")
		return buf, nil
	case api.SQL_C_WCHAR:
		if p == nil {
			return nil, nil
		}
		s := (*[1 << 20]uint16)(p)[:len(buf)/2]
		return utf16toutf8(s), nil
	case api.SQL_C_DBCHAR:
		if p == nil {
			return nil, nil
		}
		s := (*[1 << 20]uint8)(p)[:len(buf)]
		return removeNulls(s), nil
	case api.SQL_C_TYPE_TIMESTAMP:
		t := (*api.SQL_TIMESTAMP_STRUCT)(p)
		r := time.Date(int(t.Year), time.Month(t.Month), int(t.Day),
			int(t.Hour), int(t.Minute), int(t.Second), int(t.Fraction),
			time.Local)
		return r, nil
	case api.SQL_C_TYPE_DATE:
		t := (*api.SQL_DATE_STRUCT)(p)
		r := time.Date(int(t.Year), time.Month(t.Month), int(t.Day),
			0, 0, 0, 0, time.Local)
		return r, nil
	case api.SQL_C_TYPE_TIME:
		t := (*api.SQL_TIME_STRUCT)(p)
		r := time.Date(0, 0, 0,
			int(t.Hour),
			int(t.Minute),
			int(t.Second),
			0,
			time.Local)
		return r, nil
	case api.SQL_C_BINARY:
		return buf, nil
	}
	return nil, fmt.Errorf("unsupported ctype %d for OUT parameter", o.ctype)
}

// ConvertAssign function copies the database data to Dest field in stored procedure.
func (o *Out) ConvertAssign() error {
	if o.sqlOut == nil {
		return fmt.Errorf("sql.Out is nil at OUT param index %d", o.idx)
	}

	if o.sqlOut.Dest == nil {
		return fmt.Errorf("Dest is nil at OUT param index %d", o.idx)
	}

	destInfo := reflect.ValueOf(o.sqlOut.Dest)
	if destInfo.Kind() != reflect.Ptr {
		return fmt.Errorf("Dest at OUT param index %d is not a pointer", o.idx)
	}

	dv, err := o.Value()
	if err != nil {
		return err
	}
	return ConvertAssign(o.sqlOut.Dest, dv)
}

// ConvertAssign function copies the database data to Dest field in stored procedure.
func ConvertAssign(dest, src interface{}) error {
	switch s := src.(type) {
	case string:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = s
			return nil
		case *[]byte:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = []byte(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = append((*d)[:0], s...)
			return nil
		}
	case []byte:
		switch d := dest.(type) {
		case *string:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = string(s)
			return nil
		case *interface{}:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = copyBytes(s)
			return nil
		case *[]byte:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = copyBytes(s)
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = s
			return nil
		}
	case time.Time:
		switch d := dest.(type) {
		case *time.Time:
			*d = s
			return nil
		case *string:
			*d = s.Format(time.RFC3339Nano)
			return nil
		case *[]byte:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = []byte(s.Format(time.RFC3339Nano))
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = s.AppendFormat((*d)[:0], time.RFC3339Nano)
			return nil
		}
	case nil:
		switch d := dest.(type) {
		case *interface{}:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = nil
			return nil
		case *[]byte:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = nil
			return nil
		case *sql.RawBytes:
			if d == nil {
				return errors.New("destination pointer is nil")
			}
			*d = nil
			return nil
		}
	}

	var sv reflect.Value

	switch d := dest.(type) {
	case *string:
		sv = reflect.ValueOf(src)
		switch sv.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			*d = asString(src)
			return nil
		}
	case *[]byte:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes(nil, sv); ok {
			*d = b
			return nil
		}
	case *sql.RawBytes:
		sv = reflect.ValueOf(src)
		if b, ok := asBytes([]byte(*d)[:0], sv); ok {
			*d = sql.RawBytes(b)
			return nil
		}
	case *interface{}:
		*d = src
		return nil
	}

	if scanner, ok := dest.(sql.Scanner); ok {
		return scanner.Scan(src)
	}

	dpv := reflect.ValueOf(dest)
	if dpv.Kind() != reflect.Ptr {
		return errors.New("destination not a pointer")
	}
	if dpv.IsNil() {
		return errors.New("destination pointer is nil")
	}

	if !sv.IsValid() {
		sv = reflect.ValueOf(src)
	}

	dv := reflect.Indirect(dpv)
	if sv.IsValid() && sv.Type().AssignableTo(dv.Type()) {
		switch b := src.(type) {
		case []byte:
			dv.Set(reflect.ValueOf(copyBytes(b)))
		default:
			dv.Set(sv)
		}
		return nil
	}

	if dv.Kind() == sv.Kind() && sv.Type().ConvertibleTo(dv.Type()) {
		dv.Set(sv.Convert(dv.Type()))
		return nil
	}
	switch dv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		s := asString(src)
		i64, err := strconv.ParseInt(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetInt(i64)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		s := asString(src)
		u64, err := strconv.ParseUint(s, 10, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetUint(u64)
		return nil
	case reflect.Float32, reflect.Float64:
		s := asString(src)
		f64, err := strconv.ParseFloat(s, dv.Type().Bits())
		if err != nil {
			err = strconvErr(err)
			return fmt.Errorf("converting driver.Value type %T (%q) to a %s: %v", src, s, dv.Kind(), err)
		}
		dv.SetFloat(f64)
		return nil
	case reflect.String:
		switch v := src.(type) {
		case string:
			dv.SetString(v)
			return nil
		case []byte:
			dv.SetString(string(v))
			return nil
		}
	}

	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type %T", src, dest)
}

func strconvErr(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}

func copyBytes(b []byte) []byte {
	if b == nil {
		return nil
	}
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

func asString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

func asBytes(buf []byte, rv reflect.Value) (b []byte, ok bool) {
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.AppendInt(buf, rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.AppendUint(buf, rv.Uint(), 10), true
	case reflect.Float32:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 32), true
	case reflect.Float64:
		return strconv.AppendFloat(buf, rv.Float(), 'g', -1, 64), true
	case reflect.Bool:
		return strconv.AppendBool(buf, rv.Bool()), true
	case reflect.String:
		s := rv.String()
		return append(buf, s...), true
	}
	return
}

// This function is mirrored in the database/sql/driver package.
func callValuerValue(vr driver.Valuer) (v driver.Value, err error) {
	if rv := reflect.ValueOf(vr); rv.Kind() == reflect.Ptr &&
		rv.IsNil() &&
		rv.Type().Elem().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
		return nil, nil
	}
	return vr.Value()
}
