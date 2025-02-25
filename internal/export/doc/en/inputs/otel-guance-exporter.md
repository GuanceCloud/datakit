---
title     : 'OpenTelemetry Exportor for Guance Cloud'
summary   : 'Export OpenTelemetry data to GuanCe Cloud directly'
__int_icon: 'icon/opentelemetry'
tags      :
  - 'OTEL'
---

GuanCe Cloud has added a `guance-exporter` in the OTEL JAVA Agent, which can send traces and metrics directly to the GuanCe Cloud Center.

[guance-exporter](https://github.com/GuanceCloud/guance-java-exporter){:target="_blank"} is open source on GitHub and is integrated into the Guance Cloud's secondarily developed [otel-java-agent](https://github.com/GuanceCloud/opentelemetry-java-instrumentation){:target="_blank"}.

The `guance-exporter` can send data directly to GuanCe Cloud, that is, the endpoint, and the format of the sent data is InfluxDB point.

## Download {#download}

Download from [GitHub-Release](https://github.com/GuanceCloud/opentelemetry-java-instrumentation/release){:target="_blank"}, the version is **not lower than** v1.26.3-guance.

### Agent Usage {#agent}

```shell
java  -javaagent:/usr/local/opentelemetry-javaagent-1.26.3-guance.jar \
-Dotel.traces.exporter=guance \
-Dotel.metrics.exporter=guance \ 
-Dotel.exporter.guance.endpoint=https://openway.guance.com \ 
-Dotel.exporter.guance.token=<TOKEN> \
-jar app.jar
```

for k8s:

```shell
export OTEL_TRACES_EXPORTER=guance
export OTEL_METRICS_EXPORTER=guance
export OTEL_EXPORTER_GUANCE_ENDPOINT=https://openway.guance.com
export OTEL_EXPORTER_GUANCE_TOKEN=<TOKEN>
```

Parameter Description:

- `guance` exporter name.
- `endpoint` GuanCe Cloud Center address, usually `https://openway.guance.com`.
- `token` GuanCe Cloud user space token.

Note: If `otel.metrics.exporter` is not configured, metrics will not be uploaded, the same for `otel.traces.exporter`. However, `endpoint` and `token` are required.

### Integration {#code-integration}

Reference the jar package, the pom.xml section is as follows:

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
        <!--  Please confirm the version!! -->
       <version>1.4.0</version>
    </dependency>
</dependencies>
```

The version can be used in the maven2 repository with the latest version: [maven2-guance-exporter](https://repo1.maven.org/maven2/com/guance/guance-exporter/){:target="_blank"}

To initialize a global OpenTelemetry object in a SpringBoot project, you can create a singleton class to manage it. Here is an example:

First, create a class named OpenTelemetryManager:

```java
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Tracer;

public class OpenTelemetryManager {
    private static final OpenTelemetry OPEN_TELEMETRY = OpenTelemetryInitializer.initialize();

    public static OpenTelemetry getOpenTelemetry() {
        return OPEN_TELEMETRY;
    }

    public static Tracer getTracer(String name) {
        return OPEN_TELEMETRY.getTracer(name);
    }
}
```

Then, in the OpenTelemetryInitializer class, perform the initialization and configuration of OpenTelemetry:

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

Finally, in your Java files, you can directly obtain the global OpenTelemetry object through the `OpenTelemetryManager` class:

```java
import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.trace.Tracer;

public class YourClass {
    private static final OpenTelemetry openTelemetry = OpenTelemetryManager.getOpenTelemetry();
    private static final Tracer tracer = OpenTelemetryManager.getTracer("your-tracer-name");

    public void yourMethod() {
        // use tracer for tracing
        tracer.spanBuilder("your-span").startSpan().end();
        // ...
    }
}
```

## Metric {#metrics}

`guance-exporter` supports sending metric data to GuanCe Cloud, and the name of the metric set is `otel_service`.
