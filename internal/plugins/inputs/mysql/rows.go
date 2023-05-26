// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

// DB rows interface.
type rows interface {
	Next() bool
	Scan(...interface{}) error
	Close() error
	Err() error
	Columns() ([]string, error)
}

func closeRows(r rows) {
	if err := r.Close(); err != nil {
		l.Warnf("Close: %s, ignored", err)
	}
}

// type myconn interface {
// 	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
// }

// type mydb interface {
// 	Conn() (myconn, error)
// }
