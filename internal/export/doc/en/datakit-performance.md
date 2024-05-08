# Datakit Performance Explanation

---

:material-kubernetes: :fontawesome-brands-linux:

---

This document primarily showcases Datakit's performance in a real production environment. You can use the data presented here as a reference to benchmark your own environment.

## Basic Environment Information {#specs}

- Operating Environment: Kubernetes
- Datakit Version: 1.28.1
- Resource Limitations: 2C4G (2 CPU cores and 4GB of RAM)
- Runtime: 1.73 days
- Data Sources: There are numerous applications running in the cluster, and Datakit actively collects various metrics, logs, and Tracing data from these applications.

The following lists the load conditions of Datakit under high, medium, and low scenarios [^1].

<!-- markdownlint-disable MD046 -->
=== "High Load"

    Under high load conditions, the amount of data collected by Datakit and the complexity of the data itself are relatively high, which consumes more computing resources.
    
    - CPU Utilization
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/cpu-usage.png){ width="800" }
      </figure>
    
      Due to the limitation on the number of CPU cores used, the CPU usage here stabilizes around 200%.
    
    - Memory Consumption
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/mem-usage.png){ width="800" }
      </figure>
    
      The memory is limited to 4GB, which is quite close to the limit. If the memory limit is exceeded, the Datakit Pod will be restarted due to Out Of Memory (OOM).
    
      <!-- The following are the conditions for data collection, Pipeline processing, and data uploading (the interval for irate is 30 seconds). -->
    
      - Data Collection Points
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/collect-pts-irate.png){ width="800" }
      </figure>
    
      The number of data points collected by each collector, with the top one being a Prometheus metric collection that collects a relatively large number of data points each time, followed by the Tracing collector.
    
      - Pipeline Processing Points
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/pl-pts-irate.png){ width="800" }
      </figure>
    
      This shows the situation of a specific Pipeline processing data points, with one Pipeline script (*kodo.p*) being particularly busy.
    
      - Network Upload
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/upload-bytes-irate.png){ width="800" }
      </figure>
    
    The collected data points ultimately need to be sent to the center via the network. What is shown here is the upload of the GZip-compressed data point payload (HTTP Body) [^2]. Tracing has a particularly large payload because it contains a large amount of text information.

=== "Medium Load"

    Under medium load conditions, Datakit's resource consumption is significantly reduced.
    
    - CPU Utilization
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/mid-cpu-usage.png){ width="800" }
      </figure>
    
    - Memory Consumption
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/mid-mem-usage.png){ width="800" }
      </figure>
    
    - Data Collection Points
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/mid-collect-pts-irate.png){ width="800" }
      </figure>
    
    - Pipeline Processing Points
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/mid-pl-pts-irate.png){ width="800" }
      </figure>
    
    - Network Upload
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/mid-upload-bytes-irate.png){ width="800" }
      </figure>

=== "Low Load"

    Under low load conditions, Datakit only activates the basic [default collectors](datakit-input-conf.md#default-enabled-inputs), with a relatively small amount of data, thus occupying less memory [^3].
    
    - CPU Utilization
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/low-cpu-usage.png){ width="800" }
      </figure>
    
    - Memory Consumption
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/low-mem-usage.png){ width="800" }
      </figure>
    
    - Data Collection Points
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/low-collect-pts-irate.png){ width="800" }
      </figure>
    
    - Network Upload
    
      <figure markdown>
      ![](https://static.guance.com/images/datakit/performance/low-upload-bytes-irate.png){ width="800" }
      </figure>

<!-- markdownlint-enable -->

<!-- markdownlint-disable MD053 -->
[^1]: Datakit has enabled [Point Pool](datakit-conf.md#point-pool) and is using [V2 encoding](datakit-conf.md#dataway-settings) for uploads.
[^2]: This value may differ slightly from the Pod traffic, as the Pod statistics represent the network traffic information at the Kubernetes level, which will be larger than the traffic shown here.
[^3]: The low-load Datakit was tested on an additional Linux server, which only enabled basic collectors. Since there was no Pipeline involved, there is no corresponding data.
<!-- markdownlint-enable -->
