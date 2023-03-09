
# Query Data Through DQL
---

DataKit supports interactive execution of DQL queries. In interactive mode, DataKit comes with statement completion function:

> More command-line parameter help is available through datakit help dql.

```shell
datakit dql      # or datakit -Q
```

<figure markdown>
  ![](https://static.guance.com/images/datakit/dk-dql-gif.gif){ width="800" }
  <figcaption> Example of DQL Interaction Execution </figcaption>
</figure>

Tips：

- Enter `echo_explain` to see the back-end query statement
- To avoid displaying too many `nil` uery results, you can switch it through `disable_nil/enable_nil`.
- Support fuzzy search of query statement. For example, `echo_explain` only needs to input `echo` or `exp` to pop up a prompt, ==Drop-down prompt can be selected through Tab==
- DataKit automatically saves the previous DQL query history (up to 5000 queries), which can be selected by the up and down arrow keys

> Note: Under Windows, execute `datakit dql` in Powershell.

#### Execute DQL query {#dql-once}

With regard to DQL queries, DataKit supports the ability to run a single DQL statement:

```shell
# Execute one query statement at a time
datakit dql --run 'cpu limit 1'

# Write the execution results to the CSV file
datakit dql --run 'O::HOST:(os, message)' --csv="path/to/your.csv"

# Force overwrite of existing CSV files
datakit dql --run 'O::HOST:(os, message)' --csv /path/to/xxx.csv --force

# When the result is written into CSV, the query result is also displayed at the terminal
datakit dql --run 'O::HOST:(os, message)' --csv="path/to/your.csv" --vvv
```

Example of exported CSV file style:

```shell
name,active,available,available_percent,free,host,time
mem,2016870400,2079637504,24.210166931152344,80498688,achen.local,1635242524385
mem,2007961600,2032476160,23.661136627197266,30900224,achen.local,1635242534385
mem,2014437376,2077097984,24.18060302734375,73502720,achen.local,1635242544382
```

Note:

- The first column is the measurement name of the query.
- The following columns are the data corresponding to the collector.
- When the field is empty, the corresponding column is also empty.

#### DQL Query Leading to JSON Result {#json-result}

Output results in JSON, but there is no statistics in JSON mode, such as the number of rows returned and time consumption (to ensure that JSON can be parsed directly).

```shell
datakit dql --run 'O::HOST:(os, message)' --json

# Automatically do json beautification if the field value is a json string (note: in json mode (that is,--json), the `--auto-json` option is invalid).
datakit dql --run 'O::HOST:(os, message)' --auto-json
-----------------[ r1.HOST.s1 ]-----------------
message ----- json -----  # JSON is clearly marked at the beginning, where message is the field name.
{
  "host": {
    "meta": {
      "host_name": "www",
  ....                    # Omit long text here
  "config": {
    "ip": "10.100.64.120",
    "enable_dca": false,
    "http_listen": "localhost:9529",
    "api_token": "tkn_f2b9920f05d84d6bb5b14d9d39db1dd3"
  }
}
----- end of json -----   # There is a clear sign at the end of JSON
     os 'darwin'
   time 2021-09-13 16:56:22 +0800 CST
---------
8 rows, 1 series, cost 4ms
```

#### Query Data for a Specific Workspace {#query-on-wksp}

Query the data of other workspaces by specifying different Token:

```shell
datakit dql --run 'O::HOST:(os, message)' --token <your-token>
```
