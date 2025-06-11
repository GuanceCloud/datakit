// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package mysql

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	tableChunkSize             = 500
	maxExecutionTime           = 60 * time.Second
	mysqlObjectMeasurementName = "database"
	mysqlType                  = "MySQL"
)

type mysqlObjectMeasurement struct{}

//nolint:lll
func (*mysqlObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: mysqlObjectMeasurementName,
		Cat:  point.Object,
		Desc: "MySQL object metrics([:octicons-tag-24: Version-1.74.0](../datakit/changelog-2025.md#cl-1.74.0))",
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "The hostname of the MySQL server"},
			"server":        &inputs.TagInfo{Desc: "The server address of the MySQL server"},
			"version":       &inputs.TagInfo{Desc: "The version of the MySQL server"},
			"name":          &inputs.TagInfo{Desc: "The name of the database. The value is `host:port` in default"},
			"database_type": &inputs.TagInfo{Desc: "The type of the database. The value is `MySQL`"},
			"port":          &inputs.TagInfo{Desc: "The port of the MySQL server"},
		},
		Fields: map[string]interface{}{
			"message":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of database information"},
			"uptime":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "The number of seconds that the server has been up"},
			"slow_queries":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of queries that have taken more than long_query_time seconds. This counter increments regardless of whether the slow query log is enabled."},
			"avg_query_time": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.TimestampUS, Desc: "The average time taken by a query to execute"},
			"qps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of queries executed by the database per second"},
			"tps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of transactions executed by the database per second"},
			"slow_query_log": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Whether the slow query log is enabled. The value can be 0 (or OFF) to disable the log or 1 (or ON) to enable the log."},
		},
	}
}

type objectMertric struct {
	Queries     int64
	SlowQueries int64
	Trans       int64
	QPS         float64
	TPS         float64
	Time        time.Time
}

// nolint:execinquery
func (ipt *Input) collectMysqlBasicInfo() error {
	const sqlSelectUptime = "SHOW GLOBAL STATUS LIKE 'Uptime';"
	rows, err := ipt.db.Query(sqlSelectUptime)
	if err != nil {
		l.Error("collectMysqlBasicInfo fail:", err.Error())
		return fmt.Errorf("query failed: %w", err)
	}
	defer closeRows(rows)
	var variableName string
	var uptime int
	if rows.Next() {
		if err := rows.Scan(&variableName, &uptime); err != nil {
			l.Error("collectMysqlBasicInfo fail:", err.Error())
			return fmt.Errorf("scan error: %w", err)
		}
		if variableName == "Uptime" {
			ipt.Uptime = uptime
		}
	}
	if err := rows.Err(); err != nil {
		l.Error("collectMysqlBasicInfo fail:", err.Error())
		return fmt.Errorf("error iterating rows: %w", err)
	}
	return nil
}

type mysqlObjectMessage struct {
	Setting   map[string]string `json:"setting"`
	Databases []*mysqlDatabase  `json:"databases"`
}

type mysqlDatabase struct {
	Name                    string        `json:"name"`
	DefaultCharacterSetName string        `json:"default_character_set_name"`
	DefaultCollationName    string        `json:"default_collation_name"`
	Tables                  []*mysqlTable `json:"tables"`
}

type mysqlTable struct {
	Name        string            `json:"name"`
	Engine      string            `json:"engine"`
	RowFormat   string            `json:"row_format"`
	CreateTime  string            `json:"create_time"`
	Columns     []mysqlColumn     `json:"columns"`
	Indexes     []mysqlIndex      `json:"indexes"`
	ForeignKeys []mysqlForeignKey `json:"foreign_keys"`
	Partitions  []mysqlPartition  `json:"partitions"`
}

type mysqlColumn struct {
	Name            string `json:"name"`
	DataType        string `json:"data_type"`
	Default         string `json:"default"`
	Nullable        bool   `json:"nullable"`
	OrdinalPosition int64  `json:"ordinal_position"`
}

type mysqlIndex struct {
	Name        string             `json:"name"`
	Cardinality int64              `json:"cardinality"`
	IndexType   string             `json:"index_type"`
	Columns     []mysqlIndexColumn `json:"columns"`
	NonUnique   bool               `json:"non_unique"`
	Expression  string             `json:"expression"`
}

type mysqlIndexColumn struct {
	Name      string `json:"name"`
	SubPart   int64  `json:"sub_part"`
	Collation string `json:"collation"`
	Packed    string `json:"packed"`
	Nullable  bool   `json:"nullable"`
}

type mysqlForeignKey struct {
	ConstraintSchema      string `json:"constraint_schema"`
	Name                  string `json:"name"`
	ColumnNames           string `json:"column_names"`
	ReferencedTableSchema string `json:"referenced_table_schema"`
	ReferencedTableName   string `json:"referenced_table_name"`
	ReferencedColumnNames string `json:"referenced_column_names"`
	UpdateAction          string `json:"update_action"`
	DeleteAction          string `json:"delete_action"`
}

type mysqlPartition struct {
	Name                     string                     `json:"name"`
	Subpartitions            []mysqlPartionSubpartition `json:"subpartitions"`
	PartitionOrdinalPosition int64                      `json:"partition_ordinal_position"`
	PartitionMethod          string                     `json:"partition_method"`
	PartitionExpression      string                     `json:"partition_expression"`
	PartitionDescription     string                     `json:"partition_description"`
	TableRows                int64                      `json:"table_rows"`
	DataLength               int64                      `json:"data_length"`
}

type mysqlPartionSubpartition struct {
	Name                      string `json:"name"`
	SubpartionOrdinalPosition int64  `json:"subpartition_ordinal_position"`
	SubpartionMethod          string `json:"subpartition_method"`
	SubpartionExpression      string `json:"subpartition_expression"`
	TableRows                 int64  `json:"table_rows"`
	DataLength                int64  `json:"data_length"`
}

func (m *mysqlObjectMessage) String() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

func (ipt *Input) metricCollectMysqlObject() ([]*point.Point, error) {
	if !ipt.Object.lastCollectionTime.IsZero() && ipt.Object.lastCollectionTime.Add(ipt.Object.Interval.Duration).After(time.Now()) {
		l.Debugf("skip mysql_object collection, time interval not reached")
		return nil, nil
	}

	ipt.Object.lastCollectionTime = time.Now()

	opts := ipt.getKVsOpts(point.Object)
	kvs := ipt.getKVs()

	message := mysqlObjectMessage{}
	setting, err := ipt.getMysqlSetting()
	if err != nil {
		l.Warnf("getMysqlMetadata failed: %s", err.Error())
	} else {
		message.Setting = setting
	}

	databases, err := ipt.getMysqlDatabases()
	if err != nil {
		l.Warnf("getMysqlDatabases failed: %s", err.Error())
	} else {
		message.Databases = databases
	}

	kvs = kvs.AddTag("version", ipt.getVersion()).
		AddTag("type", mysqlType). // deprecated
		AddTag("database_type", mysqlType).
		AddTag("name", ipt.Object.name).
		AddTag("host", ipt.Host).
		AddTag("server", ipt.mergedTags["server"]).
		AddTag("port", fmt.Sprintf("%d", ipt.Port)).
		Add("uptime", ipt.Uptime, false, true).
		Add("message", message.String(), false, true)

	// slow query log
	if v, ok := setting["slow_query_log"]; ok {
		kvs = kvs.Add("slow_query_log", v, false, true)
	}

	// qps, tps, slow_queries, avg_query_time
	if ipt.objectMetric != nil {
		qps := fmt.Sprintf("%.2f", ipt.objectMetric.QPS)
		kvs = kvs.Add("qps", qps, false, true)

		tps := fmt.Sprintf("%.2f", ipt.objectMetric.TPS)
		kvs = kvs.Add("tps", tps, false, true)

		kvs = kvs.Add("slow_queries", ipt.objectMetric.SlowQueries, false, true)
	}

	if v, err := ipt.getAverageQueryExecutionTime(); err == nil {
		kvs = kvs.Add("avg_query_time", v, false, true)
	} else {
		l.Warnf("getAverageQueryExecutionTime failed: %s", err.Error())
	}

	return []*point.Point{point.NewPointV2("database", kvs, opts...)}, nil
}

const sqlGetAverageQueryExecutionTime = `
SELECT ROUND((SUM(sum_timer_wait) / SUM(count_star)) / 1000000) AS avg_us
	FROM performance_schema.events_statements_summary_by_digest
	WHERE schema_name IS NOT NULL;
`

func (ipt *Input) getAverageQueryExecutionTime() (float64, error) {
	rows := ipt.q(sqlGetAverageQueryExecutionTime, getMetricName("mysql_object", "mysql_average_query_execution_time"))
	if rows == nil {
		return 0, errors.New("query average query execution time failed: nil rows")
	}
	if rows.Err() != nil {
		return 0, fmt.Errorf("get average query execution time rows error: %w", rows.Err())
	}
	defer closeRows(rows)
	var avgUs sql.NullFloat64
	if rows.Next() {
		if err := rows.Scan(&avgUs); err != nil {
			return 0, fmt.Errorf("scan error: %w", err)
		}
		if avgUs.Valid {
			return avgUs.Float64, nil
		}
	}

	return 0, nil
}

func (ipt *Input) getMysqlSetting() (map[string]string, error) {
	tableName := "information_schema.GLOBAL_VARIABLES"
	if ipt.Version.flavor != strMariaDB && ipt.Version.versionCompatible([]int{5, 7, 0}) {
		tableName = "performance_schema.global_variables"
	}

	query := fmt.Sprintf("SELECT variable_name, variable_value FROM %s", tableName)

	rows := ipt.q(query, getMetricName("mysql_object", "global_variables"))

	if rows == nil {
		return nil, errors.New("query global variables failed: nil rows")
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("get global variables rows error: %w", rows.Err())
	}

	defer closeRows(rows)

	settings := make(map[string]string)
	for rows.Next() {
		var variableName, variableValue string
		if err := rows.Scan(&variableName, &variableValue); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		settings[variableName] = variableValue
	}

	return settings, nil
}

const sqlGetDatabases = `
	SELECT schema_name as name,
       default_character_set_name as default_character_set_name,
       default_collation_name as default_collation_name
       FROM information_schema.SCHEMATA
       WHERE schema_name not in ('sys', 'mysql', 'performance_schema', 'information_schema')
			 `

func (ipt *Input) getMysqlDatabases() ([]*mysqlDatabase, error) {
	rows := ipt.q(sqlGetDatabases, getMetricName("mysql_object", "mysql_databases"))

	if rows == nil {
		return nil, errors.New("query databases failed: nil rows")
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("get databases rows error: %w", rows.Err())
	}

	defer closeRows(rows)

	databases := []*mysqlDatabase{}
	for rows.Next() {
		var name, characterSetName, defaultCollationName string
		if err := rows.Scan(&name, &characterSetName, &defaultCollationName); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}

		databases = append(databases, &mysqlDatabase{
			Name:                    name,
			DefaultCharacterSetName: characterSetName,
			DefaultCollationName:    defaultCollationName,
		})
	}

	for _, database := range databases {
		tables, err := ipt.getDatabaseTables(database.Name)
		if err != nil {
			l.Errorf("get database tables failed: %s", err.Error())
			continue
		}

		database.Tables = tables
	}

	return databases, nil
}

const sqlGetDatabaseTables = `
 SELECT table_name as name,
       engine as engine,
       row_format as row_format,
       create_time as create_time
       FROM information_schema.TABLES
       WHERE TABLE_SCHEMA = '%s' AND TABLE_TYPE="BASE TABLE"
`

func (ipt *Input) getDatabaseTables(databaseName string) ([]*mysqlTable, error) {
	query := fmt.Sprintf(sqlGetDatabaseTables, databaseName)
	rows := ipt.q(query, getMetricName("mysql_object", "mysql_databases"))
	if rows == nil {
		return nil, errors.New("query tables failed: nil rows")
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("get tables rows error: %w", rows.Err())
	}
	defer closeRows(rows)
	tables := []*mysqlTable{}
	startIndex := 0
	endIndex := 0
	start := time.Now()
	for rows.Next() {
		cost := time.Since(start)
		if cost > maxExecutionTime || cost > ipt.Interval.Duration {
			l.Warnf("getDatabaseTables cost too long: %s, stop to collect more databases ", cost)
			break
		}
		var name, engine, rowFormat, createTime string

		if err := rows.Scan(&name, &engine, &rowFormat, &createTime); err != nil {
			return nil, fmt.Errorf("scan error: %w", err)
		}
		tables = append(tables, &mysqlTable{
			Name:       name,
			Engine:     engine,
			RowFormat:  rowFormat,
			CreateTime: createTime,
		})

		endIndex++

		if endIndex-startIndex >= tableChunkSize {
			if err := ipt.populateTablesData(databaseName, tables[startIndex:endIndex]); err != nil {
				return nil, fmt.Errorf("get table columns error: %w", err)
			} else {
				startIndex = endIndex
			}
		}
	}

	if endIndex-startIndex > 0 {
		if err := ipt.populateTablesData(databaseName, tables[startIndex:endIndex]); err != nil {
			return nil, fmt.Errorf("get table columns error: %w", err)
		}
	}

	return tables, nil
}

func (ipt *Input) populateTablesData(databaseName string, tables []*mysqlTable) error {
	if len(tables) == 0 {
		return nil
	}

	tablesNames := []string{}
	tableNameToTableIndex := make(map[string]int)
	for index, table := range tables {
		tablesNames = append(tablesNames, table.Name)
		tableNameToTableIndex[table.Name] = index
	}

	tableNamesString := fmt.Sprintf("\"%s\"", strings.Join(tablesNames, "\",\""))
	if err := ipt.populateColumns(tableNameToTableIndex, tables, tableNamesString, databaseName); err != nil {
		return fmt.Errorf("get table columns error: %w", err)
	}

	if err := ipt.populatePartitions(tableNameToTableIndex, tables, tableNamesString, databaseName); err != nil {
		return fmt.Errorf("get table partitions error: %w", err)
	}

	if err := ipt.populateForeignKeys(tableNameToTableIndex, tables, tableNamesString, databaseName); err != nil {
		return fmt.Errorf("get foreign keys error: %w", err)
	}

	if err := ipt.populateIndex(tableNameToTableIndex, tables, tableNamesString, databaseName); err != nil {
		return fmt.Errorf("get index error: %w", err)
	}
	return nil
}

const sqlGetTableColumns = `
SELECT table_name as table_name,
       column_name as name,
       column_type as column_type,
       column_default as column_default,
       is_nullable as nullable,
       ordinal_position as ordinal_position,
       column_key as column_key,
       extra as extra
FROM INFORMATION_SCHEMA.COLUMNS
WHERE table_schema = "%s" AND table_name IN (%s)
`

func (ipt *Input) populateColumns(tableNameToTableIndex map[string]int, tables []*mysqlTable, tableNamesString string, databaseName string) error {
	query := fmt.Sprintf(sqlGetTableColumns, databaseName, tableNamesString)
	rows := ipt.q(query, getMetricName("mysql_object", "table_columns"))
	if rows == nil {
		return errors.New("query failed: nil rows")
	}

	if rows.Err() != nil {
		return fmt.Errorf("get rows error: %w", rows.Err())
	}
	defer closeRows(rows)
	for rows.Next() {
		var tableName, name, columnType, columnDefault, nullable, columnKey, extra sql.NullString
		var oridinalPosition sql.NullInt64
		isNullable := false

		if err := rows.Scan(&tableName, &name, &columnType, &columnDefault, &nullable, &oridinalPosition, &columnKey, &extra); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}

		if strings.ToLower(nullable.String) == "yes" {
			isNullable = true
		}

		tables[tableNameToTableIndex[tableName.String]].Columns = append(tables[tableNameToTableIndex[tableName.String]].Columns, mysqlColumn{
			Name:            name.String,
			DataType:        columnType.String,
			Default:         columnDefault.String,
			Nullable:        isNullable,
			OrdinalPosition: oridinalPosition.Int64,
		})
	}
	return nil
}

const sqlGetTablePartitions = `
SELECT
    table_name as table_name,
    partition_name as name,
    subpartition_name as subpartition_name,
    partition_ordinal_position as partition_ordinal_position,
    subpartition_ordinal_position as subpartition_ordinal_position,
    partition_method as partition_method,
    subpartition_method as subpartition_method,
    partition_expression as partition_expression,
    subpartition_expression as subpartition_expression,
    partition_description as partition_description,
    table_rows as table_rows,
    data_length as data_length
FROM INFORMATION_SCHEMA.PARTITIONS
WHERE
    table_schema = "%s" AND table_name in (%s) AND partition_name IS NOT NULL
`

func (ipt *Input) populatePartitions(tableNameToTableIndex map[string]int, tables []*mysqlTable, tableNamesString string, databaseName string) error {
	query := fmt.Sprintf(sqlGetTablePartitions, databaseName, tableNamesString)
	rows := ipt.q(query, getMetricName("mysql_object", "table_partitions"))
	if rows == nil {
		return errors.New("query failed: nil rows")
	}

	if rows.Err() != nil {
		return fmt.Errorf("get rows error: %w", rows.Err())
	}
	defer closeRows(rows)
	tablePartition := map[string]map[string]mysqlPartition{}
	for rows.Next() {
		var (
			tableName, name, subPartitionName,
			partitionMethod, subpartitionMethod, partitionExpression, subpartionExpression, partitionDescription sql.NullString
			partitionOrdinalPosition, tableRows, subpartitionOrdinalPosition, dataLength sql.NullInt64
		)

		if err := rows.Scan(&tableName, &name, &subPartitionName, &partitionOrdinalPosition,
			&subpartitionOrdinalPosition, &partitionMethod, &subpartitionMethod, &partitionExpression,
			&subpartionExpression, &partitionDescription, &tableRows, &dataLength); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}

		if _, ok := tablePartition[tableName.String]; !ok {
			tablePartition[tableName.String] = map[string]mysqlPartition{}
		}

		newPartition := mysqlPartition{
			Name:                     name.String,
			PartitionOrdinalPosition: partitionOrdinalPosition.Int64,
			PartitionMethod:          partitionMethod.String,
			PartitionExpression:      strings.ToLower(strings.TrimSpace(partitionExpression.String)),
			PartitionDescription:     partitionDescription.String,
		}
		var p mysqlPartition
		if v, ok := tablePartition[tableName.String][name.String]; !ok {
			p = newPartition
		} else {
			p = v
		}

		p.Name = newPartition.Name
		p.PartitionOrdinalPosition = newPartition.PartitionOrdinalPosition
		p.PartitionMethod = newPartition.PartitionMethod
		p.PartitionExpression = newPartition.PartitionExpression
		p.PartitionDescription = newPartition.PartitionDescription

		p.TableRows += tableRows.Int64
		p.DataLength += dataLength.Int64

		if subPartitionName.String != "" {
			p.Subpartitions = append(p.Subpartitions, mysqlPartionSubpartition{
				Name:                      subPartitionName.String,
				SubpartionOrdinalPosition: subpartitionOrdinalPosition.Int64,
				SubpartionMethod:          subpartitionMethod.String,
				SubpartionExpression:      strings.ToLower(strings.TrimSpace(subpartionExpression.String)),
				TableRows:                 tableRows.Int64,
				DataLength:                dataLength.Int64,
			})
		}
		tablePartition[tableName.String][name.String] = p
	}
	for tableName, partitions := range tablePartition {
		for _, partition := range partitions {
			tables[tableNameToTableIndex[tableName]].Partitions = append(tables[tableNameToTableIndex[tableName]].Partitions, partition)
		}
	}

	return nil
}

const sqlGetTableForeignKeys = `
SELECT
    kcu.constraint_schema as constraint_schema,
    kcu.constraint_name as name,
    kcu.table_name as table_name,
    group_concat(kcu.column_name order by kcu.ordinal_position asc) as column_names,
    kcu.referenced_table_schema as referenced_table_schema,
    kcu.referenced_table_name as referenced_table_name,
    group_concat(kcu.referenced_column_name) as referenced_column_names,
    rc.update_rule as update_action,
    rc.delete_rule as delete_action
FROM
    INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
LEFT JOIN
    INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
    ON kcu.CONSTRAINT_SCHEMA = rc.CONSTRAINT_SCHEMA
    AND kcu.CONSTRAINT_NAME = rc.CONSTRAINT_NAME
WHERE
    kcu.table_schema = "%s" AND kcu.table_name in (%s)
    AND kcu.referenced_table_name is not null
GROUP BY
    kcu.constraint_schema,
    kcu.constraint_name,
    kcu.table_name,
    kcu.referenced_table_schema,
    kcu.referenced_table_name,
    rc.update_rule,
    rc.delete_rule
`

func (ipt *Input) populateForeignKeys(tableNameToTableIndex map[string]int,
	tables []*mysqlTable, tableNamesString string, databaseName string,
) error {
	query := fmt.Sprintf(sqlGetTableForeignKeys, databaseName, tableNamesString)
	rows := ipt.q(query, getMetricName("mysql_object", "table_foreign_keys"))
	if rows == nil {
		return errors.New("query failed: nil rows")
	}

	if rows.Err() != nil {
		return fmt.Errorf("get columns error: %w", rows.Err())
	}
	defer closeRows(rows)
	for rows.Next() {
		var constraintSchema, name, tableName, columnNames, referencedTableSchema,
			referencedTableName, referencedColumnNames, updateAction, deleteAction string
		if err := rows.Scan(&constraintSchema, &name, &tableName,
			&columnNames, &referencedTableSchema, &referencedTableName,
			&referencedColumnNames, &updateAction, &deleteAction); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}

		tables[tableNameToTableIndex[tableName]].ForeignKeys = append(tables[tableNameToTableIndex[tableName]].ForeignKeys, mysqlForeignKey{
			ConstraintSchema:      constraintSchema,
			Name:                  name,
			ColumnNames:           columnNames,
			ReferencedTableSchema: referencedTableSchema,
			ReferencedTableName:   referencedTableName,
			ReferencedColumnNames: referencedColumnNames,
			UpdateAction:          updateAction,
			DeleteAction:          deleteAction,
		})
	}
	return nil
}

const sqlGetIndexes = `
SELECT
    table_name as table_name,
    index_name as name,
    collation as collation,
    cardinality as cardinality,
    index_type as index_type,
    seq_in_index as seq_in_index,
    column_name as column_name,
    sub_part as sub_part,
    packed as packed,
    nullable as nullable,
    non_unique as non_unique,
    NULL as expression
FROM INFORMATION_SCHEMA.STATISTICS
WHERE table_schema = "%s" AND table_name IN (%s);
`

const sqlGetIndexes8 = `
SELECT
    table_name as table_name,
    index_name as name,
    collation as collation,
    cardinality as cardinality,
    index_type as index_type,
    seq_in_index as seq_in_index,
    column_name as column_name,
    sub_part as sub_part,
    packed as packed,
    nullable as nullable,
    non_unique as non_unique,
    expression as expression
FROM INFORMATION_SCHEMA.STATISTICS
WHERE table_schema = "%s" AND table_name IN (%s);
`

func (ipt *Input) populateIndex(tableNameToTableIndex map[string]int, tables []*mysqlTable, tableNamesString string, databaseName string) error {
	query, err := ipt.getIndexQuery()
	if err != nil {
		return fmt.Errorf("get index query error: %w", err)
	}

	query = fmt.Sprintf(query, databaseName, tableNamesString)

	rows := ipt.q(query, getMetricName("mysql_object", "table_indexes"))
	if rows == nil {
		return errors.New("query failed: nil rows")
	}

	if rows.Err() != nil {
		return fmt.Errorf("get columns error: %w", rows.Err())
	}
	defer closeRows(rows)
	tableIndex := map[string]map[string]mysqlIndex{}
	for rows.Next() {
		var tableName, name, collation, indexType, seqInIndex, columnName, packed, nullable, nonUnique, expression sql.NullString
		var subPart, cardinality sql.NullInt64
		if err := rows.Scan(&tableName, &name, &collation,
			&cardinality, &indexType, &seqInIndex, &columnName, &subPart, &packed,
			&nullable, &nonUnique, &expression); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}

		if _, ok := tableIndex[tableName.String]; !ok {
			tableIndex[tableName.String] = map[string]mysqlIndex{}
		}

		newIndex := mysqlIndex{
			Name:        name.String,
			Cardinality: cardinality.Int64,
			IndexType:   indexType.String,
			NonUnique:   nonUnique.String != "0",
			Expression:  expression.String,
		}

		var index mysqlIndex
		if v, ok := tableIndex[tableName.String][name.String]; !ok {
			index = newIndex
		} else {
			index = v
		}

		index.Name = newIndex.Name
		index.Cardinality = newIndex.Cardinality
		index.IndexType = newIndex.IndexType
		index.NonUnique = newIndex.NonUnique
		index.Expression = newIndex.Expression

		if columnName.String != "" {
			index.Columns = append(index.Columns, mysqlIndexColumn{
				Name:     columnName.String,
				Nullable: strings.ToLower(nullable.String) == "yes",
				SubPart:  subPart.Int64,
			})
		}

		tableIndex[tableName.String][name.String] = index
	}
	for tableName, indexes := range tableIndex {
		for _, index := range indexes {
			tables[tableNameToTableIndex[tableName]].Indexes = append(tables[tableNameToTableIndex[tableName]].Indexes, index)
		}
	}
	return nil
}

const sqlIndexesExpressionColumnCheck = `
    SELECT COUNT(*) as column_count
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = 'information_schema'
      AND TABLE_NAME = 'STATISTICS'
      AND COLUMN_NAME = 'EXPRESSION';
`

func (ipt *Input) getIndexQuery() (string, error) {
	rows := ipt.q(sqlIndexesExpressionColumnCheck, getMetricName("mysql_object", "indexes_expression_column_check"))

	if rows == nil {
		return "", errors.New("query failed: nil rows")
	}

	if rows.Err() != nil {
		return "", fmt.Errorf("get rows error: %w", rows.Err())
	}
	defer closeRows(rows)

	if rows.Next() {
		var columnCount int
		if err := rows.Scan(&columnCount); err != nil {
			return "", fmt.Errorf("scan error: %w", err)
		}
		if columnCount > 0 {
			return sqlGetIndexes8, nil
		}
	}

	return sqlGetIndexes, nil
}

func (ipt *Input) getVersion() string {
	if ipt.Version != nil {
		return ipt.Version.version
	}

	return ""
}
