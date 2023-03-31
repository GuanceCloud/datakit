
# Container Logging
---

Log collection on container/Pod is quite different from log collection on host, and its configuration mode and collection mechanism are different. From the data source, it can be divided into two parts:

- stdout/stderr log collection: The stdout/stderr output of the container application, which is also the most common way, can be viewed using  `docker logs` or `kubectl logs`.
- Container/Pod internal file collection: If the log is not exported to stdout/stderr, it should be stored in a file, and collecting this log needs to be slightly more complicated.

This article will introduce these two acquisition methods in detail.

## stdout/stderr Log Collection {#logging-stdout}

There are two main configurations for stdout/stderr log collection:

- Configure log collection through Pod/container *mirroring characteristics*
- Label the log collection of a specific Pod/container through Annotation/Label

### Adjust Log Collection According to Container Image {#logging-with-image-config}

By default, DataKit collects stdout/stderr logs for all containers/pods on your machine/Node, which may not be expected. Sometimes, we want to collect only (or not) the logs of some containers/pods, where the target container/Pod can be indirectly referred to by the mirror name.

=== "host installation"

    ``` toml
    ## When the container's image matches `hello*` , the container's log is collected
    container_include_logging = ["image:hello*"]
    ## Ignore all containers
    container_exclude_logging = ["image:*"]
    ```
    
    - `container_include` and `container_exclude` must begin with `image` in the form of a [regular-like Glob wildcard](https://en.wikipedia.org/wiki/Glob_(programming)){:target="_blank"}： `"image:<glob rules>"`
    
=== "Kubernetes"

    The following environment variables can be used

    - ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
    - ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG

    to configure log collection for the container. Suppose there are three Pods whose images are:

    - A：`hello/hello-http:latest`
    - B：`world/world-http:latest`
    - C：`registry.jiagouyun.com/datakit/datakit:1.2.0`

    If you want to collect only the logs of Pod A, configure  ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG.

    ``` yaml
    - env:
      - name: ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
        value: image:hello*  # Specify the image name or its wildcard
    ```

???+ tip "How to view a mirror"
    
    Docker：

    ``` shell
    docker inspect --format "{{`{{.Config.Image}}`}}" $CONTAINER_ID
    ```

    Kubernetes Pod：
    
    ``` shell
    echo `kubectl get pod -o=jsonpath="{.items[0].spec.containers[0].image}"`
    ```

### Adjust Container Log Collection Through Annotation/Label {#logging-with-annotation-or-label}

Only the target container/Pod for log collection (or non-collection) can be roughly configured through mirroring, and the corresponding annotation needs to be added to the container/Pod for more advanced configuration. Through this directional annotation, it can be used to configure functions such as log source, multi-line configuration, Pipeline cutting and so on.

In the container Label or Pod yaml, Label a specific Key (`datakit/logs`) whose Value is a JSON string, as shown below:

``` json
[
  {
    "disable"  : false,
    "source"   : "testing-source",
    "service"  : "testing-service",
    "pipeline" : "test.p",

    # The usage is exactly the same as `filter container according to image` above, and `image:` is followed by a regular expression
    "only_images"  : ["image:<your_image_regexp>"],

    "multiline_match" : "^\\d{4}-\\d{2}",

    # You can add additional labels to this container/Pod log
    "tags" : {
      "some_tag" : "some_value",
      "more_tag" : "some_other_value"
    }
  }
]
```

Value Field Description：

| Field Description            | Required | Value             | Default value | Description                                                                                                                                                             |
| -----             | ---- | ----             | ----   | ----                                                                                                                                                             |
| `disable`         | N    | true/false       | false  |  whether to disable log collection for this pod/container                                                                                                                                    |
| `source`          | N    | String           | None     | For log source, please refer to [the source setting of container log collection](container.md#config-logging-source).                                                                                   |
| `service`         | N    | String           | None     | The service to which the log belongs, the default value is log source.                                                                                                                       |
| `pipeline`        | N    | String           | None     | The Pipeline script that applies to the log, defaulting to the script name that matches the log source（`<source>.p`).                                                                                       |
| `only_images`     | N    | String array       | None     | For the multi-container scenario inside Pod, if any image wildcard is filled in, only the logs of containers that can match these images are collected, which is similar to the white list function; If the field is empty, it is considered that the logs of all containers in this Pod are collected.       |
| `multiline_match` | N    | Regular expression string | None     | Used to identify the first line in [multi-line log matching](logging.md#multiline), for example, `"multiline_match":"^\\d{4}"` means that the first line is 4 numbers. In regular expression rules, `\d` is a number, and the previous `\` is used to escape. |
| `tags`            | N    | key-value pair | None     | Add additional tags if a key with the same name already exists（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ）.                                       |

#### Configuration Sample {#logging-annotation-label-example}

You can specify the log collection configuration by configuring the Label of the container or the Annotation of the Pod.

=== "Docker"

    Docker container to add Label method, see [here](https://docs.docker.com/engine/reference/commandline/run/#set-metadata-on-container--l---label---label-file){:target="_blank"}.

=== "Kubernetes"

    At Kubernetes, you can add Pod Annotation in `template` mode when creating Deployment, for example:
    
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
            datakit/logs: |
              [
                {
                  "disable": false,
                  "source": "testing-source",
                  "service": "testing-service",
                  "pipeline": "test.p",
                  "multiline_match": "^\\d{4}-\\d{2}",
                  "only_images": ["image:.*nginx.*", "image:.*my_app.*"],
                  "tags" : {
                    "some_tag" : "some_value"
                  }
                }
              ]
    ```

???+ attention

    - If it is not necessary, do not easily configure pipeline in Annotation/Label. In general, it can be deduced automatically through the `source` field.
    - If you add Labels/Annotations in the configuration file or terminal command line, with English status double quotation marks on both sides, you need to add escape characters.

    The value of `multiline_match` is double-escaped so that four slashes can represent the actual one, for example `\"multiline_match\":\"^\\\\d{4}\"` equivalent `"multiline_match":"^\d{4}"`, for example:

    ```shell
    kubectl annotate pods my-pod datakit/logs="[{\"disable\":false,\"source\":\"testing-source\",\"service\":\"testing-service\",\"pipeline\":\"test.p\",\"only_images\":[\"image:<your_image_regexp>\"],\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]"
    ``` 

## Non-stdout/stderr Log c=Collection {#logging-not-stdout}

If the container/Pod outputs them to a file, it will not be collected using the existing stdout/stderr method.

For this type of container/Pod, the easiest way to do this is to mount the file to the host, and then Datakit opens and configures the [logging collector](logging.md). If it is inconvenient to mount files, you can use the following methods.

### Collect Pod Internal Log {#logging-with-sidecar-config}

If the Pod log does not hit stdout, the Pod internal log can be collected in Sidecar form, see [here](logfwd.md)。

### Collect Container Internal Log {#logging-with-inside-config}

Add Label to the container to specify the file path, and Datakit collects the corresponding file.

???+ attention

    - This approach only works when the Datakit host is deployed, but not when the Kubernetes DaemonSet is deployed.
    - Only Docker runtime is supported, Contaienrd is not supported for the time being.
    - Only containers where GraphDriver is `overlay2` are supported.

In the container Label, a specific Key（`datakit/logs/inside`）is labeled, and its Value is a JSON string, as shown in the following example:

``` json
[
  {
    "source"   : "testing-source",
    "service"  : "testing-service",
    "pipeline" : "test.p",
    "multiline_match" : "^\\d{4}-\\d{2}",
    "paths"    : [
        "/tmp/data*"
    ],
    "tags" : {
      "some_tag" : "some_value",
      "more_tag" : "some_other_value"
    }
  }
]
```

Value field description:

| Field Name            | Required | Value             | Default Value | Description                                                                                                                                                             |
| -----             | ---- | ----             | ----   | ----                                                                                                                                                             |
| `source`          | N    | String           | None     | For log source, please refer to the [source setting of container log collection](container.md#config-logging-source).                                                                                   |
| `service`         | N    | String           | None     | the service to which the log belongs; the default value is the script name that matches the log source (< source >. p)                                                                                                                  |
| `pipeline`        | N    | String           | None     | the Pipeline script that applies to the log, defaulting to the script name that matches the log source (< source >. p)                                                                                       |
| `paths`           | N    | String array       | None     | Configure multiple file paths and support wildcard characters. See [here](logging.md#grok-rules) for wildcard usage.                                                                                          |
| `multiline_match` | N    | Regular expression string | None     | Used to identify the first line in [multi-line log matching](logging.md#multiline), for example, `"multiline_match":"^\\d{4}"` means that the first line is 4 numbers. In regular expression rules,`\d` is a number, and the previous `\` is used to escape. |
| `tags`            | N    | key/value pair | None     | Add additional tags if a key with the same name already exists（[:octicons-tag-24: Version-1.4.6](changelog.md#cl-1.4.6) ）                                                  |

#### Configuration Sample {#logging-inside-example}

Create Dockerfile as follows: 

```
FROM ubuntu:18.04 AS base

RUN  echo 'i=0          \n\
while true              \n\
do                      \n\
    printf "[%d] Bash For Loop Examples. Hello, world! Testing output.\\n" $i >> /tmp/data1 \n\
    true $(( i++ ))     \n\
    sleep 1             \n\
done' >> /root/s.sh

LABEL datakit/logs/inside '[{"source":"testing-source","service":"testing-service","paths":["/tmp/data*"],"tags":{"some_tag":"some_value","more_tag":"some_other_value"}}]'

CMD ["/bin/bash","/root/s.sh"]
```

Create image and container:

```
$ docker build -t testing-image:v1 .
$ docker run -d testing-image:v1
```

When datakit discovers this container, it creates log collections based on its `datakit/logs/inside`.

## FAQ {#faq}

### :material-chat-question: Special ByteCode Filtering of Container Log {#special-char-filter}

Container logs may contain some unreadable bytecodes (such as the color of terminal output, etc.), which can

- Set `logging_remove_ansi_escape_codes` to `true` 
- When the DataKit DaemonSet is deployed, set  `ENV_INPUT_CONTAINER_LOGGING_REMOVE_ANSI_ESCAPE_CODES` to `true`

This configuration affects log processing performance, and the benchmark results are as follows:

```
goos: linux
goarch: amd64
pkg: gitlab.jiagouyun.com/cloudcare-tools/test
cpu: Intel(R) Core(TM) i7-4770HQ CPU @ 2.20GHz
BenchmarkRemoveAnsiCodes
BenchmarkRemoveAnsiCodes-8        636033              1616 ns/op
PASS
ok      gitlab.jiagouyun.com/cloudcare-tools/test       1.056s
```

The processing time of each text will increase by `1616 ns`. Do not turn on this function if the log is not decorated with colors.

### :material-chat-question: Source Setting for Container Log Collection {#config-logging-source}

In the container environment, the log `source` setting is a very important configuration item, which directly affects the display effect on the page. However, it would be cruel to configure a source for each container's logs one by one. Without manually configuring the container log source, DataKit has the following rule (descending priority) for automatically inferring the source of the container log:

???+ attention

    The so-called not manually specifying the container log source means that it is not specified in Pod Annotation or in container.conf (currently there is no configuration item specifying the container log source in container.conf).

- Container name: Typically take the value from the label `io.kubernetes.container.name` of the container. If it is not a container created by Kubernetes (for example, if it is just a simple Docker environment, then this label does not have it, so the container name is not used as the log source).
- short-image-name: Mirror name, `nginx` for `nginx.org/nginx:1.21.0`. In a non-Kubernetes container environment, the (reduced) image name is usually taken first.
- `unknown`: This unknown value is taken if the mirror name is invalid, such as `sha256:b733d4a32c...`.

## Extended Reading {#more-reading}

- [Pipeline: Text Data Processing](../developers/pipeline.md)
- [Overview of DataKit Log Collection](datakit-logging.md)
