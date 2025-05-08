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

Before using OTEL to send Trace to Datakit, make sure you have [configured the collector](opentelemetry.md).

Configuration: [Datakit Configuration OTEL](opentelemetry.md)

## Add Dependencies {#dependencies}

Add dependencies in pom.xml

``` xml
    <!-- add opentelemetry  -->
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
        <groupId>io.grpc</groupId>
        <artifactId>grpc-netty-shaded</artifactId>
        <version>1.41.0</version>
    </dependency>
    <dependency>
        <groupId>io.opentelemetry</groupId>
        <artifactId>opentelemetry-semconv</artifactId>
        <version>1.9.0-alpha</version>
    </dependency>
    <!-- use grpc protocol -->
    <dependency>
        <groupId>io.grpc</groupId>
        <artifactId>grpc-protobuf</artifactId>
        <version>1.36.1</version>
    </dependency>

```

## Java Agent Form {#with-agent}

There are many ways you can start an Agent. Next, we will show you how to start an Agent through environment variables, command line, and Tomcat configuration.

<!-- markdownlint-disable MD029 -->
1. Start in the form of environment variables

```shell
$export JAVA_OPTS="-javaagent:PATH/TO/opentelemetry-javaagent.jar"
$ export OTEL_TRACES_EXPORTER=otlp
```

2. Command line activation

```shell
java -javaagent:opentelemetry-javaagent-1.13.1.jar \
-Dotel.traces.exporter=otlp \
-Dotel.exporter.otlp.endpoint=http://localhost:4317 \
-jar your-server.jar
```

3. Tomcat configuration form

```shell
cd <tomcat installation directory>
cd bin
vim catalina.sh
# add at second line
CATALINA_OPTS="$CATALINA_OPTS -javaagent:PATH/TO/opentelemetry-javaagent.jar -Dotel.traces.exporter=otlp"; export CATALINA_OPTS

# restart Tomcat
```
<!-- markdownlint-enable -->
When configuring the field `exporter.otlp.endpoint`, you can dispense with the configuration and use the default value (localhost: 4317), because Datakit is on the same host as the Java program, and the default port is also 4317.

## Java 2: Code Injection Form {#with-code}

``` java
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
                    .setEndpoint("http://127.0.0.1:4317")   // if setEndpoint is configure, http/https must be added
                    .setTimeout(2, TimeUnit.SECONDS)
                    //.addHeader("header1", "1") // 添加 header
                    .build();

            String s = grpcSpanExporter.toString();
            System.out.println(s);
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
            // .build();

            Tracer tracer = openTelemetry.getTracer("instrumentation-library-name", "1.0.0");
            Span parentSpan = tracer.spanBuilder("parent").startSpan();


            Span childSpan = tracer.spanBuilder("child")
                    .setParent(Context.current().with(parentSpan))
                    .startSpan();
            childSpan.setAttribute("tagsA", "vllelel");
            // do stuff
            sleep(500);    // delay 0.5 second
            for (int i = 0; i < 10; i++) {
                Span childSpan1 = tracer.spanBuilder("child")
                        .setParent(Context.current().with(parentSpan))
                        .startSpan();
                sleep(1000);    // delay 1 second
                System.out.println(i);
                childSpan1.end();
            }
            childSpan.end();
            childSpan.end(0, TimeUnit.NANOSECONDS);
            System.out.println("span end");
            sleep(1000);    //delay 1 second
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

## View Effect {#view}

Log in to [<<<custom_key.brand_name>>>](https://console.<<<custom_key.brand_main_domain>>>/tracing/service/table?time=15m){:target="_blank"} and view `application performance monitoring` -> `links` -> Click on a single `link`

![avatar](imgs/otel-java-example.png)

In the flame diagram, you can see the execution time, call flow and so on in each module.

---

## Reference {#more-readings}

- Source sample [GitHub-OpenTelemetry-Java](https://github.com/open-telemetry/opentelemetry-java){:target="_blank"}
- [Doc](https://opentelemetry.io/docs/instrumentation/go/getting-started/){:target="_blank"}
