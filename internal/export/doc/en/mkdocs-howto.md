# MkDocs Document Writing
---

This article mainly addresses the following issues:

- The steps for writing DataKit-related documents
- How to write better documents using MkDocs

## Steps for Writing DataKit-related Documents {#steps}

The steps for writing new documents are as follows:

1. Add the document under *man/docs/zh*. If it is a collector document, add it to the *man/docs/zh/inputs* directory.
1. Write the document.
1. If necessary, add the corresponding English document under *man/docs/en*.
1. Execute the *export.sh* script in the project root directory.

### Local Debugging of Documents {#debug}

When executing *export.sh*, you can first check the supported command-line parameters:

```shell
./export.sh -h
```

The basic environment on which *export.sh* depends:

1. Clone the [document library](https://gitlab.jiagouyun.com/zy-docs/dataflux-doc){:target="_blank"} to the local directory *~/git/dataflux-doc*. This local directory is the default. *export.sh* will generate and copy the DataKit documents to the corresponding directory of this repo.
1. In the *dataflux-doc* project, there is a *requirements.txt*. Execute `pip install -r requirements.txt` to install the corresponding dependencies.
1. Return to the DataKit code directory and execute `./export.sh` in the root directory.

## MkDocs Tips {#mkdocs-tips}

### Marking Experimental Features {#experimental}

For some newly released features, if they are experimental, you can add a special mark to the section. For example:

```markdown
## This is a New Feature {#ref-to-new-feature}

[:octicons-beaker-24: Experimental](index.md#experimental)

Description of the new feature...
```

The effect is that such a legend will be added below the section:

[:octicons-beaker-24: Experimental](index.md#experimental)

Clicking on this legend will jump to the description of the experimental feature.

### Marking Version Information of Features {#version}

For the release of some new features, they are only available in specific versions. In this case, you can add some version identifiers. The method is as follows:

```markdown
## This is a New Feature {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)
```

If it is also an experimental feature, you can arrange them together and separate them with `·`:

```markdown
## This is a New Feature {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)
```

Here, taking the changelog of DataKit 1.4.6 as an example, clicking on the corresponding icon will jump to the release history of the corresponding version:

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) · [:octicons-beaker-24: Experimental](index.md#experimental)

### External Link Jumps {#outer-linkers}

In some documents, we need to add some external link descriptions. It is best to process the external links so that they open a new browser tab instead of directly jumping out of the current document:

```markdown
[Please refer to here](https://some-outer-links.com){:target="_blank"}
```

### Pre-set Section Links {#set-links}

We can pre-define the link for a section in the document. For example:

```markdown
// some-doc.md
## This is a New Section {#new-feature}
```

Then, in other places, we can directly reference it:

```markdown
Please refer to this [new feature](some-doc.md#new-feature)
```

If it is a reference within the document, we **must add the name of the current document**. The reason is explained in the [404 detection below](mkdocs-howto.md#404-check):

```markdown
Please refer to this [new feature](current.md#new-feature)
```

If it is a cross-document library reference:

```markdown
Please refer to this [new feature](../integrations/some-doc.md#new-feature)
```

### Adding Notes in Documents {#note}

When writing some documents, we need to provide some warning information. For example, for the use of a certain function, some additional conditions need to be met, or some tips need to be provided. In this case, we can use the Markdown extensions of MKDocs. For example:

```markdown
??? warning

    This is an explanation of the preconditions...
```

```markdown
??? tip

    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor massa, nec semper lorem quam in massa.
```

Rather than just a simple explanation:

```markdown
> This is a simple explanation...
```

For more beautiful warning usage, refer to [here](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}

### Tab Layout {#tab}
For some specific functions, their usage may be different in different scenarios. The general practice is to list them separately in the document, which will make the document look long. A better way is to organize the usage in different scenarios in a tag layout, which will make the document page very concise:

<!-- markdownlint-disable MD046 -->
=== "Usage in Case A"

    In case A...

=== "Usage in Case B"

    In case B...
<!-- markdownlint-enable MD046 -->

For specific usage, refer to [here](https://squidfunk.github.io/mkdocs-material/reference/content-tabs/){:target="_blank"}

### Markdown Format Check and Spelling Check {#mdlint-cspell}

To standardize the basic writing of Markdown and maintain consistent spelling in technical documents (relatively correct and consistent), DataKit documents have added layout checks and spelling checks, which are detected by the following two tools:

- [markdownlint](https://github.com/igorshubovych/markdownlint-cli){:target="_blank"}: Check whether the basic Markdown layout conforms to the existing recognized standards.
- [cspell](https://cspell.org/){:target="_blank"}: Check whether the word spelling is correct and enforce a unified spelling for some proprietary words.

#### Format Check {#mdlint}

Since the Markdown format of MkDocs introduces many extended functions, it breaks the established standards of standard Markdown, resulting in some false positives in the checks of MkDocs by markdownlint. In the following way, we can suppress specific check items for specified text blocks.

For example, the following document is to display the Tab style document. However, the standard Markdown considers the indentation here as a code block, and since the language type of the code is not specified, it will trigger the [MD046](https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md#md046---code-block-style){:target="_blank"} check item to report an error. But we can suppress this check item by adding corresponding format comments at the beginning and end:

```markdown
<!-- markdownlint-disable MD046 -->
=== "Host Deployment"

    Enter the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. The example is as follows:

    ...
=== "Kubernetes"

    In Kubernetes, configuration parameters can be modified in the form of environment variables:
    ...
<!-- markdownlint-enable MD046 -->
```

If you want to suppress multiple related checks, the format is as follows (separate the corresponding check items with spaces):

```markdown
<!-- markdownlint-disable MD046 MD047 MD048 -->
...

<!-- markdownlint-enable MD046 MD047 MD048 -->
```

Notes:

- Disable and then enable the same check items to maintain symmetry.
- Remember to enable all the check items at the appropriate position.
- If it is not a false positive and the document indeed violates the corresponding rules and can pass the detection by rewriting the document, then do not suppress the check items and bravely correct the corresponding document.

#### Spelling Check {#cspell}

cspell is very effective in detecting spelling errors of words (here mainly referring to English words, and currently it cannot detect Chinese spelling problems). Sometimes it is difficult for us to avoid misspelling some words, or we sometimes have inconsistent spellings of some standard terms (such as `DataKit/DataKit/datakit` and other various writings).

In the *scripts* directory of the project root directory, the detection settings of cspell are stored. We need to pay special attention to the glossary file *glossary.txt*, which defines several parts such as proprietary nouns and abbreviations.

In the following situations, we need to modify the *glossary.txt* file:

- If there are new proprietary nouns, such as `DataKit`, add it to the proprietary noun list.
- If there are new abbreviations, such as `JDBC`, add it to the abbreviation list.
- If there are compound words, which are relatively rare, add them to the compound word list.
- Pay special attention to extreme words. Some words that are strictly prohibited in the main text (compared to inline code and code blocks), for example, we require that `Java` cannot be written as `java/JAVA`, and `JSON` cannot be written as `Json/json`, etc.
- If the spelling cannot be avoided, in addition to adding it to the corresponding glossary, it can also be formatted in the form of inline code. In the spelling check, we have ignored the checks of text such as code snippets and URL links (refer to the `ignoreRegExpList` configuration in *scripts/cspell.json*).

<!-- markdownlint-disable MD013 -->
#### Detection of Chinese-English Mixed Arrangement {#zh-en-mix}
<!-- markdownlint-enable -->

The Chinese-English mixed arrangement involves two aspects:

- Add an English space between all texts with Chinese-English mixed arrangement (including the mixed arrangement of numbers and Chinese) to relieve reading fatigue.

For example, the following looks more spacious and visually comfortable:

```markdown
We hope to add an English space between English and Chinese...
```

However, there is no need to add a space between Chinese punctuation marks and English (including numbers) because the layout without a space will not make people feel uncomfortable:

```markdown
I write a sentence in English, but it is followed by a Chinese comma...
```

- In all Chinese contexts, use Chinese punctuation instead of English punctuation (for example, `,.:()'!` cannot directly appear before and after Chinese characters).

<!-- markdownlint-disable MD046 -->
??? warning

    All Chinese-English mixed arrangements need to follow this setting, regardless of whether it is code formatting.
<!-- markdownlint-enable MD046 -->

### Font Settings {#font}

The following types of fonts are mainly involved in the writing process:

- Code font: Inline code should be uniformly formatted as `` `this is code` ``.
- Italic font: Italic text should be uniformly formatted as `*this is italic font*`. Although Markdown also supports `_this is italic font_`, for the sake of uniformity, the latter will not be used.
- Bold font: Bold text should be uniformly formatted as `**this is bold font**`. Although Markdown also supports `__this is bold font__`, for the sake of uniformity, the latter will not be used.

### 404 Link Check {#404-check}

During the daily document writing process, we generally make the following types of document links:

- Link to other documents in the document library, in the form of: ` This is a text with an [internal document link](some-other.md#some-section) `.
- Link to an external website, in the form of: ` This is a text with an [external website link](https://host.com#some-section){:target="_blank"} `.
- Refer to other sections of the current document, in the form of: ` Please refer to the description of the [previous section](#prev-section) `, or ` Please refer to the description of the [previous section](current.md#prev-section) `.

To avoid false positives in the 404 detection program, the following specifications need to be followed:

<!-- markdownlint-disable MD038 -->
- There are technically two forms of internal links. One is in the form of `[xxx](datakit/datakit-conf/#config-http-server)`, and the other is in the form of `[xxx](datakit-conf.md#config-http-server)`. Both of these writing methods can jump normally on the page, but **the former cannot pass the 404 detection**, so the second form should be used.
- For all links referencing sections of the current document, the name of the current document must be included in the link. For example, ` Please refer to the description of the [previous section](current.md#prev-section) `. A link with only the section name will be regarded as an illegal link.
- The form of the link must be accurate. There should be no:
    - Meaningless extra spaces, such as ` Please refer to this [illegal link]( https://invalid-link.com)`.
    - Extra `#`, such as ` Please refer to this [illegal link](some.md#invalid-link/#)`.
- If there is a link description in the normal text, the link should be formatted in code font, otherwise it will trigger a 404 false positive. For example: `` Please set the host address to `http://localhost:8080` ``. After the localhost link in the text is decorated with code font, it will not trigger a 404 false positive.
<!-- markdownlint-enable MD038 -->

## Further Reading {#more-reading}

- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}
