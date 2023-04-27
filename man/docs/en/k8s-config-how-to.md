# Overview of DataKit Configuration in Kubernetes Environment
---

In K8s environment, because there may be many configuration modes of collectors, it is easy to confuse the differences between different configuration modes in the process of configuring collectors. This paper briefly introduces the best practices of configuration in K8s environment.

## Configuration Mode in K8s Environment {#intro}

The current version (> 1.2. 0) of DataKit supports the following configurations:

- Configure with [conf](datakit-daemonset-deploy.md#configmap-setting)
- Configure with [ENV](datakit-daemonset-deploy.md#using-k8-env)
- Configure with [Annotation](container-log.md#logging-with-annotation-or-label)
- Configure with [CRD](kubernetes-crd.md)
- Configure with [Git](datakit-conf.md#using-gitrepo)
- Configure with [DCA](dca.md)

If further summarized, it can be divided into two types:

- DataKit-based configuration
	- conf
	- ENV
	- Git
	- DCA
- Configuration based on **collected entity**
	- Annotation

Because there are so many different configuration modes, there are priority relationships among different configuration modes, which are decomposed one by one in the order of priority from low to high.

### Configmap {#via-configmap-conf}

When DataKit runs in the K8s environment, it is actually not much different from running on the host, and it will still read the collector configuration in the _.conf_ directory. Therefore, it is completely feasible to inject collector configuration through ConfigMap, and sometimes it is even the only way. For example, in the current DataKit version, MySQL collector can only be opened by injecting ConfigMap.

### Configure via ENV {#via-env-config}

In K8s, when we start DataKit, we can [inject many environment variables](datakit-daemonset-deploy.md#using-k8-env) into its yaml. In addition to the fact that DataKit's behavior can be intervened by injecting environment variables, some collectors also support injecting **specific environment variables**, which are generally named as follows:

```shell
ENV_INPUT_XXX_YYY
```

Here `XXX` refers to the collector name, and `YYY` is a specific configuration field in the collector configuration, such as `ENV_INPUT_CPU_PERCPU` to adjust whether [CPU collector](cpu.md) _collects metrics per CPU core_ (by default, this option is turned off by default, that is, CPU metrics per core are not collected)

It should be noted that not all collectors support ENV injection at present. The collector that supports ENV injection is generally [the collector that is turned on by default](datakit-input-conf.md#default-enabled-inputs). The collector opened through ConfigMap also supports ENV injection (see if the collector supports it), and **the default is based on ENV injection**.

> Environment variable injection mode, generally only applied in K8s mode, host installation mode cannot inject environment variables at present.

### Configure through Annotation {#annotation}

At present, Annotation configuration is more narrowly supported than ENV. It is mainly used to **mark the collected entity**, such as _whether it is necessary to turn on/off the collection of an entity (including log collection, indicator collection, etc.)_

The scenario of interfering with collector configuration through Annotation is quite special. For example, in the container (Pod) log collector, if collecting all logs is prohibited (in the container collector, `container_exclude_log = [image:*]`), but you only want to turn on log collection for certain Pods, you can append Annotation to certain Pods to mark them:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: testing-log-deployment
  labels:
    app: testing-log
spec:
  template:
    metadata:
      labels:
        app: testing-log
      annotations:
        datakit/logs: |    # <-------- 此处追加特定 Key 的 Annotation
          [
            {
              "source": "testing-source",   # 设置该 Pod 日志的 source
              "service": "testing-service", # 设置该 Pod 日志的 service
              "pipeline": "test.p"          # 设置该 Pod 日志的 Pipeline
            }
          ]
	...
```

> Note: Currently, the Annotation mode does not support the mainstream collector opening (currently only [Prom](prom.md)) is supported). More collectors will be added later.

So far, in DataKit, there are only three mainstream configuration modes in K8s environment, and their priorities are gradually improved, that is, conf mode has the lowest priority, ENV takes the second place, and Annotation mode has the highest priority.

- Configuration via CRD

CRD is a widely used configuration method of Kubernetes. Compared with Annotation, CRD does not need to change the deployment of collected objects, and it is less invasive. See [DataKit CRD Usage Documentation](kubernetes-crd.md).

### Git configuration {#git}

Git mode is supported in both host mode and K8s mode, and is essentially a conf configuration, except that its conf file is not in the default _conf. d_ directory, but in the _gitrepo_ directory of the DataKit installation directory. If Git mode is turned on, the default **collector configuration in the _conf.d_ directory is no longer valid** (except for the main configuration of _datakit.conf_), but the original _pipeline_ directory and _pythond_ directory are still valid. As you can see from this, Git is mainly used to manage various text configurations on DataKit, including various collector configurations, Pipeline scripts and Python scripts.

> Note: The DataKit master configuration (_datakit.conf_) cannot be managed by Git.

#### Configuration of Default Collector in Git Mode {#def-inputs-under-git}

In Git mode, there is a very important feature, that is, **conf files of [default collector](datakit-input-conf.md#default-enabled-inputs) are stealthy**, whether in K8s mode or host mode, so it needs some extra work to manage these collector configuration files with Git, otherwise it will cause them to be **collected repeatedly**.

In Git mode, if you want to adjust the configuration of the default collector (you don't want to turn it on or configure it accordingly), there are several ways:

- You can remove them from _datakit.conf_ or _datakit.yaml_ . **At this time, they are not the collectors turned on by default**.
-	If you want to modify the configuration of a specific collector, there are several ways:
	- Manage their conf through Git
	- Through the ENV injection mentioned above (depending on whether the collector supports ENV injection)
	- If the collector supports Annotation tags, it can also be adjusted in this way.

### DCA Configuration {#dca}

The [DCA](dca.md) configuration is actually a bit like Git, and they all affect the conf/pipeline/python file configuration on the DataKit. Just for DCA, its function is not as powerful as Git, and it is generally only used to manage files on several DataKits in a small scope.

## Summary {#summary}

So far, several configuration modes on DataKit have been basically introduced. Whether the specific collector supports specific configuration modes still needs to refer to the collector documents.

## Extended Reading {#more-readings}

- [DataKit Configuration](datakit-conf.md) 
- [DataKit Collector Configuration](datakit-input-conf.md) 
- [Daemonset Installs DataKit](datakit-daemonset-deploy.md)
