# OpenTelemetry guance-exporter

> *作者： 宋龙奇*

观测云在 OTEL JAVA agent 中添加了一个 `guance-exporter`，该 exporter 可以将链路和指标直接发送到观测云中心。

[guance-exporter](https://github.com/GuanceCloud/guance-java-exporter){:target="_blank"} 在 GitHub 中是开源的，并且集成到了观测云二次开发的 [otel-java-agent](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"} 之中。

guance-exporter 可以将数据直接发送到观测云，也就是 `endpoint`, 发送的数据格式是 InfluxDB point。

## 下载 {#download}

从 [GitHub-Release](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/release){:target="_blank"} 中下载，版本***不低于*** `v1.26.3-guance`

### Agent 使用方式 {#agent}

```shell
java  -javaagent:/usr/local/opentelemetry-javaagent-1.26.3-guance.jar \
-Dotel.traces.exporter=guance \
-Dotel.metrics.exporter=guance \ 
-Dotel.exporter.guance.endpoint=https://openway.guance.com \ 
-Dotel.exporter.guance.token=<TOKEN> \
-jar app.jar
```

如果是 k8s :

```shell
export OTEL_TRACES_EXPORTER=guance
export OTEL_METRICS_EXPORTER=guance
export OTEL_EXPORTER_GUANCE_ENDPOINT=https://openway.guance.com
export OTEL_EXPORTER_GUANCE_TOKEN=<TOKEN>
```

参数说明：

- `guance` exporter 名称。
- `endpoint` 观测云中心地址，通常为 `https://openway.guance.com`。
- `token` 观测云用户空间 token。

注意： 不配置 `otel.metrics.exporter` 则指标不会上传，`otel.traces.exporter` 同理。但是 `endpoint` 和 `token` 是必填的。

### 集成方式 {#code-integration}

引用该 jar 包， *pom.xml* 部分如下：

```xml
<dependencies>
    <dependency>
        <groupId>io.opentelemetry</groupId>
        <artifactId>opentelemetry-sdk</artifactId>
        <version>1.26.0</version>
    </dependency>

    <dependency>
        <groupId>io.opentelemetry</groupId>
        <artifactId>opentelemetry-exporter-otlp</artifactId>
       <version>1.26.0</version>
    </dependency>

    <dependency>
        <groupId>io.opentelemetry</groupId>
        <artifactId>opentelemetry-semconv</artifactId>
        <version>1.26.0-alpha</version>
    </dependency>

    <dependency>
        <groupId>com.guance</groupId>
        <artifactId>guance-exporter</artifactId>
        <!--  请确认版本！！ -->
       <version>1.4.0</version>
    </dependency>
</dependencies>
```

版本可在 maven2 仓库中使用最新版本：[maven2-guance-exporter](https://repo1.maven.org/maven2/com/guance/guance-exporter/){:target="_blank"}

要在 `SpringBoot` 项目中初始化一个全局的 OpenTelemetry 对象，你可以创建一个单例类来管理它。以下是一个示例：

首先，创建一个名为 `OpenTelemetryManager` 的类：

```java
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Tracer;

public class OpenTelemetryManager {
    private static final OpenTelemetry OPEN_TELEMETRY = OpenTelemetryInitializer.initialize(); // 初始化 OpenTelemetry

    public static OpenTelemetry getOpenTelemetry() {
        return OPEN_TELEMETRY;
    }

    public static Tracer getTracer(String name) {
        return OPEN_TELEMETRY.getTracer(name);
    }
}
```

然后，在 `OpenTelemetryInitializer` 类中进行 `OpenTelemetry` 的初始化和配置：

```java
import com.guance.exporter.guance.trace.GuanceSpanExporter;
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;
import io.opentelemetry.semconv.resource.attributes.ResourceAttributes;

public class OpenTelemetryInitializer {
    public static OpenTelemetry initialize() {
        GuanceSpanExporter guanceExporter = new GuanceSpanExporter();
        guanceExporter.setEndpoint("https://openway.guance.com"); // dataway
        guanceExporter.setToken("tkn_0d9ebb47xxxxxxxxx");    // your token

        SdkTracerProvider tracerProvider = SdkTracerProvider.builder()
                .addSpanProcessor(BatchSpanProcessor.builder(guanceExporter).build())
                .setResource(Resource.create(Attributes.builder()
                        .put(ResourceAttributes.SERVICE_NAME, "serviceForJAVA")
                        .build()))
                .build();

        return OpenTelemetrySdk.builder()
                .setTracerProvider(tracerProvider)
                .setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance()))
                .buildAndRegisterGlobal();
    }
}
```

最后，在你的 Java 文件中，你可以直接通过 `OpenTelemetryManager` 类来获取全局的 `OpenTelemetry` 对象：

```java
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Tracer;

public class YourClass {
    private static final OpenTelemetry openTelemetry = OpenTelemetryManager.getOpenTelemetry();
    private static final Tracer tracer = OpenTelemetryManager.getTracer("your-tracer-name");

    public void yourMethod() {
        // 使用 tracer 进行跟踪
        tracer.spanBuilder("your-span").startSpan().end();
        // ...
    }
}
```

## 指标 {#metrics}

guance-exporter 支持 metric 数据发送到观测云，指标集的名字是 `otel-service` 。
