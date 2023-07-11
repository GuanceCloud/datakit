# Intro

Dashboard 可以集成到 Datakit 的发布中，在当前目录下，直接添加以采集器命名的 *[input-name].json* 即可。

在采集器中，实现一下 `Dashboard` 接口即可，参见采集器 dk 的实现。注意，这里没有再区分多语言目录，跟文档的实现不同，多语言的区分，在采集器的 `Dashboard()` 接口中实现。 

## 准备工作

先拉取仓库到你的 *~/git* 目录下：

> 如果没有这个仓库的权限，请联系开通。

```shell
$ git clone git@gitee.com:dataflux/dataflux-template.git
```

执行 *mkdocs.sh* 即可生成对应的 dashboard，其目录在 *~/git/dataflux-template/dashboard/{zh,en}/[input-name]/meta.json*。

额外需要做的工作：要求设计人员（@许海浩）设计一下两个图标，一个命名为 *[input-name].png*，一个统一命名为 *icon.svg*。
