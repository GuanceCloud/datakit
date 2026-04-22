// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/util"
)

const (
	dbmPlanObjectName             = "db_exec_plan"
	defaultPlanSignatureCacheSize = 10000
)

type dbmPlanObjectMeasurement struct{}

func (*dbmPlanObjectMeasurement) Info() *inputs.MeasurementInfo {
	return dbmSampleMeasurementInfo
}

func checkPlanSignatureRate(ipt *Input, planKey string) bool {
	if planKey == "" {
		return false
	}
	if ipt.DbmSample.planSignatureCache == nil {
		ttl := ipt.DbmSample.PlanCacheTTL.Duration
		if ttl <= 0 {
			ttl = time.Hour
		}
		ipt.DbmSample.planSignatureCache = expirable.NewLRU[string, struct{}](defaultPlanSignatureCacheSize, nil, ttl)
	}
	if _, ok := ipt.DbmSample.planSignatureCache.Get(planKey); ok {
		l.Debugf("skip duplicate plan object (LRU): %s", planKey)
		return false
	}
	return true
}

func (ipt *Input) recordReportedPlanSignature(planKey string) {
	if planKey == "" || ipt.DbmSample.planSignatureCache == nil {
		return
	}
	ipt.DbmSample.planSignatureCache.Add(planKey, struct{}{})
}

func (ipt *Input) collectSamplePlans(rows []map[string]any, ptsTime time.Time, maxDuration time.Duration) []*point.Point {
	pts := make([]*point.Point, 0, len(rows))
	obfuscator := util.NewSQLPlanObfuscator()
	start := time.Now()
	for _, row := range rows {
		if time.Since(start) >= maxDuration {
			l.Warnf("stop collecting dbm plans after reaching max duration %s", maxDuration)
			break
		}

		statement := cast.ToString(row["statement"])
		datname := cast.ToString(row["datname"])
		querySignature := cast.ToString(row["query_signature"])

		if statement == "" || cast.ToString(row["backend_type"]) != postgreSQLBackendTypeClient {
			continue
		}
		if querySignature == "" {
			continue
		}

		if !ipt.DbmSample.explainCache.Acquire(querySignature) {
			continue
		}

		plan, err := ipt.getPlan(
			datname,
			cast.ToString(row["query"]),
			statement,
			querySignature,
		)
		if err != nil {
			l.Warnf("get plan failed: %v", err.Error())
			continue
		}
		if plan == "" {
			continue
		}

		normalizedPlan, err := obfuscator.ObfuscateSQLExecPlan(plan, true)
		if err != nil {
			l.Warnf("normalize dbm plan failed: %s, plan: %s", err.Error(), plan)
			continue
		}
		obfuscatedPlan, err := obfuscator.ObfuscateSQLExecPlan(plan, false)
		if err != nil {
			l.Warnf("obfuscate dbm plan failed: %s, plan: %s", err.Error(), plan)
			continue
		}

		planSignature := generatePlanSignature(normalizedPlan)
		planKey := generatePlanCacheKey(querySignature, planSignature)
		if !checkPlanSignatureRate(ipt, planKey) {
			l.Debugf("skip duplicate plan object (LRU): %s", planKey)
			continue
		}

		objectName := fmt.Sprintf("%s-%s", ipt.Object.name, planKey)
		if ipt.databaseInstance != "" {
			objectName = fmt.Sprintf("%s-%s-%s", ipt.Object.name, ipt.databaseInstance, planKey)
		}

		kvs := ipt.getKVs()
		opts := append(point.DefaultObjectOptions(), point.WithTime(ptsTime))
		kvs = kvs.AddTag("name", objectName)
		kvs = kvs.AddTag("server", ipt.Object.name)
		kvs = kvs.AddTag("database_type", "PostgreSQL")
		kvs = kvs.AddTag("plan_type", "JSON")
		kvs = kvs.AddTag("service", "postgresql")
		kvs = kvs.AddTag("plan_signature", planSignature)
		kvs = kvs.AddTag("query_signature", querySignature)
		kvs = kvs.AddTag("client_hostname", cast.ToString(row["client_hostname"]))
		kvs = kvs.AddTag("client_port", cast.ToString(row["client_port"]))
		kvs = kvs.AddTag("client_addr", cast.ToString(row["client_addr"]))
		kvs = kvs.AddTag("application_name", cast.ToString(row["application_name"]))
		kvs = kvs.AddTag("usename", cast.ToString(row["usename"]))
		kvs = kvs.AddTag("db", datname)

		kvs = kvs.Set("message", obfuscatedPlan)
		kvs = kvs.Set("statement", statement)

		pts = append(pts, point.NewPoint(dbmPlanObjectName, kvs, opts...))
		ipt.recordReportedPlanSignature(planKey)
	}

	return pts
}

func (ipt *Input) getPlan(
	datname, statement, obfuscatedStatement, querySignature string,
) (string, error) {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(obfuscatedStatement)), "set") {
		statement = TrimLeadingSetStmts(statement)
		obfuscatedStatement = TrimLeadingSetStmts(obfuscatedStatement)
	}

	if !canExplainStatement(obfuscatedStatement) {
		l.Debugf("explain statement not supported: %s", obfuscatedStatement)
		return "", nil
	}

	if _, ok := ipt.DbmSample.explainErrorCache.Get(querySignature); ok {
		l.Debugf("explain statement error in cache: %s", obfuscatedStatement)
		return "", nil
	}

	if isParameterizedQuery(statement) {
		if ipt.DbmSample.explainParameterizedInstance == nil {
			return "", fmt.Errorf("explain parameterized instance is nil")
		}

		if plan, err := ipt.DbmSample.explainParameterizedInstance.ExplainStatement(datname, statement, obfuscatedStatement, querySignature); err != nil {
			ipt.DbmSample.explainErrorCache.Add(querySignature, struct{}{})
			return "", fmt.Errorf("explain parameterized statement failed: %w", err)
		} else {
			return plan, nil
		}
	}

	if plan, err := ipt.runExplain(datname, statement, obfuscatedStatement); err != nil {
		ipt.DbmSample.explainErrorCache.Add(querySignature, struct{}{})
		return "", fmt.Errorf("run explain failed: %w", err)
	} else {
		return plan, nil
	}
}

func (ipt *Input) runExplain(datname, statement, obfuscatedStatement string) (string, error) {
	conn, err := ipt.service.GetConn(datname)
	if err != nil {
		return "", fmt.Errorf("get conn failed: %w", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), ipt.Timeout.Duration)
	defer cancel()

	var encoding string
	rows, err := conn.Query(ctx, "SHOW client_encoding")
	if err != nil {
		return "", fmt.Errorf("query encoding failed: %w", err)
	}
	if rows.Next() {
		if err := rows.Scan(&encoding); err != nil {
			return "", fmt.Errorf("scan encoding failed: %w", err)
		}
	}
	rows.Close()

	l.Debugf("run explain query on database=%s, statement=%s", datname, obfuscatedStatement)
	if encoding == "SQLASCII" {
		l.Debugf("set client_encoding to utf-8, current encoding is %s", encoding)
		if err := conn.Exec(ctx, "SET client_encoding = 'utf-8'"); err != nil {
			return "", fmt.Errorf("set client encoding failed: %w", err)
		}
	}

	query := fmt.Sprintf("SELECT %s($stmt$%s$stmt$)", "datakit.explain_statement", statement)
	var explainResult string
	rows, err = conn.Query(ctx, query)
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&explainResult); err != nil {
			return "", fmt.Errorf("scan explain result failed: %w", err)
		}
	}
	if len(explainResult) == 0 {
		return "", nil
	}

	return explainResult, nil
}

var supportedExplainStatements = map[string]struct{}{
	"select":  {},
	"table":   {},
	"delete":  {},
	"insert":  {},
	"replace": {},
	"update":  {},
	"with":    {},
}

func canExplainStatement(obfuscateSQL string) bool {
	obfuscateSQL = strings.TrimSpace(obfuscateSQL)

	if strings.HasPrefix(obfuscateSQL, "SELECT datakit.explain_statement") {
		return false
	}
	if strings.HasPrefix(obfuscateSQL, "autovacuum:") {
		return false
	}

	parts := strings.SplitN(obfuscateSQL, " ", 2)
	if len(parts) == 0 {
		return false
	}

	stmtType := strings.ToLower(parts[0])
	_, ok := supportedExplainStatements[stmtType]
	return ok
}

func generatePlanSignature(normalizedPlan string) string {
	h := xxhash.New()
	_, _ = h.WriteString(normalizedPlan)
	return fmt.Sprintf("%016x", h.Sum64())
}

// generatePlanCacheKey generates a unique signature for a execution plan.
func generatePlanCacheKey(querySignature, queryPlanHash string) string {
	h := xxhash.New()
	_, _ = h.WriteString(querySignature)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(queryPlanHash)

	return fmt.Sprintf("%016x", h.Sum64())
}
