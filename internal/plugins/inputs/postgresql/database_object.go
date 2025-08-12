// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package postgresql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/GuanceCloud/cliutils/point"
	"github.com/spf13/cast"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/internal/plugins/inputs"
)

type objectMertric struct {
	Calls int64
	Trans int64

	CallsTime time.Time
	TransTime time.Time

	AvgQueryTime float64
}

const (
	postgresqlType                  = "PostgreSQL"
	postgresqlObjectMeasurementName = "database"
)

var extensionLoader = map[string]string{
	"pg_trgm":  "SELECT word_similarity('foo', 'bar');",
	"plpgsql":  "DO $$ BEGIN PERFORM 1; END$$;",
	"pgcrypto": "SELECT armor('foo');",
	"hstore":   "SELECT 'a=>1'::hstore;",
}

type postgresqlObjectMessage struct {
	Setting   map[string]string `json:"setting"`
	Databases []*databaseInfo   `json:"databases"`
}

func (m *postgresqlObjectMessage) String() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

type postgresqlObjectMeasurement struct{}

//nolint:lll
func (*postgresqlObjectMeasurement) Info() *inputs.MeasurementInfo {
	return &inputs.MeasurementInfo{
		Name: postgresqlObjectMeasurementName,
		Cat:  point.Object,
		Desc: "PostgreSQL object metrics([:octicons-tag-24: Version-1.76.0](../datakit/changelog-2025.md#cl-1.76.0))",
		Tags: map[string]interface{}{
			"host":          &inputs.TagInfo{Desc: "The hostname of the PostgreSQL server"},
			"server":        &inputs.TagInfo{Desc: "The server address of the PostgreSQL server. The value is `host:port`"},
			"version":       &inputs.TagInfo{Desc: "The version of the PostgreSQL server"},
			"name":          &inputs.TagInfo{Desc: "The name of the database. The value is `host:port` in default"},
			"database_type": &inputs.TagInfo{Desc: "The type of the database. The value is `PostgreSQL`"},
			"port":          &inputs.TagInfo{Desc: "The port of the PostgreSQL server"},
		},
		Fields: map[string]interface{}{
			"message":        &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Summary of database information"},
			"uptime":         &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.DurationSecond, Desc: "The number of seconds that the server has been up"},
			"slow_queries":   &inputs.FieldInfo{DataType: inputs.Int, Unit: inputs.NCount, Desc: "The number of queries that have taken more than long_query_time seconds. This counter increments regardless of whether the slow query log is enabled."},
			"avg_query_time": &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.TimestampUS, Desc: "The average time taken by a query to execute"},
			"qps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of queries executed by the database per second"},
			"tps":            &inputs.FieldInfo{DataType: inputs.Float, Unit: inputs.Gauge, Desc: "The number of transactions executed by the database per second"},
			"slow_query_log": &inputs.FieldInfo{DataType: inputs.String, Unit: inputs.UnknownUnit, Desc: "Whether the slow query log is enabled. The value can OFF to disable the log or ON to enable the log."},
		},
	}
}

func (ipt *Input) collectDatabaseObject() error {
	if !ipt.Object.lastCollectionTime.IsZero() &&
		ipt.Object.lastCollectionTime.Add(ipt.Object.Interval.Duration).After(time.Now()) {
		l.Debugf("skip postgresql_object collection, time interval not reached")
		return nil
	}

	ipt.Object.lastCollectionTime = time.Now()

	kvs := ipt.getKVs()
	opts := ipt.getKVsOpts(point.Object)

	opts = append(opts, point.WithTimestamp(ipt.ptsTime.UnixNano()))

	slowQueryLog := "OFF"
	slowQueries := 0

	message := postgresqlObjectMessage{}
	setting, err := ipt.getPostgreSQLSetting()
	if err != nil {
		l.Warnf("getPostgreSQLSetting failed: %s", err.Error())
	} else {
		message.Setting = setting

		// set slow query log: logging_collector is on and log_min_duration_statement is greater than 0
		if setting["logging_collector"] == "on" {
			if v, ok := setting["log_min_duration_statement"]; ok {
				if num := cast.ToInt64(v); num > 0 {
					slowQueryLog = "ON"
					if count, err := ipt.getSlowQueries(num); err != nil {
						l.Warnf("getSlowQueries failed: %s", err.Error())
					} else {
						slowQueries = count
					}
				}
			}
		}
	}

	// collect schemas
	if ipt.Object.CollectSchemas.Enabled {
		schemas, err := ipt.getSchemas()
		if err != nil {
			l.Warnf("getSchemas failed: %s", err.Error())
		} else {
			message.Databases = schemas
		}
	}

	kvs = kvs.AddTag("version", ipt.version.String()).
		AddTag("database_type", postgresqlType).
		AddTag("name", ipt.Object.name).
		AddTag("host", ipt.host).
		AddTag("slow_query_log", slowQueryLog).
		AddTag("server", ipt.Object.name).
		AddTag("port", fmt.Sprintf("%d", ipt.port)).
		Set("uptime", ipt.Uptime).
		Set("message", message.String())

	// qps, avg_query_time
	if qps, err := ipt.getQPSAndAvgQueryTime(); err != nil {
		l.Warnf("getQPS failed: %s", err.Error())
	} else {
		kvs = kvs.Set("qps", qps)
		kvs = kvs.Set("avg_query_time", ipt.objectMetric.AvgQueryTime)
	}

	// tps
	if tps, err := ipt.getTPS(); err != nil {
		l.Warnf("getTPS failed: %s", err.Error())
	} else {
		kvs = kvs.Set("tps", tps)
	}

	// slow_queries
	if slowQueryLog == "ON" {
		kvs = kvs.Set("slow_queries", slowQueries)
	}

	p := point.NewPoint("database", kvs, opts...)
	ipt.collectCache[point.Object] = []*point.Point{p}

	return nil
}

const (
	sqlExtensions = `SELECT extname, nspname schemaname FROM pg_extension left join pg_namespace on extnamespace = pg_namespace.oid;`
	sqlSetting    = `
SELECT
name,
case when source = 'session' then reset_val else setting end as setting,
source,
sourcefile,
pending_restart
FROM pg_settings
`
)

func (ipt *Input) getPostgreSQLSetting() (map[string]string, error) {
	rows, err := ipt.getQueryResult(sqlExtensions)
	if err != nil {
		return nil, fmt.Errorf("query sqlExtensions failed: %w", err)
	}

	query := sqlSetting
	for _, row := range rows {
		var extension, schemaName string

		if row["extname"] != nil {
			extension = cast.ToString(*row["extname"])
		}

		if row["schemaname"] != nil {
			schemaName = cast.ToString(*row["schemaname"])
		}

		if v, ok := extensionLoader[extension]; ok {
			if schemaName == "pg_catalog" || schemaName == "public" {
				_, err := ipt.getQueryResult(v)
				if err != nil {
					return nil, fmt.Errorf("test extension %s failed: %w", extension, err)
				}
			} else {
				l.Warnf("extension %s is not supported in schema %s", extension, schemaName)
			}
		} else {
			l.Debugf("extension %s is not supported", extension)
		}
	}

	query += ` WHERE name NOT LIKE 'plpgsql%';`

	rows, err = ipt.getQueryResult(query)
	if err != nil {
		return nil, fmt.Errorf("query sqlSetting failed: %w", err)
	}

	setting := make(map[string]string)
	for _, row := range rows {
		var name string
		if row["name"] != nil {
			name = cast.ToString(*row["name"])
		}

		if name != "" {
			if v := row["setting"]; v != nil {
				setting[name] = cast.ToString(*v)
			}
		}
	}

	return setting, nil
}

func (ipt *Input) getQueryResult(query string, dbs ...string) ([]map[string]*interface{}, error) {
	db := ""
	if len(dbs) > 0 {
		db = dbs[0]
	}

	result := make([]map[string]*interface{}, 0)
	rows, err := ipt.service.QueryByDatabase(query, db)
	if err != nil {
		return nil, fmt.Errorf("query sqlExtensions failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("get columns failed: %w", err)
	}

	for rows.Next() {
		columnMap, err := ipt.service.GetColumnMap(rows, columns)
		if err != nil {
			return nil, fmt.Errorf("get column map failed: %w", err)
		}
		result = append(result, columnMap)
	}

	return result, nil
}

const sqlGetSlowQueries = `
SELECT
    COUNT(*) AS slow_query_count
FROM
    pg_stat_statements
WHERE
    total_exec_time / calls > %d;
`

func (ipt *Input) getSlowQueries(minDurationMilliseconds int64) (int, error) {
	if minDurationMilliseconds <= 0 {
		return 0, fmt.Errorf("minDurationMilliseconds must be greater than 0")
	}

	query := fmt.Sprintf(sqlGetSlowQueries, minDurationMilliseconds)
	rows, err := ipt.service.QueryByDatabase(query, "")
	if err != nil {
		return 0, fmt.Errorf("get slow queries failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck
	var count int
	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return 0, fmt.Errorf("scan slow query count failed: %w", err)
		}
	}

	return count, nil
}

func (ipt *Input) getSchemas() ([]*databaseInfo, error) {
	databases, err := ipt.getDatabaseList()
	if err != nil {
		return nil, fmt.Errorf("get database list failed: %w", err)
	}

	dbInfoList := []*databaseInfo{}
	for _, db := range databases {
		dbInfo, err := ipt.getDatabaseInfo(db)
		if err != nil {
			return nil, fmt.Errorf("get database info failed: %w", err)
		}

		dbInfoList = append(dbInfoList, dbInfo)
	}

	return dbInfoList, nil
}

const sqlGetDatabaseList = `
select datname from pg_catalog.pg_database where datistemplate = false;
`

func (ipt *Input) getDatabaseList() ([]string, error) {
	dbs := []string{}
	if ipt.Object.CollectSchemas.AutoDiscoveryDatabase {
		rows, err := ipt.service.QueryByDatabase(sqlGetDatabaseList, "")
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}

		defer rows.Close() //nolint:errcheck

		for rows.Next() {
			var db string
			if err := rows.Scan(&db); err != nil {
				return nil, fmt.Errorf("scan failed: %w", err)
			} else {
				dbs = append(dbs, db)
			}
		}
	} else {
		dbs = append(dbs, ipt.dbName)
	}

	return dbs, nil
}

type databaseInfo struct {
	ID          string        `json:"-"`
	Name        string        `json:"name"`
	Encoding    string        `json:"encoding"`
	Owner       string        `json:"owner"`
	Description string        `json:"description"`
	Schemas     []*schemaInfo `json:"schemas"`
}

type schemaInfo struct {
	ID     int64        `json:"-"`
	Name   string       `json:"name"`
	Owner  string       `json:"owner"`
	Tables []*tableInfo `json:"tables"`
}

type tableInfo struct {
	ID            int64             `json:"-"`
	Name          string            `json:"name"`
	Owner         string            `json:"owner"`
	HasIndexes    bool              `json:"has_indexes"`
	HasPartitions bool              `json:"has_partitions"`
	ToastTable    string            `json:"toast_table"`
	PartitionKey  string            `json:"partition_key"`
	NumPartitions int64             `json:"num_partitions"`
	ForeignKeys   []*foreignKeyInfo `json:"foreign_keys"`
	Columns       []*columnInfo     `json:"columns"`
	Indexes       []*indexInfo      `json:"indexes"`

	indexMap map[string]*indexInfo
}

type indexInfo struct {
	Name        string         `json:"name"`
	Definition  string         `json:"definition"`
	IndexType   string         `json:"index_type"`
	ColumnName  string         `json:"-"`
	IsUnique    bool           `json:"is_unique"`
	IsExclusion bool           `json:"is_exclusion"`
	IsImmediate bool           `json:"is_immediate"`
	IsValid     bool           `json:"is_valid"`
	IsClustered bool           `json:"is_clustered"`
	IsCheckxmin bool           `json:"is_checkxmin"`
	IsReady     bool           `json:"is_ready"`
	IsLive      bool           `json:"is_live"`
	IsReplident bool           `json:"is_replident"`
	IsPartial   bool           `json:"is_partial"`
	IsPrimary   bool           `json:"is_primary"`
	Columns     []*indexColumn `json:"columns"`

	tableID int64
}

type indexColumn struct {
	Name string `json:"name"`
}
type partitionKeyInfo struct {
	Relname      string `json:"relname"`
	PartitionKey string `json:"partition_key"`
}

type numPartitionInfo struct {
	ID            int64 `json:"-"`
	NumPartitions int64 `json:"num_partitions"`
}

type foreignKeyInfo struct {
	Name                  string `json:"name"`
	Definition            string `json:"definition"`
	ID                    int64  `json:"-"`
	ConstraintSchema      string `json:"constraint_schema"`
	ColumnNames           string `json:"column_names"`
	ReferencedTableSchema string `json:"referenced_table_schema"`
	ReferencedTableName   string `json:"referenced_table_name"`
	ReferencedColumnNames string `json:"referenced_column_names"`
	UpdateAction          string `json:"update_action"`
	DeleteAction          string `json:"delete_action"`
}

type columnInfo struct {
	ID       int64  `json:"-"`
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default"`
}

const sqlDatabaseInfo = `
SELECT db.oid                        AS id,
       datname                       AS NAME,
       pg_encoding_to_char(encoding) AS encoding,
       rolname                       AS owner,
       description
FROM   pg_catalog.pg_database db
       LEFT JOIN pg_catalog.pg_description dc
              ON dc.objoid = db.oid
       JOIN pg_roles a
         ON datdba = a.oid
WHERE  datname LIKE '%s';
`

func (ipt *Input) getDatabaseInfo(database string) (*databaseInfo, error) {
	query := fmt.Sprintf(sqlDatabaseInfo, database)

	rows, err := ipt.getQueryResult(query)
	if err != nil {
		return nil, fmt.Errorf("getDatabaseInfo failed: %w", err)
	}

	if len(rows) > 0 {
		row := rows[0]

		var id, name, encoding, owner, description string

		if row["id"] != nil {
			id = cast.ToString(*row["id"])
		}
		if row["name"] != nil {
			name = cast.ToString(*row["name"])
		}
		if row["encoding"] != nil {
			encoding = cast.ToString(*row["encoding"])
		}
		if row["owner"] != nil {
			owner = cast.ToString(*row["owner"])
		}
		if row["description"] != nil {
			description = cast.ToString(*row["description"])
		}

		info := &databaseInfo{
			ID:          id,
			Name:        name,
			Encoding:    encoding,
			Owner:       owner,
			Description: description,
		}

		if schemas, err := ipt.getSchemaInfo(info.Name); err != nil {
			l.Warnf("getSchemaInfo failed: %s", err.Error())
		} else {
			info.Schemas = schemas
		}

		for _, schema := range info.Schemas {
			tables, err := ipt.getSchemaTables(schema.ID, info.Name)
			if err != nil {
				l.Warnf("getSchemaTables failed: %s", err.Error())
			} else {
				if tables, err := ipt.populateTablesData(tables, info.Name); err != nil {
					l.Warnf("populateTablesData failed: %s", err.Error())
				} else {
					schema.Tables = tables
				}
			}
		}

		return info, nil
	} else {
		return nil, fmt.Errorf("database %s not found", database)
	}
}

const sqlSchemaInfo = `
SELECT nsp.oid :: bigint     AS id,
       nspname             AS name,
       nspowner :: regrole AS owner
FROM   pg_namespace nsp
       LEFT JOIN pg_roles r on nsp.nspowner = r.oid
WHERE  nspname NOT IN ( 'information_schema', 'pg_catalog' )
       AND nspname NOT LIKE 'pg_toast%'
       AND nspname NOT LIKE 'pg_temp_%';
`

func (ipt *Input) getSchemaInfo(database string) ([]*schemaInfo, error) {
	rows, err := ipt.service.QueryByDatabase(sqlSchemaInfo, database)
	if err != nil {
		return nil, fmt.Errorf("getSchemaInfo failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	schemas := make([]*schemaInfo, 0)
	for rows.Next() {
		var id int64
		var name, owner string
		if err := rows.Scan(&id, &name, &owner); err != nil {
			return nil, fmt.Errorf("scan schema info failed: %w", err)
		}
		schemas = append(schemas, &schemaInfo{
			ID:    id,
			Name:  name,
			Owner: owner,
		})
	}

	return schemas, nil
}

const sqlSchemaTables9 = `
SELECT c.oid  ::bigint         AS id,
       c.relname             AS name,
       c.relhasindex         AS has_indexes,
       c.relowner :: regrole AS owner,
       t.relname             AS toast_table
FROM   pg_class c
       left join pg_class t
              ON c.reltoastrelid = t.oid
WHERE  c.relkind IN ( 'r', 'f' )
       AND c.relnamespace = '%d' 
`

const sqlSchemaTables10Plus = `
SELECT c.oid   ::bigint        AS id,
       c.relname             AS name,
       c.relhasindex         AS has_indexes,
       c.relowner :: regrole AS owner,
       ( CASE
           WHEN c.relkind = 'p' THEN TRUE
           ELSE FALSE
         END )               AS has_partitions,
       t.relname             AS toast_table
FROM   pg_class c
       left join pg_class t
              ON c.reltoastrelid = t.oid
WHERE  c.relkind IN ( 'r', 'p', 'f' )
       AND c.relispartition != 't'
       AND c.relnamespace = '%d'
`

func (ipt *Input) getSchemaTables(schemaID int64, database string) ([]*tableInfo, error) {
	var (
		id                        sql.NullInt64
		name, owner, toastTable   sql.NullString
		hasIndexes, hasPartitions sql.NullBool
	)
	destFields := []interface{}{&id, &name, &hasIndexes, &owner, &hasPartitions, &toastTable}
	query := sqlSchemaTables10Plus
	if ipt.version.Major == 9 {
		query = sqlSchemaTables9
		destFields = []interface{}{&id, &name, &hasIndexes, &owner, &toastTable}
	}
	query = fmt.Sprintf(query, schemaID)

	rows, err := ipt.service.QueryByDatabase(query, database)
	if err != nil {
		return nil, fmt.Errorf("getSchemaTables failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck
	tables := make([]*tableInfo, 0)
	for rows.Next() {
		if err := rows.Scan(destFields...); err != nil {
			return nil, fmt.Errorf("scan table info failed: %w", err)
		}
		tables = append(tables, &tableInfo{
			ID:            id.Int64,
			Name:          name.String,
			Owner:         owner.String,
			HasIndexes:    hasIndexes.Bool,
			HasPartitions: hasPartitions.Bool,
			ToastTable:    toastTable.String,
		})
	}

	if len(tables) > ipt.Object.CollectSchemas.MaxTables {
		return tables[:ipt.Object.CollectSchemas.MaxTables], nil
	}

	return tables, nil
}

func (ipt *Input) populateTablesData(tables []*tableInfo, database string) ([]*tableInfo, error) {
	maxTables := 50
	tableIDs := []string{}
	tableNameMap := map[int64]string{}
	tableMap := map[string]*tableInfo{}
	for _, table := range tables {
		tableNameMap[table.ID] = table.Name
		tableMap[table.Name] = table
	}

	populateTable := func(ids []string, database string) error {
		// get indexes
		if indexes, err := ipt.getTableIndexes(ids, database); err != nil {
			return fmt.Errorf("get table indexes failed: %w", err)
		} else {
			for _, index := range indexes {
				// partition indexes may have appended digits for each partion
				tableName := tableNameMap[index.tableID]
				for {
					if _, ok := tableMap[tableName]; !ok && endsWithDigit(tableName) {
						tableName = tableName[:len(tableName)-1]
					} else {
						break
					}
				}

				if t, ok := tableMap[tableName]; ok {
					if t.indexMap == nil {
						t.indexMap = map[string]*indexInfo{}
					}
					lastIndex, ok := t.indexMap[index.Name]
					if !ok {
						t.indexMap[index.Name] = index
						lastIndex = index
					}
					lastIndex.Columns = append(lastIndex.Columns, &indexColumn{Name: index.ColumnName})
				}
			}
		}

		for _, t := range tableMap {
			for _, v := range t.indexMap {
				t.Indexes = append(t.Indexes, v)
			}
		}

		// get partitions
		if ipt.version.Major != 9 {
			if partitionKeys, err := ipt.getPartitionKeys(tableIDs, database); err != nil {
				return fmt.Errorf("get partition keys failed: %w", err)
			} else {
				for _, key := range partitionKeys {
					if t, ok := tableMap[key.Relname]; ok {
						t.PartitionKey = key.PartitionKey
					}
				}
			}

			if numPartitions, err := ipt.getNumPartitions(tableIDs, database); err != nil {
				return fmt.Errorf("get num partitions failed: %w", err)
			} else {
				for _, partition := range numPartitions {
					if tableName, ok := tableNameMap[partition.ID]; ok {
						if t, ok := tableMap[tableName]; ok {
							t.NumPartitions = partition.NumPartitions
						}
					}
				}
			}
		}

		// get foreinkeys
		if foreignKeys, err := ipt.getForeignKeys(tableIDs, database); err != nil {
			return fmt.Errorf("get foreign keys failed: %w", err)
		} else {
			for _, key := range foreignKeys {
				if tableName, ok := tableNameMap[key.ID]; ok {
					if t, ok := tableMap[tableName]; ok {
						t.ForeignKeys = append(t.ForeignKeys, key)
					}
				}
			}
		}

		// get columns
		if columns, err := ipt.getColumns(tableIDs, database); err != nil {
			return fmt.Errorf("get columns failed: %w", err)
		} else {
			for _, key := range columns {
				if tableName, ok := tableNameMap[key.ID]; ok {
					if t, ok := tableMap[tableName]; ok {
						t.Columns = append(t.Columns, key)
					}
				}
			}
		}

		return nil
	}

	for _, table := range tables {
		tableIDs = append(tableIDs, fmt.Sprintf("%d", table.ID))
		if len(tableIDs) >= maxTables {
			if err := populateTable(tableIDs, database); err != nil {
				return tables, fmt.Errorf("populate table data failed: %w", err)
			}

			tableIDs = tableIDs[:0]
		}
	}

	if len(tableIDs) > 0 {
		if err := populateTable(tableIDs, database); err != nil {
			return tables, fmt.Errorf("populate table data failed: %w", err)
		}
	}

	return tables, nil
}

const sqlTableIndexes = `
SELECT
    c.relname AS name,
    ix.indrelid AS table_id,
    pg_get_indexdef(c.oid) AS definition,
    ix.indisunique AS is_unique,
    ix.indisexclusion AS is_exclusion,
    ix.indimmediate AS is_immediate,
    ix.indisclustered AS is_clustered,
    ix.indisvalid AS is_valid,
    ix.indcheckxmin AS is_checkxmin,
    ix.indisready AS is_ready,
    ix.indislive AS is_live,
    ix.indisreplident AS is_replident,
    ix.indpred IS NOT NULL AS is_partial,
    ix.indisprimary AS is_primary,
    am.amname AS index_type,
    a.attname AS column_name
FROM
    pg_index ix
JOIN
  pg_class c ON c.oid = ix.indexrelid
JOIN
    pg_class t ON t.oid = ix.indrelid
JOIN
    pg_am am ON c.relam = am.oid
LEFT JOIN
    pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
WHERE
    ix.indrelid IN (%s);
`

func (ipt *Input) getTableIndexes(tableIDs []string, database string) ([]*indexInfo, error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(sqlTableIndexes, strings.Join(tableIDs, ","))
	rows, err := ipt.service.QueryByDatabase(query, database)
	if err != nil {
		return nil, fmt.Errorf("get table indexes failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	indexes := make([]*indexInfo, 0)
	for rows.Next() {
		var (
			tableID                                 sql.NullInt64
			name, definition, indexType, columnName sql.NullString
			isUnique, isExclusion, isImmediate, isClustered,
			isValid, isCheckXmin, isReady, isLive, isReplident, isPartial, isPrimary sql.NullBool
		)
		if err := rows.Scan(&name, &tableID, &definition, &isUnique,
			&isExclusion, &isImmediate, &isClustered, &isValid,
			&isCheckXmin, &isReady, &isLive, &isReplident,
			&isPartial, &isPrimary, &indexType, &columnName); err != nil {
			return nil, fmt.Errorf("scan index info failed: %w", err)
		}

		indexes = append(indexes, &indexInfo{
			Name:        name.String,
			Definition:  definition.String,
			IsUnique:    isUnique.Bool,
			IsExclusion: isExclusion.Bool,
			IsImmediate: isImmediate.Bool,
			IsClustered: isClustered.Bool,
			IsValid:     isValid.Bool,
			IsCheckxmin: isCheckXmin.Bool,
			IsReady:     isReady.Bool,
			IsLive:      isLive.Bool,
			IsReplident: isReplident.Bool,
			IsPartial:   isPartial.Bool,
			IsPrimary:   isPrimary.Bool,
			IndexType:   indexType.String,
			ColumnName:  columnName.String,
			tableID:     tableID.Int64,
		})
	}

	return indexes, nil
}

const sqlPartitionKeys = `
SELECT relname,
       pg_get_partkeydef(oid) AS partition_key
FROM   pg_class
WHERE  oid in (%s);
`

func (ipt *Input) getPartitionKeys(tableIDs []string, database string) ([]*partitionKeyInfo, error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(sqlPartitionKeys, strings.Join(tableIDs, ","))
	rows, err := ipt.service.QueryByDatabase(query, database)
	if err != nil {
		return nil, fmt.Errorf("query partion keys failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	keys := make([]*partitionKeyInfo, 0)
	for rows.Next() {
		var (
			relname      sql.NullString
			partitionKey sql.NullString
		)
		if err := rows.Scan(&relname, &partitionKey); err != nil {
			return nil, fmt.Errorf("scan partition key failed: %w", err)
		}

		keys = append(keys, &partitionKeyInfo{
			Relname:      relname.String,
			PartitionKey: partitionKey.String,
		})
	}

	return keys, nil
}

const sqlNumPartition = `
SELECT count(inhrelid :: regclass) AS num_partitions, inhparent as id
FROM   pg_inherits
WHERE  inhparent IN (%s)
GROUP BY inhparent;
`

func (ipt *Input) getNumPartitions(tableIDs []string, database string) ([]*numPartitionInfo, error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(sqlNumPartition, strings.Join(tableIDs, ","))
	rows, err := ipt.service.QueryByDatabase(query, database)
	if err != nil {
		return nil, fmt.Errorf("query num partition failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	numPartitions := make([]*numPartitionInfo, 0)
	for rows.Next() {
		var (
			id    sql.NullInt64
			count sql.NullInt64
		)
		if err := rows.Scan(&count, &id); err != nil {
			return nil, fmt.Errorf("scan num partition failed: %w", err)
		}

		numPartitions = append(numPartitions, &numPartitionInfo{
			ID:            id.Int64,
			NumPartitions: count.Int64,
		})
	}

	return numPartitions, nil
}

const sqlForeignKeys = `
SELECT con.conname                   AS name,
       pg_get_constraintdef(con.oid) AS definition,
       con.conrelid AS id,
       connamespace.nspname AS constraint_schema,
       attrel.relname AS table_name,
       string_agg(att.attname, ', ' ORDER BY u.attposition) AS column_names,
       refnamespace.nspname AS referenced_table_schema,
       refrel.relname AS referenced_table_name,
       string_agg(refatt.attname, ', ' ORDER BY u.refposition) AS referenced_column_names,
       CASE con.confupdtype
         WHEN 'a' THEN 'NO ACTION'
         WHEN 'r' THEN 'RESTRICT'
         WHEN 'c' THEN 'CASCADE'
         WHEN 'n' THEN 'SET NULL'
         WHEN 'd' THEN 'SET DEFAULT'
       END AS update_action,
       CASE con.confdeltype
        WHEN 'a' THEN 'NO ACTION'
        WHEN 'r' THEN 'RESTRICT'
        WHEN 'c' THEN 'CASCADE'
        WHEN 'n' THEN 'SET NULL'
        WHEN 'd' THEN 'SET DEFAULT'
       END AS delete_action
FROM   pg_constraint con
JOIN
    pg_namespace connamespace ON connamespace.oid = con.connamespace
JOIN
    pg_class attrel ON attrel.oid = con.conrelid
LEFT JOIN LATERAL (
    SELECT
        unnest(con.conkey) AS attnum,
        unnest(con.confkey) AS refattnum,
        generate_subscripts(con.conkey, 1) AS attposition,
        generate_subscripts(con.confkey, 1) AS refposition
) u ON true
JOIN
    pg_attribute att ON att.attrelid = con.conrelid AND att.attnum = u.attnum
JOIN
    pg_class refrel ON refrel.oid = con.confrelid
JOIN
    pg_namespace refnamespace ON refnamespace.oid = refrel.relnamespace
JOIN
    pg_attribute refatt ON refatt.attrelid = con.confrelid AND refatt.attnum = u.refattnum
WHERE  contype = 'f'
       AND conrelid IN (%s)
GROUP BY
    connamespace.nspname,
    con.conname,
    attrel.relname,
    refnamespace.nspname,
    refrel.relname,
    con.confupdtype,
    con.oid,
    con.confdeltype;
`

func (ipt *Input) getForeignKeys(tableIDs []string, database string) ([]*foreignKeyInfo, error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(sqlForeignKeys, strings.Join(tableIDs, ","))
	rows, err := ipt.service.QueryByDatabase(query, database)
	if err != nil {
		return nil, fmt.Errorf("query foreign keys failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	keys := make([]*foreignKeyInfo, 0)
	for rows.Next() {
		var (
			id sql.NullInt64
			name, definition, constraintSchema, tableName,
			columnNames, referencedTableSchema, referencedTableName,
			referencedColumnNames, updateAction, deleteAction sql.NullString
		)
		if err := rows.Scan(&name, &definition, &id, &constraintSchema,
			&tableName, &columnNames, &referencedTableSchema, &referencedTableName,
			&referencedColumnNames, &updateAction, &deleteAction); err != nil {
			return nil, fmt.Errorf("scan foreign keys failed: %w", err)
		}

		keys = append(keys, &foreignKeyInfo{
			ID:                    id.Int64,
			Definition:            definition.String,
			Name:                  name.String,
			ColumnNames:           columnNames.String,
			ReferencedColumnNames: referencedColumnNames.String,
			ReferencedTableName:   referencedTableName.String,
			ReferencedTableSchema: referencedTableSchema.String,
			UpdateAction:          updateAction.String,
			DeleteAction:          deleteAction.String,
			ConstraintSchema:      constraintSchema.String,
		})
	}

	return keys, nil
}

const sqlColumns = `
SELECT attname                          AS name,
       Format_type(atttypid, atttypmod) AS data_type,
       NOT attnotnull                   AS nullable,
       pg_get_expr(adbin, adrelid)      AS default,
       attrelid AS id
FROM   pg_attribute
       LEFT JOIN pg_attrdef ad
              ON adrelid = attrelid
                 AND adnum = attnum
WHERE  attrelid IN (%s)
       AND attnum > 0
       AND NOT attisdropped;
`

func (ipt *Input) getColumns(tableIDs []string, database string) ([]*columnInfo, error) {
	if len(tableIDs) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(sqlColumns, strings.Join(tableIDs, ","))
	rows, err := ipt.service.QueryByDatabase(query, database)
	if err != nil {
		return nil, fmt.Errorf("query columns failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	columns := make([]*columnInfo, 0)
	for rows.Next() {
		var (
			id                         sql.NullInt64
			name, dataType, defaultVal sql.NullString
			nullable                   sql.NullBool
		)
		if err := rows.Scan(&name, &dataType, &nullable, &defaultVal, &id); err != nil {
			return nil, fmt.Errorf("scan columns failed: %w", err)
		}

		columns = append(columns, &columnInfo{
			ID:       id.Int64,
			Name:     name.String,
			DataType: dataType.String,
			Nullable: nullable.Bool,
			Default:  defaultVal.String,
		})
	}

	return columns, nil
}

const sqlGetCalls = `
select sum(calls) as calls, sum(total_exec_time) as total_exec_time from pg_stat_statements;
`

func (ipt *Input) getQPSAndAvgQueryTime() (float64, error) {
	rows, err := ipt.service.QueryByDatabase(sqlGetCalls, "")
	if err != nil {
		return 0, fmt.Errorf("query columns failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	now := time.Now()
	var (
		calls         sql.NullInt64
		totalExecTime sql.NullFloat64
		qps           float64
	)

	for rows.Next() {
		if err := rows.Scan(&calls, &totalExecTime); err != nil {
			return 0, fmt.Errorf("scan columns failed: %w", err)
		}
	}

	if calls.Int64 > 0 {
		ipt.objectMetric.AvgQueryTime = totalExecTime.Float64 * 1000 / float64(calls.Int64) // us
	}

	// calculate qps
	if !ipt.objectMetric.CallsTime.IsZero() && calls.Int64 > ipt.objectMetric.Calls {
		qps = float64(calls.Int64-ipt.objectMetric.Calls) / now.Sub(ipt.objectMetric.CallsTime).Seconds()
	}
	ipt.objectMetric.Calls = calls.Int64
	ipt.objectMetric.CallsTime = now

	return qps, nil
}

const sqlGetTPS = `select sum(xact_commit) + sum(xact_rollback) from pg_stat_database;`

func (ipt *Input) getTPS() (float64, error) {
	rows, err := ipt.service.QueryByDatabase(sqlGetTPS, "")
	if err != nil {
		return 0, fmt.Errorf("query columns failed: %w", err)
	}

	defer rows.Close() //nolint:errcheck

	now := time.Now()
	var (
		sum sql.NullInt64
		tps float64
	)

	for rows.Next() {
		if err := rows.Scan(&sum); err != nil {
			return 0, fmt.Errorf("scan columns failed: %w", err)
		}
	}

	// calculate tps
	if !ipt.objectMetric.TransTime.IsZero() && sum.Int64 > ipt.objectMetric.Trans {
		tps = float64(sum.Int64-ipt.objectMetric.Trans) / now.Sub(ipt.objectMetric.TransTime).Seconds()
	}
	ipt.objectMetric.Trans = sum.Int64
	ipt.objectMetric.TransTime = now

	return tps, nil
}

func endsWithDigit(s string) bool {
	if len(s) == 0 {
		return false
	}

	lastChar := rune(s[len(s)-1])

	return unicode.IsDigit(lastChar)
}
