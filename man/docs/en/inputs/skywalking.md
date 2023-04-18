
# SkyWalking
---

:fontawesome-brands-linux: :fontawesome-brands-windows: :fontawesome-brands-apple: :material-kubernetes: :material-docker:

---

The SkyWalking Agent embedded in Datakit is used to receive, compute and analyze Skywalking Tracing protocol data.

## SkyWalking Doc {#doc}

> APM v8.8. 3 is currently incompatible and cannot be used. V8.5. 0 v8.6. 0 v8.7. 0 is currently supported.

- [Quickstart](https://skywalking.apache.org/docs/skywalking-showcase/latest/readme/){:target="_blank"}
- [Docs](https://skywalking.apache.org/docs/){:target="_blank"}
- [Clients Download](https://skywalking.apache.org/downloads/){:target="_blank"}
- [Souce Code](https://github.com/apache/skywalking){:target="_blank"}

## Configure SkyWalking Client {#client-config}

Open file /path_to_skywalking_agent/config/agent.config to configure.

```conf
# The service name in UI
agent.service_name=${SW_AGENT_NAME:your-service-name}
# Backend service addresses.
collector.backend_service=${SW_AGENT_COLLECTOR_BACKEND_SERVICES:<datakit-ip:skywalking-agent-port>}
```

## Configure SkyWalking Agent {#agent-config}

=== "Host Installation"

    Go to the `conf.d/skywalking` directory under the DataKit installation directory, copy `skywalking.conf.sample` and name it `skywalking.conf`. Examples are as follows:
    
    ```toml
        
    [[inputs.skywalking]]
      ## Skywalking grpc server listening on address.
      address = "localhost:11800"
    
      ## plugins is a list contains all the widgets used in program that want to be regarded as service.
      ## every key words list in plugins represents a plugin defined as special tag by skywalking.
      ## the value of the key word will be used to set the service name.
      # plugins = ["db.type"]
    
      ## customer_tags is a list of keys contains keys set by client code like span.SetTag(key, value)
      ## that want to send to data center. Those keys set by client code will take precedence over
      ## keys in [inputs.skywalking.tags]. DOT(.) IN KEY WILL BE REPLACED BY DASH(_) WHEN SENDING.
      # customer_tags = ["key1", "key2", ...]
    
      ## Keep rare tracing resources list switch.
      ## If some resources are rare enough(not presend in 1 hour), those resource will always send
      ## to data center and do not consider samplers and filters.
      # keep_rare_resource = false
    
      ## Ignore tracing resources map like service:[resources...].
      ## The service name is the full service name in current application.
      ## The resource list is regular expressions uses to block resource names.
      ## If you want to block some resources universally under all services, you can set the
      ## service name as "*". Note: double quotes "" cannot be omitted.
      # [inputs.skywalking.close_resource]
        # service1 = ["resource1", "resource2", ...]
        # service2 = ["resource1", "resource2", ...]
        # "*" = ["close_resource_under_all_services"]
        # ...
    
      ## Sampler config uses to set global sampling strategy.
      ## sampling_rate used to set global sampling rate.
      # [inputs.skywalking.sampler]
        # sampling_rate = 1.0
    
      # [inputs.skywalking.tags]
        # key1 = "value1"
        # key2 = "value2"
        # ...
    
      ## Storage config a local storage space in hard dirver to cache trace data.
      ## path is the local file path used to cache data.
      ## capacity is total space size(MB) used to store data.
      # [inputs.skywalking.storage]
        # path = "./skywalking_storage"
        # capacity = 5120
    
    ```
    
    For all of the following data collections, a global tag named `host` is appended by default (the tag value is the host name of the DataKit), or other tags can be specified in the configuration by `[inputs.skywalking.tags]`:
    
    ```toml
     [inputs.skywalking.tags]
      # some_tag = "some_value"
      # more_tag = "some_other_value"
      # ...
    ```

=== "Kubernetes"

    The collector can now be turned on by [ConfigMap Injection Collector Configuration](datakit-daemonset-deploy.md#configmap-setting).

## Restart Java Client {#start-java}

```command
java -javaagent:/path/to/skywalking/agent -jar /path/to/your/service.jar
```

## Send Log to Datakit {#logging}
- log4j2

The toolkit dependency package is added to the maven or gradle.
```xml
	<dependency>
      	<groupId>org.apache.skywalking</groupId>
      	<artifactId>apm-toolkit-log4j-2.x</artifactId>
      	<version>{project.release.version}</version>
	</dependency>
```

Sent through grpc protocol:
```xml
<GRPCLogClientAppender name="grpc-log">
        <PatternLayout pattern="%d{HH:mm:ss.SSS} %-5level %logger{36} - %msg%n"/>
    </GRPCLogClientAppender>
```

Others:

- [log4j-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-log4j-1.x.md){:target="_blank"}
- [logback-1.x](https://github.com/apache/skywalking-java/blob/main/docs/en/setup/service-agent/java-agent/Application-toolkit-logback-1.x.md){:target="_blank"}


## SkyWalking JVM Measurement {#jvm-measurements}



jvm metrics collected by skywalking language agent.

- Tag


| Tag Name | Description    |
|  ----  | --------|
|`service`|service name|

- Metrics List


| Metrics | Description| Data Type | Unit   |
| ---- |---- | :---:    | :----: |
|`class_loaded_count`|loaded class count.|int|count|
|`class_total_loaded_count`|total loaded class count.|int|count|
|`class_total_unloaded_class_count`|total unloaded class count.|int|count|
|`cpu_usage_percent`|cpu usage percentile|float|percent|
|`gc_phrase_old/new_count`|gc old or new count.|int|count|
|`heap/stack_committed`|heap or stack committed amount of memory.|int|count|
|`heap/stack_init`|heap or stack initialized amount of memory.|int|count|
|`heap/stack_max`|heap or stack max amount of memory.|int|count|
|`heap/stack_used`|heap or stack used amount of memory.|int|count|
|`pool_*_committed`|committed amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).|int|count|
|`pool_*_init`|initialized amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).|int|count|
|`pool_*_max`|max amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).|int|count|
|`pool_*_used`|used amount of memory in variety of pool(code_cache_usage,newgen_usage,oldgen_usage,survivor_usage,permgen_usage,metaspace_usage).|int|count|
|`thread_blocked_state_count`|blocked state thread count|int|count|
|`thread_daemon_count`|thread daemon count.|int|count|
|`thread_live_count`|thread live count.|int|count|
|`thread_peak_count`|thread peak count.|int|count|
|`thread_runnable_state_count`|runnable state thread count.|int|count|
|`thread_time_waiting_state_count`|time waiting state thread count.|int|count|
|`thread_waiting_state_count`|waiting state thread count.|int|count|


