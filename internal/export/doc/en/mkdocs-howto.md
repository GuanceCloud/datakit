# MkDocs documentation writing

---

This paper mainly addresses the following issues:

- Documentation steps related to Datakit
- How to write better documentation with MkDocs

## DataKit related writing steps {#steps}

The steps to write a new document are:

1. Add documents under *man/docs/zh*, if it is a collector document, add it to *man/docs/zh/inputs* directory
1. Write your documents
1. If necessary, add the corresponding English document under *man/docs/en*
1. Execute the *export.sh* script in the project root directory

### Document local debugging {#debug}

When executing *export.sh*, you can first look at the command line parameters it supports:

```shell
$ ./export.sh -h
...
```

*export.sh* depends on the basic environment:

1. First clone the [document library](https://gitlab.jiagouyun.com/zy-docs/dataflux-doc){:target="_blank"} to the local directory *~/git/dataflux-doc*, here This local directory is used by default. *export.sh* will generate and copy the Datakit documentation to the corresponding directory of the repo.

1. Under the *dataflux-doc* project, there is a *requirements.txt*, execute `pip install -r requirements.txt` to install the corresponding dependencies

1. Go back to the Datakit code directory and execute `./export.sh` in the root directory

## MkDocs Tips Sharing {#mkdocs-tips}

### Mark experimental features {#experimental}

In some newly released functions, if it is an experimental function, a special mark can be added to the chapter, such as:

```markdown
## This is a new feature {#ref-to-new-feature}

[:octicons-beaker-24: Experimental](index.md#experimental)

New Feature Text Description...
```

The effect is to add a legend like this below the chapter:

[:octicons-beaker-24: Experimental](index.md#experimental)

Clicking on the legend will jump you to a description of the experimental feature.

### Mark the version information of the function {#version}

The release of some new features is only available in a specific version. In this case, we can add some version identifiers. The method is as follows:

```markdown
## This is a new feature {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6)
```

If this happens to be an experimental feature, you can line them up and separate them with `·`:

```markdown
## This is a new feature {#ref-to-new-feature}

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)
```

Here, we take the changelog of DataKit 1.4.6 as an example, click the corresponding icon to jump to the corresponding version release history:

[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

### Outer link jump {#outer-linkers}

In some documents, we need to add some external link instructions. It is best to do some processing on the external link so that it opens a new browser tab instead of directly jumping out of the current document library:

```markdown
[Please refer here](https://some-outer-links.com){:target="_blank"}
```

### Preset chapter links {#set-links}

We can pre-define its links at the chapters of the document, for example:

```markdown
// some-doc.md
## This is a new section {#new-feature}
```

Then in other places, we can directly quote here:

```markdown
Please refer to this [new feature](some-doc.md#new-feature)
```

If it is referenced in the document, it must also add the name of the current document**. For the reason, see [404 detection later](mkdocs-howto.md#404-check) description:

```markdown
Please refer to this [new feature](current.md#new-feature)
```

If referenced across document libraries:

```markdown
Please refer to this [new feature](../integrations/some-doc.md#new-feature) in the integrations repository
```

### Add notes to documentation {#note}

The writing of some documents requires some warning information, such as the use of a certain function, which needs to meet certain conditions, or give some technical instructions. In this case, we can use the markdown extension of MKDocs, for example:

```markdown
??? attention

    Here is a description of the preconditions...
```

```markdown
??? tip

    Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nulla et euismod nulla. Curabitur feugiat, tortor non consequat finibus, justo purus auctor massa, nec semper lorem quam in massa.
```

Rather than just a simple statement:

```markdown
> Here's a rough explanation...
```

For more beautiful alert usage, see [here](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}

### Tab Typesetting {#tab}

Some specific functions may be used in different ways in different scenarios. The general method is to list them separately in the document, which will make the document lengthy. A better way is to use tags in different scenarios. Organize it in such a way that the documentation page will be very concise:

<!-- markdownlint-disable MD046 -->
=== "Use this in case A"

    In case A...

=== "Use this in case B"

    In case B...
<!-- markdownlint-enable -->

For specific usage, see [here](https://squidfunk.github.io/mkdocs-material/reference/content-tabs/){:target="_blank"}

### Markdown format check and spell check {#mdlint-cspell}

In order to standardize the basic writing of Markdown and keep the spelling of technical documents consistent (relatively correct and consistent), Datakit's documents have added typesetting checks and spelling checks, which are detected by the following two tools:

- [markdownlint](https://github.com/igorshubovych/markdownlint-cli){:target="_blank"}: checks whether basic Markdown typography conforms to existing recognized standards
- [cspell](https://cspell.org/){:target="_blank"}: Check whether the word spelling is correct, and make mandatory unification of some proprietary spellings

#### Format checking {#mdlint}

Since the Markdown format of MkDocs introduces a lot of extended functions, it breaks the established standard of standard Markdown, which leads to some false positives in markdownlint's current inspection of MkDocs. Through the following methods, we can block the specified text blocks. check item.

For example, the following documents are for displaying Tab-style documents, but the standard Markdown considers the indentation here to be a code block, but does not specify the language type of the code, which will trigger [MD046](https://github.com/DavidAnson /markdownlint/blob/main/doc/Rules.md#md046---code-block-style){:target="_blank"} The inspection item reports an error, but we can block it by adding comments in the corresponding format at the beginning and end This check item:

```markdown
<!-- markdownlint-disable MD046 -->
=== "Host Deployment"

    Enter the `conf.d/{{.Catalog}}` directory under the DataKit installation directory, copy `{{.InputName}}.conf.sample` and name it `{{.InputName}}.conf`. Examples are as follows:

    ...
=== "Kubernetes"

    Kubernetes supports modifying configuration parameters in the form of environment variables:

    | Environment variable name | Corresponding configuration parameter item | Parameter example |
    | :--- | --- | --- |
    | `ENV_INPUT_CPU_PERCPU` | `percpu` | `true/false` |
    ...
<!-- markdownlint-enable -->
```

If you want to block multiple related checks, the format is as follows (separate the corresponding detection items with spaces):

```markdown
<!-- markdownlint-disable MD046 MD047 MD048 -->
...
<!-- markdownlint-enable -->
```

Precautions:

- Be sure to remember to enable all detection items in the appropriate position
- If it is not a false positive, but because the document does violate the corresponding rules, and the document can pass the test by rewriting the document, then do not block the detection item, and bravely correct the corresponding document

#### Spell Check {#cspell}

CSpell is very effective in detecting misspellings of words (here mainly refers to English words, and Chinese spelling problems cannot be detected at present). Sometimes it is difficult for us to avoid spelling some words wrong, or we sometimes have inconsistencies in the spelling of some standard terms Situation (such as `Datakit/DataKit/datakit` and other ways of writing).

The detection settings of cspell are stored in the *scripts* directory of the project root directory. We need to focus on the glossary file *glossary.txt*, in which we define proper nouns, abbreviations and other parts.

In the following situations, we need to modify the *glossary.txt* file:

- If there is a new proper noun, like `Datakit`, we add it to the list of proper nouns
- If there is a new abbreviation, like `JDBC`, we add it to the list of abbreviations
- If there is a compound word, which is relatively rare, just add it to the compound word
- Pay attention to limit words, some words that are strictly prohibited in the text (relative to inline codes and code blocks), for example, we require that `Java` cannot be written as `java/JAVA`, and `JSON` cannot be written as `Json /json` etc.
- If the spelling is really unavoidable, in addition to adding it to the corresponding vocabulary, it can also be typeset in the form of inline code. In the spell check, we ignore the check of text such as code fragments and URL links (see *scripts `ignoreRegExpList` configuration in /cspell.json*)

#### Mixed Chinese and English detection {#zh-en-mix}

Mixed Chinese and English involves two aspects:

- Add an English space between all Chinese and English mixed (including numbers and Chinese mixed) text to relieve reading fatigue

For example, the following will look more sparse and not visually cramped:

```markdown
We want to add 1 English space between English and Chinese ...
```

However, there is no need to add spaces between Chinese punctuation marks and English (including numbers), because without spaces, this kind of layout will not make people feel uncomfortable:

```markdown
I write an English sentence, but it is followed by a Chinese comma ...
```

- In all Chinese contexts, use Chinese punctuation instead of English punctuation (such as `,.:()'!` cannot appear directly before or after Chinese characters)

<!-- markdownlint-disable MD046 -->
??? warning

    All Chinese and English mixed layouts need to follow this setting, whether it is code layout or not.
<!-- markdownlint-enable -->

### 404 link check {#404-check}

In the daily document writing process, we generally make the following types of document links:

- Links to other documents within the docbase in the form: `This is a section of text with [internal document link](some-other.md#some-section)`

- Links to external sites in the form: `This is a text with [external site link](https://host.com#some-section){:target="_blank"}`

- References to other sections of the current document, such as: `See description of [previous chapter](#prev-section)`, or `See description of [previous chapter](current.md#prev-section) describe `

In order to avoid false positives from the 404 detection program, the following specifications must be followed:

<!-- markdownlint-disable MD038 -->

- Links in the site can have two forms technically, one is in the form of `[xxx](datakit/datakit-conf/#config-http-server)`, and the other is in the form of `[xxx](datakit-conf.md# config-http-server)`, these two writing methods can jump normally on the page, but **the former cannot pass the 404 detection**, please use the second form.

- All links that refer to the current document chapter must have the current document name in the link, such as `see the description of [previous chapter](current.md#prev-section)`, not just the chapter name. Only section names are considered illegal links.

- The form of the link must be exact and not:

    - With meaningless extra spaces like ` see this [invalid link](https://invalid-link.com)`
    - Extra `#` like ` see this [invalid link](some.md#invalid-link/#)`

- If there is a link description in the normal text, the link needs to be formatted with a code font, otherwise a 404 false positive will be triggered. For example: `` Please set the host address to `http://localhost:8080` ``, after the localhost link in this article is decorated with code fonts, it will not trigger 404 false positives.
<!-- markdownlint-enable -->

## More Reading {#more-reading}

- [Material for MkDocs](https://squidfunk.github.io/mkdocs-material/reference/admonitions/){:target="_blank"}
