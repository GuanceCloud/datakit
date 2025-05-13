---
title     : 'SCheck'
summary   : '接收 SCheck 采集的数据'
tags:
  - '安全'
__int_icon: 'icon/scheck'
---

操作系统支持：:fontawesome-brands-linux: :fontawesome-brands-windows:

---

DataKit 可以直接接入 Security Checker 的数据。Security Checker 具体使用，参见[这里](../scheck/scheck-install.md)。

## 通过 DataKit 安装 Security Checker 安装 {#install}

```shell
sudo datakit install --scheck
```

安装完后，Security Checker 默认将数据发送给 DataKit `:9529/v1/write/security` 接口。
