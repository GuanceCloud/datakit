// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package oracle

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/cespare/xxhash/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/datakit"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/metrics"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"

	"github.com/DataDog/datadog-agent/pkg/obfuscate"
)

const (
	dbmPlanObjectName = "db_exec_plan"
)

type dbmPlanObjectMeasurement struct{}

func (*dbmPlanObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: dbmPlanObjectName,
		Cat:  point.Object,
		//nolint:lll
		Desc:   "Oracle DBM plan objects. Each object represents a unique execution plan identified by query_signature:plan_hash_value, containing the obfuscated plan content.",
		DescZh: "Oracle DBM 执行计划对象。每个对象代表一个由 query_signature:plan_hash_value 唯一标识的执行计划，包含脱敏后的计划内容。",
		Tags: map[string]interface{}{
			"name":                     inputs.NewTagInfo("Hash signature generated from query_signature:plan_hash_value"),
			"query_signature":          inputs.NewTagInfo("Hash signature generated from pdb_name:query_hash to link metrics and objects"),
			"plan_hash_value":          inputs.NewTagInfo("The hash value of the query execution plan"),
			"server":                   inputs.NewTagInfo("The server address (host:port)"),
			"database_instance":        inputs.NewTagInfo("Oracle instance identifier, derived from v$instance.host_name"),
			"database_type":            inputs.NewTagInfo("The type of the database. The value is `Oracle`"),
			"plan_type":                inputs.NewTagInfo("The format of the plan content. The value is `JSON`"),
			"con_id":                   inputs.NewTagInfo("The container ID (con_id) in Oracle multi tenant architecture"),
			"pdb_name":                 inputs.NewTagInfo("The name of the PDB (Pluggable Database)"),
			"cdb_name":                 inputs.NewTagInfo("The name of the CDB (Container Database)"),
			"force_matching_signature": inputs.NewTagInfo("The force matching signature of the query"),
			"sql_id":                   inputs.NewTagInfo("The SQL ID of the query "),
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The obfuscated/normalized execution plan content (full content)",
			},
			"timestamp": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The timestamp when the execution plan was created",
			},
			"optimizer_mode": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "The optimizer mode used for the execution plan",
			},
			"other": &inputs.FieldInfo{
				DataType: inputs.String,
				Unit:     inputs.UnknownUnit,
				Desc:     "Other information about the execution plan",
			},
		},
	}
}

// statementRowWithPlan stores statement row with plan information.
type statementRowWithPlan struct {
	*OracleRow
	planObfuscated string
	timestamp      string
	optimizerMode  string
	other          string
}

// collectDbmPlans collects execution plans from statements results and feeds database_plan DBO objects.
func (ipt *Input) collectDbmPlans(ctx context.Context, oracleRows []*OracleRow, ptsTime time.Time) {
	if len(oracleRows) == 0 {
		return
	}

	// Initialize plan object cache if not already initialized
	if ipt.dbmPlanObjectCache == nil {
		planTTL := dbmPlanCacheTTL.Duration // default TTL: 1 hour
		if ipt.Dbm.Metric.PlanCacheTTL.Duration > 0 {
			planTTL = ipt.Dbm.Metric.PlanCacheTTL.Duration
		}
		ipt.dbmPlanObjectCache = expirable.NewLRU[string, struct{}](
			10000,
			nil,
			planTTL,
		)
	}

	start := time.Now()
	// Collect plans for statement rows
	rowsWithPlans := ipt.collectPlansForStatements(ctx, oracleRows, start)
	if len(rowsWithPlans) == 0 {
		return
	}
	l.Debugf("collected %d plans, time taken: %s", len(rowsWithPlans), time.Since(start))

	// Build and feed database_plan DBO objects
	pts := ipt.buildAndFeedDatabasePlanObjects(rowsWithPlans, ptsTime)
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
			l.Errorf("feed database_plan DBO failed: %s", err.Error())
		}
	}
}

type PlanRows struct {
	PlanGlobalRow
	PlanStepRows
}

type PlanGlobalRow struct {
	SQLID         string         `db:"SQL_ID"`
	ChildNumber   sql.NullInt64  `db:"CHILD_NUMBER"`
	PlanCreated   sql.NullString `db:"TIMESTAMP"`
	OptimizerMode sql.NullString `db:"OPTIMIZER"`
	Other         sql.NullString `db:"OTHER"`
	Executions    sql.NullString `db:"EXECUTIONS"`
	PDBName       sql.NullString `db:"PDB_NAME"`
}

type PlanStepRows struct {
	Operation        sql.NullString  `db:"OPERATION"`
	Options          sql.NullString  `db:"OPTIONS"`
	ObjectOwner      sql.NullString  `db:"OBJECT_OWNER"`
	ObjectName       sql.NullString  `db:"OBJECT_NAME"`
	ObjectAlias      sql.NullString  `db:"OBJECT_ALIAS"`
	ObjectType       sql.NullString  `db:"OBJECT_TYPE"`
	PlanStepID       sql.NullInt64   `db:"ID"`
	ParentID         sql.NullInt64   `db:"PARENT_ID"`
	Depth            sql.NullInt64   `db:"DEPTH"`
	Position         *uint64         `db:"POSITION"`
	SearchColumns    sql.NullInt64   `db:"SEARCH_COLUMNS"`
	Cost             sql.NullFloat64 `db:"COST"`
	Cardinality      sql.NullFloat64 `db:"CARDINALITY"`
	Bytes            sql.NullFloat64 `db:"BYTES"`
	PartitionStart   sql.NullString  `db:"PARTITION_START"`
	PartitionStop    sql.NullString  `db:"PARTITION_STOP"`
	CPUCost          sql.NullFloat64 `db:"CPU_COST"`
	IOCost           sql.NullFloat64 `db:"IO_COST"`
	TempSpace        sql.NullFloat64 `db:"TEMP_SPACE"`
	AccessPredicates sql.NullString  `db:"ACCESS_PREDICATES"`
	FilterPredicates sql.NullString  `db:"FILTER_PREDICATES"`
	Projection       sql.NullString  `db:"PROJECTION"`
	LastStarts       *uint64         `db:"LAST_STARTS"`
	LastOutputRows   *uint64         `db:"LAST_OUTPUT_ROWS"`
	LastCRBufferGets *uint64         `db:"LAST_CR_BUFFER_GETS"`
	LastDiskReads    *uint64         `db:"LAST_DISK_READS"`
	LastDiskWrites   *uint64         `db:"LAST_DISK_WRITES"`
	LastElapsedTime  *uint64         `db:"LAST_ELAPSED_TIME"`
	LastMemoryUsed   *uint64         `db:"LAST_MEMORY_USED"`
	LastDegree       *uint64         `db:"LAST_DEGREE"`
	LastTempsegSize  *uint64         `db:"LAST_TEMPSEG_SIZE"`
}

type PlanDefinition struct {
	Operation   string `json:"operation,omitempty"`
	Options     string `json:"options,omitempty"`
	ObjectOwner string `json:"object_owner,omitempty"`
	ObjectName  string `json:"object_name,omitempty"`
	ObjectAlias string `json:"object_alias,omitempty"`
	ObjectType  string `json:"object_type,omitempty"`
	//nolint:revive // TODO(DBM) Fix revive linter
	PlanStepID       int64   `json:"id"`
	ParentID         int64   `json:"parent_id"`
	Depth            int64   `json:"depth"`
	Position         uint64  `json:"position"`
	SearchColumns    int64   `json:"search_columns,omitempty"`
	Cost             float64 `json:"cost"`
	Cardinality      float64 `json:"cardinality,omitempty"`
	Bytes            float64 `json:"bytes,omitempty"`
	PartitionStart   string  `json:"partition_start,omitempty"`
	PartitionStop    string  `json:"partition_stop,omitempty"`
	CPUCost          float64 `json:"cpu_cost,omitempty"`
	IOCost           float64 `json:"io_cost,omitempty"`
	TempSpace        float64 `json:"temp_space,omitempty"`
	AccessPredicates string  `json:"access_predicates,omitempty"`
	FilterPredicates string  `json:"filter_predicates,omitempty"`
	Projection       string  `json:"projection,omitempty"`
	LastStarts       uint64  `json:"actual_starts,omitempty"`
	LastOutputRows   uint64  `json:"actual_rows,omitempty"`
	LastCRBufferGets uint64  `json:"actual_cr_buffer_gets,omitempty"`
	LastDiskReads    uint64  `json:"actual_disk_reads,omitempty"`
	LastDiskWrites   uint64  `json:"actual_disk_writes,omitempty"`
	LastElapsedTime  uint64  `json:"actual_elapsed_time,omitempty"`
	LastMemoryUsed   uint64  `json:"actual_memory_used,omitempty"`
	LastDegree       uint64  `json:"actual_parallel_degree,omitempty"`
	LastTempsegSize  uint64  `json:"actual_tempseg_size,omitempty"`
}

// collectPlansForStatements collects execution plans for statement rows.
func (ipt *Input) collectPlansForStatements(ctx context.Context, oracleRows []*OracleRow, startTime time.Time) []*statementRowWithPlan {
	var resultRows []*statementRowWithPlan

	// Get max runtime configuration
	maxRunTime := ipt.Dbm.Metric.MaxRunTime
	if maxRunTime <= 0 {
		maxRunTime = 30 // Default to 30 seconds if not configured
	}
	maxRunTimeDuration := time.Duration(maxRunTime) * time.Second

	obfuscator := obfuscate.NewObfuscator(obfuscate.Config{})
	for i, row := range oracleRows {
		select {
		case <-ctx.Done():
			l.Info("context done, collect plans for statements exit")
			return resultRows

		case <-datakit.Exit.Wait():
			l.Info("datakit exit, collect plans for statements exit")
			return nil

		case <-ipt.semStop.Wait():
			l.Info("oracle return, collect plans for statements exit")
			return nil

		default:
		}

		// Check if plan collection time exceeded max runtime (every 5 iterations)
		if i > 0 && i%5 == 0 {
			elapsed := time.Since(startTime)
			if elapsed > maxRunTimeDuration {
				l.Warnf("plan collection time (%s) exceeded max_run_time (%ds), stopping plan collection", elapsed, maxRunTime)
				return resultRows
			}
		}

		// Check if SQL_ID is available
		// Note: When using force_matching_signature, SQLID might be empty or invalid
		// We need SQLID to query v$sql_plan_statistics_all, so skip if it's empty
		if row.RawData.SQLID == "" || row.RawData.SQLID == "0" {
			l.Debugf("sql_id is empty or invalid, plan_hash_value: %d, force_matching_signature: %s",
				row.RawData.PlanHashValue, row.RawData.ForceMatchingSignature)
			continue
		}

		// Check if plan_hash_value is valid
		// plan_hash_value = 0 indicates no valid execution plan exists
		if row.RawData.PlanHashValue == 0 {
			l.Debugf("plan_hash_value is 0 (invalid),sql_id: %s, force_matching_signature: %s",
				row.RawData.SQLID, row.RawData.ForceMatchingSignature)
			continue
		}

		// Check cache before querying to avoid unnecessary queries
		planKey := generatePlanCacheKey(row.querySignature, fmt.Sprintf("%d", row.RawData.PlanHashValue))
		if ipt.dbmPlanObjectCache != nil {
			if _, ok := ipt.dbmPlanObjectCache.Get(planKey); ok {
				continue
			}
		}

		var planStepsPayload []PlanDefinition
		var planStepsDB []PlanRows
		var planTimestamp, planOptimizerMode, planOther string

		var planQuery string
		var err error
		if isDBVersionGreaterOrEqualThan(ipt.dbVersion, "12") {
			planQuery = planQuery12
			err = selectWrapperWithBinds(ipt, ctx, &planStepsDB, planQuery, row.RawData.SQLID, row.RawData.PlanHashValue, row.RawData.ConID)
		} else {
			planQuery = planQuery11
			err = selectWrapperWithBinds(ipt, ctx, &planStepsDB, planQuery, row.RawData.SQLID, row.RawData.PlanHashValue)
		}
		if err != nil {
			l.Warnf("failed to query execution plan: %v", err)
			continue
		}

		if len(planStepsDB) > 0 {
			var firstChildNumber int64
			for i, stepRow := range planStepsDB {
				if !stepRow.ChildNumber.Valid {
					l.Errorf("invalid child number in execution plan for sql_id: %s", row.RawData.SQLID)
					break
				}
				if i == 0 {
					firstChildNumber = stepRow.ChildNumber.Int64
				} else if firstChildNumber != stepRow.ChildNumber.Int64 {
					break
				}
				var stepPayload PlanDefinition
				if stepRow.Operation.Valid {
					stepPayload.Operation = stepRow.Operation.String
				}
				if stepRow.Options.Valid {
					stepPayload.Options = stepRow.Options.String
				}
				if stepRow.ObjectOwner.Valid {
					stepPayload.ObjectOwner = stepRow.ObjectOwner.String
				}
				if stepRow.ObjectName.Valid {
					stepPayload.ObjectName = stepRow.ObjectName.String
				}
				if stepRow.ObjectAlias.Valid {
					stepPayload.ObjectAlias = stepRow.ObjectAlias.String
				}
				if stepRow.ObjectType.Valid {
					stepPayload.ObjectType = stepRow.ObjectType.String
				}
				if stepRow.PlanStepID.Valid {
					stepPayload.PlanStepID = stepRow.PlanStepID.Int64
				}
				if stepRow.ParentID.Valid {
					stepPayload.ParentID = stepRow.ParentID.Int64
				}
				if stepRow.Depth.Valid {
					stepPayload.Depth = stepRow.Depth.Int64
				}
				if stepRow.Position != nil {
					stepPayload.Position = *stepRow.Position
				}
				if stepRow.SearchColumns.Valid {
					stepPayload.SearchColumns = stepRow.SearchColumns.Int64
				}
				if stepRow.Cost.Valid {
					stepPayload.Cost = stepRow.Cost.Float64
				}
				if stepRow.Cardinality.Valid {
					stepPayload.Cardinality = stepRow.Cardinality.Float64
				}
				if stepRow.Bytes.Valid {
					stepPayload.Bytes = stepRow.Bytes.Float64
				}
				if stepRow.PartitionStart.Valid {
					stepPayload.PartitionStart = stepRow.PartitionStart.String
				}
				if stepRow.PartitionStop.Valid {
					stepPayload.PartitionStop = stepRow.PartitionStop.String
				}
				if stepRow.CPUCost.Valid {
					stepPayload.CPUCost = stepRow.CPUCost.Float64
				}
				if stepRow.IOCost.Valid {
					stepPayload.IOCost = stepRow.IOCost.Float64
				}
				if stepRow.TempSpace.Valid {
					stepPayload.TempSpace = stepRow.TempSpace.Float64
				}
				handlePredicate("access", stepRow.AccessPredicates, &stepPayload.AccessPredicates, row, obfuscator)
				handlePredicate("filter", stepRow.FilterPredicates, &stepPayload.FilterPredicates, row, obfuscator)
				if stepRow.Projection.Valid {
					stepPayload.Projection = stepRow.Projection.String
				}
				if stepRow.LastStarts != nil {
					stepPayload.LastStarts = *stepRow.LastStarts
				}
				if stepRow.LastOutputRows != nil {
					stepPayload.LastOutputRows = *stepRow.LastOutputRows
				}
				if stepRow.LastCRBufferGets != nil {
					stepPayload.LastCRBufferGets = *stepRow.LastCRBufferGets
				}
				if stepRow.LastDiskReads != nil {
					stepPayload.LastDiskReads = *stepRow.LastDiskReads
				}
				if stepRow.LastDiskWrites != nil {
					stepPayload.LastDiskWrites = *stepRow.LastDiskWrites
				}
				if stepRow.LastElapsedTime != nil {
					stepPayload.LastElapsedTime = *stepRow.LastElapsedTime
				}
				if stepRow.LastMemoryUsed != nil {
					stepPayload.LastMemoryUsed = *stepRow.LastMemoryUsed
				}
				if stepRow.LastDegree != nil {
					stepPayload.LastDegree = *stepRow.LastDegree
				}
				if stepRow.LastTempsegSize != nil {
					stepPayload.LastTempsegSize = *stepRow.LastTempsegSize
				}
				// Store plan metadata (only from first row)
				if i == 0 {
					if stepRow.PlanCreated.Valid && stepRow.PlanCreated.String != "" {
						planTimestamp = stepRow.PlanCreated.String
					}
					if stepRow.OptimizerMode.Valid && stepRow.OptimizerMode.String != "" {
						planOptimizerMode = stepRow.OptimizerMode.String
					}
					if stepRow.Other.Valid && stepRow.Other.String != "" {
						planOther = stepRow.Other.String
					}
				}
				planStepsPayload = append(planStepsPayload, stepPayload)
			}

			planJSON, err := json.Marshal(planStepsPayload)
			if err != nil {
				l.Warnf("failed to marshal plan to JSON for sql_id %s: %v", row.RawData.SQLID, err)
				continue
			}

			rowWithPlan := &statementRowWithPlan{
				OracleRow:      row,
				planObfuscated: string(planJSON),
				timestamp:      planTimestamp,
				optimizerMode:  planOptimizerMode,
				other:          planOther,
			}

			resultRows = append(resultRows, rowWithPlan)
		}
	}

	return resultRows
}

// handlePredicate handles access or filter predicates by obfuscating them.
func handlePredicate(predicateType string, predicate sql.NullString, target *string, row *OracleRow, obfuscator *obfuscate.Obfuscator) {
	if !predicate.Valid || predicate.String == "" {
		return
	}

	// Obfuscate the predicate using the obfuscator
	obfResult, err := obfuscator.ObfuscateSQLString(predicate.String)
	if err != nil {
		l.Warnf("failed to obfuscate %s predicate for sql_id %s: %v", predicateType, row.RawData.SQLID, err)
		// If obfuscation fails, use the original predicate
		*target = predicate.String
		return
	}

	*target = obfResult.Query
}

// buildAndFeedDatabasePlanObjects builds and feeds database_plan DBO objects.
func (ipt *Input) buildAndFeedDatabasePlanObjects(rowsWithPlans []*statementRowWithPlan, ptsTime time.Time) []*point.Point {
	if len(rowsWithPlans) == 0 {
		return nil
	}

	opts := append(point.DefaultObjectOptions(), point.WithTime(ptsTime))
	var pts []*point.Point

	for _, row := range rowsWithPlans {
		// Tags - name is query_signature:plan_hash_value
		planName := generatePlanCacheKey(row.querySignature, fmt.Sprintf("%d", row.RawData.PlanHashValue))

		kvs := ipt.getKVs()

		kvs = kvs.AddTag("name", planName)
		kvs = kvs.AddTag("query_signature", row.querySignature)
		kvs = kvs.AddTag("plan_hash_value", fmt.Sprintf("%d", row.RawData.PlanHashValue))
		kvs = kvs.AddTag("server", ipt.Object.name)
		kvs = kvs.AddTag("database_type", "Oracle")
		kvs = kvs.AddTag("plan_type", "JSON")
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
		kvs = kvs.AddTag("sql_id", row.RawData.SQLID)

		// Fields
		// message contains the obfuscated plan content
		kvs = kvs.Set("message", row.planObfuscated)
		kvs = kvs.Set("timestamp", row.timestamp)
		kvs = kvs.Set("optimizer_mode", row.optimizerMode)
		kvs = kvs.Set("other", row.other)

		pt := point.NewPoint(dbmPlanObjectName, kvs, opts...)
		pts = append(pts, pt)

		// Add to reported cache after successfully building the point
		if ipt.dbmPlanObjectCache != nil {
			ipt.dbmPlanObjectCache.Add(planName, struct{}{})
		}
	}
	return pts
}

// generatePlanCacheKey generates a unique signature for an execution plan.
func generatePlanCacheKey(querySignature, planHashValue string) string {
	h := xxhash.New()
	_, _ = h.WriteString(querySignature)
	_, _ = h.WriteString(":")
	_, _ = h.WriteString(planHashValue)

	return fmt.Sprintf("%016x", h.Sum64())
}
