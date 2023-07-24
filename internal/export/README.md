# dashboard 和 monitor 的构建方式

一个具体的采集器，有如下两种方式来构建其 dashboard 和 monitor

1. 在采集器中，实现 Monitor 和 Dashboard 两个接口，在接口中导出其 monitor 和 dashboard 所需的数据。这种是最终的理想状态。同时在 *dashboards* 和 *monitors* 目录下，直接放置单独的一个 json 模板即可。 Datakit 的导出模块能根据该 json 模板分别导致不同语种的内置视图和告警模板。参考采集器 dk 的实现。
1. 不改任何采集器代码，在 *dashboards* 和 *monitors* 目录下，分别在各自的 *{zh,en}* 存放不同语种的内置视图和告警模版。

## 准备工作

先拉取仓库到你的 *~/git* 目录下：

> 如果没有这个仓库的权限，请联系开通。

```shell
$ git clone git@gitee.com:dataflux/dataflux-template.git
```

执行 *mkdocs.sh* 即可生成对应的 dashboard，其目录在 *~/git/dataflux-template/{dashboard,monitor}/{zh,en}/[input-name]/meta.json*。

额外需要做的工作：要求设计人员（@许海浩）设计一下两个图标，放到 *~/git/dataflux-template/icon/* 对应目录下。
