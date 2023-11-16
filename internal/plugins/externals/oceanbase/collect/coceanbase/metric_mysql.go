// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

//go:build linux
// +build linux

package coceanbase

import (
	"database/sql"
	"fmt"
)

// MySQL [oceanbase]> select * from gv$server_schema_info\G;
// *************************** 1. row ***************************
//                     SVR_IP: 10.200.14.240
//                   SVR_PORT: 2882
//                  TENANT_ID: 1001
//   REFRESHED_SCHEMA_VERSION: 1699950217506448
//    RECEIVED_SCHEMA_VERSION: 1699950217506448
//               SCHEMA_COUNT: 22
//                SCHEMA_SIZE: 3822
// MIN_SSTABLE_SCHEMA_VERSION: -1
// 1 row in set (0.014 sec)

//nolint:stylecheck
const MYSQL_SERVER_SCHEMA_INFO = `select SVR_IP, SVR_PORT from gv$server_schema_info`

type msServerSchemaInfo struct {
	SvrIP   sql.NullString `db:"SVR_IP"`
	SvrPort sql.NullInt64  `db:"SVR_PORT"`
}

func msGetServerSchemaInfo(ipt *Input) (string, error) {
	rows := []msServerSchemaInfo{}
	if err := selectWrapper(ipt, &rows, MYSQL_SERVER_SCHEMA_INFO); err != nil {
		return "", err
	}

	for _, v := range rows {
		ip := v.SvrIP.String
		port := v.SvrPort.Int64

		if len(ip) > 0 && port > 0 {
			return fmt.Sprintf("%s:%d", v.SvrIP.String, v.SvrPort.Int64), nil
		}
	}

	return "", errEmptyRow
}

//nolint:lll
// MySQL [oceanbase]> SHOW VARIABLES;
// +--------------------------------------+---------------------------------------------------------------------------------------------------------------+
// | Variable_name                        | Value                                                                                                         |
// +--------------------------------------+---------------------------------------------------------------------------------------------------------------+
// | version                              | 5.7.25-OceanBase-v3.2.4.3                                                                                     |
// | version_comment                      | OceanBase 3.2.4.3 (r103020012023051810-a562da822afa7979edf77bb00d0b4f8e222dc937) (Built May 18 2023 10:41:35) |

//nolint:stylecheck
const MYSQL_SHOW_VARIABLES = `SHOW VARIABLES`

type msShowVariables struct {
	VariableName sql.NullString `db:"Variable_name"`
	Value        sql.NullString `db:"Value"`
}

func msGetShowVariables(ipt *Input) (string, error) {
	rows := []msShowVariables{}
	if err := selectWrapper(ipt, &rows, MYSQL_SHOW_VARIABLES); err != nil {
		return "", err
	}

	for _, v := range rows {
		if v.VariableName.String == "version" {
			return v.Value.String, nil
		}
	}

	return "", errEmptyRow
}

////////////////////////////////////////////////////////////////////////////////

func getMySQLStatus(ipt *Input) (*dbState, error) {
	hostName, err := msGetServerSchemaInfo(ipt)
	if err != nil {
		return nil, err
	}

	version, err := msGetShowVariables(ipt)
	if err != nil {
		return nil, err
	}

	return &dbState{
		Hostname: hostName,
		Version:  version,
		Status:   OK,
	}, nil
}
