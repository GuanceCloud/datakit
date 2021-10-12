{{.CSS}}

- DataKit 版本：{{.Version}}
- 文档发布日期：{{.ReleaseDate}}
- 操作系统支持：Linux

# Security Checker 接入

Datakit 可以直接接入 Security Checker的数据。Security Checker 具体使用，参见[这里](https://www.yuque.com/dataflux/sec_checker/install) 

## 通过 DataKit 安装 Security Checker 安装

```shell
$ sudo datakit --install scheck
```

安装完后，Security Checker 默认将数据发送给 DataKit `:9529/v1/write/security` 接口
