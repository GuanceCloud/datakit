
# 集成文档合并

---

本文档主要介绍如何将现有的集成文档合并进 DataKit 的文档中。现有集成文档在[这里](https://www.yuque.com/dataflux/integrations){:target="_blank"}。

<!-- markdownlint-disable MD046 -->
???+ note

    DataKit 集成相关的文档，均不建议直接在 *dataflux-doc/docs/integrations* 中修改，因为 datakit 自身的文档导出是覆盖式写到该目录的，可能导致手动添加到 *dataflux-doc/docs/integrations* 的文档被覆盖。
<!-- markdownlint-enable -->

名词定义：

- 文档库：指新的文档库 `dataflux-doc`

大概看了下，集成文档合并到 DataKit 文档，有几种可能：

- 合并集成文档：直接扩展现有采集器文档，比如 [CPU 集成文档](https://www.yuque.com/dataflux/integrations/fyiw75){:target="_blank"}，可以直接合并到采集器的 cpu.md （man/manuals/cpu.md）中
- 新增 DataKit 文档：如果 DataKit 中并无对应的文档，那么需要手动在 DataKit 中新增文档

下面将针对以上几种情况，分别说明如何合并。

## 合并集成文档 {#merge}

已有的 DataKit 文档，大部分集成文档中的内容都已具备，主要缺少的是截图信息以及场景导航，除此之外，环境配置和指标信息基本都已具备。故合并的时候，只需要添加一些截图信息即可：

- 在现有语雀的集成文档中，获取截图的链接地址，在当前的集成文档库中直接下载即可：

```shell
cd dataflux-doc/docs/integrations
wget http://yuque-img-url.png -O imgs/input-xxx-0.png
wget http://yuque-img-url.png -O imgs/input-xxx-1.png
wget http://yuque-img-url.png -O imgs/input-xxx-2.png
...
```

> 注意：不要将图片下载到 DataKit 项目所在的文档目录中。

对某个具体的采集器而言，此处可能有多张截图，建议这里以固定的命名规范来保存这些图片，即图片都保存在集成文档库的 *imgs* 目录下，并且每张采集器有关的图片都以 `input-` 为前缀，并且按照编号来命名。

下载完图片后，DataKit 文档中，将图片添加进去即可，具体可查看现有 CPU 采集器示例（man/manuals/cpu.md）

- 编译 DataKit

由于修改的是 DataKit 自身的文档，故需要编译才能生效。DataKit 编译，参见[这里](https://github.com/GuanceCloud/datakit/blob/github-mirror/README.zh_CN.md){:target="_blank"}。

如果编译过程有困难，可以暂时不管，直接将上述修改提交 merge request 到 DataKit 仓库即可，暂时可以由开发这边编译并最终同步到文档库。

## 新增 DataKit 文档 {#add}

对于 DataKit 中没有直接采集器支持的集成文档，添加起来会简单一点，下面以现有集成库中的 Resin 为例，分别说明上述过程。

- 从语雀现有页面获取 markdown 原文，保存到 *man/manuals/* 目录下

直接在 Resin 集成页面的 URL 后加上 markdown 后，[访问即可得到其 Markdown 原文](https://www.yuque.com/dataflux/integrations/resin/markdown){:target="_blank"}，全选拷贝，保存到 *man/manuals/resin.md* 中。

下载下来之后，要修改里面的排版，具体而言，去掉一些无谓的 HTML 修饰（可看下当前 *resin.md* 是怎么改的），另外就是将那些图片全部下载下来（跟上面 CPU 的示例一样）保存，然后在新的 *resin.md* 中引用这些图片。

- 修改 *man/manuals/integrations.pages* 中的目录结构，新增对应的文档

由于 Resin 是一类 Web 服务器，故在现有 *integrations.pages* 文件中，我们将其跟 Nginx/Apache 放在一起：

```yaml
- "Web 服务器"
  - 'Nginx': nginx.md
  - apache.md
  - resin.md
```

- 修改 *mkdocs.sh* 脚本

修改 *mkdocs.sh* 脚本，将新增的文档增加到导出列表中：

```shell
cp man/manuals/resin.md $integration_docs_dir/
```

## 文档生成和导出 {#export}

在 DataKit 现有仓库中，直接执行 *mkdocs.sh* 即可实现编译、发布两个步骤。在 *mkdocs.sh* 中，目前直接将文档分成两份导出，分别同步到文档库的 *datakit* 和 *developers* 两个目录下。

如果要在文档中插入图片，在 *datakit* 和 *integrations* 各自的 *imgs* 目录下放置图片即可。如何引用图片，参考[上面的例子](integrations-to-dk-howto.md#merge)。

下面具体说下文档库的本地操作方式。主要以下几个步骤。

- Clone 现有文档库并安装对应依赖

``` shell
cd ~/ && mkdir -p git && cd git
git clone ssh://git@gitlab.jiagouyun.com:40022/zy-docs/dataflux-doc.git
cd dataflux-doc
pip install -r requirements.txt # 期间可能要求你更新 pip 版本
```

<!-- markdownlint-disable MD046 -->
???+ note

    MKDocs 安装完成后，可能需要设置 `$PATH`，Mac 的设置可能是这样的（具体可以 find 下 `mkdocs` 二进制位置）：
    
    ``` shell
    PATH="/System/Volumes/Data/Users/<user-name>/Library/Python/3.8/bin:$PATH"
    ```
<!-- markdownlint-enable -->

- 熟悉 *mkdocs.sh*

在 DataKit 根目录下有一个 *mkdocs.sh* 脚本，它负责导出 DataKit 所有文档，并拷贝到文档库的不同目录，最后启动本地文档服务。

- 访问本地 `http://localhost:8000` 即可看到
- 调试完成后，提交 Merge Request 到 DataKit 项目的 `mkdocs` 分支
