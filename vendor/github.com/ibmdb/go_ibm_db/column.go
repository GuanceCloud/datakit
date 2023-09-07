// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package go_ibm_db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"
        "bytes"

	"github.com/ibmdb/go_ibm_db/api"
)

type BufferLen api.SQLLEN

func (l *BufferLen) IsNull() bool {
	return int16(*l) == api.SQL_NULL_DATA
}

func (l *BufferLen) GetData(h api.SQLHSTMT, idx int, ctype api.SQLSMALLINT, buf []byte) api.SQLRETURN {
	return api.SQLGetData(h, api.SQLUSMALLINT(idx+1), ctype,
		api.SQLPOINTER(unsafe.Pointer(&buf[0])), api.SQLLEN(len(buf)),
		(*api.SQLLEN)(l))
}

func (l *BufferLen) Bind(h api.SQLHSTMT, idx int, ctype api.SQLSMALLINT, buf []byte) api.SQLRETURN {
	if len(buf) <= 2147483647 {
		return api.SQLBindCol(h, api.SQLUSMALLINT(idx+1), ctype,
			buf, api.SQLLEN(len(buf)),
			(*api.SQLLEN)(l))
	}
	return api.SQLBindCol(h, api.SQLUSMALLINT(idx+1), ctype,
		buf, api.SQLLEN(len(buf)-1),
		(*api.SQLLEN)(l))
}

// Column provides access to row columns.
type Column interface {
	Name() string
	TypeScan() reflect.Type
	Bind(h api.SQLHSTMT, idx int) (bool, error)
	Value(h api.SQLHSTMT, idx int) (driver.Value, error)
}

func describeColumn(h api.SQLHSTMT, idx int, namebuf []uint16) (namelen int, sqltype api.SQLSMALLINT, size api.SQLULEN, ret api.SQLRETURN) {
	var l, decimal, nullable api.SQLSMALLINT
	ret = api.SQLDescribeCol(h, api.SQLUSMALLINT(idx+1),
		(*api.SQLWCHAR)(unsafe.Pointer(&namebuf[0])),
		api.SQLSMALLINT(len(namebuf)), &l,
		&sqltype, &size, &decimal, &nullable)
	return int(l), sqltype, size, ret
}

// TODO(brainman): did not check for MS SQL timestamp

func NewColumn(h api.SQLHSTMT, idx int) (Column, error) {
	namebuf := make([]uint16, 150)
	namelen, sqltype, size, ret := describeColumn(h, idx, namebuf)
	if ret == api.SQL_SUCCESS_WITH_INFO && namelen > len(namebuf) {
		// try again with bigger buffer
		namebuf = make([]uint16, namelen)
		namelen, sqltype, size, ret = describeColumn(h, idx, namebuf)
	}
	if IsError(ret) {
		return nil, NewError("SQLDescribeCol", h)
	}
	if namelen > len(namebuf) {
		// still complaining about buffer size
		return nil, errors.New("Failed to allocate column name buffer")
	}
	b := &BaseColumn{
		name:  api.UTF16ToString(namebuf[:namelen]),
		SType: sqltype,
	}
	switch sqltype {
	case api.SQL_BIT, api.SQL_BOOLEAN:
		return NewBindableColumn(b, api.SQL_C_BIT, 1), nil
	case api.SQL_TINYINT, api.SQL_SMALLINT, api.SQL_INTEGER:
		return NewBindableColumn(b, api.SQL_C_LONG, 4), nil
	case api.SQL_BIGINT:
		return NewBindableColumn(b, api.SQL_C_SBIGINT, 8), nil
	case api.SQL_NUMERIC, api.SQL_FLOAT, api.SQL_REAL, api.SQL_DOUBLE:
		return NewBindableColumn(b, api.SQL_C_DOUBLE, 8), nil
	case api.SQL_TYPE_TIMESTAMP:
		var v api.SQL_TIMESTAMP_STRUCT
		return NewBindableColumn(b, api.SQL_C_TYPE_TIMESTAMP, int(unsafe.Sizeof(v))), nil
	case api.SQL_TYPE_DATE:
		var v api.SQL_DATE_STRUCT
		return NewBindableColumn(b, api.SQL_C_TYPE_DATE, int(unsafe.Sizeof(v))), nil
	case api.SQL_TYPE_TIME:
		var v api.SQL_TIME_STRUCT
		return NewBindableColumn(b, api.SQL_C_TYPE_TIME, int(unsafe.Sizeof(v))), nil
	case api.SQL_CHAR, api.SQL_VARCHAR, api.SQL_CLOB, api.SQL_DECFLOAT, api.SQL_DECIMAL:
		return NewVariableWidthColumn(b, api.SQL_C_CHAR, size), nil
	case api.SQL_WCHAR, api.SQL_WVARCHAR:
		return NewVariableWidthColumn(b, api.SQL_C_WCHAR, size), nil
	case api.SQL_BINARY, api.SQL_VARBINARY, api.SQL_BLOB:
		return NewVariableWidthColumn(b, api.SQL_C_BINARY, size), nil
	case api.SQL_LONGVARCHAR:
		return NewVariableWidthColumn(b, api.SQL_C_CHAR, size), nil
	case api.SQL_WLONGVARCHAR, api.SQL_SS_XML:
		return NewVariableWidthColumn(b, api.SQL_C_WCHAR, size), nil
	case api.SQL_LONGVARBINARY:
		return NewVariableWidthColumn(b, api.SQL_C_BINARY, 0), nil
	case api.SQL_DBCLOB:
		return NewVariableWidthColumn(b, api.SQL_C_DBCHAR, size), nil
	case api.SQL_XML:
		return NewVariableWidthColumn(b, api.SQL_C_BINARY, 31457280), nil
	default:
		return nil, fmt.Errorf("unsupported column type %d", sqltype)
	}
	panic("unreachable")
}

// BaseColumn implements common column functionality.
type BaseColumn struct {
	name  string
	CType api.SQLSMALLINT
	SType api.SQLSMALLINT
}

func (c *BaseColumn) Name() string {
	return c.name
}

func (c *BaseColumn) TypeScan() reflect.Type {
	//TODO(Akhil):This will return the golang type of a variable
	switch c.CType {
	case api.SQL_C_BIT:
		return reflect.TypeOf(false)
	case api.SQL_C_LONG:
		return reflect.TypeOf(int32(0))
	case api.SQL_C_SBIGINT:
		return reflect.TypeOf(int64(0))
	case api.SQL_C_DOUBLE:
		return reflect.TypeOf(float64(0.0))
	case api.SQL_C_CHAR, api.SQL_C_WCHAR:
		if c.SType == api.SQL_DECFLOAT {
			return reflect.TypeOf(float64(0.0))
		}
		return reflect.TypeOf(string(""))
	case api.SQL_C_TYPE_DATE, api.SQL_C_TYPE_TIME, api.SQL_C_TYPE_TIMESTAMP:
		return reflect.TypeOf(time.Time{})
	case api.SQL_C_BINARY:
		return reflect.TypeOf([]byte(nil))
	default:
		return reflect.TypeOf(new(interface{}))
	}
	return reflect.TypeOf(new(interface{}))
}

func (c *BaseColumn) Value(buf []byte) (driver.Value, error) {
	var p unsafe.Pointer
	if len(buf) > 0 {
		p = unsafe.Pointer(&buf[0])
	}
	switch c.CType {
	case api.SQL_C_BIT:
		return buf[0] != 0, nil
	case api.SQL_C_LONG:
		return *((*int32)(p)), nil
	case api.SQL_C_SBIGINT:
		return *((*int64)(p)), nil
	case api.SQL_C_DOUBLE:
		return *((*float64)(p)), nil
	case api.SQL_C_CHAR:
		if c.SType == api.SQL_DECIMAL {
			    return bytes.Replace(buf, []byte(","), []byte("."), 1), nil
		}
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
		r := time.Date(1, 1, 1,
			int(t.Hour),
			int(t.Minute),
			int(t.Second),
			0,
			time.Local)
		return r, nil
	case api.SQL_C_BINARY:
		return buf, nil
	}
	return nil, fmt.Errorf("unsupported column ctype %d", c.CType)
}

// BindableColumn allows access to columns that can have their buffers
// bound. Once bound at start, they are written to by odbc driver every
// time it fetches new row. This saves on syscall and, perhaps, some
// buffer copying. BindableColumn can be left unbound, then it behaves
// like NonBindableColumn when user reads data from it.
type BindableColumn struct {
	*BaseColumn
	IsBound         bool
	IsVariableWidth bool
	Size            int
	Len             BufferLen
	Buffer          []byte
	smallBuf        [8]byte // small inline memory buffer, so we do not need allocate external memory all the time
}

func NewBindableColumn(b *BaseColumn, ctype api.SQLSMALLINT, bufSize int) *BindableColumn {
	b.CType = ctype
	c := &BindableColumn{BaseColumn: b, Size: bufSize}
	if c.Size <= len(c.smallBuf) {
		// use inline buffer
		c.Buffer = c.smallBuf[:c.Size]
	} else {
		c.Buffer = make([]byte, c.Size)
	}
	return c
}

func NewVariableWidthColumn(b *BaseColumn, ctype api.SQLSMALLINT, colWidth api.SQLULEN) Column {
	if colWidth == 0 {
		b.CType = ctype
		return &NonBindableColumn{b}
	}
	l := int(colWidth)
	switch ctype {
	case api.SQL_C_WCHAR, api.SQL_C_DBCHAR:
		l++    // room for null-termination character
		l *= 2 // wchars take 2 bytes each
	case api.SQL_C_CHAR:
		if b.SType == api.SQL_DECIMAL {
			l = l + 4 // adding 4 as decimal has '.' which takes 1 byte
		} else {
			l++     // room for null-termination character
			l *= 2  //chars take 2 bytes each
		}
	case api.SQL_C_BINARY:
		// nothing to do
	default:
		panic(fmt.Errorf("do not know how wide column of ctype %d is", ctype))
	}
	c := NewBindableColumn(b, ctype, l)
	c.IsVariableWidth = true
	return c
}

func (c *BindableColumn) Bind(h api.SQLHSTMT, idx int) (bool, error) {
	ret := c.Len.Bind(h, idx, c.CType, c.Buffer)
	if IsError(ret) {
		return false, NewError("SQLBindCol", h)
	}
	c.IsBound = true
	return true, nil
}

func (c *BindableColumn) Value(h api.SQLHSTMT, idx int) (driver.Value, error) {
	if !c.IsBound {
		ret := c.Len.GetData(h, idx, c.CType, c.Buffer)
		if IsError(ret) {
			return nil, NewError("SQLGetData", h)
		}
	}
	if c.Len.IsNull() {
		// is NULL
		return nil, nil
	}
	if !c.IsVariableWidth && int(c.Len) != c.Size {
		panic(fmt.Errorf("wrong column #%d length %d returned, %d expected", idx, c.Len, c.Size))
	}
	// check buffer len
	bufferLen := int(c.Len)
	if len(c.Buffer) < bufferLen {
		bufferLen = len(c.Buffer)
	}

	return c.BaseColumn.Value(c.Buffer[:bufferLen])
}

// NonBindableColumn provide access to columns, that can't be bound.
// These are of character or binary type, and, usually, there is no
// limit for their width.
type NonBindableColumn struct {
	*BaseColumn
}

func (c *NonBindableColumn) Bind(h api.SQLHSTMT, idx int) (bool, error) {
	return false, nil
}

func (c *NonBindableColumn) Value(h api.SQLHSTMT, idx int) (driver.Value, error) {
	var l BufferLen
	var total []byte
	b := make([]byte, 1024)
loop:
	for {
		ret := l.GetData(h, idx, c.CType, b)
		switch ret {
		case api.SQL_SUCCESS:
			if l.IsNull() {
				// is NULL
				return nil, nil
			}
			total = append(total, b[:l]...)
			break loop
		case api.SQL_SUCCESS_WITH_INFO:
			err := NewError("SQLGetData", h).(*Error)
			if len(err.Diag) > 0 && err.Diag[0].State != "01004" {
				return nil, err
			}
			i := len(b)
			switch c.CType {
			case api.SQL_C_WCHAR, api.SQL_C_DBCHAR:
				i -= 2 // remove wchar (2 bytes) null-termination character
			case api.SQL_C_CHAR:
				i-- // remove null-termination character
			}
			total = append(total, b[:i]...)
			if l != api.SQL_NO_TOTAL {
				// odbc gives us a hint about remaining data,
				// lets get it in one go.
				n := int(l) // total bytes for our data
				n -= i      // subtract already received
				n += 2      // room for biggest (wchar) null-terminator
				if len(b) < n {
					b = make([]byte, n)
				}
			}
		default:
			return nil, NewError("SQLGetData", h)
		}
	}
	return c.BaseColumn.Value(total)
}
