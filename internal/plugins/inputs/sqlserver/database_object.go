// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package sqlserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	dkio "gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/io"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

const (
	tableChunkSize                 = 500
	sqlserverObjectMeasurementName = "database"
	sqlserverType                  = "SQLServer"

	counterTypeBulkCount = 272696576
)

var (
	settingColumns = map[string]bool{
		"name":         true,
		"value":        true,
		"minimum":      true,
		"maximum":      true,
		"value_in_use": true,
		"is_dynamic":   true,
		"is_advanced":  true,
	}

	settingColumnCast = map[string]string{
		"minimum":      "varchar(max)",
		"maximum":      "varchar(max)",
		"value_in_use": "varchar(max)",
		"value":        "varchar(max)",
	}
)

type sqlserverObjectMeasurement struct{}

//nolint:lll
func (*sqlserverObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: sqlserverObjectMeasurementName,
		Cat:  point.Object,
		Desc: "SQLServer object metrics([:octicons-tag-24: Version-1.78.0](../datakit/changelog-2025.md#cl-1.78.0))",
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "The hostname of the SQLServer server"},
			"server":        &inputs.TagInfo{Desc: "The server address of the SQLServer server"},
			"version":       &inputs.TagInfo{Desc: "The version of the SQLServer server"},
			"name":          &inputs.TagInfo{Desc: "The name of the database. The value is `host:port` in default"},
			"database_type": &inputs.TagInfo{Desc: "The type of the database. The value is `SQLServer`"},
			"port":          &inputs.TagInfo{Desc: "The port of the SQLServer server"},
		},
		Fields: map[string]interface{}{
			"message": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of database information"},
			"uptime":  &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "The number of seconds that the server has been up"},
			// "slow_queries":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of queries that have taken more than long_query_time seconds. This counter increments regardless of whether the slow query log is enabled."},
			"avg_query_time": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.TimestampUS, Desc: "The average time taken by a query to execute"},
			"qps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of queries executed by the database per second"},
			"tps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of transactions executed by the database per second"},
			// "slow_query_log": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Whether the slow query log is enabled. The value can be 0 (or OFF) to disable the log or 1 (or ON) to enable the log."},
		},
	}
}

type metricValue struct {
	LastValue float64
	CntrType  int
	Value     float64

	LastTime time.Time
}

type objectMertric struct {
	Queries       int64
	SlowQueries   int64
	Transactions  *metricValue
	BatchRequests *metricValue
	Time          time.Time
}

type sqlserverObjectMessage struct {
	Setting   []map[string]*interface{} `json:"setting"`
	Databases []*sqlserverDatabase      `json:"databases"`
}

type sqlserverDatabase struct {
	Name      string             `json:"name"`
	ID        string             `json:"-"`
	OwnerName string             `json:"owner_name"`
	Collation string             `json:"collation"`
	Schemas   []*sqlserverSchema `json:"schemas"`
}

type sqlserverSchema struct {
	Name      string            `json:"name"`
	ID        string            `json:"-"`
	OwnerName string            `json:"owner_name"`
	Tables    []*sqlserverTable `json:"tables"`
}

type sqlserverTable struct {
	Name        string                 `json:"name"`
	ID          string                 `json:"-"`
	Columns     []*sqlserverColumn     `json:"columns"`
	Indexes     []*sqlserverIndex      `json:"indexes"`
	ForeignKeys []*sqlserverForeignKey `json:"foreign_keys"`
	Partitions  *sqlserverPartition    `json:"partitions"`
}

type sqlserverColumn struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Default  string `json:"default"`
	Nullable bool   `json:"nullable"`

	tableName string
}

type sqlserverIndex struct {
	Name               string `json:"name"`
	Type               string `json:"type"`
	IsUnique           bool   `json:"is_unique"`
	IsPrimaryKey       bool   `json:"is_primary_key"`
	IsUniqueConstraint bool   `json:"is_unique_constraint"`
	IsDisabled         bool   `json:"is_disabled"`
	ColumnNames        string `json:"column_names"`

	TableID string `json:"-"`
}

type sqlserverForeignKey struct {
	ForeignKeyName    string `json:"foreign_key_name"`
	ReferencingTable  string `json:"referencing_table"`
	ReferencingColumn string `json:"referencing_column"`
	ReferencedTable   string `json:"referenced_table"`
	ReferencedColumn  string `json:"referenced_column"`
	UpdateAction      string `json:"update_action"`
	DeleteAction      string `json:"delete_action"`

	TableID string `json:"-"`
}

type sqlserverPartition struct {
	PartitionCount int64  `json:"partition_count"`
	TableID        string `json:"-"`
}

func (m *sqlserverObjectMessage) String() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

func (ipt *Input) metricCollectSqlserverObject() {
	start := time.Now()
	if !ipt.Object.lastCollectionTime.IsZero() &&
		ipt.Object.lastCollectionTime.Add(ipt.Object.Interval.Duration).After(start) {
		l.Debugf("skip sqlserver_object collection, time interval not reached")
	}

	ipt.Object.lastCollectionTime = start

	opts := ipt.getKVsOpts(point.Object)
	kvs := ipt.getKVs()

	message := sqlserverObjectMessage{}
	setting, err := ipt.getSetting()
	if err != nil {
		l.Warnf("getSetting failed: %s", err.Error())
	} else {
		message.Setting = setting
	}

	databases, err := ipt.getSqlserverDatabases()
	if err != nil {
		l.Warnf("getSqlserverDatabases failed: %s", err.Error())
	} else {
		message.Databases = databases
	}

	version := ipt.Version
	if ipt.MajorVersion > 0 {
		version = fmt.Sprintf("%d", ipt.MajorVersion)
	}

	kvs = kvs.AddTag("version", version).
		AddTag("database_type", sqlserverType).
		AddTag("name", ipt.Object.name).
		AddTag("host", ipt.Object.host).
		AddTag("server", ipt.Object.name).
		AddTag("port", ipt.Object.port).
		AddV2("uptime", ipt.Uptime, false).
		AddV2("message", message.String(), false)

	if ipt.objectMetric != nil {
		if ipt.objectMetric.BatchRequests != nil {
			kvs = kvs.AddV2("qps", ipt.objectMetric.BatchRequests.Value, false)
		}
		if ipt.objectMetric.Transactions != nil {
			kvs = kvs.AddV2("tps", ipt.objectMetric.Transactions.Value, false)
		}
	}

	if avgQueryTime, err := ipt.getAvgQueryTime(); err != nil {
		l.Warnf("failed to get avg query time: %s", err)
	} else {
		kvs = kvs.AddV2("avg_query_time", avgQueryTime, false)
	}

	pts := []*point.Point{point.NewPointV2("database", kvs, opts...)}

	if err := ipt.feeder.Feed(point.Object,
		pts,
		dkio.WithCollectCost(time.Since(start)),
		dkio.WithElection(ipt.Election),
		dkio.WithSource(objectFeedName)); err != nil {
		l.Warnf("feeder.Feed: %s, ignored", err)
	}
}

const sqlAvgQueryTime = `
SELECT SUM(total_elapsed_time) / Sum(execution_count) AS [avg_cost_ms] FROM sys.dm_exec_query_stats where execution_count > 0;
`

func (ipt *Input) getAvgQueryTime() (float64, error) {
	rows, err := ipt.query("avg_query_time", sqlAvgQueryTime)
	if err != nil {
		return 0, fmt.Errorf("query failed: %w", err)
	}

	for _, row := range rows {
		if v, ok := row["avg_cost_ms"]; ok {
			return cast.ToFloat64(v), nil
		}
	}

	return 0, fmt.Errorf("no avg_cost_ms found")
}

const (
	sqlSetting = `
select %s from sys.configurations
`
	sqlSettingColumns    = "select top 0 * from sys.configurations"
	cacheSQLSettingLabel = "setting"
)

func (ipt *Input) getSetting() ([]map[string]*interface{}, error) {
	query, ok := ipt.Object.queryCache[cacheSQLSettingLabel]
	if !ok {
		ctx, cancel := context.WithTimeout(context.Background(), ipt.timeoutDuration)
		defer cancel()
		rows, err := ipt.db.QueryContext(ctx, sqlSettingColumns)
		if err != nil {
			return nil, fmt.Errorf("query columns error: %w", err)
		}
		defer rows.Close() //nolint:errcheck

		if err = rows.Err(); err != nil {
			return nil, fmt.Errorf("query columns rows error: %w", err)
		}

		columns, err := rows.Columns()
		if err != nil {
			return nil, fmt.Errorf("rows columns error: %w", err)
		}

		availableColumns := []string{}

		for _, column := range columns {
			if _, ok := settingColumns[column]; ok {
				availableColumns = append(availableColumns, column)
			} else {
				l.Debugf("column %s not found in sys.configurations", column)
			}
		}

		formattedColumns := []string{}
		for _, column := range availableColumns {
			if v, ok := settingColumnCast[column]; ok {
				formattedColumns = append(formattedColumns, fmt.Sprintf("CAST(%s AS %s) AS %s", column, v, column))
			} else {
				formattedColumns = append(formattedColumns, column)
			}
		}

		query = fmt.Sprintf(sqlSetting, strings.Join(formattedColumns, ", "))
		l.Infof("query setting sql: %s", query)
		ipt.Object.queryCache[cacheSQLSettingLabel] = query
	}

	rows, err := ipt.query("sqlserver_get_setting", query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return rows, nil
}

func (ipt *Input) getSqlserverDatabases() ([]*sqlserverDatabase, error) {
	dbs, err := ipt.getDatabasesInfo([]string{ipt.Database})
	if err != nil {
		return nil, fmt.Errorf("get databases info failed: %w", err)
	}

	return dbs, nil
}

const sqlDatabaseInfo = `
SELECT
    db.database_id AS id, db.name AS name, db.collation_name AS collation, dp.name AS owner
FROM
    sys.databases db LEFT JOIN sys.database_principals dp ON db.owner_sid = dp.sid
WHERE db.name IN ('%s');
`

func (ipt *Input) getDatabasesInfo(dbs []string) ([]*sqlserverDatabase, error) {
	if len(dbs) == 0 {
		return nil, fmt.Errorf("no database name")
	}

	query := fmt.Sprintf(sqlDatabaseInfo, strings.Join(dbs, "','"))
	rows, err := ipt.query("database_info", query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	databases := make([]*sqlserverDatabase, 0)

	for _, row := range rows {
		database := &sqlserverDatabase{
			ID:        getFieldString(row["id"]),
			Name:      getFieldString(row["name"]),
			OwnerName: getFieldString(row["owner"]),
			Collation: getFieldString(row["collation"]),
		}

		if schemas, err := ipt.getSchemas(database.Name); err != nil {
			l.Errorf("getSchemas failed: %s", err.Error())
		} else {
			database.Schemas = schemas
		}

		databases = append(databases, database)
	}

	return databases, nil
}

const sqlSchema = `
SELECT
    s.name AS name, s.schema_id AS id, dp.name AS owner_name
FROM
    sys.schemas AS s JOIN sys.database_principals dp ON s.principal_id = dp.principal_id
WHERE s.name NOT IN ('sys', 'information_schema')
`

func (ipt *Input) getSchemas(database string) ([]*sqlserverSchema, error) {
	var schemas []*sqlserverSchema

	rows, err := ipt.query("sqlserver_schemas", sqlSchema, database)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	for _, row := range rows {
		schema := &sqlserverSchema{
			Name:      getFieldString(row["name"]),
			OwnerName: getFieldString(row["owner_name"]),
			ID:        getFieldString(row["id"]),
		}

		schema.Tables, err = ipt.getTables(database, schema)
		if err != nil {
			l.Warnf("getTables failed: %s", err.Error())
		}
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

const sqlGetTables = `
SELECT
    object_id AS id, name
FROM
    sys.tables
WHERE schema_id=%s
`

func (ipt *Input) getTables(database string, schema *sqlserverSchema) ([]*sqlserverTable, error) {
	var tables []*sqlserverTable
	query := fmt.Sprintf(sqlGetTables, schema.ID)
	rows, err := ipt.query("get_tables", query, database)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	for _, row := range rows {
		tables = append(tables, &sqlserverTable{
			ID:   getFieldString(row["id"]),
			Name: getFieldString(row["name"]),
		})
	}

	startIndex := 0

	for {
		if startIndex >= len(tables) {
			break
		}

		endIndex := startIndex + tableChunkSize
		if endIndex > len(tables) {
			endIndex = len(tables)
		}

		if err := ipt.populateTablesData(database, schema, tables[startIndex:endIndex]); err != nil {
			return nil, fmt.Errorf("populateTablesData failed: %w", err)
		}

		startIndex = endIndex
	}

	return tables, nil
}

func (ipt *Input) populateTablesData(database string, schema *sqlserverSchema, tables []*sqlserverTable) error {
	nameToID := map[string]string{}
	idToTable := map[string]*sqlserverTable{}
	tableIDs := []string{}
	tableIDsObject := []string{}

	for _, table := range tables {
		nameToID[table.Name] = table.ID
		idToTable[table.ID] = table
		tableIDs = append(tableIDs, table.ID)
		tableIDsObject = append(tableIDsObject, fmt.Sprintf("OBJECT_NAME(%s)", table.ID))
	}

	// get columns
	if columns, err := ipt.getColumns(database, schema, tableIDsObject); err != nil {
		return fmt.Errorf("getColumns failed: %w", err)
	} else {
		for _, column := range columns {
			tableID := nameToID[column.tableName]
			table := idToTable[tableID]
			if table == nil {
				continue
			}
			table.Columns = append(table.Columns, column)
		}
	}

	// get partitions
	if partitions, err := ipt.getPartitions(database, tableIDs); err != nil {
		return fmt.Errorf("getPartitions failed: %w", err)
	} else {
		for _, partition := range partitions {
			table := idToTable[partition.TableID]
			if table == nil {
				continue
			}
			table.Partitions = partition
		}
	}

	// get foreign keys
	if foreignKeys, err := ipt.getForeignKeys(database, tableIDs); err != nil {
		return fmt.Errorf("getForeignKeys failed: %w", err)
	} else {
		for _, foreignKey := range foreignKeys {
			table := idToTable[foreignKey.TableID]
			if table == nil {
				continue
			}
			table.ForeignKeys = append(table.ForeignKeys, foreignKey)
		}
	}

	// get indexes
	if indexes, err := ipt.getIndexes(database, tableIDs); err != nil {
		return fmt.Errorf("getIndexes failed: %w", err)
	} else {
		for _, index := range indexes {
			table := idToTable[index.TableID]
			if table == nil {
				continue
			}
			table.Indexes = append(table.Indexes, index)
		}
	}

	return nil
}

const sqlColumns = `
SELECT
    column_name AS name, data_type, column_default, is_nullable AS nullable , table_name, ordinal_position
FROM
    INFORMATION_SCHEMA.COLUMNS
WHERE
    table_name IN (%s) and table_schema='%s';
`

func (ipt *Input) getColumns(database string, schema *sqlserverSchema, tableIDObjects []string) (columns []*sqlserverColumn, err error) {
	if len(tableIDObjects) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(sqlColumns, strings.Join(tableIDObjects, ","), schema.Name)
	rows, err := ipt.query("get_columns", query, database)
	if err != nil {
		return nil, fmt.Errorf("query columns failed: %w", err)
	}

	columns = make([]*sqlserverColumn, 0)
	for _, row := range rows {
		column := &sqlserverColumn{
			Name:      getFieldString(row["name"]),
			DataType:  getFieldString(row["data_type"]),
			Default:   getFieldString(row["column_default"]),
			tableName: getFieldString(row["table_name"]),
		}

		nullable := strings.ToLower(getFieldString(row["nullable"]))

		if nullable == "no" || nullable == "false" {
			column.Nullable = false
		} else {
			column.Nullable = true
		}

		columns = append(columns, column)
	}

	return columns, nil
}

const sqlPartitions = `
SELECT
    object_id AS id, COUNT(*) AS partition_count
FROM
    sys.partitions
WHERE
    object_id IN (%s) GROUP BY object_id;
`

func (ipt *Input) getPartitions(database string, tableIDs []string) (partitions []*sqlserverPartition, err error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(sqlPartitions, strings.Join(tableIDs, ","))
	rows, err := ipt.query("get_partitions", query, database)
	if err != nil {
		return nil, fmt.Errorf("query partitions failed: %w", err)
	}

	for _, row := range rows {
		partition := &sqlserverPartition{
			TableID:        getFieldString(row["id"]),
			PartitionCount: getFieldInt64(row["partition_count"]),
		}
		partitions = append(partitions, partition)
	}

	return partitions, nil
}

const sqlForeignKeys = `
SELECT
    FK.parent_object_id AS table_id,
    FK.name AS foreign_key_name,
    OBJECT_NAME(FK.parent_object_id) AS referencing_table,
    STRING_AGG(COL_NAME(FKC.parent_object_id, FKC.parent_column_id),',') AS referencing_column,
    OBJECT_NAME(FK.referenced_object_id) AS referenced_table,
    STRING_AGG(COL_NAME(FKC.referenced_object_id, FKC.referenced_column_id),',') AS referenced_column,
    FK.delete_referential_action_desc AS delete_action,
    FK.update_referential_action_desc AS update_action
FROM
    sys.foreign_keys AS FK
    JOIN sys.foreign_key_columns AS FKC ON FK.object_id = FKC.constraint_object_id
WHERE
    FK.parent_object_id IN (%s)
GROUP BY
    FK.name,
    FK.parent_object_id,
    FK.referenced_object_id,
    FK.delete_referential_action_desc,
    FK.update_referential_action_desc;
`

const sqlForeignKeys2016 = `
SELECT
    FK.parent_object_id AS table_id,
    FK.name AS foreign_key_name,
    OBJECT_NAME(FK.parent_object_id) AS referencing_table,
    STUFF((
        SELECT ',' + COL_NAME(FKC.parent_object_id, FKC.parent_column_id)
        FROM sys.foreign_key_columns AS FKC
        WHERE FKC.constraint_object_id = FK.object_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS referencing_column,
    OBJECT_NAME(FK.referenced_object_id) AS referenced_table,
    STUFF((
        SELECT ',' + COL_NAME(FKC.referenced_object_id, FKC.referenced_column_id)
        FROM sys.foreign_key_columns AS FKC
        WHERE FKC.constraint_object_id = FK.object_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS referenced_column,
    FK.delete_referential_action_desc AS delete_action,
    FK.update_referential_action_desc AS update_action
FROM
    sys.foreign_keys AS FK
WHERE
    FK.parent_object_id IN (%s)
GROUP BY
    FK.name,
    FK.parent_object_id,
    FK.referenced_object_id,
    FK.delete_referential_action_desc,
    FK.update_referential_action_desc;
`

func (ipt *Input) getForeignKeys(database string, tableIDs []string) (foreignKeys []*sqlserverForeignKey, err error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	query := sqlForeignKeys
	if ipt.MajorVersion > 0 && ipt.MajorVersion <= 2016 {
		query = sqlForeignKeys2016
	}

	query = fmt.Sprintf(query, strings.Join(tableIDs, ","))
	rows, err := ipt.query("get_foreign_keys", query, database)
	if err != nil {
		return nil, fmt.Errorf("query foreign keys failed: %w", err)
	}

	for _, row := range rows {
		foreignKey := &sqlserverForeignKey{
			TableID:           getFieldString(row["table_id"]),
			ForeignKeyName:    getFieldString(row["foreign_key_name"]),
			ReferencingTable:  getFieldString(row["referencing_table"]),
			ReferencingColumn: getFieldString(row["referencing_column"]),
			ReferencedTable:   getFieldString(row["referenced_table"]),
			ReferencedColumn:  getFieldString(row["referenced_column"]),
			UpdateAction:      getFieldString(row["update_action"]),
			DeleteAction:      getFieldString(row["delete_action"]),
		}
		foreignKeys = append(foreignKeys, foreignKey)
	}

	return foreignKeys, nil
}

const sqlIndex = `
SELECT
    i.object_id AS id, i.name, i.type, i.is_unique, i.is_primary_key, i.is_unique_constraint,
    i.is_disabled, STRING_AGG(c.name, ',') AS column_names
FROM
    sys.indexes i JOIN sys.index_columns ic ON i.object_id = ic.object_id
    AND i.index_id = ic.index_id JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
WHERE
    i.object_id IN (%s) GROUP BY i.object_id, i.name, i.type,
    i.is_unique, i.is_primary_key, i.is_unique_constraint, i.is_disabled;
`

const sqlIndex2016 = `
SELECT
    i.object_id AS id,
    i.name,
    i.type,
    i.is_unique,
    i.is_primary_key,
    i.is_unique_constraint,
    i.is_disabled,
    STUFF((
        SELECT ',' + c.name
        FROM sys.index_columns ic
        JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
        WHERE ic.object_id = i.object_id AND ic.index_id = i.index_id
        FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS column_names
FROM
    sys.indexes i
WHERE
    i.object_id IN (i%s)
GROUP BY
    i.object_id,
    i.name,
    i.index_id,
    i.type,
    i.is_unique,
    i.is_primary_key,
    i.is_unique_constraint,
    i.is_disabled;
`

func (ipt *Input) getIndexes(database string, tableIDs []string) (indexes []*sqlserverIndex, err error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}
	query := sqlIndex
	if ipt.MajorVersion > 0 && ipt.MajorVersion <= 2016 {
		query = sqlIndex2016
	}
	query = fmt.Sprintf(query, strings.Join(tableIDs, ","))
	rows, err := ipt.query("get_indexes", query, database)
	if err != nil {
		return nil, fmt.Errorf("query indexes failed: %w", err)
	}

	for _, row := range rows {
		index := &sqlserverIndex{
			TableID:            getFieldString(row["id"]),
			Name:               getFieldString(row["name"]),
			Type:               getFieldString(row["type"]),
			IsUnique:           getFieldBool(row["is_unique"]),
			IsPrimaryKey:       getFieldBool(row["is_primary_key"]),
			IsUniqueConstraint: getFieldBool(row["is_unique_constraint"]),
			IsDisabled:         getFieldBool(row["is_disabled"]),
			ColumnNames:        getFieldString(row["column_names"]),
		}
		indexes = append(indexes, index)
	}

	return indexes, nil
}

func getFieldString(f *interface{}) string {
	if f == nil {
		return ""
	}

	return cast.ToString(*f)
}

func getFieldInt64(f *interface{}) int64 {
	if f == nil {
		return 0
	}

	return cast.ToInt64(*f)
}

func getFieldBool(f *interface{}) bool {
	if f == nil {
		return false
	}

	return cast.ToBool(*f)
}
