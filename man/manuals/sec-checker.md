{{.CSS}}

- 版本：{{.Version}}
- 发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux


Datakit 可以直接接入 [Security Checker](http://doc/to/sec-checker) 的数据。

> Security Checker 具体使用，参见[这里](http://doc/to/sec-checker)

编辑 sec-checker 配置文件（一般位于 `/usr/local/security-checker/checker.conf`），将 `output` 指向 DataKit 的时序数据接口即可：

```toml
output = 'http://localhost:9529/v1/write/security' # datakit 1.1.6(含)以上版本才支持
```
