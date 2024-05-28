# Managing DataKit Configuration with Git
---

This article explains how to use Git to manage DataKit configurations, including collection configurations and Pipeline scripts. By maintaining a local or remote Git repository, you can manage changes to DataKit's configurations while leveraging Git's version control features to track historical changes.

## Operating Mechanism {#mechanism}

DataKit integrates Git client functionality, regularly (by default every 1 minute) pulling the latest configuration data from the Git repository. By loading these up-to-date configurations, DataKit achieves configuration updates.

## Usage Example {#example}

The complete steps for the usage example are as follows:

1. Create a Git repository
2. Plan the repository's configuration according to established directory rules
3. Push the configuration to the Git repository
4. Add the Git repository to the main DataKit configuration
5. Restart DataKit

<!-- markdownlint-disable MD046 -->
???+ note

    The Git repository does not have to be created in this order. For example, you can first create a remote repository address and then clone it to make changes. The following example creates a local Git repository first, then pushes it to a remote repository.
<!-- markdownlint-enable -->

### Create a Git Repository {#new-repo}

First, create a local Git repository:

```shell
mkdir datakit-repo
git init
```

### Directory Planning {#dir-naming}

Create various [basic directories](git-config-how-to.md#repo-dirs):

```shell
mkdir -p conf.d   && touch conf.d/.gitkeep
mkdir -p pipeline && touch pipeline/.gitkeep
mkdir -p python.d && touch python.d/.gitkeep
```

### Push Configuration {#repo-push}

Use common Git commands to push configuration changes to the repository:

```shell
# cd your/path/to/repo
git add conf.d pipeline python.d

# Add any conf or pipeline to path conf.d/pipeline/python.d...

git commit -m "init datakit repo"

# Push the repo to YOUR GitHub (ssh or https)
git remote add origin ssh://git@github.com/PATH/TO/datakit-confs.git
git push origin --all
```

### Configure the Repository in DataKit {#config-git-repo}

Enable the *git_repos* feature in the *datakit.conf* configuration file, locate `git_repos`, as shown below:

```toml
[[git_repos.repo]]
    enable = true # Enable the repo

    ###########################################
    # Git support http/git/ssh authentication
    ###########################################
    url = "http://username:password@github.com/PATH/TO/datakit-confs.git" 

    branch = "master" # Specify which branch to pull

    # git/ssh authentication requires key-path key-password configuration
    # url = "git@github.com:PATH/TO/datakit-confs.git"
    # url = "ssh://git@github.com/PATH/TO/datakit-confs.git"
    # ssh_private_key_path = "/Users/username/.ssh/id_rsa"
    # ssh_private_key_password = "<YOUR-PASSWORD>"
```

If the password contains special characters, refer to [here](datakit-input-conf.md#password-encode).

### Restart DataKit {#restart}

After the configuration is complete, [restart Datakit](datakit-service-how-to.md#manage-service). After a short wait, you can check the status of the collectors through [Datakit Monitor](datakit-monitor.md).

## Git Usage in Kubernetes {#k8s}

Refer to [here](datakit-daemonset-deploy.md#env-git).

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: Error: authentication required {#auth-required}
<!-- markdownlint-enable -->

This error may occur in the following situations.

If using SSH, it is generally because the provided key is incorrect. If using HTTP, it may be due to:

1. Incorrect username and password provided
2. The protocol for the git address is filled in incorrectly

For example, the original address is

```not-set
https://username:password@github.com/path/to/repository.git 
```

And it was written as

```not-set
http://username:password@github.com/path/to/repository.git 
```

That is, `https` was changed to `http`, which will also result in this error. Change `http` to `https` here to resolve it.

<!-- markdownlint-disable MD013 -->
### :material-chat-question: Repository Directory Constraints {#repo-dirs}
<!-- markdownlint-enable -->

The Git repository must be stored with the following directory structure for various configurations:

```shell
+── conf.d    # 
├── pipeline  # Dedicated to storing pipeline scripts
└── python.d  # Store python scripts
```

Among them:

- *conf.d* is dedicated to storing collector configurations, and its subdirectories can be planned arbitrarily (subdirectories are allowed), any collector configuration file just needs to end with `.conf`
- *pipeline* is used to store Pipeline scripts, and it is recommended to plan Pipeline scripts according to [data type](../developers/pipeline/pipeline-category.md#store-and-index)
- *python.d* is used to store Python scripts

Here is an example of DataKit's directory structure after Git synchronization is enabled:

```shell
DataKit root directory
├── conf.d   # Default main configuration directory
├── pipeline # Top-level Pipeline scripts
├── python.d # Top-level python scripts
└── gitrepos
    ├── repo-1        # Repository 1
    │   ├── conf.d    # Dedicated to storing collector configurations
    │   ├── pipeline  # Dedicated to storing pipeline scripts
    │   └── python.d  # Store python scripts
    └── repo-2        # Repository 2
        ├── ...
```

<!-- markdownlint-disable MD013 -->
### :material-chat-question: Git Configuration Loading Mechanism {#repo-apply-rules}
<!-- markdownlint-enable -->

After Git synchronization is enabled, the configuration (*.conf/pipeline*) priority is defined as follows:

1. All collector configurations are loaded from the *gitrepos* directory
1. The order of Git repository loading is based on the order in which they appear in *datakit.conf*
1. For Pipeline, the first found Pipeline file is used. As shown in the example, when looking for *nginx.p*, if found in `repo-1`, it will **not** look in `repo-2`. If neither of these repositories has *nginx.p*, then look in the top-level Pipeline directory. The search mechanism for Python is the same.

<!-- markdownlint-disable MD046 -->
???+ attention

    After enabling the remote Pipeline feature, the first loaded Pipeline is the one synchronized from the center.

    After enabling Git synchronization, the original collector configurations in the *conf.d* directory will no longer be effective. In addition, the main configuration *datakit.conf* **cannot** be managed through Git.
<!-- markdownlint-enable -->
