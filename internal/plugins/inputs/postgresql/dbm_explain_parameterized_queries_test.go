// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type prepareProtocolConn struct {
	Conn
	extendedErr error
	simpleCalls int
}

func (c *prepareProtocolConn) ExecExtended(context.Context, string) error {
	return c.extendedErr
}

func (c *prepareProtocolConn) Exec(context.Context, string, ...any) error {
	c.simpleCalls++
	return nil
}

func TestPrepareDoesNotFallbackToSimpleProtocol(t *testing.T) {
	rejected := errors.New("extended protocol rejected statement")
	conn := &prepareProtocolConn{extendedErr: rejected}
	explainer := &ExplainParameterizedQueries{}
	err := explainer.createPreparedStatement(context.Background(), conn,
		"SELECT 1 /* $1 */; SET application_name = 'unexpected'", "SELECT ?", "test")
	require.ErrorIs(t, err, rejected)
	require.Zero(t, conn.simpleCalls)
}

// Set DATAKIT_POSTGRES_TEST_DSN to a disposable PostgreSQL 12+ database.
// All test objects are session-local; the pool has one connection to verify
// that rejection leaves the same session usable and free of side effects.
func TestExplainStatementSingleStatementIntegration(t *testing.T) {
	for _, mode := range []pgx.QueryExecMode{pgx.QueryExecModeCacheStatement, pgx.QueryExecModeSimpleProtocol} {
		t.Run(mode.String(), func(t *testing.T) {
			pool := newPrepareTestPool(t, mode)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := pool.Exec(ctx, "CREATE TEMP TABLE dk_plan_test_session (id int)")
			require.NoError(t, err)
			_, err = pool.Exec(ctx, `CREATE FUNCTION pg_temp.dk_explain_statement(q text)
RETURNS json LANGUAGE plpgsql AS $$
DECLARE p json;
BEGIN
  EXECUTE 'EXPLAIN (FORMAT JSON) ' || q INTO p;
  RETURN p;
END $$`)
			require.NoError(t, err)
			_, err = pool.Exec(ctx, "SET datakit.prepare_regression = 'unchanged'")
			require.NoError(t, err)

			var originalPID int
			require.NoError(t, pool.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&originalPID))
			ipt := &Input{
				service: &SQLService{pool: pool},
				version: V120,
				Timeout: datakit.Duration{Duration: 5 * time.Second},
			}
			explainer, err := NewExplainParameterizedQueries(ipt, "pg_temp.dk_explain_statement")
			require.NoError(t, err)

			for _, tc := range []struct {
				name      string
				statement string
				reject    bool
			}{
				{name: "parameter", statement: "SELECT $1::int"},
				{name: "trailing_semicolon", statement: "SELECT $1::int;"},
				{name: "literal_semicolon", statement: "SELECT $1::int, ';'::text"},
				{name: "comment_semicolon", statement: "SELECT $1::int /* ; SELECT 2 */"},
				{name: "dollar_quote", statement: "SELECT $1::int, $text$; SELECT 2$text$::text"},
				{name: "extra_set", statement: "SELECT 1 /* $1 */; SET datakit.prepare_regression = 'changed'", reject: true},
				{name: "extra_create", statement: "SELECT 1; CREATE TEMP TABLE dk_injection_marker(x int); --$1", reject: true},
				{name: "parameter_and_extra_set", statement: "SELECT $1::int; SET datakit.prepare_regression = 'changed'", reject: true},
				{name: "after_rejection", statement: "SELECT $1::int, $2::text"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					// These include comment-only placeholders that currently enter the parameterized path.
					require.True(t, isParameterizedQuery(tc.statement))
					plan, err := explainer.ExplainStatement("", tc.statement, "SELECT ?", tc.name)
					if tc.reject {
						require.Error(t, err)
						var pgErr *pgconn.PgError
						require.ErrorAs(t, err, &pgErr)
						require.Equal(t, "42601", pgErr.Code)
						require.Contains(t, pgErr.Message, "cannot insert multiple commands into a prepared statement")
						require.Empty(t, plan)
					} else {
						require.NoError(t, err)
						require.True(t, json.Valid([]byte(plan)), plan)
						require.Contains(t, plan, `"Plan"`)
					}

					var pid, preparedCount int
					var marker string
					var tableName *string
					err = pool.QueryRow(ctx, `SELECT pg_backend_pid(),
current_setting('datakit.prepare_regression'), to_regclass('pg_temp.dk_injection_marker')::text,
(SELECT count(*) FROM pg_prepared_statements WHERE name LIKE 'dk_%')`).Scan(&pid, &marker, &tableName, &preparedCount)
					require.NoError(t, err)
					require.Equal(t, originalPID, pid, "must check the same session for side effects")
					require.Equal(t, "unchanged", marker)
					require.Nil(t, tableName)
					require.Zero(t, preparedCount)
				})
			}
		})
	}
}

func TestExecExtendedTimeoutIntegration(t *testing.T) {
	pool := newPrepareTestPool(t, pgx.QueryExecModeCacheStatement)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer timeoutCancel()
	err = (&pgxConn{conn}).ExecExtended(timeoutCtx, "SELECT pg_sleep(10)")
	conn.Release()
	require.ErrorIs(t, err, context.DeadlineExceeded)
	// A canceled connection may be discarded; the pool must still be usable.
	require.NoError(t, pool.Ping(ctx))
}

func newPrepareTestPool(t *testing.T, mode pgx.QueryExecMode) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("DATAKIT_POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("set DATAKIT_POSTGRES_TEST_DSN to run PostgreSQL protocol integration tests")
	}
	config, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	config.MaxConns = 1
	config.ConnConfig.DefaultQueryExecMode = mode
	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return pool
}
