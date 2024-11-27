// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	ws "gitlab.jiagouyun.com/cloudcare-tools/datakit/dca/websocket"
	_ "modernc.org/sqlite"
)

func TestDB(t *testing.T) {
	db := NewDB()
	err := db.Init()

	assert.NoError(t, err)

	err = db.Insert(&ws.DataKit{})
	assert.NoError(t, err)

	res := []ws.DataKit{}
	err = db.Select("select * from datakit", &res)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res))

	t.Log(res)
}
