{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：`{{.AvailableArchs}}`

# 简介

Oracle 监控指标采集，具有以下数据收集功能

- process 相关
- tablespace 相关数据
- system 数据采集
- 自定义查询数据采集

## 前置条件

- 创建监控账号

```sql
-- Enable Oracle Script.
ALTER SESSION SET "_ORACLE_SCRIPT"=true;

-- Create the datakit user. Replace the password placeholder with a secure password.
CREATE USER datakit IDENTIFIED BY <PASSWORD>;

-- Grant access to the datakit user.
GRANT CONNECT TO datakit;
GRANT SELECT ON GV_$PROCESS TO datakit;
GRANT SELECT ON gv_$sysmetric TO datakit;
GRANT SELECT ON sys.dba_data_files TO datakit;
GRANT SELECT ON sys.dba_tablespaces TO datakit;
GRANT SELECT ON sys.dba_tablespace_usage_metrics TO datakit;
```

备注：oracle 11g, 需要以下设定

```sql
ALTER SESSION SET "_ORACLE_SCRIPT"=true;
```

- 安装依赖包

根据操作系统和oracle版本选择安装对应的安装包,参考[这里](https://oracle.github.io/odpi/doc/installation.html)，如：

```shell
$ cat /etc/redhat-release
$ rpm -ivh oracle-instantclient11.2-basic-11.2.0.4.0-1.x86_64.rpm
$ echo /usr/lib/oracle/11.2/client64/lib > /etc/ld.so.conf.d/oracle-instantclient.conf
$ ldconfig

$ yum install libaio # 对应 ubuntu: apt-get install libaio1
```

## 配置

进入 DataKit 安装目录下的 `conf.d/{{.Catalog}}` 目录，复制 `{{.InputName}}.conf.sample` 并命名为 `{{.InputName}}.conf`。示例如下：

```toml
{{.InputSample}}
```

配置好后，重启 DataKit 即可。

## 指标集

以下所有指标集，默认会追加名为 `host` 的全局 tag（tag 值为 DataKit 所在主机名），也可以在配置中通过 `[[inputs.{{.InputName}}.tags]]` 另择 host 来命名。

{{ range $i, $m := .Measurements }}

### `{{$m.Name}}`

-  标签

{{$m.TagsMarkdownTable}}

- 指标列表

{{$m.FieldsMarkdownTable}}

{{ end }}
