# influxdb  VS  clickhouse
## FT 通用函数列表

| type | influxdb | clickhose|
|  ----  | ----  | ----|
|位置函数| first | any |
||last|anyLast|
||min|min|
||max|max|
|聚合函数|mean|avg|
||count|count|
||distinct|distinct|
||median| median |
||mode|anyHeavy|
||sum|sum|
||stddev|stddevPop|
||spread|max()-min()|

## 降精度方法示例

### clickhouse

    select sum(f2), time from (select *, toStartOfInterval(toDateTime(time), `INTERVAL 30 minute`) as time from test0227i_rp0_all) group by time order by time

### influxdb

    select sum(f2) from test0227i_rp0 group by time(1800s) order by time

## 指标查询

### clickhouse

    show tables； 获取当前数据库的所有表名，即指标集；
    show create table tableName_xxxx;获取表结构信息
    
### influxdb

    show measurements；获取当前数据库的指标集；
    show field keys from measeurement_xxx；获取指标集的field字段；
    show tag keys from measurement_xxx；获取指标集的tag字段；
    
