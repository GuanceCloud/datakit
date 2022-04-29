# 如何通过 Git 来管理 DataKit 的配置

## Git 的工作原理

Git 是用于版本控制的一项技术, 同 SVN。更多介绍详见[这里](https://www.runoob.com/git/git-tutorial.html)。

Git 组件分为 Git Server 和 Git Client。在远程服务器上运行的是 Git Server, 即远程仓库。在本地 (或 Kubernates 容器里面。以下说的 "本地" 都是这层意思。) 运行的是 Git Client, 即本地副本。

Git 管理的内容分为本地副本和远程仓库两份。在执行 commit 操作的时候会把改动提交到本地作为副本, 只有当执行 push 操作时才会把改动提交到远程仓库。

## 如何创建一个 Git 仓库?

一般可在 Github/Gitlab 中使用 `New Project` 中即可创建一个 Git 仓库。

创建 Git 仓库后可以获得一个地址，类似 `http://github.com/path/to/repository.git` 这样的，Git Client 通过该地址 push 或 pull 内容。

## Git 的操作流程

一般 Git 的操作流程大致如下:

第 1 步: 添加改动文件。如:

```shell
git add clickhouse.conf
```

第 2 步: 说明此次改动, 并提交到本地副本(commit 操作)。如:

```shell
git commit -m "修改了 Exporter 的 IP 地址"
```

第 3 步: 把改动提交到远程仓库(push 操作)。如:

```shell
git push origin master
```

## Git 仓库的目录要求

- `gitrepos/repo-name/conf.d` 用来放采集器配置文件，其下的子目录不做限制（`datakit.conf` 不在 `gitrepos` 管理）
- `gitrepos/repo-name/pipeline` 用来放 pipeline 脚本，且只有该目录下第一层的 `.p` 才生效，其下的子目录均不生效
- `gitrepos/repo-name/python.d` 用来放 python 脚本

## 如何远程提交一个 conf 文件以及目录?

下面以 [clickhouse](https://www.yuque.com/dataflux/datakit/clickhousev1) 采集器为例进行演示。

第 1 步: 切换到 `/root` 目录下，使用 `git clone http://github.com/path/to/repository.git` 命令拉取远程仓库到本地。

选取想要开启的采集器，这里是 clickhouse。复制 `[Datakit 安装目录]/conf.d/db/clickhousev1.conf.sample` 到上面的 `/root/repository` 目录下。

备注: 所有采集器配置文件样本在 `[Datakit 安装目录]/conf.d` 目录下。

文件名去掉 `.sample`，文件结构如下:

```shell
.
├── repository
│   └── clickhousev1.conf
```

根据自己的实际情况，修改 `clickhousev1.conf` 的各项配置、保存。

第 2 步: 提交改动到远程仓库。

```shell
$ git add clickhousev1.conf              # 添加改动文件
$ git commit -m "new clickhousev1.conf"  # 添加改动说明
$ git push origin master                 # 提交改动到远程仓库
```

至此，已经将编辑好的 `clickhousev1.conf` 文件成功推送到了远程仓库。

## 如何在 dk 上配置该仓库? 

这里演示采用的是宿主机的方式，不适应于 Kubernates 环境。Kubernates 环境下的操作在下面单独介绍。

这里演示采用的 Git 验证方式是用户名和密码方式。

第 1 步: 需要在 `datakit.conf` 中开启 gitrepos 功能。

在 `datakit.conf` 中找到 `git_repos` 进行配置，如下所示:

```toml
[git_repos]
  pull_interval = "1m"  # 每分钟拉一次更新

  [[git_repos.repo]]
    enable = true                                                       # 开启拉取这个 Git 分支。
    url = "http://username:password@github.com/path/to/repository.git"  # 使用 用户名/密码 验证方式。
    branch = "master"                                                   # 要拉取的分支名。一般为 master。
```

第 2 步: 配置完成后，重启 datakit 即可。

```shell
$ sudo datakit --restart
```

第 3 步: 观察 Git 是否已拉取更新并加载配置。

可以通过观察新增/修改的采集器是否生效:

```shell
$ sudo datakit -M --vvv
```

## 更新 Git 仓库, 演示一下 dk 拉取到了新的 conf 并生效

上面我们在 `/root/repository` 里面存有一份本地副本。我们在那里对 `clickhousev1.conf` 文件进行一下修改。

修改完成后进行提交:

```shell
$ git add clickhousev1.conf                 # 添加改动文件
$ git commit -m "modify clickhousev1.conf"  # 添加改动说明
$ git push origin master                    # 提交改动到远程仓库
```

提交完成后。datakit 根据配置里面 `pull_interval` 设定的拉取间隔，间隔时间到了即会自动拉取最新的 `clickhousev1.conf` 并使其生效。

## 在 Kubernates 中, 又如何使用 Git?

由于 Kubernates 环境的特殊性，采用环境变量传递的安装/配置方式最为简单。

git 验证方式采用用户名和密码方式。

在 Kubernates 里面安装的时候需要设置如下的环境变量，把 Git 配置信息带进去:

|  环境变量名   | 环境变量值  |
|  ----  | ----  |
| ENV_GIT_URL  | `http://username:password@github.com/path/to/repository.git` |
| ENV_GIT_BRANCH  | `master` |
| ENV_GIT_INTERVAL  | `1m` |

更多关于 Datakit 的 Kubernates 环境下面的配置可以参见[这个文档](https://www.yuque.com/dataflux/datakit/k8s-config-how-to#40747310)。
