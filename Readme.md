# DataKit

## 安装手册

参见[这里](https://gitlab.jiagouyun.com/zy-docs/pd-forethought-helps/blob/dev/03-%E6%95%B0%E6%8D%AE%E9%87%87%E9%9B%86/02-datakit%E9%87%87%E9%9B%86%E5%99%A8/index.md)

## 编译

### 选择不同的编译输出

```
$ make test     # 编译测试环境
$ make pub_test # 发布 datakit 到测试环境

$ make release  # 编译线上发布版本
$ make pub_test # 发布 datakit 到线上环境

# 将 datakit 以镜像方式发布到 https://registry.jiagouyun.com
# 注意：registry.jiagouyun.com 需要一定的权限才能发布镜像
$ make pub_image

$ make agent # 编译不同平台的 telegraf 到 embed 目录
```

> 注意：datakit 没有预发发布

## datakit 使用示例

列举当前 datakit 支持的采集器列表，可 `grep` 输出，采集器带 `[d]` 前缀的为 datakit 采集器，带 `[t]` 为 telegraf 采集器

```
$ ./datakit -tree | grep aliyun
aliyun
  |--[d] aliyunddos
  |--[d] aliyunsecurity
  |--[d] aliyunlog
  |--[d] aliyuncms
  |--[d] aliyuncdn
  |--[d] aliyuncost
  |--[d] aliyunprice
  |--[d] aliyunrdsslowLog
  |--[d] aliyunactiontrail

$ ./datakit -tree | grep cpu
cpu
  |--[t] cpu
```
