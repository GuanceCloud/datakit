// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"fmt"
	"testing"

	_ "github.com/denisenkom/go-mssqldb"
)

func TestCon(t *testing.T) {
	n := Input{
		Host:     "10.100.64.109:1433",
		User:     "_",
		Password: "_",
	}
	if err := n.initDB(); err != nil {
		l.Error(err.Error())
		return
	}

	n.getMetric()
	for _, v := range collectCache {
		fmt.Println(v.String())
	}
}
