# Mkdocs 文档撰写
---

本文主要阐述以下几个问题：

- DataKit 相关的文档编写步骤
- 如何用 Mkdocs 写出更好的文档

## DataKit 相关编写步骤 {#steps}

由于 DataKit 文档大部分都是用代码生成的，还有部分是纯手写的实践性文档，对不同的文档我们需要区别对待。目前 DataKit 相关的文档主要分为两类（即两大类集成），一类是 DataKit 文档，一个是集成文档：

- DataKit：跟具体数据采集不直接相关的文档，主要是一些 DataKit 总体使用上的文档
- 集成文档：主要是采集器相关的文档，这种文档又细分为俩类：
    - 一类是 DataKit 内置采集器相关的文档
    - 一类是衍生出来的数据采集文档，它们不通过 DataKit 生成，只是简单的添加在 man/docs 目录下。这些采集器的数据采集，一般通过 DataKit 已有采集器（比如 prom 等）来采集

对于一篇新的文档，作者应该辨明应该存放在哪个文档库中，在 mkdocs.sh 脚本中，对文档分成了三类，这三类分别对应上述三类文档：

- `datakit_docs`
- `integrations_files_from_datakit`
- `integrations_extra_files`

文档作者确定了具体文档的归属后，将其添加到这三个数组中即可，mkdocs.sh 脚本会自动将文档发布到正确的文档库。故新文档的撰写步骤为：

1. 文档编写
2. 确定归属
3. 将文档路径添加到上述三种文档之一中
4. 执行 mkdocs.sh 脚本

## Mkdocs 技巧分享 {#mkdocs-tips}

### 标记实验性功能 {#experimental}

在一些新发布的功能中，如果是实验性的功能，可以在章节中加入特殊的标记，比如：

```markdown
## 这是一个新功能 {#ref-to-new-feature}

[:octicons-beaker-24: Experimental](index.md#experimental)

新功能正文描述...
```

其效果就是会在章节的下面增加一个这样的图例：

[:octicons-beaker-24: Experimental](index.md#experimental)

点击该图例，就会跳转到实验性功能的说明。

### 标记功能的版本信息 {#version}

某些新功能的发布，是在特定版本中才有的，这种情况下，我们可以添加一些版本标识，其做法如下：

```markdown
## 这是一个新功能 {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)
```

如果恰好这还是个实验性功能，可以将它们排列在一起，用 `·` 分割：

```markdown
## 这是一个新功能 {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)
```

此处，我们以 DataKit 1.4.6 的 changelog 为例，点击对应的图标，即可跳转到对应的版本发布历史：

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

### 外链跳转 {#outer-linkers}

部分文档中，我们需要增加一些外链说明，最好对外链做一些处理，使得其新开一个浏览器 tab，而不是直接跳出当前文档库：

```markdown
[请参考这里](https://some-outer-links.com){:target="_blank"}
```

### 预置章节链接 {#set-links}

我们可以在文档的章节处预先定义其链接，比如：

```markdown
// some-doc.md
## 这是一个新的章节 {#new-feature}
```

那么在其他地方，我们就能直接引用到这里：

```markdown
请参考这个[新功能](some-doc.md#new-feature)
```

如果是在文档内引用，也**必须加上当前文档名字**，原因参见[后面的 404 检测](mkdocs-howto.md#404-check)说明：

```markdown
请参考这个[新功能](current.md#new-feature)
```

如果跨文档库引用：

```markdown
请参考集成库中的这个[新功能](../integrations/some-doc.md#new-feature)
```

### 在文档中增加注意事项 {#note}

部分文档的编写，需提供一些警告信息，比如某功能的使用，需额外满足某些条件，或者给出一些技巧性的说明。这种情况下，我们可以使用 Mkdocs
的 markdown 扩展，比如 

```markdown
??? attention

    这里是一段前置条件的说明...
```

```markdown
??? tip

    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor massa, nec semper lorem quam in massa.
```

而不仅仅只是一个简单的说明:

```
> 这里是一个简陋的说明...
```

更多漂亮的警示用法，参见[这里](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}

### Tab 排版 {#tab}

某些具体的功能，在不同的场景下其使用方式可能不同，一般的做法是在文档中分别罗列，这样会开起来文档冗长，一种更好的方式是将不同场景的使用以 tag 排版的方式组织一下，这样文档页面会非常简洁：

=== "A 情况下这么使用"
     
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor massa, nec semper lorem quam in massa.

=== "B 情况下这么使用"

    .assam ni mauq merol repmes cen ,assam rotcua surup otsuj ,subinif tauqesnoc non rotrot ,taiguef rutibaruC .allun domsiue te alluN .tile gnicsipida rutetcesnoc ,tema tis rolod muspi meroL

具体用法，参见[这里](https://squidfunk.github.io/mkdocs-material/reference/content-tabs/){:target="_blank"}

### 404 链接检查 {#404-check}

在日常的文档编写过程中，我们一般会做如下几类文档链接：

- 链接文档库内的其他文档，其形式为：`这是一段带[内部文档链接](some-other.md#some-section)的文本`
- 链接外站，其形式为：`这是一段带[外站链接](https://host.com#some-section){:target="_blank"}的文本`
- 引用当前文档的其它章节，形如：`请参见[前一章节](#prev-section)的描述`，或者 `请参见[前一章节](current.md#prev-section)的描述`

为了避免 404 检测程序误报，需遵循如下规范：

- 站内链接技术上可以有两种形式，一种形如 `[xxx](datakit/datakit-conf/#config-http-server)`，一种形如 `[xxx](datakit-conf.md#config-http-server)`，这两种写法，在页面上都能正常跳转，但**前者不能通过 404 检测**，请使用第二种形式。
- 所有引用当前文档章节的链接，链接中必须带当前文档名，比如 `请参见[前一章节](current.md#prev-section)的描述`，不能只有章节名。只有章节名会被视为非法的连接。
- 链接的形式必须准确，不能：
    - 带有无意义的多余空格，如 `请参见这个[非法链接]( https://invalid-link.com)`
    - 多余的 `#`，如 `请参见这个[非法链接](https://invalid-link.com/#)`
- 如果普通文本中带有链接说明，需用代码字体来排版这个链接，不然会触发 404 误报。比如：`` 请将主机地址设置为 `http://localhost:8080` ``，文中的这个 localhost 链接用代码字体修饰后，就不会触发 404 误报了。

## 更多阅读 {#more-reading}

- [Material for  Mkdocs](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}
