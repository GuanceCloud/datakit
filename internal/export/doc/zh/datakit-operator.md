# DataKit Operator

---

:material-kubernetes:

---

## 概述和安装 {#datakit-operator-overview-and-install}

DataKit Operator 是 DataKit 在 Kubernetes 编排的联动项目，旨在协助 DataKit 更方便的部署，以及其他诸如验证、注入的功能。

目前 DataKit-Operator 提供以下功能：

- 注入 DDTrace Java SDK 以及对应环境变量信息，参见[文档](datakit-operator.md#datakit-operator-inject-lib)
- 注入 Sidecar logfwd 服务以采集容器内日志，参见[文档](datakit-operator.md#datakit-operator-inject-logfwd)
- 支持 DataKit 采集器的任务选举，参见[文档](election.md#plugins-election)

先决条件：

- 推荐 Kubernetes v1.24.1 及以上版本，且能够访问互联网（下载 yaml 文件并拉取对应镜像）
- 确保启用 `MutatingAdmissionWebhook` 和 `ValidatingAdmissionWebhook` [控制器](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/extensible-admission-controllers/#prerequisites){:target="_blank"}
- 确保启用了 `admissionregistration.k8s.io/v1` API

### 安装步骤 {#datakit-operator-install}

<!-- markdownlint-disable MD046 -->
=== "Deployment"

    下载 [*datakit-operator.yaml*](https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml){:target="_blank"}，步骤如下：
    
    ``` shell
    $ kubectl create namespace datakit
    $ wget https://static.<<<custom_key.brand_main_domain>>>/datakit-operator/datakit-operator.yaml
    $ kubectl apply -f datakit-operator.yaml
    $ kubectl get pod -n datakit
    
    NAME                               READY   STATUS    RESTARTS   AGE
    datakit-operator-f948897fb-5w5nm   1/1     Running   0          15s
    ```

=== "Helm"

    前提条件

    * Kubernetes >= 1.14
    * Helm >= 3.0+

    ```shell
    $ helm install datakit-operator datakit-operator \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        <<<% else -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit --create-namespace
    ```

    查看部署状态：

    ```shell
    $ helm -n datakit list
    ```

    可以通过如下命令来升级：

    ```shell
    $ helm -n datakit get values datakit-operator -a -o yaml > values.yaml
    $ helm upgrade datakit-operator datakit-operator \
        <<<% if custom_key.brand_key == 'guance' -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/datakit-operator \
        <<<% else -%>>>
        --repo https://pubrepo.<<<custom_key.brand_main_domain>>>/chartrepo/truewatch \
        <<<% endif -%>>>
        -n datakit \
        -f values.yaml
    ```

    可以通过如下命令来卸载：

    ```shell
    $ helm uninstall datakit-operator -n datakit
    ```
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ note

    - DataKit-Operator 有严格的程序和 yaml 对应关系，如果使用一份过旧的 yaml 可能无法安装新版 DataKit-Operator，请重新下载最新版 yaml。
    - 如果出现 `InvalidImageName` 报错，可以手动 pull 镜像。
<!-- markdownlint-enable -->

## 配置说明 {#datakit-operator-jsonconfig}

DataKit Operator 配置是 JSON 格式，在 Kubernetes 中单独以 ConfigMap 存放，以环境变量方式加载到容器中。

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本（推荐）"

    从 DataKit-Operator v1.7.0 版本开始，推荐使用 `admission_inject_v2` 配置项。新配置采用数组结构，支持更灵活的配置方式。

    默认配置如下：

    ```json
    {
        "server_listen": "0.0.0.0:9543",
        "log_level":     "info",
        "admission_inject_v2": {
            "ddtraces": [
                {
                    "namespace_selectors": [],
                    "label_selectors":     [],
                    "image":    "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:latest",
                    "language": "java",
                    "envs": {
                        "DD_AGENT_HOST":           "datakit-service.datakit.svc.cluster.local",
                        "DD_TRACE_AGENT_PORT":     "9529",
                        "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc.cluster.local",
                        "DD_JMXFETCH_STATSD_PORT": "8125",
                        "DD_SERVICE":              "{fieldRef:metadata.labels['app']}",
                        "POD_NAME":                "{fieldRef:metadata.name}",
                        "POD_NAMESPACE":           "{fieldRef:metadata.namespace}",
                        "NODE_NAME":               "{fieldRef:spec.nodeName}",
                        "DD_TAGS":                 "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
                    },
                    "resources": {
                        "requests": {
                            "cpu":    "100m",
                            "memory": "64Mi"
                        },
                        "limits": {
                           "cpu":    "200m",
                           "memory": "128Mi"
                         }
                    }
                }
            ],
            "logfwds": [
                {
                    "namespace_selectors": [],
                    "label_selectors":     [],
                    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:1.86.0",
                    "envs": {
                        "LOGFWD_DATAKIT_HOST":              "{fieldRef:status.hostIP}",
                        "LOGFWD_DATAKIT_PORT":              "9533",
                        "LOGFWD_DATAKIT_OPERATOR_ENDPOINT": "datakit-operator.datakit.svc:443",
                        "LOGFWD_GLOBAL_SERVICE":            "{fieldRef:metadata.labels['app']}",
                        "LOGFWD_POD_NAME":                  "{fieldRef:metadata.name}",
                        "LOGFWD_POD_NAMESPACE":             "{fieldRef:metadata.namespace}",
                        "LOGFWD_POD_IP":                    "{fieldRef:status.podIP}"
                    },
                    "resources": {
                        "requests": {
                            "cpu":    "100m",
                            "memory": "128Mi"
                        },
                        "limits": {
                           "cpu":    "200m",
                           "memory": "256Mi"
                        }
                    },
                    "log_configs": "",
                    "log_volume_paths": []
                }
            ],
            "flameshots": [
                {
                    "namespace_selectors": [],
                    "label_selectors":     [],
                    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/flameshot:latest",
                    "envs": {
                        "FLAMESHOT_DATAKIT_ADDR":     "http://datakit-service.datakit:9529/profiling/v1/input",
                        "FLAMESHOT_MONITOR_INTERVAL": "10s",
                        "FLAMESHOT_LOG_LEVEL":        "info",
                        "FLAMESHOT_PROFILING_PATH":   "/flameshot-data",
                        "FLAMESHOT_LOG_PATH":         "/var/log/flameshot.log",
                        "FLAMESHOT_HTTP_LOCAL_IP":    "{fieldRef:status.podIP}",
                        "FLAMESHOT_HTTP_LOCAL_PORT":  "8089"
                    },
                    "resources": {
                        "requests": {
                            "cpu":    "100m",
                            "memory": "128Mi"
                        },
                        "limits": {
                           "cpu":    "200m",
                           "memory": "256Mi"
                        }
                    },
                    "processes": "",
                    "enable_prometheus_annotations": true
                }
            ]
        },
        "admission_mutate": {
            "loggings": [
                {
                    "namespace_selectors": [],
                    "label_selectors":     [],
                    "config": ""
                }
            ]
        }
    }
    ```

    主要配置项说明：

    - `ddtraces`：DDtrace 注入配置数组，目前仅支持 Java 语言的 trace agent
    - `logfwds`：logfwd sidecar 注入配置数组，支持配置多个日志采集规则
    - `flameshots`：Flameshot 注入配置数组（替代原有的 profiler），用于性能分析数据采集

    配置特点：

    - 采用数组结构，支持配置多个相同类型的注入规则
    - `namespace_selectors` 和 `label_selectors` 同时配置时，两者的关系为"且"（必须同时满足）
    - 镜像配置直接在数组中通过 `image` 字段指定，不再使用嵌套的 `images` 对象

=== "DataKit-Operator v1.7.0 之前版本（兼容）"

    <!-- markdownlint-disable MD046 -->
    ???+ attention

        此配置方式在 DataKit-Operator v1.7.0 版本之前使用，v1.7.0 及以后版本推荐使用 `admission_inject_v2` 配置。旧配置在 v1.7.0 版本中仍保持向后兼容。
    <!-- markdownlint-enable -->

    默认配置如下：

    ```json
    {
        "server_listen": "0.0.0.0:9543",
        "log_level":     "info",
        "admission_inject": {
            "ddtrace": {
                "enabled_namespaces":     [],
                "enabled_labelselectors": [],
                "images": {
                    "java_agent_image":   "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:latest"
                },
                "envs": {
                    "DD_AGENT_HOST":           "datakit-service.datakit.svc.cluster.local",
                    "DD_TRACE_AGENT_PORT":     "9529",
                    "DD_JMXFETCH_STATSD_HOST": "datakit-service.datakit.svc.cluster.local",
                    "DD_JMXFETCH_STATSD_PORT": "8125",
                    "DD_SERVICE":              "{fieldRef:metadata.labels['service']}",
                    "POD_NAME":                "{fieldRef:metadata.name}",
                    "POD_NAMESPACE":           "{fieldRef:metadata.namespace}",
                    "NODE_NAME":               "{fieldRef:spec.nodeName}",
                    "DD_TAGS":                 "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
                },
                "resources": {
                    "requests": {
                        "cpu":    "100m",
                        "memory": "64Mi"
                    },
                    "limits": {
                       "cpu":    "200m",
                       "memory": "128Mi"
                     }
                }
            },
            "profiler": {
                "images": {
                    "java_profiler_image":   "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/async-profiler:0.5.0",
                    "python_profiler_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/py-spy:0.1.0",
                    "golang_profiler_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/go-pprof:0.1.0"
                },
                "envs": {
                    "DK_AGENT_HOST":  "datakit-service.datakit.svc.cluster.local",
                    "DK_AGENT_PORT":  "9529",
                    "DK_PROFILE_VERSION":  "1.2.333",
                    "DK_PROFILE_ENV":      "prod",
                    "DK_PROFILE_DURATION": "240",
                    "DK_PROFILE_SCHEDULE": "0 * * * *"
                },
                "resources": {
                    "requests": {
                        "cpu":    "100m",
                        "memory": "64Mi"
                    },
                    "limits": {
                       "cpu":    "500m",
                       "memory": "512Mi"
                     }
                }
            },
            "logfwd": {
                "images": {
                    "logfwd_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:1.86.0"
                },
                "envs": {
                    "LOGFWD_DATAKIT_HOST":              "{fieldRef:status.hostIP}",
                    "LOGFWD_DATAKIT_PORT":              "9533",
                    "LOGFWD_DATAKIT_OPERATOR_ENDPOINT": "datakit-operator.datakit.svc:443",
                    "LOGFWD_GLOBAL_SERVICE":            "{fieldRef:metadata.labels['app']}",
                    "LOGFWD_POD_NAME":                  "{fieldRef:metadata.name}",
                    "LOGFWD_POD_NAMESPACE":             "{fieldRef:metadata.namespace}",
                    "LOGFWD_POD_IP":                    "{fieldRef:status.podIP}"
                },
                "resources": {
                    "requests": {
                        "cpu":    "100m",
                        "memory": "64Mi"
                    },
                    "limits": {
                       "cpu":    "500m",
                       "memory": "512Mi"
                     }
                }
            }
        },
        "admission_mutate": {
            "loggings": [
                {
                    "namespace_selectors": ["test01"],
                    "label_selectors":     ["app=logging"],
                    "config":"[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"storage_index\":\"logging-index\"\"source\":\"logging-tmp\"},{\"disable\":true,\"type\":\"file\",\"path\":\"/var/log/opt/**/*.log\",\"source\":\"logging-var\"}]"
                }
            ]
        }
    }
    ```

    主要配置项是 `ddtrace`、`logfwd` 和 `profiler`，指定注入的镜像和环境变量。此外，ddtrace 还支持根据 `enabled_namespaces` 和 `enabled_selectors` 批量注入，详见后文的"注入方式"。

    <!-- markdownlint-disable MD046 -->
    ???+ note

        在旧配置中，`namespace_selectors` 和 `label_selectors` 同时配置时，两者的关系为"或"（满足任一条件即可）。DataKit-Operator v1.7.0 版本的新配置中，两者的关系改为"且"（必须同时满足）。
    <!-- markdownlint-enable -->
<!-- markdownlint-enable -->

### 指定镜像地址 {#datakit-operator-config-images}

DataKit Operator 主要作用就是注入镜像和环境变量。镜像配置方式在新旧版本中有所不同。

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    在 `admission_inject_v2` 配置中，镜像地址直接在数组项中通过 `image` 字段指定：

    ```json
    {
        "admission_inject_v2": {
            "ddtraces": [
                {
                    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:latest",
                    "language": "java"
                }
            ],
            "logfwds": [
                {
                    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:1.86.0"
                }
            ],
            "flameshots": [
                {
                    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/flameshot:latest"
                }
            ]
        }
    }
    ```

=== "DataKit-Operator v1.7.0 之前版本"

    在 `admission_inject` 配置中，使用 `images` 对象配置镜像地址。`images` 是多个 Key/Value，Key 是固定的，修改 Value 值指定镜像地址。

    例如配置 ddtrace 的 Java Agent 镜像：

    ```json
    {
        "admission_inject": {
            "ddtrace": {
                "images": {
                    "java_agent_image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:latest"
                }
            }
        }
    }
    ```
<!-- markdownlint-enable -->

正常情况下，镜像统一存放在 `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator`，对于一些特殊环境不方便访问此镜像库，可以使用以下方法（以 `dd-lib-java-init` 镜像为例）：

1. 在可以访问 `pubrepo.<<<custom_key.brand_main_domain>>>` 的环境中，pull 镜像 `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:v1.30.1-ext`，并将其转存到自己的镜像库，例如 `inside.image.hub/datakit-operator/dd-lib-java-init:v1.30.1-ext`
1. 修改 JSON 配置中的镜像地址，应用此 yaml
1. 此后 DataKit Operator 会使用的新的 Java Agent 镜像路径

**DataKit Operator 不检查镜像，如果该镜像路径错误，Kubernetes 创建 Pod 会报错。**


### 添加环境变量 {#datakit-operator-config-envs}

所有需要注入的环境变量，都必须在配置文件指定，DataKit Operator 不默认添加任何环境变量。

环境变量配置项是 `envs`，由多个 Key/Value 组成：Key 是固定值；Value 可以是固定值，也可以是占位符，根据实际情况取值。

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    在 `admission_inject_v2` 配置中，环境变量在数组项的 `envs` 字段中配置：

    ```json
    {
        "admission_inject_v2": {
            "ddtraces": [
                {
                    "envs": {
                        "DD_AGENT_HOST":       "datakit-service.datakit.svc",
                        "DD_TRACE_AGENT_PORT": "9529",
                        "testing-env":         "ok"
                    }
                }
            ]
        }
    }
    ```

=== "DataKit-Operator v1.7.0 之前版本"

    在 `admission_inject` 配置中，环境变量在对应对象的 `envs` 字段中配置：

    ```json
    {
        "admission_inject": {
            "ddtrace": {
                "envs": {
                    "DD_AGENT_HOST":       "datakit-service.datakit.svc",
                    "DD_TRACE_AGENT_PORT": "9529",
                    "testing-env":         "ok"
                }
            }
        }
    }
    ```
<!-- markdownlint-enable -->

所有注入 `ddtrace` agent 的容器，都会添加 `envs` 中配置的环境变量。

在 DataKit-Operator v1.4.2 及以后版本，`envs` 支持 Kubernetes Downward API 的 [环境变量取值字段](https://kubernetes.io/zh-cn/docs/concepts/workloads/pods/downward-api/#downwardapi-fieldRef)。现支持以下几种：

- `metadata.name`：Pod 的名称
- `metadata.namespace`： Pod 的命名空间
- `metadata.uid`： Pod 的唯一 ID
- `metadata.annotations['<KEY>']`： Pod 的注解 `<KEY>` 的值（例如：metadata.annotations['myannotation']）
- `metadata.labels['<KEY>']`： Pod 的标签 `<KEY>` 的值（例如：metadata.labels['mylabel']）
- `spec.serviceAccountName`： Pod 的服务账号名称
- `spec.nodeName`： Pod 运行时所处的节点名称
- `status.hostIP`： Pod 所在节点的主 IP 地址
- `status.hostIPs`： 这组 IP 地址是 status.hostIP 的双协议栈版本，第一个 IP 始终与 status.hostIP 相同。 该字段在启用了 PodHostIPs 特性门控后可用。
- `status.podIP`： Pod 的主 IP 地址（通常是其 IPv4 地址）
- `status.podIPs`： 这组 IP 地址是 status.podIP 的双协议栈版本，第一个 IP 始终与 status.podIP 相同。

举个例子，现有一个 Pod 名称是 nginx-123，namespace 是 middleware，要给它注入环境变量 `POD_NAME` 和 `POD_NAMESPACE`，参考以下：

```json
{
    "admission_inject": {
        "ddtrace": {
            "envs": {
                "POD_NAME":      "{fieldRef:metadata.name}",
                "POD_NAMESPACE": "{fieldRef:metadata.namespace}"
            }
        }
    }
}
```

最终在该 Pod 可以看到：

``` shell
$ env | grep POD
POD_NAME=nginx-123
POD_NAMESPACE=middleware
```

<!-- markdownlint-disable MD046 -->
???+ note

    如果该 Value 占位符无法识别，会以纯字符串添加到环境变量。例如 `"POD_NAME": "{fieldRef:metadata.PODNAME}"`，这是错误的写法，在环境变量是 `POD_NAME={fieldRef:metadata.PODNAME}`。
<!-- markdownlint-enable -->

## 注入方式 {#datakit-operator-inject}

DataKit-Operator 支持两种资源输入方式，分别是“全局配置 namespaces 和 selectors”，以及在目标 Pod 添加指定 Annotation。它们的区别如下：

- 全局配置 namespace 和 selector：通过修改 DataKit-Operator config，指定目标 Pod 的 Namespace 和 Selector，如果发现 Pod 符合条件，就执行注入资源。
    - 优点：不需要在目标 Pod 添加 Annotation（但是需要重启目标 Pod）
    - 缺点：范围不够精确，可能存在无效注入

- 在目标 Pod 添加 Annotation：在目标 Pod 添加 Annotation 可以拒绝注入（设置为 `"false"`），但不能单独触发注入。实际的注入需要配置匹配规则和相应的配置字段。
    - 优点：可以通过 Annotation 精确控制是否拒绝注入
    - 缺点：不能单独通过 Annotation 触发注入，仍需配置匹配规则

<!-- markdownlint-disable MD046 -->
???+ note

    截止到 DataKit-Operator v1.5.8，全局配置 namespaces 和 selectors 方式只在注入 DDtrace 生效，对于 logfwd 和 profiler 无效，后者仍需添加 annotation 注入。
<!-- markdownlint-enable -->


<!-- markdownlint-disable MD013 -->
### 全局配置 namespaces 和 selectors 配置 {#datakit-operator-config-ddtrace-enabled}
<!-- markdownlint-enable -->

通过配置 `namespace_selectors` 和 `label_selectors` 可以实现批量注入。配置方式在新旧版本中有所不同，且选择器之间的关系也有变化。

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    在 `admission_inject_v2` 配置中，`namespace_selectors` 和 `label_selectors` 直接在数组项中配置：

    ```json
    {
        "admission_inject_v2": {
            "ddtraces": [
                {
                    "namespace_selectors": ["testns"],
                    "label_selectors":     ["app=log-output"],
                    "language": "java",
                    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:latest"
                }
            ]
        }
    }
    ```

    ???+ important

        **重要变更**：从 DataKit-Operator v1.7.0 版本开始，`namespace_selectors` 和 `label_selectors` 同时配置时，两者的关系由"或"改为"且"（必须同时满足）。如果只配置其中一个，则按该选择器匹配。

    - `namespace_selectors`：命名空间选择器数组，支持正则表达式匹配。如需精确匹配，请使用 `^` 和 `$` 将模式包围，例如 `^testns$`
    - `label_selectors`：标签选择器数组，使用 Kubernetes Label Selector 语法
    - `language`：指定需要注入的 agent 语言，例如 `java`

=== "DataKit-Operator v1.7.0 之前版本"

    在 `admission_inject` 配置中，使用 `enabled_namespaces` 和 `enabled_labelselectors` 配置：

    ```json
    {
        "admission_inject": {
            "ddtrace": {
                "enabled_namespaces": [
                    {
                        "namespace": "testns",  # 指定 namespace，支持正则表达式匹配。如需精确匹配，请使用 ^ 和 $ 将模式包围，例如 ^testns$
                        "language": "java"      # 指定需要注入的 agent 语言
                    }
                ],
                "enabled_labelselectors": [
                    {
                        "labelselector": "app=log-output",  # 指定 labelselector
                        "language": "java"                  # 指定需要注入的 agent 语言
                    }
                ]
            }
        }
    }
    ```

    ???+ note

        在旧版本中，`enabled_namespaces` 和 `enabled_labelselectors` 数组之间是"或"的关系。如果一个 Pod 既满足 `enabled_namespaces` 规则，又满足 `enabled_labelselectors`，以 `enabled_labelselectors` 配置为准（通常在 `language` 取值用到）。
<!-- markdownlint-enable -->

关于 labelselector 的编写规范，可参考此[官方文档](https://kubernetes.io/zh-cn/docs/concepts/overview/working-with-objects/labels/#label-selectors){:target="_blank"}。

<!-- markdownlint-disable MD046 -->
???+ note

    - 在 Kubernetes 1.16.9 或更早版本，Admission 不记录 Pod Namespace，所以无法使用 `enabled_namespaces` 功能。
<!-- markdownlint-enable -->

### 添加 Annotation 配置注入 {#datakit-operator-config-annotation}

<!-- markdownlint-disable MD046 -->
???+ important

    **重要说明**：Annotation 只能用于拒绝注入（设置为 `"false"`），不能单独触发注入。实际的注入需要：
    1. 在 DataKit-Operator 配置中设置匹配规则（`namespace_selectors`/`label_selectors`）和相应的配置字段
    2. Pod 匹配配置中的选择器规则
    3. Annotation 不为 `"false"`（如果设置为 `"false"` 则会拒绝注入）
<!-- markdownlint-enable -->

在 Deployment 添加指定 Annotation 可以控制是否允许注入。注意 Annotation 要添加在 template 中。

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    ???+ attention

        从 DataKit-Operator v1.7.0 版本开始，移除了对 `admission.datakit/java-lib.version` Annotation 的支持。如需禁用 ddtrace 注入，可以使用 `admission.datakit/ddtrace.enabled:"false"`。

    支持的 Annotation：

    - `admission.datakit/ddtrace.enabled`：控制是否允许注入 ddtrace，值为 `"false"` 时拒绝注入，值为 `"true"` 或不设置时允许注入（但需配置匹配规则才能实际触发注入）
    - `admission.datakit/logfwd.enabled`：控制是否允许注入 logfwd sidecar，值为 `"false"` 时拒绝注入，值为 `"true"` 或不设置时允许注入（但需配置匹配规则和 `log_configs` 字段才能实际触发注入）
    - `admission.datakit/flameshot.enabled`：控制是否允许注入 Flameshot（替代原有的 profiler），值为 `"false"` 时拒绝注入，值为 `"true"` 或不设置时允许注入（但需配置匹配规则和 `processes` 字段才能实际触发注入）
    - `admission.datakit/enabled`：控制是否启用所有注入功能，值为 `"false"` 时拒绝所有注入（优先级最高）

    示例：

    ```yaml
      annotations:
        admission.datakit/ddtrace.enabled: "true"
        admission.datakit/logfwd.enabled: "true"
    ```

=== "DataKit-Operator v1.7.0 之前版本"

    在旧版本中，可以通过 `admission.datakit/java-lib.version` Annotation 指定 ddtrace 的镜像版本：

    - key 是 `admission.datakit/%s-lib.version`，`%s` 需要替换成指定的语言，目前支持 `java`
    - value 是指定版本号。默认是 DataKit-Operator 配置 `java_agent_image` 指定的版本

    例如添加 Annotation 如下：

    ```yaml
      annotations:
        admission.datakit/java-lib.version: "v1.36.2-ext"
    ```

    表示这个 Pod 需要注入的镜像版本是 v1.36.2-ext，镜像地址取自配置 `admission_inject`->`ddtrace`->`images`->`java_agent_image`，替换镜像版本为"v1.36.2-ext"，即 `pubrepo.<<<custom_key.brand_main_domain>>>/datakit-operator/dd-lib-java-init:v1.36.2-ext`。
<!-- markdownlint-enable -->

## DataKit Operator 注入 {#datakit-operator-inject-sidecar}

在大型 Kubernetes 集群中，批量修改配置是比较麻烦的事情。DataKit-Operator 会根据配置中的匹配规则（`namespace_selectors`/`label_selectors`）决定是否注入，Annotation 只能用于拒绝注入（设置为 `"false"`），不能单独触发注入。

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    目前支持的功能有：

    - 注入 `ddtrace` agent 和 environment 的功能
    - 挂载 `logfwd` sidecar 并开启日志采集的功能
    - 注入 Flameshot 采集性能分析数据（替代原有的 Profiler）[:octicons-beaker-24: Experimental](index.md#experimental)

=== "DataKit-Operator v1.7.0 之前版本"

    目前支持的功能有：

    - 注入 `ddtrace` agent 和 environment 的功能
    - 挂载 `logfwd` sidecar 并开启日志采集的功能
    - 注入 [`async-profiler`](https://github.com/async-profiler/async-profiler){:target="_blank"}  采集 JVM 程序的 profile 数据 [:octicons-beaker-24: Experimental](index.md#experimental)
    - 注入 [`py-spy`](https://github.com/benfred/py-spy){:target="_blank"} 采集 Python 应用的 profile 数据 [:octicons-beaker-24: Experimental](index.md#experimental)
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ info

    只支持 v1 版本的 `deployments/daemonsets/cronjobs/jobs/statefulsets` 这五类 Kind，且因为 DataKit-Operator 实际对 PodTemplate 操作，所以不支持 Pod。 在本文中，以 `Deployment` 代替描述这五类 Kind。
<!-- markdownlint-enable -->

### DDtrace Agent {#datakit-operator-inject-lib}

#### 使用说明 {#datakit-operator-inject-lib-usage}

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
    1. 在 DataKit-Operator 配置中设置 `ddtraces` 数组，配置 `namespace_selectors`/`label_selectors` 匹配规则。
    1. （可选）在 deployment 添加 Annotation `admission.datakit/ddtrace.enabled: "true"` 允许注入（如果设置为 `"false"` 则会拒绝注入）。

=== "DataKit-Operator v1.7.0 之前版本"

    1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
    1. 在 deployment 添加指定 Annotation `admission.datakit/java-lib.version: ""`，表示需要注入默认版本的 DDtrace Java Agent。
<!-- markdownlint-enable -->

#### 用例 {#datakit-operator-inject-lib-example}

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    下面是一个 Deployment 示例，给 Deployment 创建的所有 Pod 注入 `dd-java-lib`：

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: nginx-deployment
      labels:
        app: nginx
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: nginx
      template:
        metadata:
          labels:
            app: nginx
          annotations:
            admission.datakit/ddtrace.enabled: "true"
        spec:
          containers:
          - name: nginx
            image: nginx:1.22
            ports:
            - containerPort: 80
    ```

=== "DataKit-Operator v1.7.0 之前版本"

    下面是一个 Deployment 示例，给 Deployment 创建的所有 Pod 注入 `dd-java-lib`：

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: nginx-deployment
      labels:
        app: nginx
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: nginx
      template:
        metadata:
          labels:
            app: nginx
          annotations:
            admission.datakit/java-lib.version: ""
        spec:
          containers:
          - name: nginx
            image: nginx:1.22
            ports:
            - containerPort: 80
    ```
<!-- markdownlint-enable -->

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f nginx.yaml
...
```

验证如下：

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
nginx-deployment-7bd8dd85f-fzmt2       1/1     Running   0             4s

$ kubectl get pod nginx-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.initContainers\[\*\].name}

datakit-lib-init
```

### logfwd {#datakit-operator-inject-logfwd}

#### 前置条件 {#datakit-operator-inject-logfwd-prerequisites}

[logfwd](../integrations/logfwd.md#using) 是 DataKit 的专属日志采集应用，需要先在同一个 Kubernetes 集群中部署 DataKit，且达成以下两点：

1. DataKit 开启 `logfwdserver` 采集器，例如监听端口是 `9533`
1. DataKit service 需要开放 `9533` 端口，使得其他 Pod 能访问 `datakit-service.datakit.svc:9533`

#### 使用说明 {#datakit-operator-inject-logfwd-instructions}

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    DataKit-Operator v1.7.0 对 logfwd 注入进行了优化，采用 `ClusterLoggingConfig` CRD 集中管理日志采集配置：

    - **集中管理采集配置**：支持监听 Kubernetes `ClusterLoggingConfig` CRD，并暴露匹配结果供 logfwd sidecar 轮询获取（sidecar 默认每 60 秒向 Operator 发起 HTTP 请求，logfwd 需 ≥ 1.86.0）。
    - **热更新 & 精细匹配**：CRD selector（Namespace/Pod/Label/Container）随改随生效，无需重建 Workload。
    - **简化配置**：日志采集配置完全通过 CRD 管理，不再支持通过 Annotation 覆盖配置。

    > 若尚未了解 ClusterLoggingConfig 的定义及编写方式，请先阅读[容器日志采集 CRD 配置文档](../integrations/container-log-for-k8s-crd.md)。

    操作流程：

    1. 注册 `ClusterLoggingConfig` CRD（如 DataKit 文档中所述）。
    1. 升级/安装 DataKit-Operator v1.7.0，并添加 CRD 的 RBAC 读权限。
    1. 在 DataKit-Operator 配置中设置 `logfwds` 数组，配置 `namespace_selectors`/`label_selectors` 匹配规则和 `log_configs` 字段。
    1. （可选）在目标 Pod 添加 Annotation `admission.datakit/logfwd.enabled: "true"` 允许注入（如果设置为 `"false"` 则会拒绝注入）。
    1. 创建 `ClusterLoggingConfig` 资源，logfwd sidecar 将定期（默认 60 秒）向 DataKit-Operator 发送 HTTP 请求并拉取采集配置。

    安装最新的 `datakit-operator.yaml` 即可带上必要权限，或参考下列最小示例：

    ```yaml
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: datakit-operator
    rules:
    - apiGroups: ["logging.datakits.io"]
      resources: ["clusterloggingconfigs"]
      verbs: ["get", "list", "watch"]

    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRoleBinding
    metadata:
      name: datakit-operator
    roleRef:
      apiGroup: rbac.authorization.k8s.io
      kind: ClusterRole
      name: datakit-operator
    subjects:
    - kind: ServiceAccount
      name: datakit-operator
      namespace: datakit

    ---
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: datakit-operator
      namespace: datakit

    ---
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: datakit-operator
      namespace: datakit
      labels:
        app: datakit-operator
    spec:
      replicas: 1  # Do not change the ReplicaSet number!
      selector:
         matchLabels:
           app: datakit-operator
      template:
        metadata:
          labels:
            app: datakit-operator
        spec:
          serviceAccountName: datakit-operator
          containers:
          - name: operator
            # other..
    ```

    logfwd 注入新增若干环境变量与镜像版本要求，可在 `datakit-operator-config` ConfigMap 中配置：

    ```json
                "logfwds": [
                    {
                        "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/logfwd:1.86.0",
                        "envs": {
                            "LOGFWD_DATAKIT_HOST":              "{fieldRef:status.hostIP}",
                            "LOGFWD_DATAKIT_PORT":              "9533",
                            "LOGFWD_DATAKIT_OPERATOR_ENDPOINT": "datakit-operator.datakit.svc:443",
                            "LOGFWD_GLOBAL_SERVICE":            "{fieldRef:metadata.labels['app']}",
                            "LOGFWD_POD_NAME":                  "{fieldRef:metadata.name}",
                            "LOGFWD_POD_NAMESPACE":             "{fieldRef:metadata.namespace}",
                            "LOGFWD_POD_IP":                    "{fieldRef:status.podIP}"
                        },
                        "log_configs": "",
                        "log_volume_paths": []
                    }
                ]
    ```

    | 环境变量名                         | 配置项含义                                                                                                                                                                              |
    | :---                               | :---                                                                                                                                                                                    |
    | `LOGFWD_DATAKIT_HOST`              | DataKit 实例地址（IP 或可解析域名）。                                                                                                                                                   |
    | `LOGFWD_DATAKIT_PORT`              | DataKit `logfwdserver` 监听端口，例如 `9533`。                                                                                                                                          |
    | `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` | DataKit-Operator Endpoint，形如 `datakit-operator.datakit.svc:443` 或 `https://datakit-operator.datakit.svc:443`，用于查询 CRD 配置；留空则不会尝试拉取。支持自动添加 `https://` 前缀。 |
    | `LOGFWD_GLOBAL_SOURCE`             | 全局 `source`，优先级高于单条配置中的 `source` 字段。                                                                                                                                   |
    | `LOGFWD_GLOBAL_SERVICE`            | 全局 `service`，若单条配置中未指定 `service`，则使用全局值；若全局值也为空，则回退为 `source`。                                                                                         |
    | `LOGFWD_GLOBAL_STORAGE_INDEX`      | 全局 `storage_index`，优先级高于单条配置中的 `storage_index` 字段。                                                                                                                     |
    | `LOGFWD_POD_NAME`                  | 自动写入 `pod_name` tag，通常通过 Downward API 注入。                                                                                                                                   |
    | `LOGFWD_POD_NAMESPACE`             | 自动写入 `namespace` tag。                                                                                                                                                              |
    | `LOGFWD_POD_IP`                    | 自动写入 `pod_ip` tag，便于定位容器实例。                                                                                                                                               |

    配置字段说明：

    - `log_configs`：日志采集配置（JSON 字符串），用于调试或覆盖 CRD 内容。**如果 `log_configs` 为空，logfwd 注入将被跳过**。结构示例：

    ```json
    [
      {
        "type": "file",
        "disable": false,
        "source": "nginx-access",
        "service": "nginx",
        "path": "/var/log/nginx/access.log",
        "pipeline": "nginx-access.p",
        "storage_index": "app-logs",
        "multiline_match": "^\\d{4}-\\d{2}-\\d{2}",
        "remove_ansi_escape_codes": false,
        "from_beginning": false,
        "character_encoding": "utf-8",
        "tags": {
          "env": "production",
          "team": "backend"
        }
      }
    ]
    ```

    | 字段                            | 类型    | 必填     | 说明                                                                                                  | 示例                      |
    | ------                          | ------  | ------   | ------                                                                                                | ------                    |
    | `type`                          | string  | 是       | logfwd 采集类型只能是 `"file"`                                                                        | `"file"`                  |
    | `disable`                       | boolean | 否       | 是否禁用此采集配置                                                                                    | `false`                   |
    | `source`                        | string  | 是       | 日志来源标识，用于区分不同日志流                                                                      | `"nginx-access"`          |
    | `service`                       | string  | 否       | 日志隶属的服务，默认值为日志来源（source）                                                            | `"nginx"`                 |
    | `path`                          | string  | 条件必填 | 日志文件路径（支持 glob 模式），type=file 时必填                                                      | `"/var/log/nginx/*.log"`  |
    | `multiline_match`               | string  | 否       | 多行日志起始行的正则表达式，注意 JSON 中需要转义反斜杠                                                | `"^\\d{4}-\\d{2}-\\d{2}"` |
    | `pipeline`                      | string  | 否       | 日志解析管道配置文件名称（需在 DataKit 端配置）                                                       | `"nginx-access.p"`        |
    | `storage_index`                 | string  | 否       | 日志存储的索引名称                                                                                    | `"app-logs"`              |
    | `remove_ansi_escape_codes`      | boolean | 否       | 是否删除日志数据的 ANSI 转义字符（颜色代码等）                                                        | `false`                   |
    | `from_beginning`                | boolean | 否       | 是否从文件首部开始采集日志（默认从文件末尾开始）                                                      | `false`                   |
    | `from_beginning_threshold_size` | int     | 否       | 搜寻到文件时，如果文件 size 小于此值就从文件首部采集日志，单位字节，默认 20MB                         | `1000`                    |
    | `character_encoding`            | string  | 否       | 字符编码，支持 `utf-8`, `utf-16le`, `utf-16be`, `gbk`, `gb18030` 或空字符串（自动检测）。默认为空即可 | `"utf-8"`                 |
    | `tags`                          | object  | 否       | 额外的标签键值对，会附加到每条日志记录上                                                              | `{"env": "prod"}`         |

    - `log_volume_paths`：需要挂载的宿主路径列表（字符串数组），用于让 sidecar 能访问真实日志文件，例如 `["/var/log", "/data/log"]`。请避免父子路径同时存在，防止 Volume 冲突。

    支持的 Annotation：

    - `admission.datakit/logfwd.enabled`：控制是否允许注入，值为 `"false"` 时拒绝注入，值为 `"true"` 或不设置时允许注入（但需配置匹配规则和 `log_configs` 字段才能实际触发注入）。

    <!-- markdownlint-disable MD046 -->
    ???+ attention

        从 DataKit-Operator v1.7.0 版本开始，移除了对 `admission.datakit/logfwd.log_configs` 和 `admission.datakit/logfwd.volume_paths` Annotation 的支持。日志采集配置应完全通过 `ClusterLoggingConfig` CRD 进行管理。

    ???+ important

        **重要说明**：如果配置中的 `log_configs` 字段为空，logfwd 注入将被跳过。即使 Pod 添加了 `admission.datakit/logfwd.enabled: "true"` Annotation 且匹配了选择器规则，也需要确保 `log_configs` 字段不为空才能成功注入。
    <!-- markdownlint-enable -->

    ClusterLoggingConfig 示例：

    ```yaml
    apiVersion: logging.datakits.io/v1alpha1
    kind: ClusterLoggingConfig
    metadata:
      name: nginx-logs
    spec:
      selector:
        namespaceRegex: "^(middleware)$"
        podLabelSelector: "app=logging"
      podTargetLabels:
        - app
        - env
      configs:
        - type: file
          source: nginx-access
          service: nginx
          path: /var/log/nginx/access.log
          pipeline: nginx-access.p
          storage_index: app-logs
          multiline_match: "^\\d{4}-\\d{2}-\\d{2}"
          tags:
            team: web
    ```

    应用上述资源后，DataKit-Operator 会：

    1. 监听 Deployment 创建事件并注入 `datakit-logfwd` 容器。
    1. 根据 `ClusterLoggingConfig` 选择器匹配 Pod，持续维护匹配结果，供 sidecar 在轮询时读取。
    1. sidecar 启动后通过 `LOGFWD_DATAKIT_OPERATOR_ENDPOINT` 与 Operator 交互，每 60 秒拉取一次 CRD 配置，并将任务转发至 DataKit `logfwdserver`。

=== "DataKit-Operator v1.6.0 及之前版本"

    <!-- markdownlint-disable MD046 -->
    ???+ attention

        此配置方式在 DataKit-Operator v1.6.0 及之前版本使用。v1.7.0 版本采用了新的 CRD 配置方式，v1.6.0 版本中引入的 CRD + Annotation 混合方案已被废弃。
    <!-- markdownlint-enable -->

    1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
    1. 在 deployment 添加指定 Annotation，表示需要挂载 logfwd sidecar。注意 Annotation 要添加在 template 中
        - key 统一是 `admission.datakit/logfwd.instances`
        - value 是一个 JSON 字符串，是具体的 logfwd 配置，示例如下：

    ```json
    [
        {
            "datakit_addr": "datakit-service.datakit.svc:9533",
            "loggings": [
                {
                    "logfiles":      ["<your-logfile-path>"],
                    "ignore":        [],
                    "storage_index": "<your-storage-index>",
                    "source":        "<your-source>",
                    "service":       "<your-service>",
                    "pipeline":      "<your-pipeline.p>",
                    "character_encoding": "",
                    "multiline_match": "<your-match>",
                    "tags": {}
                },
                {
                    "logfiles": ["<your-logfile-path-2>"],
                    "source": "<your-source-2>"
                }
            ]
        }
    ]
    ```

    参数说明，可参考 [logfwd 配置](../integrations/logfwd.md#config)：

    - `datakit_addr` 是 DataKit logfwdserver 地址
    - `loggings` 为主要配置，是一个数组，可参考 [DataKit logging 采集器](../integrations/logging.md)
        - `logfiles` 日志文件列表，可以指定绝对路径，支持使用 glob 规则进行批量指定，推荐使用绝对路径
        - `ignore` 文件路径过滤，使用 glob 规则，符合任意一条过滤条件将不会对该文件进行采集
        - `storage_index` 指定日志存储索引
        - `source` 数据来源，如果为空，则默认使用 'default'
        - `service` 新增标记 tag，如果为空，则默认使用 $source
        - `pipeline` Pipeline 脚本路径，如果为空将使用 $source.p，如果 $source.p 不存在将不使用 Pipeline（此脚本文件存在于 DataKit 端）
        - `character_encoding` 选择编码，如果编码有误会导致数据无法查看，默认为空即可。支持 `utf-8/utf-16le/utf-16le/gbk/gb18030`
        - `multiline_match` 多行匹配，详见 [DataKit 日志多行配置](../integrations/logging.md#multiline)，注意因为是 JSON 格式所以不支持 3 个单引号的"不转义写法"，正则 `^\d{4}` 需要添加转义写成 `^\\d{4}`
        - `tags` 添加额外 `tag`，书写格式是 JSON map，例如 `{ "key1":"value1", "key2":"value2" }`

    <!-- markdownlint-disable MD046 -->
    ???+ note

        注入 logfwd 时，DataKit Operator 默认复用相同路径的 volume，避免因为存在同样路径的 volume 而注入报错。

        路径末尾有斜线和无斜线的意义不同，例如 `/var/log` 和 `/var/log/` 是不同路径，不能复用。
    <!-- markdownlint-enable -->
<!-- markdownlint-enable -->

#### 用例 {#datakit-operator-inject-logfwd-example}

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    下面是一个 Deployment 示例，使用 CRD 方式配置日志采集：

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: logging-demo
      namespace: middleware
      labels:
        app: logging
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: logging
      template:
        metadata:
          labels:
            app: logging
          annotations:
            admission.datakit/logfwd.enabled: "true"
        spec:
          containers:
          - name: log-app
            image: nginx:1.25
    ```

    同时需要创建对应的 `ClusterLoggingConfig` CRD 资源来配置日志采集规则。

=== "DataKit-Operator v1.6.0 及之前版本"

    下面是一个 Deployment 示例，使用 shell 持续向文件写入数据，且配置该文件的采集：

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: logging-deployment
      labels:
        app: logging
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: logging
      template:
        metadata:
          labels:
            app: logging
          annotations:
            admission.datakit/logfwd.instances: '[{"datakit_addr":"datakit-service.datakit.svc:9533","loggings":[{"logfiles":["/var/log/log-test/*.log"],"source":"deployment-logging","tags":{"key01":"value01"}}]}]'
        spec:
          containers:
          - name: log-container
            image: busybox
            args: [/bin/sh, -c, 'mkdir -p /var/log/log-test; i=0; while true; do printf "$(date "+%F %H:%M:%S") [%-8d] Bash For Loop Examples.\\n" $i >> /var/log/log-test/1.log; i=$((i+1)); sleep 1; done']
    ```
<!-- markdownlint-enable -->

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f logging.yaml
...
```

验证如下：

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
logging-deployment-5d48bf9995-vt6bb       1/1     Running   0             4s

$ kubectl get pod logging-deployment-5d48bf9995-vt6bb -o=jsonpath={.spec.containers\[\*\].name}
log-container datakit-logfwd
```

最终可以在<<<custom_key.brand_name>>>日志平台查看日志是否采集。

### Flameshot {#inject-flameshot}

<!-- markdownlint-disable MD046 -->
???+ attention

    Flameshot 是 DataKit-Operator v1.7.0 引入的性能分析工具，用于替代原有的 Profiler（async-profiler、py-spy 等）。原有的 Profiler 注入功能在 v1.7.0 版本中已被移除。
<!-- markdownlint-enable -->

#### 前置条件 {#flameshot-prerequisites}

- 集群已安装 [DataKit](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/){:target="_blank"}。
- [开启 profile](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env){:target="_blank"} 采集器。
- （可选）如需使用 Prometheus Annotations 自动注入功能，需要开启 DataKit 的 KubernetesPrometheus 采集器，并配置 `EnableDiscoveryOfPrometheusPodAnnotations = true` 启用 Pod Annotations 自动发现功能。

#### 使用说明 {#flameshot-usage}

<!-- markdownlint-disable MD046 -->
=== "DataKit-Operator v1.7.0 及以后版本"

    1. 在目标 Kubernetes 集群，[下载和安装 DataKit-Operator](datakit-operator.md#datakit-operator-overview-and-install)
    1. 在 DataKit-Operator 配置中设置 `flameshots` 数组，配置 `namespace_selectors`/`label_selectors` 匹配规则和 `processes` 字段指定要监控的进程。
    1. 在 deployment 添加指定 Annotation `admission.datakit/flameshot.enabled: "true"`，允许注入 Flameshot（如果设置为 `"false"` 则会禁用注入）。

    Flameshot 配置示例：

    ```json
    {
        "admission_inject_v2": {
            "flameshots": [
                {
                    "namespace_selectors": [],
                    "label_selectors":     [],
                    "image": "pubrepo.<<<custom_key.brand_main_domain>>>/datakit/flameshot:latest",
                    "envs": {
                        "FLAMESHOT_DATAKIT_ADDR":     "http://datakit-service.datakit:9529/profiling/v1/input",
                        "FLAMESHOT_MONITOR_INTERVAL": "10s",
                        "FLAMESHOT_LOG_LEVEL":        "info",
                        "FLAMESHOT_PROFILING_PATH":   "/flameshot-data",
                        "FLAMESHOT_LOG_PATH":         "/var/log/flameshot.log",
                        "FLAMESHOT_HTTP_LOCAL_IP":    "{fieldRef:status.podIP}",
                        "FLAMESHOT_HTTP_LOCAL_PORT":  "8089",
                        "FLAMESHOT_SERVICE":  "{fieldRef:metadata.labels['app']}",
                        "FLAMESHOT_TAGS": "pod_name:$(POD_NAME),pod_namespace:$(POD_NAMESPACE),host:$(NODE_NAME)"
                    },
                    "resources": {
                        "requests": {
                            "cpu":    "100m",
                            "memory": "128Mi"
                        },
                        "limits": {
                           "cpu":    "200m",
                           "memory": "256Mi"
                        }
                    },
                    "processes": "",
                    "enable_prometheus_annotations": true
                }
            ]
        }
    }
    ```

    配置字段说明：

    | 字段                            | 类型    | 必填     | 说明                                                                                                  |
    | ------                          | ------  | ------   | ------                                                                                                |
    | `namespace_selectors`           | array   | 否       | 命名空间选择器数组，支持正则表达式匹配                                                                |
    | `label_selectors`               | array   | 否       | 标签选择器数组，使用 Kubernetes Label Selector 语法                                                  |
    | `image`                         | string  | 是       | Flameshot 容器镜像地址                                                                                |
    | `envs`                          | object  | 否       | 环境变量配置，支持 Downward API                                                                       |
    | `resources`                     | object  | 否       | 资源限制配置（requests 和 limits）                                                                   |
    | `processes`                     | string  | 是       | 进程监控配置（JSON 字符串），会作为 `FLAMESHOT_PROCESSES` 环境变量注入到 Flameshot 容器中。格式请参考 [Flameshot 相关文档](../integrations/flameshot.md) |
    | `enable_prometheus_annotations` | boolean | 否       | 是否自动添加 Prometheus 相关 Annotations。在默认配置模板中为 `true`，如果用户自定义配置且不设置该字段，则默认为 `false`。如果 Pod 已存在任意 `prometheus.io/` 开头的 Annotation，则不会注入 |

    <!-- markdownlint-disable MD046 -->
    ???+ important

        **重要说明**：`processes` 字段是一个 JSON 字符串，该值会直接作为 `FLAMESHOT_PROCESSES` 环境变量注入到 Flameshot 容器中。`processes` 字段的格式和含义请参考 [Flameshot 相关文档](../integrations/flameshot.md)。如果 `processes` 为空，Flameshot 注入将被跳过。
    <!-- markdownlint-enable -->

    环境变量说明：

    | 环境变量名                    | 说明                                                                                                 |
    | :---                          | :---                                                                                                 |
    | `FLAMESHOT_DATAKIT_ADDR`     | DataKit profiling 接收地址，例如 `http://datakit-service.datakit:9529/profiling/v1/input`          |
    | `FLAMESHOT_MONITOR_INTERVAL`  | 监控间隔，例如 `10s`                                                                                 |
    | `FLAMESHOT_LOG_LEVEL`        | 日志级别，例如 `info`                                                                                |
    | `FLAMESHOT_PROFILING_PATH`    | Profiling 数据存储路径，例如 `/flameshot-data`                                                       |
    | `FLAMESHOT_LOG_PATH`          | 日志文件路径，例如 `/var/log/flameshot.log`                                                          |
    | `FLAMESHOT_HTTP_LOCAL_IP`     | HTTP 服务本地 IP，通常通过 Downward API 注入，例如 `{fieldRef:status.podIP}`                       |
    | `FLAMESHOT_HTTP_LOCAL_PORT`   | HTTP 服务端口，例如 `8089`                                            |
    | `FLAMESHOT_PROCESSES`         | 进程监控配置（由 `processes` 字段自动注入），JSON 字符串格式                                        |

    **Prometheus Annotations 自动注入**：

    当 `enable_prometheus_annotations` 设置为 `true` 时（在默认配置模板中为 `true`），DataKit-Operator 会自动为注入 Flameshot 的 Pod 添加以下 Prometheus 相关 Annotations，便于 DataKit 的 KubernetesPrometheus 采集器自动发现并采集 Flameshot 暴露的指标：

    - `prometheus.io/scrape: "true"`：标识该 Pod 需要被采集
    - `prometheus.io/port: "<port>"`：指标暴露端口，取值来自环境变量 `FLAMESHOT_HTTP_LOCAL_PORT`（例如 `"8089"`）
    - `prometheus.io/scheme: "http"`：指标采集协议
    - `prometheus.io/path: "/metrics"`：指标路径
    - `prometheus.io/param_measurement: "flameshot"`：指定 measurement 名称

    <!-- markdownlint-disable MD046 -->
    ???+ important

        **重要说明**：
        
        1. 如果 Pod 已存在任意一个 `prometheus.io/` 开头的 Annotation，DataKit-Operator 将不会注入上述 Prometheus Annotations，避免覆盖已有的指标采集配置。
        2. 要使用此功能，需要在 DataKit 中开启 KubernetesPrometheus 采集器，并配置 `EnableDiscoveryOfPrometheusPodAnnotations = true` 启用 Pod Annotations 自动发现功能。
    <!-- markdownlint-enable -->

=== "DataKit-Operator v1.7.0 之前版本"

    <!-- markdownlint-disable MD046 -->
    ???+ attention

        在 DataKit-Operator v1.7.0 之前版本，使用 async-profiler（Java）和 py-spy（Python）进行性能分析。这些功能在 v1.7.0 版本中已被移除，请使用 Flameshot 替代。
    <!-- markdownlint-enable -->
<!-- markdownlint-enable -->

#### 用例 {#flameshot-example}

<!-- markdownlint-disable MD046 -->
???+ note

    注意：仅添加 `admission.datakit/flameshot.enabled: "true"` Annotation 不足以触发注入，还需要在 DataKit-Operator 配置中设置匹配的 `flameshots` 规则（包括 `namespace_selectors`/`label_selectors` 和 `processes` 字段）。如果 `processes` 字段为空，注入将被跳过。
<!-- markdownlint-enable -->

下面是一个 Deployment 示例，给 Deployment 创建的所有 Pod 注入 Flameshot（前提是 DataKit-Operator 配置中已设置匹配的规则）：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app-deployment
  labels:
    app: myapp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
      annotations:
        admission.datakit/flameshot.enabled: "true"
    spec:
      containers:
      - name: app
        image: myapp:latest
        ports:
        - containerPort: 8080
```

使用 yaml 文件创建资源：

```shell
$ kubectl apply -f app-deployment.yaml
...
```

验证如下：

```shell
$ kubectl get pod

NAME                                   READY   STATUS    RESTARTS      AGE
app-deployment-7bd8dd85f-fzmt2          2/2     Running   0             4s

$ kubectl get pod app-deployment-7bd8dd85f-fzmt2 -o=jsonpath={.spec.containers\[\*\].name}
app datakit-flameshot
```

稍等几分钟后即可在<<<custom_key.brand_name>>>控制台 [应用性能检监测-Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    若无法看到数据，可以进入 `datakit-flameshot` 容器查看相应日志进行排查：
    ```shell
    $ kubectl exec -it app-deployment-7bd8dd85f-fzmt2 -c datakit-flameshot -- bash
    $ cat /var/log/flameshot.log
    ```
<!-- markdownlint-enable -->

### async-profiler {#inject-async-profiler}

#### 前置条件 {#async-profiler-prerequisites}

- 集群已安装 [DataKit](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/){:target="_blank"}。
- [开启 profile](https://docs.<<<custom_key.brand_main_domain>>>/datakit/datakit-daemonset-deploy/#using-k8-env){:target="_blank"} 采集器。
- Linux 内核参数 [kernel.perf_event_paranoid](https://www.kernel.org/doc/Documentation/sysctl/kernel.txt){:target="_blank"} 值设置为 2 及以下。

<!-- markdownlint-disable MD046 -->
???+ note

    `async-profiler` 使用 [`perf_events`](https://perf.wiki.kernel.org/index.php/Main_Page){:target="_blank"} 工具来抓取 Linux 的内核调用堆栈，非特权进程依赖内核的相应设置，可以使用以下命令来修改内核参数：
    ```shell
    $ sudo sysctl kernel.perf_event_paranoid=1
    $ sudo sysctl kernel.kptr_restrict=0
    # 或者
    $ sudo sh -c 'echo 1 >/proc/sys/kernel/perf_event_paranoid'
    $ sudo sh -c 'echo 0 >/proc/sys/kernel/kptr_restrict'
    ```
<!-- markdownlint-enable -->

在你的 [Pod 控制器](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} 资源配置文件中的
`.spec.template.metadata.annotations` 节点下添加 annotation：`admission.datakit/java-profiler.version: "latest"`，然后应用该资源配置文件，
DataKit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。

接下来以一个名为 `movies-java` 的 `Deployment` 资源配置文件为例进行说明。

```yaml
kind: Deployment
metadata:
  name: movies-java
  labels:
    app: movies-java
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-java
  template:
    metadata:
      name: movies-java
      labels:
        app: movies-java
      annotations:
        admission.datakit/java-profiler.version: "0.4.4"
    spec:
      containers:
        - name: movies-java
          image: zhangyicloud/movies-java:latest
          imagePullPolicy: IfNotPresent
          securityContext:
            seccompProfile:
              type: Unconfined
          env:
            - name: JAVA_OPTS
              value: ""

      restartPolicy: Always
```

应用配置文件并检查是否生效：

```shell
$ kubectl apply -f deployment-movies-java.yaml

$ kubectl get pods | grep movies-java
movies-java-784f4bb8c7-59g6s   2/2     Running   0          47s

$ kubectl describe pod movies-java-784f4bb8c7-59g6s | grep datakit-profiler
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler:
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler-volume:
  Normal  Created    12m   kubelet            Created container datakit-profiler
  Normal  Started    12m   kubelet            Started container datakit-profiler
```

稍等几分钟后即可在<<<custom_key.brand_name>>>控制台 [应用性能检监测-Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    默认使用命令 `jps -q -J-XX:+PerfDisableSharedMem | head -n 20` 来查找容器中的 JVM 进程，出于性能的考虑，最多只会采集 20 个进程的数据。

???+ note

    可以通过修改 `datakit-operator.yaml` 配置文件中的 `datakit-operator-config` 下的环境变量来配置 profiling 的行为。
    
    | 环境变量              | 说明                                                                                                                                               | 默认值                        |
    | ----                  | --                                                                                                                                                 | -----                         |
    | `DK_PROFILE_SCHEDULE` | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` | `0 * * * *`（每小时调度一次） |
    | `DK_PROFILE_DURATION` | 每次 profiling 持续的时间，单位秒                                                                                                                  | 240（4 分钟）                 |


???+ note

    若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：
    ```shell
    $ kubectl exec -it movies-java-784f4bb8c7-59g6s -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->

### py-spy {#inject-py-spy}

#### 前置条件 {#py-spy-prerequisites}

- 当前只支持 Python 官方解释器（`CPython`）

在你的 [Pod 控制器](https://kubernetes.io/docs/concepts/workloads/controllers/){:target="_blank"} 资源配置文件中的
`.spec.template.metadata.annotations` 节点下添加 annotation：`admission.datakit/python-profiler.version: "latest"`，然后应用该资源配置文件，
DataKit-Operator 会自动在相应的 Pod 中创建一个名为 `datakit-profiler` 的容器来辅助进行 profiling。

接下来将以一个名为 "movies-python" 的 `Deployment` 资源配置文件为例进行说明。

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: movies-python
  labels:
    app: movies-python
spec:
  replicas: 1
  selector:
    matchLabels:
      app: movies-python
  template:
    metadata:
      name: movies-python
      labels:
        app: movies-python
      annotations:
        admission.datakit/python-profiler.version: "latest"
    spec:
      containers:
        - name: movies-python
          image: zhangyicloud/movies-python:latest
          imagePullPolicy: Always
          command:
            - "gunicorn"
            - "-w"
            - "4"
            - "--bind"
            - "0.0.0.0:8080"
            - "app:app"
```

应用资源配置并验证是否生效：

```shell
$ kubectl apply -f deployment-movies-python.yaml

$ kubectl get pods | grep movies-python
movies-python-78b6cf55f-ptzxf   2/2     Running   0          64s
 
$ kubectl describe pod movies-python-78b6cf55f-ptzxf | grep datakit-profiler
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler:
      /app/datakit-profiler from datakit-profiler-volume (rw)
  datakit-profiler-volume:
  Normal  Created    98s   kubelet            Created container datakit-profiler
  Normal  Started    97s   kubelet            Started container datakit-profiler
```

稍等几分钟后即可在<<<custom_key.brand_name>>>控制台 [应用性能检监测-Profiling](https://console.<<<custom_key.brand_main_domain>>>/tracing/profile){:target="_blank"} 页面查看应用性能数据。

<!-- markdownlint-disable MD046 -->
???+ note

    默认使用命令 `ps -e -o pid,cmd --no-headers | grep -v grep | grep "python" | head -n 20` 来查找容器中的 `Python` 进程，出于性能考虑，最多只会采集 20 个进程的数据。

???+ note

    可以通过修改 `datakit-operator.yaml` 配置文件中的 ConfigMap `datakit-operator-config`  下的环境变量来配置 profiling 的行为。

    | 环境变量              | 说明                                                                                                                                               | 默认值                        |
    | ----                  | --                                                                                                                                                 | -----                         |
    | `DK_PROFILE_SCHEDULE` | profiling 的运行计划，使用与 Linux [Crontab](https://man7.org/linux/man-pages/man5/crontab.5.html){:target="_blank"} 相同的语法，如 `*/10 * * * *` | `0 * * * *`（每小时调度一次） |
    | `DK_PROFILE_DURATION` | 每次 profiling 持续的时间，单位秒                                                                                                                  | 240（4 分钟）                 |


???+ note

    若无法看到数据，可以进入 `datakit-profiler` 容器查看相应日志进行排查：
    ```shell
    $ kubectl exec -it movies-python-78b6cf55f-ptzxf -c datakit-profiler -- bash
    $ tail -n 2000 log/main.log
    ```
<!-- markdownlint-enable -->

## DataKit Operator 资源变动 {#datakit-operator-mutate-resource}

### 添加 DataKit Logging 采集所需的配置 {#add-logging-configs}

DataKit Operator 可以为指定的 Pod 自动添加 DataKit Logging 采集所需的配置，包括 `datakit/logs` 注解和对应的文件路径 volume/volumeMount，简化了手动配置的繁杂步骤。这样，用户无需手动干预每个 Pod 配置即可自动启用日志采集功能。

以下是一个配置示例，展示了如何通过 DataKit Operator 的 `admission_mutate` 配置来实现日志采集配置的自动注入：

```json
{
    "server_listen": "0.0.0.0:9543",
    "log_level":     "info",
    "admission_inject": {
        # 其他配置
    },
    "admission_mutate": {
        "loggings": [
            {
                "namespace_selectors": ["middleware"],
                "label_selectors":     ["app=logging"],
                "config": "[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/**/*.log\",\"source\":\"logging-tmp\"}]"
            }
        ]
    }
}
```

`admission_mutate.loggings`：这是一个对象数组，包含多个日志采集配置。每个日志配置包括以下字段：

- `namespace_selectors`：限定符合条件的 Pod 所在的 Namespacce。可以设置多个 Namespace，Pod 必须匹配至少一个 Namespace 才会被选中。与 `label_selectors` 是“或”的关系。
- `label_selectors`：限定符合条件的 Pod 的 label。Pod 必须匹配至少一个 label selector 才会被选中。与 `namespace_selectors` 是“或”的关系。
- `config`：这是一个 JSON 字符串，它将被添加到 Pod 的注解中，注解的 Key 是 `datakit/logs`。如果该 Key 已经存在，它不会被覆盖或重复添加。这个配置将告诉 DataKit 如何采集日志。

DataKit Operator 会自动解析 `config` 配置，并根据其中的路径（`path`）为 Pod 创建对应的 volume 和 volumeMount。

以上述 DataKit Operator 配置为例，如果发现某个 Pod 的 Namespace 是 `middleware`，或 Labels 匹配 `app=logging`，就在 Pod 新增注解和挂载。例如：

```yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    datakit/logs: '[{"disable":false,"type":"file","path":"/tmp/opt/**/*.log","source":"logging-tmp"}]'
  labels:
    app: logging
  name: logging-test
  namespace: default
spec:
  containers:
  - args:
    - |
      mkdir -p /tmp/opt/log1;
      i=1;
      while true; do
        echo "Writing logs to file ${i}.log";
        for ((j=1;j<=10000000;j++)); do
          echo "$(date +'%F %H:%M:%S')  [$j]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log1/file_${i}.log;
          sleep 1;
        done;
        echo "Finished writing 5000000 lines to file_${i}.log";
        i=$((i+1));
      done
    command:
    - /bin/bash
    - -c
    - --
    image: pubrepo.<<<custom_key.brand_main_domain>>>/base/ubuntu:18.04
    imagePullPolicy: IfNotPresent
    name: demo
    volumeMounts:
    - mountPath: /tmp/opt
      name: datakit-logs-volume-0
  volumes:
  - emptyDir: {}
    name: datakit-logs-volume-0
```

这个 Pod 存在 label `app=logging`，能够匹配上，于是 DataKit Operator 就给它添加了 `datakit/logs` 注解，并且将路径 `/tmp/opt` 添加 EmptyDir 挂载。

DataKit 日志采集发现到 Pod 后，就会根据 `datakit/logs` 内容进行定制化采集。

### FAQ {#datakit-operator-faq}

- 怎样指定某个 Pod 不注入？给该 Pod 添加 Annotation `"admission.datakit/enabled": "false"`，将不再为它执行任何操作，此优先级最高。

- DataKit-Operator 使用 Kubernetes Admission Controller 功能进行资源注入，详细机制请查看[官方文档](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/admission-controllers/){:target="_blank"}

- 在 AWS EKS 环境部署，可能导致 DataKit-Operator 不生效，需要在安全组开启 `9543` 端口。
