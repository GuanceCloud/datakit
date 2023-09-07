// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package odbc implements database/sql driver to access data via odbc interface.
//
package go_ibm_db

import (
	"database/sql"
	"fmt"

	"github.com/ibmdb/go_ibm_db/api"
)

var drv Driver

type Driver struct {
	Stats
	h api.SQLHENV // environment handle
}

func initDriver() error {

	//Allocate environment handle
	var out api.SQLHANDLE
	in := api.SQLHANDLE(api.SQL_NULL_HANDLE)
	ret := api.SQLAllocHandle(api.SQL_HANDLE_ENV, in, &out)
	if IsError(ret) {
		return NewError("SQLAllocHandle", api.SQLHENV(in))
	}
	drv.h = api.SQLHENV(out)
	drv.Stats.updateHandleCount(api.SQL_HANDLE_ENV, 1)

	// will use ODBC v3
	ret = api.SQLSetEnvUIntPtrAttr(drv.h, api.SQL_ATTR_ODBC_VERSION,
		api.SQL_OV_ODBC3, 0)
	if IsError(ret) {
		defer releaseHandle(drv.h)
		return NewError("SQLSetEnvAttr ODBC v3", drv.h)
	}

	return nil
}

func (d *Driver) Close() error {
	// TODO(brainman): who will call (*Driver).Close (to dispose all opened handles)?
	h := d.h
	d.h = api.SQLHENV(api.SQL_NULL_HENV)
	return releaseHandle(h)
}

func init() {

	// Recover from panic to avoid stop an application when can't get the db2 cli
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(fmt.Sprintf("%s\nThe go_ibm_db driver cannot be registered", err))
		}
	}()

	err := initDriver()
	if err != nil {
		panic(err)
	}
	//go's to databse/sql/sql.go 43 line
	sql.Register("go_ibm_db", &drv)

}
