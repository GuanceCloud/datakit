

# Reference Table

[:octicons-tag-24: Version-1.4.11](../datakit/changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

---

Through the Reference Table function, Pipeline supports importing external data for data processing.

???+ attention

    This feature consumes high memory, with reference to 1.5 million rows of disk occupying about 200MB (JSON file) of non-repetitive data (string type two columns; int, float, bool), the memory footprint is maintained at 950MB ~ 1.2 GB, and the peak memory at update is 2.2 GB ~ 2.7 GB.

## Table Structure and Column Data Type {#table-struct}

The table structure is a two-dimensional table, which is distinguished from each other by table name. At least one column needs to exist. The data types of elements in each column must be consistent, and the data types must be one of int (int 64), float (float 64), string and bool.

Setting primary keys to tables is not supported yet, but you can query through any column and take the first row of all the results found as the query result. The following is an example of a table structure:

- Table name: `refer_table_abc`

- Column name(col1, col2, ...), column data type(int, float, ...), line data:

| col1: int | col2: float | col3: string | col4: bool |
| ---       | ---         | ---          | ---        |
| 1         | 1.1         | "abc"        | true       |
| 2         | 3           | "def"        | false      |

## Import Data from Outside {#import}


=== "Host Installation"

    Configure reference table url and pull interval in configuration file `datakit.conf` (default interval is 5 minutes)
    
    ```toml
    [pipeline]
      refer_table_url = "http[s]://host:port/path/to/resource"
      refer_table_pull_interval = "5m"
    ```

=== "Kubernetes"

    [see here](../datakit/datakit-daemonset-deploy.md#env-reftab)

---

Supported data formats:

Content-Type: application/json ：

* The data consists of a list of tables, and each table consists of a map with the fields in the map:

| Field Name   | table_name | column_name | column_type                                                         | row_data                                                                                                             |
| ---      | ---        | --          | --                                                                  | ---                                                                                                                  |
| Description     | Table Name       | All Column Names    | Column data type, need to correspond to column name, value range "int", "float", "string", "bool" | Multiple rows of data, for int, float, bool types can use corresponding type data or converted to string representation; Elements in [] any must correspond to column names and column types one by one. |
| Data Type | string     | [ ]string   | [ ]string                                                           | [ ][ ]any                                                                                                            |

* JSON structure:
  
```json
[
    {
        "table_name":  string,
        "column_name": []string{},
        "column_type": []string{},
        "row_data": [
            []any{},
            ...
        ]
    },
    ...
]
```

* example:

```json
[
    {
        "table_name": "table_abc",
        "column_name": ["col", "col2", "col3", "col4"],
        "column_type": ["string", "float", "int", "bool"],
        "row_data": [
            ["a", 123, "123", "false"],
            ["ab", "1234.", "123", true],
            ["ab", "1234.", "1235", "false"]
        ]
    },
    {
        "table_name": "table_ijk",
        "column_name": ["name", "id"],
        "column_type": ["string", "string"],
        "row_data": [
            ["a", "12"],
            ["a", "123"],
            ["ab", "1234"]
        ]
    }
]
```

## Practice Example {#example}

Write the json text above as the file `test.json` and place the file under/var/www/html after installing nginx with apt in Ubuntu 18.04 +

Execute `curl -v localhost/test.json` to test whether the file can be obtained via HTTP GET, and the output is roughly

```txt
...
< Content-Type: application/json
< Content-Length: 522
< Last-Modified: Tue, 16 Aug 2022 06:20:52 GMT
< Connection: keep-alive
< ETag: "62fb3744-20a"
< Accept-Ranges: bytes
< 
[
    {
        "table_name": "table_abc",
        "column_name": ["col", "col2", "col3", "col4"],
        "column_type": ["string", "float", "int", "bool"],
        "row_data": [
...
```

Modify the value of refer_table_url in the configuration file `datakit.conf`:

```toml
[pipeline]
  refer_table_url = "http://localhost/test.json"
  refer_table_pull_interval = "5m"
```

Go into the datakit pipeline logging directory and create the test script `refer_table_for_test.p` and write the following

```python
# Extract table name, column name and column value from input
json(_, table)
json(_, key)
json(_, value)

# Query and append the data of the current column, which is added to the data as field by default
query_refer_table(table, key, value)
```

```shell
cd /usr/local/datakit/pipeline/logging

vim refer_table_for_test.p

datakit pipeline -P refer_table_for_test.p -T '{"table": "table_abc", "key": "col2", "value": 1234.0}' --date
```

As can be seen from the following output results, coll, col2, col3 and col4 of the columns in the table were successfully appended to the output results:

```shell
2022-08-16T15:02:14.150+0800  DEBUG  refer-table  refertable/cli.go:26  performing request[method GET url http://localhost/test.json]
{
  "col": "ab",
  "col2": 1234,
  "col3": 123,
  "col4": true,
  "key": "col2",
  "message": "{\"table\": \"table_abc\", \"key\": \"col2\", \"value\": 1234.0}",
  "status": "unknown",
  "table": "table_abc",
  "time": "2022-08-16T15:02:14.158452592+08:00",
  "value": 1234
}
```
