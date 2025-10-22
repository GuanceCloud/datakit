// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
)

const (
	prepareStatementQuery         = "PREPARE dk_%s AS %s"
	paramTypesCountQuery          = "SELECT CARDINALITY(parameter_types) FROM pg_prepared_statements WHERE name = 'dk_%s'"
	executePreparedStatementQuery = "EXECUTE dk_%s%s"
	explainQuery                  = "SELECT %s($1)"
)

var parameterizedQueryPattern = regexp.MustCompile(`\$\d+`)

type ExplainParameterizedQueries struct {
	input           *Input
	explainFunction string
}

func NewExplainParameterizedQueries(input *Input, explainFunction string) (*ExplainParameterizedQueries, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}
	return &ExplainParameterizedQueries{
		input:           input,
		explainFunction: explainFunction,
	}, nil
}

func (e *ExplainParameterizedQueries) ExplainStatement(dbname, statement, obfuscatedStatement string) (string, error) {
	if e.input.version.LessThan(*V120) {
		return "", fmt.Errorf("database version %s does not support plan_cache_mode", e.input.version.String())
	}

	conn, err := e.input.service.GetConn(dbname)
	if err != nil {
		return "", fmt.Errorf("get conn error: %w", err)
	}
	defer conn.Close()

	if err := e.setPlanCacheMode(conn); err != nil {
		return "", fmt.Errorf("set plan cache mode error: %w", err)
	}

	querySignature := util.ComputeSQLSignature(obfuscatedStatement)

	if err := e.createPreparedStatement(conn, statement, obfuscatedStatement, querySignature); err != nil {
		return "", fmt.Errorf("create prepared statement error: %w", err)
	}

	defer e.deallocatePreparedStatement(conn, querySignature)

	plan, err := e.explainPreparedStatement(conn, statement, obfuscatedStatement, querySignature)
	if err != nil {
		return "", fmt.Errorf("failed to explain with prepared statement error: %w", err)
	}

	return plan, nil
}

func (e *ExplainParameterizedQueries) setPlanCacheMode(conn Conn) error {
	query := "SET plan_cache_mode = force_generic_plan"
	return conn.Exec(context.Background(), query)
}

func (e *ExplainParameterizedQueries) createPreparedStatement(conn Conn, statement, obfuscatedStatement, querySignature string) error {
	query := fmt.Sprintf(prepareStatementQuery, querySignature, statement)
	err := conn.Exec(context.Background(), query)
	if err != nil {
		return fmt.Errorf("create prepared statement error: %w", err)
	}
	return nil
}

func (e *ExplainParameterizedQueries) getNumberOfParametersForPreparedStatement(conn Conn, querySignature string) (int, error) {
	query := fmt.Sprintf(paramTypesCountQuery, querySignature)
	rows, err := e.executeQueryAndFetchRows(conn, query)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}

	var paramCount int
	if err := rows.Scan(&paramCount); err != nil {
		return 0, err
	}
	return paramCount, nil
}

func (e *ExplainParameterizedQueries) generatePreparedStatementQuery(conn Conn, querySignature string) (string, error) {
	paramCount, err := e.getNumberOfParametersForPreparedStatement(conn, querySignature)
	if err != nil {
		return "", err
	}

	params := ""
	if paramCount > 0 {
		nullParams := make([]string, paramCount)
		for i := range nullParams {
			nullParams[i] = "null"
		}
		params = fmt.Sprintf("(%s)", strings.Join(nullParams, ","))
	}

	return fmt.Sprintf(executePreparedStatementQuery, querySignature, params), nil
}

func (e *ExplainParameterizedQueries) explainPreparedStatement(conn Conn, statement, obfuscatedStatement, querySignature string) (string, error) {
	preparedQuery, err := e.generatePreparedStatementQuery(conn, querySignature)
	if err != nil {
		return "", err
	}

	query := fmt.Sprintf(explainQuery, e.explainFunction)
	rows, err := e.executeQueryAndFetchRows(conn, query, preparedQuery)
	if err != nil {
		return "", fmt.Errorf("fetch rows failed: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return "", nil
	}

	var plan string
	if err := rows.Scan(&plan); err != nil {
		return "", err
	}
	return plan, nil
}

func (e *ExplainParameterizedQueries) deallocatePreparedStatement(conn Conn, querySignature string) {
	query := fmt.Sprintf("DEALLOCATE PREPARE dk_%s", querySignature)
	if err := conn.Exec(context.Background(), query); err != nil {
		l.Errorf("deallocate prepared statement error: %s", err.Error())
	}
}

func (e *ExplainParameterizedQueries) executeQueryAndFetchRows(conn Conn, query string, args ...interface{}) (Rows, error) {
	rows, err := conn.Query(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	return rows, err
}

func isParameterizedQuery(statement string) bool {
	matches := parameterizedQueryPattern.FindAllStringIndex(statement, -1)
	if len(matches) == 0 {
		return false
	}

	for _, match := range matches {
		start, end := match[0], match[1]
		leftQuote := strings.LastIndex(statement[:start], "'")
		rightQuote := strings.Index(statement[end:], "'")

		if leftQuote == -1 || (rightQuote != -1 && leftQuote%2 == 0) {
			return true
		}
	}
	return false
}
