// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapMysqlWaitEventToGroup(t *testing.T) {
	cases := []struct {
		name      string
		waitEvent string
		want      string
	}{
		{name: "cpu exact", waitEvent: "CPU", want: "CPU"},
		{name: "lock path", waitEvent: "wait/lock/table/sql/handler", want: "Lock"},
		{name: "sync prefix", waitEvent: "wait/synch/mutex/sql/LOCK_table_cache", want: "Concurrency"},
		{name: "table handler", waitEvent: "wait/io/table/sql/handler", want: "Concurrency"},
		{name: "network socket", waitEvent: "wait/io/socket/sql/client_connection", want: "Network"},
		{name: "redo log", waitEvent: "wait/io/file/innodb/innodb_log_file", want: "Commit/Log"},
		{name: "io generic", waitEvent: "wait/io/file/sql/binlog", want: "Commit/Log"},
		{name: "memory", waitEvent: "innodb_buffer_pool", want: "Memory"},
		{name: "idle", waitEvent: "idle", want: "Other"},
		{name: "sleep", waitEvent: "User sleep", want: "Other"},
		{name: "unknown", waitEvent: "some/custom/event", want: "Other"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapMysqlWaitEventToGroup(tc.waitEvent)
			assert.Equal(t, tc.want, got)
		})
	}
}
