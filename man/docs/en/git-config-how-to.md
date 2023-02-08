# Managing Configuration with Git
---

## How Git Works {#intro}

Git is a technology for version control, the same as SVN. For more information, see [here](https://www.runoob.com/git/git-tutorial.html){:target="_blank"}.

Git components are divided into Git Server and Git Client. Running on the remote server is the Git Server, the remote repository. In the local (or Kubernates container). The following words "local" all mean this. ) is running the Git Client, the local copy.

The content managed by Git is divided into local copy and remote warehouse. Changes are submitted locally as a copy during commit operations and to the remote repository only when push operations are performed.

## Create A Git Repository {#new-repo}

You can generally create a Git repository using `New Project` in Github/Gitlab.

After creating the Git repository, you can get an address like `http://github.com/path/to/repository.git`, through which the Git Client pushes or pulls the content.

## Git Operation Flow {#steps}

Generally Git operation flow is roughly as follows:

Step 1: Add the change file. Such as:

```shell
git add clickhouse.conf
```

Step 2: Explain this change and submit it to the local copy (commit operation). Such as:

```shell
git commit -m "Modified the IP address of Exporter"
```

Step 3: Commit the changes to the remote repository (push operation). Such as:

```shell
git push origin master
```

## Directory Requirements for Git Repositories {#dir-naming}

- `gitrepos/repo-name/conf.d` is used to place collector configuration files with unrestricted subdirectories (`datakit.conf` is not managed by `gitrepos`)
- `gitrepos/repo-name/pipeline` is used to put pipeline scripts, and only `.p` in the first tier of this directory will take effect, and none of its subdirectories will take effect
- `gitrepos/repo-name/python.d` for python scripts

## Submit A conf File and Directory {#commit-conf}

The following is an example of the [clickhouse](clickhousev1.md) collector.

Step 1: Switch to the `/root` directory and use the `git clone http://github.com/path/to/repository.git` command to pull the remote repository locally.

Select the collector you want to open, here is clickhouse. Copy `[Datakit 安装目录]/conf.d/db/clickhousev1.conf.sample` to the `/root/repository` directory above.

Note: All collector configuration file samples are in the `[Datakit 安装目录]/conf.d` directory.

The file name is removed from `.sample`, and the file structure is as follows:

```shell
.
└── repository
    └── conf.d
        └── clickhousev1.conf
```

According to their actual situation, modify the `clickhousev1.conf` configuration, save.

Step 2: Commit changes to the remote repository.

```shell
$ git add clickhousev1.conf              # Add change file
$ git commit -m "new clickhousev1.conf"  # Add change description
$ git push origin master                 # Commit changes to remote repository
```

At this point, the edited `clickhousev1.conf` file has been successfully pushed to the remote repository.

## Configure the Repository on the DataKit {#config-git-repo}

The demonstration here adopts the host mode, which is not suitable for Kubernates environment. Operations in the Kubernates environment are described separately below.

The Git authentication method demonstrated here is user name and password.

Step 1: You need to turn on the gitrepos functionality in `datakit.conf`.

Find `git_repos` in `datakit.conf` to configure as follows:

```toml
[git_repos]
  pull_interval = "1m"  # Pull updates every minute

  [[git_repos.repo]]
    enable = true                                                       # Open to pull this Git branch.
    url = "http://username:password@github.com/path/to/repository.git"  # User name/password authentication is used.
    branch = "master"                                                   # The name of the branch to pull. Usually master.
```

Step 2: After configuration, restart datakit.

```shell
$ sudo datakit service -R
```

Step 3: Observe whether Git has pulled updates and loaded the configuration.

You can observe whether the newly added/modified collector is effective:

```shell
$ sudo datakit monitor -V
```

## Update and Pull Warehouse {#git-pull}

We have a local copy in `/root/repository` above. There we made some modifications to the `clickhousev1.conf` file.

Submit after modification is completed:

```shell
$ git add clickhousev1.conf                 # Add change file
$ git commit -m "modify clickhousev1.conf"  # Add change description
$ git push origin master                    # Commit changes to remote repository
```

After the submission is completed. datakit pulls according to the interval set by `pull_interval` in the configuration, and when the interval expires, it automatically pulls the latest `clickhousev1.conf` and makes it effective.

## Git Uses in Kubernates {#k8s}

Because of the particularity of Kubernates environment, the installation/configuration mode with environment variable passing is the simplest.

The git authentication method is user name and password.

When installing in Kubernates, you need to set the following environment variables to bring Git configuration information into it:

| Environment Variable Name       | Environment Variable Value                                                   |
| ----             | ----                                                         |
| ENV_GIT_URL      | `http://username:password@github.com/path/to/repository.git` |
| ENV_GIT_BRANCH   | `master`                                                     |
| ENV_GIT_INTERVAL | `1m`                                                         |

For more information on the configuration under Datakit's Kubernates environment, see [this document](k8s-config-how-to.md#via-env-config).

## FAQ {#faq}

### Error: Authentication Required {#auth-required}

This error may be reported in the following situations.

If SSH is used:

1. The key provided is wrong;

If you are using HTTP:

1. The user name and password provided are wrong;
2. The protocol of git address is incorrect;
For example, if the original address is `https://username:password@github.com/path/to/repository.git`, and then it is written as `http://username:password@github.com/path/to/repository.git`, that is, if `https` is changed to `http`, this error will also be reported.
