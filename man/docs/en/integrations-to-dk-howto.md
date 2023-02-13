# Integrated Document Merge
---

This document focuses on how to incorporate existing integration documents into datakit's documentation. The existing integration documentation is at [here](https://www.yuque.com/dataflux/integrations){:target="_blank"}.

???+ Attention

    Documents related to datakit integration are not recommended to be modified directly in *dataflux-doc/docs/integrations*, because datakit's own document export is overwritten to this directory, which may cause documents manually added to *dataflux-doc/docs/integrations* to be overwritten.

Noun definition:

- Document library: Refers to the new document library dataflux-doc

There are several possibilities for merging the integration document into the datakit document:

- Merge integration documents: Extend existing collector documents directly, such as [CPU integration documents](https://www.yuque.com/dataflux/integrations/fyiw75){:target="_blank"}, which can be merged directly into the collector's cpu.md (man/manuals/cpu.md)
- Add a datakit document: If there is no corresponding document in the datakit, you need to add a document in the datkit manually

The following will explain how to merge for the above situations.

## Merge Integration Document {#merge}

In the existing datakit documents, most of the contents in the integrated documents are already available, but the main missing information is screenshot information and scene navigation. In addition, the environment configuration and metric information are basically available. Therefore, when merging, you only need to add some screenshot information:

- In the integrated document of the existing language sparrow, get the link address of the screenshot and download it directly from the current integrated document library:

```shell
cd dataflux-doc/docs/integrations
wget http://yuque-img-url.png -O imgs/input-xxx-0.png
wget http://yuque-img-url.png -O imgs/input-xxx-1.png
wget http://yuque-img-url.png -O imgs/input-xxx-2.png
...
```

> Note: Do not download images to the same documentation directory as the datakit project.

For a specific collector, there may be multiple screenshots here. It is recommended to save these pictures with a fixed naming convention, that is, all the pictures are saved in the *imgs* directory of the integrated document library, and each collector-related picture is prefixed with `input-` and named according to the number.

After downloading the picture, add it to the datakit document, as shown in the existing CPU collector sample (man/manuals/cpu.md).

- Compile DataKit

As the document of datakit itself is modified, it needs to be compiled to take effect. datakit compilation, see [here](https://github.com/GuanceCloud/datakit/blob/github-mirror/README.zh_CN.md){:target="_blank"}.

If the compilation process is difficult, you can ignore it for the time being, and directly submit the above changes to the merge request to the datakit repository, which can be compiled by the development side for the time being and finally synchronized to the document library.

## Add datakit Doc {#add}

For integration documents that are not supported by direct collectors in datakit, it will be easier to add them. Let's take resin in the existing integration library as an example to illustrate the above process.

- Get the markdown text from the existing page of the language sparrow and save it to the *man/manuals/* directory

Add markdown directly to the URL of the resin integration page, [visit to get its original markdown text](https://www.yuque.com/dataflux/integrations/resin/markdown){:target="_blank"}, select all copies, and save them to *man/manuals/resin.md*.

After downloading, modify the layout, specifically, remove some unnecessary html decorations (see how the current resin.md is changed), and download all those pictures (as in the CPU example above), save them, and then reference them in the new resin.md.

- Modify the directory structure in *man/manuals/integrations.pages* to add corresponding documents

Because resin is a kind of web server, we put it with nginx/apache in the existing *integrations.pages* file:

```yaml
- 'Web server'
  - 'Nginx': nginx.md
  - apache.md
  - resin.md
```

- Modify mkdocs.sh script

Modify the mkdocs.sh script to add the new document to the export list:

```
cp man/manuals/resin.md $integration_docs_dir/
```

## Document Generation and Export {#export}

In datakit's existing repository, you can implement the two steps of compiling and publishing by directly executing mkdocs.sh. In mkdocs.sh, the document is currently exported directly into two copies, synchronized to the datakit and integrations directories of the document library.

If you want to insert pictures into your document, you can place them in the *imgs* directories of datakit and integrations, respectively. For how to reference pictures, refer to [example above](#merge).

Let's talk about the local operation mode of the document library. The main steps are as follows.

- Clone existing document libraries and install corresponding dependencies

``` shell
cd ~/ && mkdir -p git && cd git
git clone ssh://git@gitlab.jiagouyun.com:40022/zy-docs/dataflux-doc.git
cd dataflux-doc
pip install -r requirements.txt # You may be asked to update the pip version during the period
```

???+ attention

    After mkdocs is installed, you may need to set $PATH, and the setting of Mac may be like this (you can find the binary location of mkdocs under find):
    
    ``` shell
    PATH="/System/Volumes/Data/Users/<user-name>/Library/Python/3.8/bin:$PATH"
    ```
- Familiar with *mkdocs.sh*

There is a mkdocs.sh script in the DataKit root directory, which exports all DataKit documents, copies them to different directories in the document library and finally starts the local document service.

- Visit local http://localhost:8000
- After debugging, submit the Merge Request to the `mkdocs` branch of the datakit project
