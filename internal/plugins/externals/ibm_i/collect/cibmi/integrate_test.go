// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package cibmi

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIntegrationCollect(t *testing.T) {
	dsn := os.Getenv("IBM_I_DSN")
	host := os.Getenv("IBM_I_HOST")
	user := os.Getenv("IBM_I_USER")
	password := os.Getenv("IBM_I_PASSWORD")
	if dsn == "" && (host == "" || user == "") {
		t.Skip("set IBM_I_DSN or IBM_I_HOST/IBM_I_USER to run integration test")
	}
	if !driverRegistered(defaultDriver) {
		t.Skip("ODBC driver is not registered; run with unixODBC, IBM i Access ODBC, and -tags ibm_i")
	}

	ipt := defaultInput()
	ipt.DSN = dsn
	ipt.Host = host
	ipt.User = user
	ipt.Password = password
	require.NoError(t, ipt.setup())

	db, err := ipt.connectDB()
	require.NoError(t, err)
	defer db.Close() //nolint:errcheck
	ipt.db = db

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pts, logPts, err := ipt.collectMetricPoints(ctx, time.Now())
	require.NoError(t, err)
	require.NotEmpty(t, pts)
	require.Empty(t, logPts)
}

func driverRegistered(name string) bool {
	for _, driver := range sql.Drivers() {
		if driver == name {
			return true
		}
	}
	return false
}
