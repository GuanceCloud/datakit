# 使用 Git 管理配置
---

本文介绍如何使用 Git 来管理 DataKit 的配置，这些配置包括采集配置、Pipeline 脚本等。我们可以通过维护一个本地或远程的 Git 仓库来管理 DataKit 的配置变更，同时借用 Git 的版本管理功能，用以追踪配置的历史变更。


## 运行机制 {#mechanism}

DataKit 集成了 Git 客户端功能，它会定期（默认 1min）拉取 Git 仓库里最新的配置数据，通过加载这些最新的配置来实现 DataKit 的配置更新。

## 使用示例 {#example}

完整的使用示例步骤如下：

1. 创建 Git 仓库
1. 按照既定的目录规则，规划仓库中的配置
1. 将配置推送到 Git 仓库
1. 在 DataKit 主配置中添加 Git 仓库
1. 重启 DataKit

???+ note

    Git 仓库的创建不必以这个顺序。比如可以先创建远程仓库地址，然后将该仓库 clone 下来进行更改。以下示例是先创建本地 Git 仓库，然后再将其推送到远程仓库中。

### 创建 Git 仓库 {#new-repo}

先在本地创建一个 Git 仓库：

```shell
mkdir datakit-repo
git init
```

### 目录规划 {#dir-naming}

创建各种[基本目录](git-config-how-to.md#repo-dirs)：

```shell
mkdir -p conf.d   && touch conf.d/.gitkeep
mkdir -p pipeline && touch pipeline/.gitkeep
mkdir -p python.d && touch python.d/.gitkeep
```

### 推送配置 {#repo-push}

通过常用的 Git 命令将配置变更推送到仓库：

```shell
# cd your/path/to/repo
git add conf.d pipeline python.d

# Add any conf or pipeline to path conf.d/pipeline/python.d...

git commit -m "init datakit repo"

# Push the repo to YOUR GitHub(ssh or https)
git remote add origin ssh://git@github.com/PATH/TO/datakit-confs.git
git push origin --all
```

### 在 DataKit 上配置仓库 {#config-git-repo}

在 *datakit.conf* 中开启 gitrepos 功能，找到 `git_repos`，如下所示:

```toml
[[git_repos.repo]]
	enable = true # Enable the repo

	###########################################
	# Git support http/git/ssh authentication
	###########################################
	url = "http://username:password@github.com/PATH/TO/datakit-confs.git"

	branch = "master" # Specify which branch to pull

	# git/ssh authentication require key-path key-password configure
	# url = "git@github.com:PATH/TO/datakit-confs.git"
	# url = "ssh://git@github.com/PATH/TO/datakit-confs.git"
	# ssh_private_key_path = "/Users/username/.ssh/id_rsa"
	# ssh_private_key_password = "<YOUR-PASSSWORD>"
```

### 重启 DataKit {#restart}

配置完成后，[重启 datakit](datakit-service-how-to.md#manage-service) 即可。稍等片刻后，通过 [DataKit Monitor](datakit-monitor.md) 即可查看采集器的开启和运行情况。

## Kubernates 中的 Git 使用 {#k8s}

参见[这里](datakit-daemonset-deploy.md#env-git)。

## FAQ {#faq}

### 报错: authentication required {#auth-required}

出现这个报错可能是以下几种情况。

如果用的是 SSH 方式，一般是因为提供的密钥有错。如果用的是 HTTP 方式，则可能因为：

1. 提供的用户名和密码有错
1. git 地址的协议填错了

比如说, 原地址是

```
https://username:password@github.com/path/to/repository.git
```

然后被写成了

```
http://username:password@github.com/path/to/repository.git
```

即把 `https` 改成了 `http`, 则也会报出这个错误。此处将 `http` 改成 `https` 即可。

### 仓库目录约束 {#repo-dirs}

Git 仓库中必须以如下目录结构来存放各种配置：

```
+── conf.d    # 
├── pipeline  # 专门存放 pipeline 切割脚本
└── python.d  # 存放 python.d 脚本
```

其中

- *conf.d* 专门存放采集器配置，其下的子目录可以任意规划（可以有子目录），任何采集器配置文件，只需要以 `.conf` 结尾即可
- *pipeline* 用来放 pipeline 脚本，pipeline 脚本建议以[数据类型来做规划](datakit-pl-global.md#loading)
- *python.d* 用来放 python 脚本

以下是开启 Git 同步后 DataKit 目录结构示例：

```
datakit 根目录
├── conf.d   # 默认主配置目录
├── pipeline # 顶层 Pipeline 脚本
├── python.d # 顶层 python.d 脚本
└── gitrepos
    ├── repo-1        # 仓库 1
    │   ├── conf.d    # 专门存放采集器配置
    │   ├── pipeline  # 专门存放 pipeline 切割脚本
    │   └── python.d  # 存放 python.d 脚本
    └── repo-2        # 仓库 2
        ├── ...
```

### Git 配置的加载机制 {#repo-apply-rules}

Git 同步开启后，配置（conf/pipeline）优先级定义如下：

1. 所有采集器的配置，都会从 gitrepos 目录下加载
1. Git 仓库加载顺序以其在 *datakit.conf* 中出现的次序为准
1. 对 Pipeline 而言，以第一个找到的 Pipeline 文件为准。以上例所示，查找 *nginx.p* 时，如果在 `repo-1` 中找到了，则 **不会** 再去 `repo-2` 中查找。当这两个仓库都没找到 *nginx.p* 时，才去顶层目录的 pipeline 目录查找。Pythond 的查找机制也一样。

???+ note

    开启远程 Pipeline 功能后，最先加载的是从中心同步下来的 Pipeline。

???+ attention

    开启 Git 同步后，原 *conf.d* 目录下的采集器配置将不再生效。另外，主配置 *datakit.conf* **不能** 通过 Git 来管理。
