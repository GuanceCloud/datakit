---
title     : 'OpenTelemetry Java'
summary   : 'OpenTelemetry Java 集成'
tags      :
  - 'JAVA'
  - 'OTEL'
  - '链路追踪'
  - 'APM'
__int_icon: 'icon/opentelemetry'
---

使用 OpenTelemetry 时，请先完成 [OpenTelemetry 输入配置](opentelemetry.md)，再选择下文方式之一接入 Java。

## Java Agent 方式 {#with-agent}

最常用的是 Java Agent 自动埋点，启动方式可分为三类。

### 1) 环境变量 {#environment-variable-mode}

```shell
export JAVA_OPTS="-javaagent:/path/to/opentelemetry-javaagent.jar"
export OTEL_TRACES_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_ENDPOINT="http://127.0.0.1:4317"
export OTEL_SERVICE_NAME=service-name
```

### 2) 命令行启动 {#command-line-mode}

```shell
java -javaagent:/path/to/opentelemetry-javaagent.jar \
  -Dotel.traces.exporter=otlp \
  -Dotel.exporter.otlp.endpoint=http://127.0.0.1:4317 \
  -Dotel.service.name=service-name \
  -jar your-server.jar
```

### 3) Tomcat 配置 {#tomcat-mode}

```shell
cd <tomcat 安装目录>/bin
vim catalina.sh

# 在 CATALINA_OPTS 中增加
CATALINA_OPTS="$CATALINA_OPTS -javaagent:/path/to/opentelemetry-javaagent.jar -Dotel.traces.exporter=otlp -Dotel.service.name=service-name"; export CATALINA_OPTS
```

> 如果 DataKit 与应用在同主机，且使用默认端口，可不设置 `OTEL_EXPORTER_OTLP_ENDPOINT`（默认 `http://localhost:4317`）。

如果使用 OTEL Agent V2 的 HTTP transport，需要设置：`-Dotel.exporter.otlp.protocol=http/protobuf`，并为 traces/metrics/logs 配置各自 endpoint（例如 `/otel/v1/...`）。

## 代码方式接入 {#with-code}

如果不适合自动埋点，可通过代码方式集成 OTEL SDK。

Maven 依赖示例：

```xml
<dependency>
    <groupId>io.opentelemetry</groupId>
    <artifactId>opentelemetry-sdk</artifactId>
    <version>1.9.0</version>
</dependency>
<dependency>
    <groupId>io.opentelemetry</groupId>
    <artifactId>opentelemetry-exporter-otlp</artifactId>
    <version>1.9.0</version>
</dependency>
<dependency>
    <groupId>io.opentelemetry</groupId>
    <artifactId>opentelemetry-semconv</artifactId>
    <version>1.9.0-alpha</version>
</dependency>
<dependency>
    <groupId>io.grpc</groupId>
    <artifactId>grpc-netty-shaded</artifactId>
    <version>1.41.0</version>
</dependency>
<dependency>
    <groupId>io.grpc</groupId>
    <artifactId>grpc-protobuf</artifactId>
    <version>1.36.1</version>
</dependency>
```

示例代码：

```java
package com.example;

import io.opentelemetry.api.OpenTelemetry;
import io.opentelemetry.api.common.Attributes;
import io.opentelemetry.api.trace.Span;
import io.opentelemetry.api.trace.Tracer;
import io.opentelemetry.api.trace.propagation.W3CTraceContextPropagator;
import io.opentelemetry.context.Context;
import io.opentelemetry.context.propagation.ContextPropagators;
import io.opentelemetry.exporter.otlp.trace.OtlpGrpcSpanExporter;
import io.opentelemetry.sdk.OpenTelemetrySdk;
import io.opentelemetry.sdk.resources.Resource;
import io.opentelemetry.semconv.resource.attributes.ResourceAttributes;
import io.opentelemetry.sdk.trace.SdkTracerProvider;
import io.opentelemetry.sdk.trace.export.BatchSpanProcessor;
import java.util.concurrent.TimeUnit;
import static java.lang.Thread.sleep;

public class otlpdemo {
    public static void main(String[] args) {
        try {
            OtlpGrpcSpanExporter grpcSpanExporter = OtlpGrpcSpanExporter.builder()
                    .setEndpoint("http://127.0.0.1:4317")
                    .setTimeout(2, TimeUnit.SECONDS)
                    .build();

            SdkTracerProvider tracerProvider = SdkTracerProvider.builder()
                    .addSpanProcessor(BatchSpanProcessor.builder(grpcSpanExporter).build())
                    .setResource(Resource.create(Attributes.builder()
                        .put(ResourceAttributes.SERVICE_NAME, "serviceForJAVA")
                        .put(ResourceAttributes.SERVICE_VERSION, "1.0.0")
                        .put(ResourceAttributes.HOST_NAME, "host")
                        .build()))
                    .build();

            OpenTelemetry openTelemetry = OpenTelemetrySdk.builder()
                    .setTracerProvider(tracerProvider)
                    .setPropagators(ContextPropagators.create(W3CTraceContextPropagator.getInstance()))
                    .buildAndRegisterGlobal();

            Tracer tracer = openTelemetry.getTracer("instrumentation-library-name", "1.0.0");
            Span parentSpan = tracer.spanBuilder("parent").startSpan();


            Span childSpan = tracer.spanBuilder("child")
                    .setParent(Context.current().with(parentSpan))
                    .startSpan();
            childSpan.setAttribute("tagsA", "example");
            sleep(500);
            for (int i = 0; i < 10; i++) {
                Span childSpan1 = tracer.spanBuilder("child")
                        .setParent(Context.current().with(parentSpan))
                        .startSpan();
                sleep(1000);
                System.out.println(i);
                childSpan1.end();
            }
            childSpan.end();
            childSpan.end(0, TimeUnit.NANOSECONDS);
            System.out.println("span end");
            sleep(1000);
            parentSpan.end();
            tracerProvider.shutdown();

        } catch (InterruptedException e) {
            e.printStackTrace();
        } finally {
            System.out.println("finally end");
        }
    }
}
```


## 参考 {#more-readings}

- [OpenTelemetry Java 示例](https://github.com/open-telemetry/opentelemetry-java){:target="_blank"}
- [OpenTelemetry Java 官方文档](https://opentelemetry.io/docs/instrumentation/java/getting-started/){:target="_blank"}
