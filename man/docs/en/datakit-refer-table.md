{{.CSS}}

# Reference Table

[:octicons-tag-24: Version-1.4.11](../datakit/changelog.md#cl-1.4.11) ·
[:octicons-beaker-24: Experimental](../datakit/index.md#experimental)

---

通过 Reference Table 功能，Pipeline 支持导入外部数据进行数据处理。

???+ attention

    该功能内存消耗较高，参考 150 万行磁盘占用约 200MB (JSON 文件) 的不重复数据 (string 类型两列; int, float, bool 各一列) 为例，其内存占用维持在 950MB ～ 1.2GB, 更新时的峰值内存 2.2GB ~ 2.7GB

## 表结构与列的数据类型 {#table-struct}

表结构为一个二维表，表与表之间通过表名区分，需要至少存在一列，各列内的元素的数据类型必须一致，且数据类型需为 int(int64), float(float64), string, bool 之一。

暂未支持给表设置主键，但是可以通过任意列进行查询，并将查到的所有结果中的第一行作为查询结果。以下为一个表结构示例：

- 表名： `refer_table_abc`

- 列名(col1, col2, ...)、列数据类型(int, float, ...)、行数据：

| col1: int | col2: float | col3: string | col4: bool |
| ---       | ---         | ---          | ---        |
| 1         | 1.1         | "abc"        | true       |
| 2         | 3           | "def"        | false      |

## 从外部导入数据 {#import}

=== "主机安装"

    在配置文件 `datakit.conf` 中配置 reference table url 与拉取间隔(默认间隔为 5 分钟)
    
    ```toml
    [pipeline]
      refer_table_url = "http[s]://host:port/path/to/resource"
      refer_table_pull_interval = "5m"
    ```

=== "Kubernetes"

    [参见这里](datakit-daemonset-deploy.md#env-reftab)

---

???+ attention

    目前要求 refer_table_url 指定的地址，其 HTTP 返回的 Content-Type 必须为 `Content-Type: application/json`。

---

* 数据由多个 table 构成列表，每个表由一个 map 构成，map 中的字段为：

| 字段名   | table_name | column_name | column_type                                                         | row_data                                                                                                             |
| ---      | ---        | --          | --                                                                  | ---                                                                                                                  |
| 描述     | 表名       | 所有列名    | 列数据类型，需要与列名对应，值范围 "int", "float", "string", "bool" | 多个行数据，对于 int, float, bool 类型可以使用对应类型数据或转换成字符串表示; []any 中元素需与列名以及列类型一一对应 |
| 数据类型 | string     | [ ]string   | [ ]string                                                           | [ ][ ]any                                                                                                            |

* JSON 结构：
  
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

* 示例：

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

## 实践示例 {#example}

将上面的 json 文本写成文件 `test.json`，在 Ubuntu18.04+ 使用 apt 安装 nginx 后将文件放置于 /var/www/html 下

执行 `curl -v localhost/test.json` 测试文件是否能通过 HTTP GET 获取到，输出结果大致为

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

在配置文件 `datakit.conf` 修改 refer_table_url 的值为：

```toml
[pipeline]
  refer_table_url = "http://localhost/test.json"
  refer_table_pull_interval = "5m"
```

进入 datakit pipeline loggging 目录, 并创建测试脚本 `refer_table_for_test.p`，并写入以下内容

```python
# 从输入中提取 表名，列名，列值
json(_, table)
json(_, key)
json(_, value)

# 查询并追加当前列的数据，默认作为 field 添加到数据中
query_refer_table(table, key, value)
```

```shell
cd /usr/local/datakit/pipeline/logging

vim refer_table_for_test.p

datakit pipeline refer_table_for_test.p -T '{"table": "table_abc", "key": "col2", "value": 1234.0}' --date
```

由以下输出结果可知，表中列的 col, col2, col3, col4 成功被追加到输出的结果中：

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
