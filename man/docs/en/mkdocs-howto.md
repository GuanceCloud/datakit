# Documentation Guide
---

This article mainly expounds the following problems:

- How to create Datakit's documents
- How to write better documents under Mkdocs

## DataKit Related Writing Steps {#steps}

We can following these steps to create new Datakit documents:

1. Create markdown under *man/docs/zh*, if the new document related to inputs(collector), create the new documents under *man/docs/zh/inputs*
1. Write and review the documents
1. Add translated version of the new document under  *man/docs/en*(or *man/docs/en/inputs*)
1. Run *mkdocs.sh* under project root path.

### Debugging your local documents {#debug}

We can show more flags that "mkdocs.sh" available:

```shell
./mkdocs.sh -h
```

*mkdocs.sh* depends on some settings, following these steps to setup them:

1. Clone our [big documents repository](https://gitlab.jiagouyun.com/zy-docs/dataflux-doc){:target="_blank"} to your local *~/git/dataflux-doc*. Here *./mkdocs.sh* use these path default to export documents
1. Under *dataflux-doc*, there was a *requirements.txt*, we can run `pip install -r requirements.txt` to setup various dependencies
1. Back to Datakit repository, run `./mkdocs.sh`

## Mkdocs Tips Sharing {#mkdocs-tips}

### Tag Experimental Functionality {#experimental}

In some newly released functions, if they are experimental functions, special tags can be added in the chapters, such as:

```markdown
## This is a new feature {#ref-to-new-feature}

[:octicons-beaker-24: Experimental](index.md#experimental)

New feature text description...
```

The effect is to add such a legend below the chapter:

[:octicons-beaker-24: Experimental](index.md#experimental)

Clicking on the legend will jump to the description of the experimental function.

### Tags Feature Version Information {#version}

Some new features are released in a specific version. In this case, we can add some version logos as follows:

```markdown
## This is a new feature {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)
```

If it happens that this is still an experimental function, you can arrange them together and divide them with `·`:

```markdown
## This is a new feature {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)
```

Here, we take changelog of DataKit 1.4. 6 as an example, and click the corresponding icon to jump to the corresponding version release history:

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

### Outer-linkers {#outer-linkers}

In some documents, we need to add some instructions on the external chain, and it is best to do some processing on the external chain, so that it can open a new browser tab instead of jumping out of the current document library directly:

```markdown
[See here](https://some-outer-links.com){:target="_blank"}
```

### Preset Chapter Links {#set-links}

We can predefine its links at sections of the document, such as:

```markdown
// some-doc.md
## This is a new chapter {#new-feature}
```

Then in other places, we can quote directly here:

```markdown
Please refer to this [new function](some-doc.md#new-feature)
```

If referenced within a document:

```markdown
Please refer to this [new function](#new-feature)
```

If cross-document library references:

```markdown
Refer to this [new function](../integrations/some-doc.md#new-feature) in the integration library 
```

### Add a Note to the Document {#note}

The preparation of some documents, need to provide some warning information, such as the use of a certain function, need to meet some additional conditions, or give some technical instructions. In this case, we can use markdown extension of Mkdocs, such as

```markdown
??? attention

    Here's a precondition note...
```

```markdown
??? tip

    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor massa, nec semper lorem quam in massa.
```

This is not just a simple explanation:

```
> Here is a simple explanation...
```

For more flexible warning usage, see [here](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}

### Tab Typesetting {#tab}

Some specific functions may be used differently in different scenarios. The general practice is to list them separately in the document, which will make the document lengthy. A better way is to organize the use of different scenarios in the way of tag typesetting, so that the document page will be very concise:

=== "A 情况下这么使用"
     
    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor massa, nec semper lorem quam in massa.

=== "B 情况下这么使用"

    .assam ni mauq merol repmes cen ,assam rotcua surup otsuj ,subinif tauqesnoc non rotrot ,taiguef rutibaruC .allun domsiue te alluN .tile gnicsipida rutetcesnoc ,tema tis rolod muspi meroL

For specific usage, see [here](https://squidfunk.github.io/mkdocs-material/reference/content-tabs/){:target="_blank"}

## More Readings {#more-reading}

- [Material for  Mkdocs](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}
