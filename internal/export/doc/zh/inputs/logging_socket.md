---
title     : 'Socket'
summary   : '主要用于 Java/Go/Python 日志框架如何配置 Socket，将日志发送给 Datakit 日志采集器中。'
__int_icon      : 'icon/socket'
dashboard :
  - desc  : '暂无'
    path  : '-'
monitor   :
  - desc  : '暂无'
    path  : '-'
---

<!-- markdownlint-disable MD025 -->
# Socket 日志接入示例
<!-- markdownlint-enable -->


---

本篇主要介绍 Java/Go/Python 日志框架如何配置 Socket，将日志发送给 Datakit 日志采集器中。

> 文件采集和 Socket 采集是互斥的，开启 Socket 采集之前，请先关闭文件采集，参见[日志采集配置](logging.md)  

## 配置
### Java {#java}

在配置 log4j 的时候需要注意，log4j v1，默认是使用 *.properties* 文件进行配置；而目前 log4j-v2 使用 XML 文件进行配置。

虽然文件名有区别，但是 log4j 查找配置文件时，都是去 Class Path 目录下查找，按照规范：v1 的配置在 *resources/log4j.properties*，v2 配置在 *resources/log4j.xml*。

#### log4j(v2) {#log4j-v2}

在 Maven 的配置中导入 log4j 2.x 的 jar 包：

``` xml
<dependency>
   <groupId>org.apache.logging.log4j</groupId>
   <artifactId>log4j-api</artifactId>
   <version>2.6.2</version>
</dependency>

<dependency>
   <groupId>org.apache.logging.log4j</groupId>
   <artifactId>log4j-core</artifactId>
   <version>2.6.2</version>
</dependency>
```

在 *resources* 中配置 *log4j.xml*，添加 `Socket Appender`：

``` xml
 <!-- Socket appender socket 配置日志传输到本机 9530 端口，protocol 默认 TCP -->
 <Socket name="socketname" host="localhost" port="9530" charset="utf8">
     <!-- 自定义输出格式和序列布局 -->
     <PatternLayout pattern="%d{yyyy.MM.dd 'at' HH:mm:ss z} %-5level [traceId=%X{trace_id} spanId=%X{span_id}] %class{36} %L %M - %msg%xEx%n"/>

     <!--注意：不要开启序列化传输到 socket 采集器上，目前 DataKit 无法反序列化，请使用纯文本形式传输-->
     <!-- <SerializedLayout/>-->

     <!-- 第二种输出格式 json-->
     <!-- 注意：配置 compact eventEol 一定要是 true  这样单条日志输出为一行-->
     <!-- 将日志发送到观测云上后会自动将 JSON 展开，所以在这里建议您将日志单条单行输出 -->
     <!-- <JsonLayout  properties="true" compact="true" complete="false" eventEol="true"/>-->
 </Socket>

 <!-- 然后定义 logger，只有定义了 logger 并引入的 appender，appender 才会生效 -->
 <loggers>
      <root level="trace">
          <appender-ref ref="Console"/>
          <appender-ref ref="socketname"/>
      </root>
 </loggers>
```

Java 代码示例：

``` java
package com.example;

import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import java.io.PrintWriter;
import java.io.StringWriter;

public class logdemo {
    public static void main(String[] args) throws InterruptedException {
        Logger logger = LogManager.getLogger(logdemo.class);
       for (int i = 0; i < 5; i++) {
            logger.debug("this is log msg to  datakt");
        }

        try {
            int i = 0;
            int a = 5 / i; // 除 0 异常
        } catch (Exception e) {
            StringWriter sw = new StringWriter();
            e.printStackTrace(new PrintWriter(sw));
            String exceptionAsString = sw.toString();
            logger.error(exceptionAsString);
        }
    }
}
```

#### log4j(v1) {#log4j-v1}

在 maven 的配置中导入 log4j 1.x 的 jar 包

``` xml
 <dependency>
    <groupId>org.apache.logging.log4j</groupId>
    <artifactId>log4j-api</artifactId>
    <version>1.2.17</version>
 </dependency>
```

到 resources 目录下 创建 log4j.properties 文件

``` text
log4j.rootLogger=INFO,server
# ... 其他配置

log4j.appender.server=org.apache.log4j.net.SocketAppender
log4j.appender.server.Port=<dk socket port>
log4j.appender.server.RemoteHost=<dk socket ip>
log4j.appender.server.ReconnectionDelay=10000

# 可配置成 JSON 格式
# log4j.appender.server.layout=net.logstash.log4j.JSONEventLayout
...
```

#### Logback {#logback}

Logback 中的 `SocketAppender` [无法将纯文本发送到 Socket 上](https://logback.qos.ch/manual/appenders.html#SocketAppender){:target="_blank"}。

> 问题是 `SocketAppender` 发送序列化 Java 对象而不是纯文本。您可以使用 `log4j` 输入，但并不建议更换日志组件，而是重写一个将日志数据发送为纯文本的 `Appender`，并且您将其与 JSON 格式化一起使用。

替代方案： [Logback Logstash 插件最佳实践](../best-practices/cloud-native/k8s-logback-socket.md#spring-boot){:target="_blank"}


### Golang {#golang}

#### Zap {#zap}

Golang 中最常用的是 Uber 的 Zap 开源日志框架，Zap 支持自定义 Output 注入。

自定义日志输出器并注入到 `zap.core`：

``` golang
type soceketOutput struct {
    conn net.Conn
}

func (s *soceketOutput) Write(b []byte) (int, error) {
    return s.conn.Write(b)
}

func zapcal() {
    conn, _ := net.DialTCP("tcp", nil, DK_LOG_PORT)
    socket := &soceketOutput{
        conn: conn,
    }

    w := zapcore.AddSync(socket)

    core := zapcore.NewCore(zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
        TimeKey:        "ts",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        FunctionKey:    zapcore.OmitKey,
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.LowercaseLevelEncoder,
        EncodeTime:     zapcore.EpochTimeEncoder,
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }),
        w,
        zapcore.InfoLevel)
    
    l := zap.New(core, zap.AddCaller())

    l.Info("======= message =======")
}
```

### Python  {#python}

#### logging.handlers.SocketHandler {#socket-handler}

原生的 `socketHandler` 通过 socket 发送的是日志对象，并不是纯文本形式，所以需要自定义 Handler 并重写 `socketHandler` 中的 `makePickle(slef,record)` 方法。

代码仅供参考：

```python
import logging
import logging.handlers

logger = logging.getLogger("") # 实例化 logging

#自定义 class 并重写 makePickle 方法
class PlainTextTcpHandler(logging.handlers.SocketHandler):
    """ Sends plain text log message over TCP channel """

    def makePickle(self, record):
     message = self.formatter.format(record) + "\r\n"
     return message.encode()


def logging_init():
    # 创建文件 handler
    fh = logging.FileHandler("test.log", encoding="utf-8")
    # 创建自定义 handler
    plain = PlainTextTcpHandler("10.200.14.226", 9540)

    # 设置 logger 日志等级
    logger.setLevel(logging.INFO)

    # 设置输出日志格式
    formatter = logging.Formatter(
        fmt="%(asctime)s - %(filename)s line:%(lineno)d - %(levelname)s: %(message)s"
    )

    # 为 handler 指定输出格式，注意大小写
    fh.setFormatter(formatter)
    plain.setFormatter(formatter)
  
    
    # 为 logger 添加的日志处理器
    logger.addHandler(fh)
    logger.addHandler(plain)
    
    return True

if __name__ == '__main__':
    logging_init()

    logger.debug(u"debug")
    logger.info(u"info")
    logger.warning(u"warning")
    logger.error(u"error")
    logger.critical(u"critical")
    
```

TODO: 后续会慢慢补充其他语言的日志框架去使用 socket 将日志发送到 DataKit 上。
