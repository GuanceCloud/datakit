
# DataKit 已有指标集列表

---

DataKit 内置的数据采集能采集到如下这些指标。以下这些指标在开启采集器之后，不一定全部都能采集到，某些指标的采集需要满足一定的条件才能采集到。

如果您有自定义的数据要上报给 DataKit，可以参考这个列表，以规避一些已有的指标集名字（相同的指标集名字容易导致数据混乱）。

<!-- markdownlint-disable MD046 -->
???+ info

    - 对于日志而言，下表中的指标集名字对应 Studio 中的 `source` 字段。对对象/自定义对象而言，则对应 Studio 中的 `class` 字段。
<!-- markdownlint-enable -->

{{ .AllMeasurements }}
