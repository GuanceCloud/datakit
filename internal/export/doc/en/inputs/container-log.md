
# Container Logs

---

Datakit supports collecting Kubernetes and host container logs, which can be classified into the following two types based on the data source:

- Console output: This refers to the stdout/stderr output of the container application, which is the most common way. It can be viewed using commands like `docker logs` or `kubectl logs`.

- Container internal files: If the logs are not output to stdout/stderr, they are usually stored in files. Collecting this type of logs requires mounting.

This article will provide a detailed introduction to these two collection methods.

## Logging Collection for Console stdout/stderr {#logging-stdout}

Console output (stdout/stderr) is written to files by the container runtime, and Datakit automatically fetches the LogPath of the container for collection.

If you want to customize the collection configuration, it can be done through adding container environment variables or Kubernetes Pod Annotations.

-The following are the key scenarios for custom configurations:
    - For container environment variables, the key must be set as `DATAKIT_LOGS_CONFIG`.
    - For Pod Annotations, there are two possible formats:
        - `datakit/$CONTAINER_NAME.logs`, where `$CONTAINER_NAME` needs to be replaced with the current Pod's container name. This format is used in multi-container environments.
        - `datakit/logs` applies to all containers of the Pod.

<!-- markdownlint-disable MD046-->
???+ info

    If a container has an environment variable `DATAKIT_LOGS_CONFIG` and can also find the Annotation `datakit/logs` of its corresponding Pod, the configuration from the container environment variable will take precedence.
<!-- markdownlint-enable -->

- The value for custom configurations is as follows:

``` json
[
  {
    "disable" : false,
    "source"  : "<your-source>",
    "service" : "<your-service>",
    "pipeline": "<your-pipeline.p>",
    "tags" : {
      "<some-key>" : "<some_other_value>"
    }
  }
]
```

Field explanations:

| Field Name           | Possible Values           | Explanation                                                                                                                                                                                                                                                                                    |
| -------------------- | ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `disable`            | true/false                | Whether to disable log collection for the container. The default value is `false`.                                                                                                                                                                                                             |
| `type`               | `file`/empty              | The type of collection. If collecting logs from container internal files, it must be set as `file`. The default value is empty, which means collecting `stdout/stderr`.                                                                                                                        |
| `path`               | string                    | The configuration file path. If collecting logs from container internal files, it should be set as the path of the volume, which is accessible from outside the container. The default is not required when collecting `stdout/stderr`.                                                        |
| `source`             | string                    | The source of the logs. Refer to [Configuring the Source for Container Log Collection](container.md#config-logging-source).                                                                                                                                                                    |
| `service`            | string                    | The service to which the logs belong. The default value is the log source (`source`).                                                                                                                                                                                                          |
| `pipeline`           | string                    | The Pipeline script for processing the logs. The default value is the script name that matches the log source (`<source>.p`).                                                                                                                                                                  |
| `multiline_match`    | regular expression string | The pattern used for recognizing the first line of a [multiline log match](logging.md#multiline), e.g., `"multiline_match":"^\\d{4}"` indicates that the first line starts with four digits. In regular expression rules, `\d` represents a digit, and the preceding `\` is used for escaping. |
| `character_encoding` | string                    | The character encoding. If the encoding is incorrect, the data may not be viewable. Supported values are `utf-8`, `utf-16le`, `utf-16le`, `gbk`, `gb18030`, or an empty string. The default is empty.                                                                                          |
| `tags`               | key/value pairs           | Additional tags to be added. If there are duplicate keys, the value in this configuration will take precedence ([:octicons-tag-24: Version-1.4.6](../datakit/changelog.md#cl-1.4.6)).                                                                                                          |

Below is a complete example:

<!-- markdownlint-disable MD046 -->
=== "Container Environment Variables"

    ``` shell
    $ cat Dockerfile
    FROM pubrepo.guance.com/base/ubuntu:18.04 AS base
    RUN mkdir -p /opt
    RUN echo 'i=0; \n\
    while true; \n\
    do \n\
        echo "$(date +"%Y-%m-%d %H:%M:%S")  [$i]  Bash For Loop Examples. Hello, world! Testing output."; \n\
        i=$((i+1)); \n\
        sleep 1; \n\
    done \n'\
    >> /opt/s.sh
    CMD ["/bin/bash", "/opt/s.sh"]

    ## Build the image
    $ docker build -t testing/log-output:v1 .

    ## Start the container, add the environment variable DATAKIT_LOGS_CONFIG (note the character escaping)
    $ docker run --name log-output -env DATAKIT_LOGS_CONFIG='[{"disable":false,"source":"log-source","service":"log-service"}]' -d testing/log-output:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="log-output.yaml"
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: log-demo-deployment
      labels:
        app: log-demo
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: log-demo
      template:
        metadata:
          labels:
            app: log-demo
          annotations:
            ## Add the configuration and specify the container as log-output
            datakit/log-output.logs: |
              [{
                  "disable": false,
                  "source":  "log-output-source",
                  "service": "log-output-service",
                  "tags" : {
                    "some_tag": "some_value"
                  }
              }]
        spec:
          containers:
          - name: log-output
            image: pubrepo.guance.com/base/ubuntu:18.04
            args:
            - /bin/sh
            - -c
            - >
              i=0;
              while true;
              do
                echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output.";
                i=$((i+1));
                sleep 1;
              done
    ```

    ``` yaml
    $ kubectl apply -f log-output.yaml
    ...
    ```

???+ attention

    - If not necessary, avoid configuring the Pipeline in environment variables and Pod Annotations. In general, it can be automatically inferred through the `source` field.
    - When adding Env/Annotations in configuration files or terminal commands, both sides should be enclosed in double quotes with escape characters.
    
    The value of `multiline_match` requires double escaping, with 4 backslashes representing a single one. For example, `\"multiline_match\":\"^\\\\d{4}\"` is equivalent to `"multiline_match":"^\d{4}"`. Here's an example:

    ```shell
    kubectl annotate pods my-pod datakit/logs="[{\"disable\":false,\"source\":\"log-source\",\"service\":\"log-service\",\"pipeline\":\"test.p\",\"only_images\":[\"image:<your_image_regexp>\"],\"multiline_match\":\"^\\\\d{4}-\\\\d{2}\"}]"
    ```
    If a Pod/Container log is already being collected, adding configuration via the `kubectl annotate` command does not take effect.

<!-- markdownlint-enable -->

## Logging for Log Files Inside Containers {#logging-with-inside-config}

For log files inside containers, the configuration is similar to logging console output, except that you need to specify the file path. Other configurations are mostly the same.

Similarly, you can add the configuration either as a container environment variable or a Kubernetes Pod Annotation. The key and value remain the same as mentioned earlier. Please refer to the previous section for details.

Here is a complete example:

<!-- markdownlint-disable MD046 -->
=== "Container Environment Variables"

    ``` shell
    $ cat Dockerfile
    FROM pubrepo.guance.com/base/ubuntu:18.04 AS base
    RUN mkdir -p /opt
    RUN echo 'i=0; \n\
    while true; \n\
    do \n\
        echo "$(date +"%Y-%m-%d %H:%M:%S")  [$i]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log; \n\
        i=$((i+1)); \n\
        sleep 1; \n\
    done \n'\
    >> /opt/s.sh
    CMD ["/bin/bash", "/opt/s.sh"]
    
    ## Build the image
    $ docker build -t testing/log-to-file:v1 .
    
    ## Start the container, add the environment variable DATAKIT_LOGS_CONFIG (note the character escaping).
    ## Unlike configuring stdout, "type" and "path" are mandatory fields, and add the path volume.
    ## Path `/tmp/opt/log` add the `/tmp/opt` anonymous volumes.
    $ docker run --env DATAKIT_LOGS_CONFIG="[{\"disable\":false,\"type\":\"file\",\"path\":\"/tmp/opt/log\",\"source\":\"log-source\",\"service\":\"log-service\"}]" -v /tmp/opt -d testing/log-to-file:v1
    ```

=== "Kubernetes Pod Annotation"

    ``` yaml title="logging.yaml"
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: log-demo-deployment
      labels:
        app: log-demo
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: log-demo
      template:
        metadata:
          labels:
            app: log-demo
          annotations:
            ## Add the configuration and specify the container as logging-demo.
            ## Configure both file and stdout collection, need to add the emptyDir volume to "/tmp/opt" first.
            datakit/logging-demo.logs: |
              [
                {
                  "disable": false,
                  "type": "file",
                  "path":"/tmp/opt/log",
                  "source":  "logging-file",
                  "tags" : {
                    "some_tag": "some_value"
                  }
                },
                {
                  "disable": false,
                  "source":  "logging-output"
                }
              ]
        spec:
          containers:
          - name: logging-demo
            image: pubrepo.guance.com/base/ubuntu:18.04
            args:
            - /bin/sh
            - -c
            - >
              i=0;
              while true;
              do
                echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output.";
                echo "$(date +'%F %H:%M:%S')  [$i]  Bash For Loop Examples. Hello, world! Testing output." >> /tmp/opt/log;
                i=$((i+1));
                sleep 1;
              done
            volumeMounts:
            - mountPath: /tmp/opt
              name: datakit-vol-opt
          volumes:
          - name: datakit-vol-opt
            emptyDir: {}
    ```

    ``` yaml
    $ kubectl apply -f logging.yaml
    ```
<!-- markdownlint-enable -->

For log files inside containers, in a Kubernetes environment, you can also achieve collection by adding a sidecar. Please refer to [here](logfwd.md) for more information.

<!-- markdownlint-disable MD013 -->
### Adjust Log Collection According to Container Image {#logging-with-image-config}
<!-- markdownlint-enable -->
By default, DataKit collects stdout/stderr logs for all containers on your machine/Node, which may not be expected. Sometimes, we want to collect only (or not) the logs of some containers, where the target container/Pod can be indirectly referred to by the mirror name.

<!-- markdownlint-disable MD046 -->
=== "host installation"

    ```toml
    ## Take image for example.
    ## Capture a container's log when its image matches `datakit`.
    container_include_log = ["image:datakit"]
    ## Ignore all kodo containers
    container_exclude_log = ["image:kodo"]
    ```

    `container_include` and `container_exclude` must start with an attribute field in a sort of [Glob wildcard for class regularity](https://en.wikipedia.org/wiki/Glob_(programming)){:target="_ blank"}: `"<field name>:<glob rule>"`

    The following 4 field rules are now supported, all of which are infrastructure attribute fields:

    - image : `image:pubrepo.guance.com/datakit/datakit:1.18.0`
    - image_name : `image_name:pubrepo.guance.com/datakit/datakit`
    - image_short_name : `image_short_name:datakit`
    - namespace : `namespace:datakit-ns`

    For the same type of rule (`image` or `namespace`), if both `include` and `exclude` exist, the condition that `include` holds and `exclude` does not hold needs to be satisfied. For example:
    ```toml
    ## This causes all containers to be filtered. If there is a container ``datakit`` that satisfies both ``include`` and ``exclude``, then it will be filtered out of log collection; if there is a container ``nginx`` that does not satisfy ``include`` in the first place, it will be filtered out of log collection.

    container_include_log = ["image_name:datakit"]
    container_exclude_log = ["image_name:*"]
    ```

    Any one of the field rules for multiple types matches and its logs are no longer captured. Example:
    ```toml
    ## The container only needs to match either `image_name` and `namespace` to stop collecting logs.

    container_include_log = []
    container_exclude_log = ["image_name:datakit", "namespace:datakit-ns"]
    ```

    The configuration rules for `container_include_log` and `container_exclude_log` are complex, and their simultaneous use can result in a variety of priority cases. It is recommended to use only `container_exclude_log`.

=== "Kubernetes"

    The following environment variables can be used

    - ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
    - ENV_INPUT_CONTAINER_CONTAINER_EXCLUDE_LOG

    to configure log collection for the container. Suppose there are three Pods whose images are:

    - A：`hello/hello-http:latest`
    - B：`world/world-http:latest`
    - C：`pubrepo.guance.com/datakit/datakit:1.2.0`

    If you want to collect only the logs of Pod A, configure  ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG.

    ``` yaml
    - env:
      - name: ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
        value: image:hello*  # Specify the image name or its wildcard
    ```

    Or namespace:

    ``` yaml
    - env:
      - name: ENV_INPUT_CONTAINER_CONTAINER_INCLUDE_LOG
        value: namespace:foo  # Specify the namespace or its wildcard
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

<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ attention

    The priority of the global configuration `container_exclude_log` is lower than the custom configuration `disable` within the container. For example, if `container_exclude_log = ["image:*"]` is configured to exclude all logs, but there is a Pod Annotation as follows:
    
    ``` JSON
    [
      {
          "disable": false,
          "type": "file",
          "path":"/tmp/opt/log",
          "source":  "logging-file",
          "tags" : {
            "some_tag": "some_value"
          }
      },
      {
          "disable": true,
          "source":  "logging-output"
      }
    ]
    ```
    
    This configuration is closer to the container and has a higher priority. The `disable=false` in the configuration indicates that log files should be collected, overriding the global configuration.

    Therefore, the log files for this container will still be collected, but the stdout/stderr console output will not be collected because of disable=true.

<!-- markdownlint-enable -->

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: Issue with Soft Links in Log Directories {#log-path-link}
<!-- markdownlint-enable -->

Normally, Datakit retrieves the path of log files from the container/Kubernetes API and collects the file accordingly.

However, in some special environments, a soft link may be created for the directory containing the log file, and Datakit is unable to know the target of the soft link in advance, which prevents it from mounting the directory and collecting the log file.

For example, suppose a container log file is located at `/var/log/pods/default_log-demo_f2617302-9d3a-48b5-b4e0-b0d59f1f0cd9/log-output/0.log`. In the current environment, `/var/log/pods` is a soft link pointing to `/mnt/container_logs`, as shown below:

```shell
root@node-01:~# ls /var/log -lh
total 284K
lrwxrwxrwx 1 root root   20 Oct  8 10:06 pods -> /mnt/container_logs/
```

To enable Datakit to collect the log file, `/mnt/container_logs` hostPath needs to be mounted. For example, the following can be added to `datakit.yaml`:

``` yaml
# .. omitted..
spec:
  containers:
  - name: datakit
    image: pubrepo.guance.com/datakit/datakit:1.16.0
    volumeMounts:
    - mountPath: /mnt/container_logs
      name: container-logs
  # .. omitted..
  volumes:
  - hostPath:
      path: /mnt/container_logs
    name: container-logs
```

This situation is not very common and is usually only executed when it is known in advance that there is a soft link in the path or when Datakit logs indicate collection errors.
<!-- markdownlint-disable MD013 -->
### :material-chat-question: Source Setting for Container Log Collection {#config-logging-source}
<!-- markdownlint-enable -->
In the container environment, the log `source` setting is a very important configuration item, which directly affects the display effect on the page. However, it would be cruel to configure a source for each container's logs one by one. Without manually configuring the container log source, DataKit has the following rule (descending priority) for automatically inferring the source of the container log:

<!-- markdownlint-disable MD046 -->
???+ attention

    The so-called not manually specifying the container log source means that it is not specified in Pod Annotation or in container.conf (currently there is no configuration item specifying the container log source in container.conf).
<!-- markdownlint-enable -->

- Container's own name: The name that can be seen through `docker ps` or `crictl ps`.
- Container name specified by Kubernetes: Obtained from the `io.kubernetes.container.name` label of the container.
- `default`: Default `source`.

<!-- markdownlint-disable MD013 -->
### :material-chat-question: Wildcard Collection of Log Files in Containers {#config-logging-source}
<!-- markdownlint-enable -->

To collect log files within a container, you need to add a configuration in Annotations/Labels and specify the `path` as follows:

```yaml
[
  {
    "disable": false,
    "type": "file",
    "path":"/tmp/opt/log",
    "source":  "logging-file",
    "tags" : {
      "some_tag": "some_value"
    }
  }
]
```

The `path` configuration supports [glob rules](logging.md#glob-rules) for batch specification. For example, if you want to collect `/var/top/mysql/1.log` and `/var/opt/mysql/errors/2.log`, you can write it like this:

```yaml
[
  {
    "disable": false,
    "type": "file",
    "path":"/tmp/opt/**/*.log",
    "source":  "logging-file",
    "tags" : {
      "some_tag": "some_value"
    }
  }
]
```

The `path` configuration uses doublestar (`**`) to match multiple directories, and `*.log` will match all files ending with `.log`. This way, log files with different directories and names will be collected.

Note that the mounting directory for the emptyDir volume must be higher than the directory to be matched. Taking the example of collecting `/tmp/opt/**/*.log`, you must mount `/tmp/opt` or a higher-level directory like `/tmp`, otherwise, the corresponding files will not be found.

## Extended Reading {#more-reading}

- [Pipeline: Text Data Processing](../developers/pipeline/index.md)
- [Overview of DataKit Log Collection](datakit-logging.md)
Therefore, the log files for this container will still be collected, but the stdout/stderr console output will not be collected because of `disable=true`.
