---
title     : 'OpenTelemetry Java'
summary   : 'Tracing Java applications with OpenTelemetry'
tags      :
  - 'JAVA'
  - 'OTEL'
  - 'APM'
  - 'TRACING'
__int_icon: 'icon/opentelemetry'
---

After DataKit OpenTelemetry input is configured ([OpenTelemetry](opentelemetry.md)), Java apps can send traces through Agent mode or SDK mode.

## Java Agent {#with-agent}

The fastest way is the Java Agent, which instruments libraries automatically.

### 1) Environment variable mode {#environment-variable-mode}

```shell
export JAVA_OPTS="-javaagent:/path/to/opentelemetry-javaagent.jar"
export OTEL_TRACES_EXPORTER=otlp
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_ENDPOINT="http://127.0.0.1:4317"
export OTEL_SERVICE_NAME=service-name
```

### 2) Command line mode {#command-line-mode}

```shell
java -javaagent:/path/to/opentelemetry-javaagent.jar \
  -Dotel.traces.exporter=otlp \
  -Dotel.exporter.otlp.protocol=grpc \
  -Dotel.exporter.otlp.endpoint=http://127.0.0.1:4317 \
  -Dotel.service.name=service-name \
  -jar your-server.jar
```

### 3) Tomcat mode {#tomcat-mode}

```shell
cd <tomcat-install-dir>/bin
vim catalina.sh

# add to CATALINA_OPTS
CATALINA_OPTS="$CATALINA_OPTS -javaagent:/path/to/opentelemetry-javaagent.jar -Dotel.traces.exporter=otlp -Dotel.exporter.otlp.protocol=grpc -Dotel.exporter.otlp.endpoint=http://127.0.0.1:4317 -Dotel.service.name=service-name"; export CATALINA_OPTS
```

Always set the OTLP protocol and endpoint explicitly to avoid SDK- or Agent-version defaults. The examples above use DataKit's OTLP/gRPC receiver on port `4317`. Java Agent 2.x defaults to `http/protobuf`; if you keep that protocol, set the traces, metrics, and logs endpoints to `/otel/v1/traces`, `/otel/v1/metrics`, and `/otel/v1/logs` on DataKit port `9529`.

## Code instrumentation {#with-code}

If bytecode instrumentation is not possible, integrate OTEL SDK directly in code.

Maven dependency example:

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

Example code:

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
            sleep(500); // do work
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


## Reference {#more-readings}

- [OpenTelemetry Java sample](https://github.com/open-telemetry/opentelemetry-java){:target="_blank"}
- [OpenTelemetry Java instrumentation docs](https://opentelemetry.io/docs/instrumentation/java/getting-started/){:target="_blank"}
