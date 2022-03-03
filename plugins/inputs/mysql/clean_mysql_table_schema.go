package mysql

import "fmt"

// port from getTableSchema.
func getCleanTableSchema(r rows) []map[string]interface{} {
	if r == nil {
		return nil
	}

	var list []map[string]interface{}

	defer closeRows(r)

	for r.Next() {
		res := make(map[string]interface{})

		var (
			tableSchema string
			tableName   string
			tableType   string
			engine      string
			version     int
			rowFormat   string
			tableRows   int
			dataLength  int
			indexLength int
			dataFree    int
		)

		err := r.Scan(
			&tableSchema,
			&tableName,
			&tableType,
			&engine,
			&version,
			&rowFormat,
			&tableRows,
			&dataLength,
			&indexLength,
			&dataFree,
		)
		if err != nil {
			l.Warnf("Scan: %s, ignored", err)
			continue
		}

		// tags
		res["table_schema"] = tableSchema
		res["table_name"] = tableName
		res["table_type"] = tableType
		res["engine"] = engine
		res["version"] = fmt.Sprintf("%d", version)

		// fields
		res["table_rows"] = tableRows
		res["data_length"] = dataLength
		res["index_length"] = indexLength
		res["data_free"] = dataFree

		list = append(list, res)
	}

	return list
}
