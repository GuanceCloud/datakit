// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
)

type explainQueryService struct {
	Service
	conn Conn
}

func (s *explainQueryService) GetConn(string) (Conn, error) {
	return s.conn, nil
}

type explainQueryConn struct {
	Conn
	query func(context.Context, string, ...any) (Rows, error)
}

func (c *explainQueryConn) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return c.query(ctx, sql, args...)
}

func (c *explainQueryConn) Close() {}

type explainQueryRows struct {
	Rows
	value string
	read  bool
}

func (r *explainQueryRows) Next() bool {
	if r.read {
		return false
	}
	r.read = true
	return true
}

func (r *explainQueryRows) Scan(dest ...interface{}) error {
	*dest[0].(*string) = r.value
	return nil
}

func (r *explainQueryRows) Close() {}

func TestRunExplainBindsStatement(t *testing.T) {
	for _, statement := range []string{
		"SELECT 1",
		"SELECT '$stmt$'",
		"SELECT 1$stmt$) UNION ALL SELECT current_user--",
	} {
		t.Run(statement, func(t *testing.T) {
			const plan = `[{"Plan":{"Node Type":"Result"}}]`
			var explainCalls int
			conn := &explainQueryConn{query: func(_ context.Context, query string, args ...any) (Rows, error) {
				if query == "SHOW client_encoding" {
					return &explainQueryRows{value: "UTF8"}, nil
				}
				explainCalls++
				require.Equal(t, "SELECT datakit.explain_statement($1)", query)
				require.Equal(t, []any{statement}, args, "sampled SQL must be passed unchanged as a parameter")
				return &explainQueryRows{value: plan}, nil
			}}
			ipt := &Input{
				service: &explainQueryService{conn: conn},
				Timeout: datakit.Duration{Duration: time.Second},
			}
			actual, err := ipt.runExplain("", statement, "SELECT ?")
			require.NoError(t, err)
			require.Equal(t, plan, actual)
			require.Equal(t, 1, explainCalls)
		})
	}
}

// Requires DATAKIT_POSTGRES_TEST_DSN pointing to a disposable database without a datakit schema.
func TestRunExplainQuotedStatementIntegration(t *testing.T) {
	for _, mode := range []pgx.QueryExecMode{pgx.QueryExecModeCacheStatement, pgx.QueryExecModeSimpleProtocol} {
		t.Run(mode.String(), func(t *testing.T) {
			pool := newPrepareTestPool(t, mode)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			// Require a new schema so an existing explain function is never replaced.
			_, err := pool.Exec(ctx, "CREATE SCHEMA datakit")
			require.NoError(t, err)
			t.Cleanup(func() {
				cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cleanupCancel()
				_, err := pool.Exec(cleanupCtx, "DROP SCHEMA datakit CASCADE")
				require.NoError(t, err)
			})
			_, err = pool.Exec(ctx, `CREATE FUNCTION datakit.explain_statement(q text)
RETURNS json LANGUAGE plpgsql AS $$
DECLARE p json;
BEGIN
  EXECUTE 'EXPLAIN (FORMAT JSON) ' || q INTO p;
  RETURN p;
END $$`)
			require.NoError(t, err)
			ipt := &Input{
				service: &SQLService{pool: pool},
				Timeout: datakit.Duration{Duration: 5 * time.Second},
			}
			for _, statement := range []string{
				"SELECT 1",
				"SELECT '$stmt$'",
				"SELECT $stmt$hello$stmt$",
				"SELECT 1 /* $stmt$ */",
				"SELECT 'it''s $stmt$ 中文'",
				`SELECT E'back\\slash; $stmt$'`,
			} {
				t.Run(statement, func(t *testing.T) {
					plan, err := ipt.runExplain("", statement, "SELECT ?")
					require.NoError(t, err)
					require.True(t, json.Valid([]byte(plan)), plan)
					require.Contains(t, plan, `"Plan"`)
				})
			}
		})
	}
}
