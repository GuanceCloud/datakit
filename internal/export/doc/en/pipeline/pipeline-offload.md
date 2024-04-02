# DataKit Pipeline Offload

[:octicons-tag-24: Version-1.9.2](changelog.md#cl-1.9.2) ·
[:octicons-beaker-24: Experimental](index.md#experimental)

---

You can use DataKit's Pipeline Offload function to reduce high data latency and high host load caused by data processing.

## Configuration Method

It needs to be configured and enabled in the `datakit.conf` main configuration file. See below for the configuration. Currently supported targets `receiver` are `datakit-http` and `ploffload`, which allows multiple `DataKit` addresses to be configured to achieve load balancing.

Notice:

- Currently only supports unloading **logging (`Logging`) category** data processing tasks;
- **The address of the current `DataKit` cannot be filled in the `addresses` configuration item**, otherwise a loop will be formed, causing the data to always be in the current `DataKit`;
- Please make the `DataWay` configuration of the target `DataKit` consistent with the current `DataKit`, otherwise the data recipient sends to its `DataWay` address;
- If `receiver` is configured as `ploffload`, the DataKit on the receiving end needs to have the `ploffload` collector enabled.

> Please check whether the target network address is locally accessible. The target cannot be reached if it is listening on the loopback address.

Reference configuration:

```txt
[pipeline]

  # Offload data processing tasks to post-level data processors.
  [pipeline.offload]
    receiver = "datakit-http"
    addresses = [
      # "http://<ip>:<port>"
    ]
```

If the receiving end DataKit turns on the `ploffload` collector, it can be configured as:

```txt
[pipeline]

  # Offload data processing tasks to post-level data processors.
  [pipeline.offload]
    receiver = "ploffload"
    addresses = [
      # "http://<ip>:<port>"
    ]
```

## Working Principle

After `DataKit` finds the `Pipeline` data processing script, it will judge whether it is a remote script from `Observation Cloud`, and if so, forward the data to the post-level data processor for processing (such as `DataKit`). The load balancing method is round robin.

![`pipeline-offload`](img/pipeline-offload.drawio.png)

## Deploy post-level data processor

There are several ways to deploy the data processor (DataKit) for receiving computing tasks:

- host deployment

  DataKit dedicated to data processing is not currently supported; host deployment DataKit see [documentation](../../datakit/datakit-install.md)

- container deployment

  The environment variables `ENV_DATAWAY` and `ENV_HTTP_LISTEN` need to be set, and the DataWay address needs to be consistent with the DataKit configured with the Pipeline Offload function; it is recommended to map the listening port of the DataKit running in the container to the host.

  Reference command:

  ```sh
  docker run --ulimit nofile=64000:64000  -e ENV_DATAWAY="https://openway.guance.com?token=<tkn_>" -e ENV_HTTP_LISTEN="0.0.0.0:9529" \
  -p 9590:9529 -d pubrepo.guance.com/datakit/datakit:<version>
  ```
