
# DataKit 基础运行情况以及性能说明

---

:material-kubernetes: :fontawesome-brands-linux:

---

本文档主要展示 DataKit 在真实生产环境下的运行表现，大家可依据这里展示的数据作为参考，对标各自的环境。

## 基础环境信息 {#specs}

- 运行环境     ： Kubernetes
- DataKit 版本 ： 1.28.1
- 资源限制     ： 2C4G
- 运行时长     ： 1.73d
- 数据源       ： 集群中有大量的应用在运行，DataKit 会主动采集各种应用的指标、日志以及 Tracing 数据

以下列举了高、中、低三种情况下的 DataKit 负载情况 [^1]。

<!-- markdownlint-disable MD046 -->
=== "高负载"

    高负载情况下，DataKit 采集的数据量以及数据本身都比较复杂，会消耗更多的计算资源。
    
    - CPU 占比
    
    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/cpu-usage.png){ width="800" }
    </figure>
    
    由于限制了 CPU 使用核数，这里 CPU 稳定在 200% 附近。

    - 内存消耗
    
    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/mem-usage.png){ width="800" }
    </figure>
    
    内存限制了 4GB，这里已经比较接近限制。当超过内存限制，DataKit POD 将被 OOM 重启。
    
    <!-- 以下是数据采集、Pipeline 处理以及数据上传的情况（此处 irate 间隔为 30S）。 -->
    
    - 数据采集点数
    
    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/collect-pts-irate.png){ width="800" }
    </figure>
    
    每个采集器采集的数据点数，排在首位的是一个 Prometheus 指标采集，每次采集的数据点比较多，其次是 Tracing 采集器。
    
    - Pipeline 处理点数

    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/pl-pts-irate.png){ width="800" }
    </figure>
    
    这里是某个具体 Pipeline 处理数据点的情况，有一个 Pipeline 脚本（*kodo.p*）业务比较繁忙。
    
    - 网络发送

    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/upload-bytes-irate.png){ width="800" }
    </figure>
    
    采集到的数据点最终都要通过网络发送到中心，这里展示的是 GZip 之后的数据点 Payload （HTTP Body）上传 [^2] 情况。Tracing 因为其含有大量文本信息，所以其 Payload 特别大。

=== "中负载"

    中度负载情况下，DataKit 的资源消耗大大降低。
    
    - CPU 占比
    
    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/mid-cpu-usage.png){ width="800" }
    </figure>

    - 内存消耗

    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/mid-mem-usage.png){ width="800" }
    </figure>
    
    - 数据采集点数
    
    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/mid-collect-pts-irate.png){ width="800" }
    </figure>
    
    - Pipeline 处理点数

    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/mid-pl-pts-irate.png){ width="800" }
    </figure>
    
    - 网络发送

    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/mid-upload-bytes-irate.png){ width="800" }
    </figure>

=== "低负载"

    低负载情况下，DataKit 只开启了基本的[默认采集器](datakit-input-conf.md#default-enabled-inputs)，其数据量比较小，所以占用较小的内存 [^3]。

    - CPU 占比
    
    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/low-cpu-usage.png){ width="800" }
    </figure>

    - 内存消耗

    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/low-mem-usage.png){ width="800" }
    </figure>

    - 数据采集点数
    
    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/low-collect-pts-irate.png){ width="800" }
    </figure>

    - 网络发送

    <figure markdown>
    ![](https://static.<<<custom_key.brand_main_domain>>>/images/datakit/performance/low-upload-bytes-irate.png){ width="800" }
    </figure>

<!-- markdownlint-enable -->

<!-- markdownlint-disable MD053 -->
[^1]: DataKit 都开启了 [Point Pool](datakit-conf.md#point-pool)，且使用 [V2 的编码](datakit-conf.md#dataway-settings)上传
[^2]: 该数值跟 Pod 流量会有一定的出入，Pod 统计的是 Kubernetes 层面网络流量信息，它的值会比此处的流量要大
[^3]: 该低负载的 DataKit 是在额外的一台 Linux 服务器上测试的，它只开启了基础的采集器。由于没有 Pipeline 参与，所以没有对应的数据
<!-- markdownlint-enable -->
