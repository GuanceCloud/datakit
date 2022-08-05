# Dataway
---

Dataway 支持所有种类的数据。

## 第一步: 搭建后端存储

使用[观测云](https://console.guance.com/)的 Dataway, 或者自己搭建一个 Dataway 的 server 环境。

## 第二步: 增加配置

在 `datakit.conf` 中增加以下片段:

```conf
...
[sinks]
  [[sinks.sink]]
    categories = ["M", "N", "K", "O", "CO", "L", "T", "R", "S", "P"]
    target = "dataway"
    url = "https://openway.guance.com?token=<YOUR-TOKEN>"
    ; proxy = "127.0.0.1:1080"
    filters = ["{ host = 'user-ubuntu' }"] # 这里是举例。这里填写的是过滤条件, 满足该条件的就会往上述 url 里面打数据。
...
```

除了 Sink 必须配置[通用参数](datakit-sink-guide.md)外, Dataway 的 Sink 实例目前支持以下参数:

- `url`(必须): 这里填写 dataway 的全地址(带 token)。
- `token`(可选): 工作空间的 token。如果在 `url` 里面写了这里就可以不用填。
- `filters`(可选): 过滤规则。类似于 io 的 `filters`, 但功能是截然相反的。sink 里面的 filters 匹配满足了才写数据; io 里面的 filters 匹配满足了则丢弃数据。前者是 `include` 后者是 `exclude`。
- `proxy`(可选): 代理地址, 如 `127.0.0.1:1080`。

## 第三步: 重启 DataKit

`$ sudo datakit --restart`

## 安装阶段指定 Dataway Sink 设置

Dataway 支持安装时环境变量开启的方式。

```shell
# 这里为了演示，写了 2 个 filters, 只要满足任何一个条件就会写数据。
DK_SINK_M="dataway://?url=https://openway.guance.com&token=<YOUR-TOKEN>&filters={host='user-ubuntu'}&filters={cpu='cpu-total'}" \
bash -c "$(curl -L https://static.guance.com/datakit/install.sh)"
```
