---
title     : 'DDTrace Java'
summary   : 'DDTrace Java 集成'
tags      :
  - 'DDTRACE'
  - 'JAVA'
  - '链路追踪'
__int_icon: 'icon/ddtrace'
---


Java 的 APM 接入很方便，无需修改业务代码，只需要注入对应的 agent 即可。

## 前置条件 {#requirements}

安装 DataKit 并开启 [DDTrace 采集器](ddtrace.md){:target="_blank"} ，如果想要采集 JVM 运行时的一些指标需要开启 [StatsD 采集器](statsd.md){:target="_blank"} 。

要求 jdk 版本在 1.8 及以上。

## 安装依赖 {#dependence}

<!-- markdownlint-disable MD046 -->
=== "<<<custom_key.brand_name>>>版本"

    为了增加更多中间件支持，<<<custom_key.brand_name>>>增强了 [DDTrace-Java 实现](ddtrace-ext-java.md)。

    ```shell
    wget -O dd-java-agent.jar 'https://static.<<<custom_key.brand_main_domain>>>/dd-image/dd-java-agent.jar'
    ```

=== "Datadog 版本"

    ```shell
    wget -O dd-java-agent.jar 'https://dtdg.co/latest-java-tracer'
    ```
<!-- markdownlint-enable -->

## 运行应用 {#instrument}

<!-- markdownlint-disable MD046 -->
=== "主机应用"

    可以通过多种途径运行你的 Java Code，如 IDE，Maven，Gradle 或直接通过 `java -jar` 命令，以下通过 `java` 命令启动应用：
    
    ```shell hl_lines="2-7" linenums="1"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.logs.injection=true \
      -Ddd.service.name=<YOUR-SERVICE-NAME> \
      -Ddd.env=<YOUR-ENV-NAME> \
      -Ddd.agent.host=<YOUR-DATAKIT-HOST> \
      -Ddd.trace.agent.port=9529 \
      -jar path/to/your/app.jar
    ```
    
    此处请补充您的各种基本参数配置 `<YOUR-...>`，除此之外，还有如下一些可选参数：
    
    ### 开启 profiling {#instrument-profiling}

    > 此处需开启 [Profiling 采集器](profile.md)。
    
    开启 Profiling 后，我们能看到更多 Java 运行时的信息：
    
    ```shell linenums="1" hl_lines="3-4"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.profiling.enabled=true \
      -XX:FlightRecorderOptions=stackdepth=256 \
      ...
    ```

    ### 开启采样率 {#instrument-sampling}
    
    我们可以开启采样率，来降低实际产生的数据量：

    ```shell hl_lines="3" linenums="1"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.trace.sample.rate=0.8 \
      ...
    ```

    ### 开启 JVM 指标采集 {#instrument-jvm-metrics}
    
    > 此处需开启 [statsd 采集器](statsd.md)。
    
    ```shell hl_lines="3-6" linenums="1"
    java \
      -javaagent:/path/to/dd-java-agent.jar \
      -Ddd.jmxfetch.enabled=true \
      -Ddd.jmxfetch.check-period=1000 \
      -Ddd.jmxfetch.statsd.host=<YOUR-DATAKIT-HOST>  \
      -Ddd.jmxfetch.statsd.port=8125 \
      ...
    ```

=== "Kubernetes"

    在 Kubernetes 中，可以通过 [DataKit Operator 来注入 trace agent](../datakit/datakit-operator.md#datakit-operator-inject-lib)，也可以手动挂载 trace agent 到应用容器中。

    ```yaml hl_lines="10-19" linenums="1"
    apiVersion: apps/v1
    kind: Deployment
    spec:
      template:
        spec:
          containers:
            - name: <CONTAINER_NAME>
              image: <CONTAINER_IMAGE>/<TAG>
              env:
                - name: DD_AGENT_HOST
                  value: "datakit-service.datakit.svc"
                - name: DD_TRACE_AGENT_PORT
                  value: "9529"
                - name: DD_ENV
                  value: <YOUR-ENV-NAME>
                - name: DD_SERVICE
                  value: <YOUR-SERVICE-NAME>
                - name: DD_LOGS_INJECTION
                  value: "true"
    ```

    更多其它参数设置，参见下面[参数解释](ddtrace-java.md#start-options) 中的对应 ENV 字段。
<!-- markdownlint-enable -->

## 参数解释 {#start-options}

以下是各个命令行参数的说明以及对应的环境变量配置方式，完整参数支持，参见 [DataDog 官方文档](https://docs.datadoghq.com/tracing/trace_collection/library_config/java){:target="_blank"}。

- **`dd.env`**

    **ENV**：`DD_ENV`

    为服务设置环境信息，比如 `testing/prod` 等。

- **`dd.version`**

    **ENV**：`DD_VERSION`

    APP 版本号

- **`dd.service.name`**

    设置服务名
    **ENV**：`DD_SERVICE`

- **`dd.trace.agent.timeout`**

    **ENV**：`DD_TRACE_AGENT_TIMEOUT`

    客户端网络发送超时默认 10s

- **`dd.logs.injection`**

    **ENV**：`DD_LOGS_INJECTION`

    是否开启 Java 应用日志注入，让日志与链路数据进行关联，默认为 true

- **`dd.tags`**

    **ENV**：`DD_TAGS`

    为每个 Span 添加默认 Tags

- **`dd.agent.host`**

    **ENV**：`DD_AGENT_HOST`

    DataKit 监听的地址名，默认 localhost

- **`dd.trace.agent.port`**

    **ENV**：`DD_TRACE_AGENT_PORT`

    DataKit 监听的端口号，默认 9529

- **`dd.trace.sample.rate`**

    **ENV**：`DD_TRACE_SAMPLE_RATE`

    设置采样率从 0.0(0%) ~ 1.0(100%)

- **`dd.jmxfetch.enabled`

    **ENV**：`DD_JMXFETCH_ENABLED`

    开启 JMX metrics 采集，默认值 true

- **`dd.jmxfetch.config.dir`**

    **ENV**：`DD_JMXFETCH_CONFIG_DIR`

    额外的 JMX metrics 采集配置目录。Java Agent 将会在 yaml 配置文件中的 instance section 寻找 `jvm_direct:true` 来修改配置


- **`dd.jmxfetch.config`**

    **ENV**：`DD_JMXFETCH_CONFIG`

    额外的 JMX metrics 采集配置文件。JAVA agent 将会在 yaml 配置文件中的 instance section 寻找 `jvm_direct: true` 来修改配置

- **`dd.jmxfetch.check-period`**

    **ENV**：`DD_JMXFETCH_CHECK_PERIOD`

    JMX metrics 发送频率(ms)，默认值 1500

- **`dd.jmxfetch.refresh-beans-period`**

    **ENV**：`DD_JMXFETCH_REFRESH_BEANS_PERIOD`

    刷新 JMX beans 频率(s)，默认值 600

- **`dd.jmxfetch.statsd.host`**

    **ENV**：`DD_JMXFETCH_STATSD_HOST`

    Statsd 主机地址用来接收 JMX metrics，如果使用 Unix Domain Socket 请使用形如 `unix://PATH_TO_UDS_SOCKET` 的主机地址。默认值同 `agent.host`

- **`dd.jmxfetch.statsd.port`**

    **ENV**：`DD_JMXFETCH_STATSD_PORT`

    StatsD 端口号用来接收 JMX metrics ，如果使用 Unix Domain Socket 请使填写 0。默认值同 agent.port

- **`dd.profiling.enabled`**

    **ENV**：`DD_PROFILING_ENABLED`

    开启 Profiling 控制，开启之后，Java 应用运行期间的 Profiling 信息也会收集，并且上报给 DataKit

## 链路错误情况说明 {#error}

一个 span 表示了一个单一的逻辑操作单元，它可以是一个数据库查询、一个 HTTP 请求或者是任何其他类型的操作。当这个操作出现问题或者未按预期执行时，span 的状态就会被标记为 `error`。

### 错误产生的原因 {#error_reason}

具体来说，span 会在以下情况下被标记为 `error` 状态：

1. **异常抛出**：如果在 span 的执行期间，有任何异常被抛出（不论是已检查异常还是运行时异常），并且这个异常没有被应用代码恰当地处理（被捕获而没有重新抛出），那么这个 span 会被标记为 `error`。这是最常见的情况。
1. **手动标记**：开发者可以通过 Datadog 提供的 SDK 手动将某个 span 标记为 `error`。这在某些逻辑错误或特殊条件下非常有用，即使这些情况并不抛出异常。
1. **HTTP 请求错误**：对于 HTTP 客户端或服务端的 span，如果响应状态码表示一个错误（例如，4xx 或 5xx），span 通常会被标记为 `error`。这取决于具体的集成和配置。
1. **自定义错误条件**：通过配置或自定义的集成，开发者可以定义特定的条件或检查来标记 span 为错误。例如，如果一个数据库查询返回的结果不符合预期的格式或内容，即使没有抛出异常，也可以将其标记为 `error`。
1. **超时**：某些操作可能因为超时而失败，如数据库查询、远程调用等。如果这些操作被配置为有超时限制，且实际执行超过了这个限制，span 也可能被标记为 `error`。

标记 span 为 `error` 状态是为了帮助开发者和运维人员快速定位和解决应用中的问题。通过观察那些被标记为错误的 spans，我们可以更容易地发现性能瓶颈、失败的服务调用、异常抛出等问题，并采取相应的优化或修复措施。

### 常见的错误类型 {#error_type}

在 Java 编程中，异常（Exception）是在程序执行期间发生的问题，它们会打断正常的程序流程。Java 中的异常可以分为两大类：

- 已检查异常（Checked Exceptions）
- 未检查异常（Unchecked Exceptions）：未检查异常进一步分为

    - 运行时异常（Runtime Exceptions）
    - 错误（Errors）

这里列出一些常见的异常类型：

已检查异常（Checked Exceptions）：这些异常必须在代码中被显式地处理（捕获或声明抛出）。如果一个方法可能产生这类异常，但没有处理它（即没有捕获或没有在其方法声明中用 `throws` 关键字声明抛出），编译器将报错。

- `IOException`：处理输入输出操作失败或中断时抛出，比如文件读写操作失败。
- `SQLException`：处理数据库访问错误或其他与数据库相关的问题时抛出。
- `ClassNotFoundException`：当应用尝试加载类通过其字符串名称，但找不到对应的类时抛出。

运行时异常（Runtime Exceptions）：这些是未检查异常的一种，程序可以选择捕获它们，但不是强制性的。它们通常表示程序逻辑错误，应该在开发过程中被修复。

- `NullPointerException`：尝试使用 `null` 对象时发生。
- `ArrayIndexOutOfBoundsException`：尝试访问数组的非法索引时抛出。
- `ClassCastException`：尝试将对象强制转换为不是实例的类时抛出。
- `ArithmeticException`：数学运算异常，比如除以零。
- `IllegalArgumentException`：向方法传递了一个不合法或不适当的参数。

错误（Errors）：错误表示严重的问题，不是设计用来被应用程序捕获的。它们通常与底层资源的问题有关，比如系统资源不足，虚拟机问题等。

- `OutOfMemoryError`：Java 虚拟机（JVM）没有足够的内存来为对象分配空间。
- `StackOverflowError`：应用递归调用过深，导致堆栈溢出。
- `NoClassDefFoundError`：当 Java 虚拟机或 `ClassLoader` 实例尝试加载类的定义，但找不到对应的类时抛出。

理解这些常见的异常及其使用场景对于编写健壯和可靠的 Java 应用程序至关重要。正确地处理异常可以使你的程序更加稳定，并提供更好的用户体验。

### 示例 {#error_exception}

这是一个 Java 代码的除 0 异常：

```java
@RequestMapping("/billing")
@ResponseBody
public AjaxResult billing(String tag) {
    logger.info("this is method3,{}", tag);
    sleep();
    if (Optional.ofNullable(tag).get().equalsIgnoreCase("error")) {
        System.out.println(1 / 0);
    }
    return AjaxResult.success("下单成功");
}
```

请求该接口触发除零异常：`http://localhost:8080/billing?tag=error`

这时候可以从<<<custom_key.brand_name>>>上看到 span 的信息：`error_message` `error_stack`:

```txt
  error_message Request processing failed; nested exception is java.lang.ArithmeticException: / by zero
  error_stack  org.springframework.web.util.NestedServletException: Request processing failed; nested exception is java.lang.ArithmeticException: / by zero
    at org.springframework.web.servlet.FrameworkServlet.processRequest(FrameworkServlet.java:1014)
    at org.springframework.web.servlet.FrameworkServlet.doGet(FrameworkServlet.java:898)
    at javax.servlet.http.HttpServlet.service(HttpServlet.java:670)
    at org.springframework.web.servlet.FrameworkServlet.service(FrameworkServlet.java:883)
    at javax.servlet.http.HttpServlet.service(HttpServlet.java:779)
    at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:227)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
    at org.apache.tomcat.websocket.server.WsFilter.doFilter(WsFilter.java:53)
    at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
    at org.springframework.web.filter.RequestContextFilter.doFilterInternal(RequestContextFilter.java:100)
    at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:117)
    at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
    at org.springframework.web.filter.FormContentFilter.doFilterInternal(FormContentFilter.java:93)
    at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:117)
    at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
    at datadog.trace.instrumentation.springweb.HandlerMappingResourceNameFilter.doFilterInternal(HandlerMappingResourceNameFilter.java:50)
    at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:117)
    at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
    at org.springframework.web.filter.CharacterEncodingFilter.doFilterInternal(CharacterEncodingFilter.java:201)
    at org.springframework.web.filter.OncePerRequestFilter.doFilter(OncePerRequestFilter.java:117)
    at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
    at org.springframework.web.filter.ServletRequestPathFilter.doFilter(ServletRequestPathFilter.java:56)
    at org.springframework.web.filter.DelegatingFilterProxy.invokeDelegate(DelegatingFilterProxy.java:354)
    at org.springframework.web.filter.DelegatingFilterProxy.doFilter(DelegatingFilterProxy.java:267)
    at org.apache.catalina.core.ApplicationFilterChain.internalDoFilter(ApplicationFilterChain.java:189)
    at org.apache.catalina.core.ApplicationFilterChain.doFilter(ApplicationFilterChain.java:162)
    at org.apache.catalina.core.StandardWrapperValve.invoke(StandardWrapperValve.java:177)
    at org.apache.catalina.core.StandardContextValve.invoke(StandardContextValve.java:97)
    at org.apache.catalina.authenticator.AuthenticatorBase.invoke(AuthenticatorBase.java:541)
    at org.apache.catalina.core.StandardHostValve.invoke(StandardHostValve.java:135)
    at org.apache.catalina.valves.ErrorReportValve.invoke(ErrorReportValve.java:92)
    at org.apache.catalina.core.StandardEngineValve.invoke(StandardEngineValve.java:78)
    at org.apache.catalina.connector.CoyoteAdapter.service(CoyoteAdapter.java:360)
    at org.apache.coyote.http11.Http11Processor.service(Http11Processor.java:399)
    at org.apache.coyote.AbstractProcessorLight.process(AbstractProcessorLight.java:65)
    at org.apache.coyote.AbstractProtocol$ConnectionHandler.process(AbstractProtocol.java:891)
    at org.apache.tomcat.util.net.NioEndpoint$SocketProcessor.doRun(NioEndpoint.java:1784)
    at org.apache.tomcat.util.net.SocketProcessorBase.run(SocketProcessorBase.java:49)
    at org.apache.tomcat.util.threads.ThreadPoolExecutor.runWorker(ThreadPoolExecutor.java:1191)
    at org.apache.tomcat.util.threads.ThreadPoolExecutor$Worker.run(ThreadPoolExecutor.java:659)
    at org.apache.tomcat.util.threads.TaskThread$WrappingRunnable.run(TaskThread.java:61)
    at java.lang.Thread.run(Thread.java:750)
  Caused by: java.lang.ArithmeticException: / by zero
    at com.zy.observable.server.controller.ServerController.billing(ServerController.java:99)
    at sun.reflect.NativeMethodAccessorImpl.invoke0(Native Method)
    at sun.reflect.NativeMethodAccessorImpl.invoke(NativeMethodAccessorImpl.java:62)
    at sun.reflect.DelegatingMethodAccessorImpl.invoke(DelegatingMethodAccessorImpl.java:43)
    at java.lang.reflect.Method.invoke(Method.java:498)
    at org.springframework.web.method.support.InvocableHandlerMethod.doInvoke(InvocableHandlerMethod.java:205)
    at org.springframework.web.method.support.InvocableHandlerMethod.invokeForRequest(InvocableHandlerMethod.java:150)
    at org.springframework.web.servlet.mvc.method.annotation.ServletInvocableHandlerMethod.invokeAndHandle(ServletInvocableHandlerMethod.java:117)
    at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.invokeHandlerMethod(RequestMappingHandlerAdapter.java:895)
    at org.springframework.web.servlet.mvc.method.annotation.RequestMappingHandlerAdapter.handleInternal(RequestMappingHandlerAdapter.java:808)
    at org.springframework.web.servlet.mvc.method.AbstractHandlerMethodAdapter.handle(AbstractHandlerMethodAdapter.java:87)
    at org.springframework.web.servlet.DispatcherServlet.doDispatch(DispatcherServlet.java:1071)
    at org.springframework.web.servlet.DispatcherServlet.doService(DispatcherServlet.java:964)
    at org.springframework.web.servlet.FrameworkServlet.processRequest(FrameworkServlet.java:1006)
    ... 46 more

  meta{
  error.type  org.springframework.web.util.NestedServletException
}
```

可以从 `error_message` 中看到这是一个 `Request processing failed; nested exception is java.lang.ArithmeticException: / by zero` 异常。

其中 `error.type` 就是异常的类名称，`error_stack` 就是异常的堆栈信息。

再次，修改代码并使用 try/catch 捕捉异常信息：

```java
 try {
    if (Optional.ofNullable(tag).get().equalsIgnoreCase("error")) {
        System.out.println(1 / 0);
    }
 }catch (Exception e){
     System.out.println(e);
 }
```

此时再次请求接口则不会产生异常信息，这是因为在方法内部捕捉异常并处理之后就不会抛出被探针捕获。

## 更多 {#more-reading}

- 二开版本的 [DDTrace JAVA 扩展内容](ddtrace-ext-java.md)
- 二开版本的 [DDTrace JAVA 更新日志](ddtrace-ext-changelog.md)
- 默认会采集指标，具体 [指标集列表](jvm.md#metric)
- [Java 接入与异常说明](ddtrace-java.md)
- [DDTrace 采样策略和多链路工具串联注意事项](tracing-sample.md)
