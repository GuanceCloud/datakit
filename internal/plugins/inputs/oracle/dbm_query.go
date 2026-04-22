// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/hashicorp/golang-lru/v2/expirable"
	go_ora "github.com/sijms/go-ora/v2"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

const (
	dbmQueryObjectName            = "db_query"
	MaxSQLFullTextVSQLStats int16 = 1000 // SQL_FULLTEXT size in V$SQLSTATS is limited to 1000 characters
)

type dbmQueryObjectMeasurement struct{}

func (*dbmQueryObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dbmQueryObjectName,
		Cat:  point.Object,
		//nolint:lll
		Desc:   "Oracle DBM query objects. Each object represents a unique SQL query identified by server:database_instance:query_signature, containing the obfuscated SQL text.",
		DescZh: "Oracle DBM SQL 文本对象。每个对象代表一个由 server:database_instance:query_signature 唯一标识的 SQL 查询，包含脱敏后的 SQL 文本。",
		Tags: map[string]interface{}{
			"name":                     inputs.NewTagInfo("Object identifier generated from server:database_instance:query_signature"),
			"query_signature":          inputs.NewTagInfo("Hash signature generated from pdb_name:query_hash to link metrics and objects"),
			"server":                   inputs.NewTagInfo("The server address (host:port)"),
			"database_instance":        inputs.NewTagInfo("Oracle instance identifier from configured tag or v$instance.host_name."),
			"database_type":            inputs.NewTagInfo("The type of the database. The value is `Oracle`"),
			"con_id":                   inputs.NewTagInfo("The container ID (con_id) in Oracle multi tenant architecture"),
			"cdb_name":                 inputs.NewTagInfo("The name of the CDB (Container Database)"),
			"pdb_name":                 inputs.NewTagInfo("The name of the PDB (Pluggable Database)"),
			"force_matching_signature": inputs.NewTagInfo("The force matching signature of the query"),
			"sql_id":                   inputs.NewTagInfo("The SQL ID of the query "),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The obfuscated/normalized SQL text (full text)",
			},
		},
	}
}

func (ipt *Input) collectDbmQueries(oracleRows []*OracleRow, ptsTime time.Time) {
	if len(oracleRows) == 0 {
		return
	}

	// Initialize query object cache if not already initialized
	if ipt.dbmQueryObjectCache == nil {
		ipt.dbmQueryObjectCache = expirable.NewLRU[string, struct{}](
			100000,
			nil,
			24*time.Hour,
		)
	}

	start := time.Now()

	opts := append(point.DefaultObjectOptions(), point.WithTime(ptsTime))
	var pts []*point.Point

	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{})
	for _, row := range oracleRows {
		// Use query signature from OracleRow
		querySignature := row.querySignature

		if querySignature == "" {
			continue
		}

		// Check cache to avoid duplicate reporting
		if ipt.dbmQueryObjectCache != nil {
			if _, ok := ipt.dbmQueryObjectCache.Get(querySignature); ok {
				continue
			}
		}

		kvs := ipt.getKVs()
		// Tags
		objectName := fmt.Sprintf("%s-%s", ipt.Object.name, querySignature)
		if ipt.databaseInstance != "" {
			objectName = fmt.Sprintf("%s-%s-%s", ipt.Object.name, ipt.databaseInstance, querySignature)
		}
		kvs = kvs.AddTag("name", objectName)
		kvs = kvs.AddTag("query_signature", querySignature)
		kvs = kvs.AddTag("server", ipt.Object.name)
		kvs = kvs.AddTag("database_type", "Oracle")
		if row.RawData.ConID > 0 {
			kvs = kvs.AddTag("con_id", fmt.Sprintf("%d", row.RawData.ConID))
		}
		if ipt.cdbName != "" {
			kvs = kvs.AddTag("cdb_name", ipt.cdbName)
		}
		if row.RawData.PDBName != "" {
			kvs = kvs.AddTag("pdb_name", row.RawData.PDBName)
		}
		if row.RawData.ForceMatchingSignature != "" {
			kvs = kvs.AddTag("force_matching_signature", row.RawData.ForceMatchingSignature)
		} else {
			kvs = kvs.AddTag("force_matching_signature", "0")
		}
		if row.RawData.SQLID != "" {
			kvs = kvs.AddTag("sql_id", row.RawData.SQLID)
		}

		// Fields - obfuscate SQL text
		sqlStatement := row.RawData.SQLText
		// If SQL text is truncated (length == 1000), get full text from v$sql
		if row.RawData.SQLTextLength == MaxSQLFullTextVSQLStats {
			err := ipt.getFullSQLText(&sqlStatement, "sql_id", row.RawData.SQLID)
			if err != nil {
				l.Warnf("failed to get full SQL text for sql_id %s: %v", row.RawData.SQLID, err)
			}
		}

		obfuscatedText := sqlStatement
		if sqlStatement != "" {
			obfResult, err := obfuscator.ObfuscateSQLString(sqlStatement)
			if err != nil {
				l.Warnf("failed to obfuscate SQL for query_signature %s: %v", querySignature, err)
			} else {
				obfuscatedText = obfResult.Query
			}
		}

		kvs = kvs.Set("message", obfuscatedText)

		pt := point.NewPoint(dbmQueryObjectName, kvs, opts...)
		pts = append(pts, pt)

		// Add to reported cache after successfully building the point
		if ipt.dbmQueryObjectCache != nil {
			ipt.dbmQueryObjectCache.Add(querySignature, struct{}{})
		}
	}

	if len(pts) > 0 {
		if err := ipt.feeder.Feed(point.Object, pts,
			dkio.WithCollectCost(time.Since(start)),
			dkio.WithElection(ipt.Election),
			dkio.WithSource(dbmFeedName),
		); err != nil {
			ipt.feeder.FeedLastError(err.Error(),
				metrics.WithLastErrorInput(inputName),
				metrics.WithLastErrorCategory(point.Object),
			)
			l.Errorf("feed dbm query object failed: %s", err.Error())
		}
	}
}

// getFullSQLText retrieves the full SQL text from v$sql when SQL text is truncated in v$sqlstats.
func (ipt *Input) getFullSQLText(sqlStatement *string, key string, value string) error {
	// Initialize go-ora connection if not already initialized (lazy loading)
	if ipt.goOraConnection == nil {
		conn, err := ipt.connectGoOra()
		if err != nil {
			return fmt.Errorf("failed to create go-ora connection: %w", err)
		}
		ipt.goOraConnection = conn
	}

	// Use PL/SQL block with go_ora.Clob to handle CLOB type
	var sqlFullText go_ora.Clob
	sql := fmt.Sprintf("BEGIN SELECT /* DK */ sql_fulltext INTO :sql_fulltext FROM v$sql WHERE %s = :v AND rownum = 1; END;", key)
	queryStart := time.Now()
	_, err := ipt.goOraConnection.Exec(sql, go_ora.Out{Dest: &sqlFullText, Size: 8000}, value)
	dbmSQLQueryDuration.WithLabelValues("query", "full_sql_text").Observe(time.Since(queryStart).Seconds())
	if err != nil {
		// Close connection on error so it will be recreated on next call
		if isConnectionError(err) {
			l.Debugf("Connection error detected, closing go-ora connection")
			ipt.closeGoOraConnection()
		}
		return fmt.Errorf("failed to query sql full text for %s = %s: %w", key, value, err)
	}

	if sqlFullText.String == "" {
		l.Warnf("The SQL text for the statement %s = %s couldn't be fetched because the SQL was evicted from shared pool", key, value)
		return nil
	}

	*sqlStatement = sqlFullText.String

	return nil
}
