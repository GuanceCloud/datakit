# Socket 日志接入示例
---

本篇主要介绍 Java Go Python 日志框架 如何配置 socket 输出到 datakit socket 日志采集器中。

> 文件采集和socket是互斥开启socket之前 请先关闭文件采集 请先配置好 `logging.conf` [具体配置说明](logging.md)  

## Java {#java}

在配置log4j的时候需要注意，log4j v1，默认是使用*.properties文件进行配置；而目前log4j v2使用*.xml文件进行配置。

虽然文件名有区别，但是log4j查找配置文件时，都是去classpath目录下查找,按照规范:v1的配置在 resources/log4j.properties, v2配置在resources/log4j.xml。

### log4j(v2) {#log4j-v2}

在maven的配置中导入log4j 2.x 的jar包:
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

在 resources 中配置 log4j.xml，添加 socket appender：

``` xml
 <!-- Socket appender socket 配置日志传输到本机9540端口，protocol默认tcp -->
 <Socket name="socketname" host="localHost" port="9540" charset="utf8">
     <!-- 自定义 输出格式  序列布局-->
     <PatternLayout pattern="%d{yyyy.MM.dd 'at' HH:mm:ss z} %-5level %class{36} %L %M - %msg%xEx%n"/>

     <!--注意：不要开启序列化传输到 socket 采集器上，目前 DataKit 无法反序列化，请使用纯文本形式传输-->
     <!-- <SerializedLayout/>-->

     <!-- 注意: 配置 compact eventEol 一定要是true  这样单条日志输出为一行-->
     <!-- 将日志发送到观测云上后会自动将json展开 所以在这里建议您将日志单条单行输出 -->
     <!-- <JsonLayout  properties="true" compact="true" complete="false" eventEol="true"/>-->
 </Socket>

 <!-- 然后定义logger，只有定义了logger并引入的appender，appender才会生效 -->
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
            int a = 5 / i; // 除0异常
        } catch (Exception e) {
            StringWriter sw = new StringWriter();
            e.printStackTrace(new PrintWriter(sw));
            String exceptionAsString = sw.toString();
            logger.error(exceptionAsString);
        }
    }
}

```
 
### log4j(v1) {#log4j-v1}

在maven的配置中导入log4j 1.x 的jar包

``` xml
 <dependency>
    <groupId>org.apache.logging.log4j</groupId>
    <artifactId>log4j-api</artifactId>
    <version>1.2.17</version>
 </dependency>
```

到 resources 目录下 创建log4j.properties文件

``` text
log4j.rootLogger=INFO,server
# ... 其他配置

log4j.appender.server=org.apache.log4j.net.SocketAppender
log4j.appender.server.Port=<dk socket port>
log4j.appender.server.RemoteHost=<dk socket ip>
log4j.appender.server.ReconnectionDelay=10000

# 可配置成json格式
# log4j.appender.server.layout=net.logstash.log4j.JSONEventLayout
...
```

### logback {#logback}

logback 中的`SocketAppender` 无法将纯文本发送到 socket上  [官方文档说明](https://logback.qos.ch/manual/appenders.html#SocketAppender){:target="_blank"}

> 问题是 SocketAppender发送序列化Java对象而不是纯文本。您可以使用log4j输入，但我并不建议更换日志组件，而是重写一个将日志数据发送为纯文本的Appender，并且您将其与JSON格式化一起使用。

datakit 同时支持从文件中采集日志 [从文本中采集日志](logging.md) ,可作为socket采集不可用时的最佳方案。 

## Golang {#golang}

### zap {#zap}

Golang 中最常用的是uber的zap开源日志框架，zap支持自定义output注入

自定义日志输出器并注入到`zap.core`

``` go

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

## Python  {#python}

### logging.handlers.SocketHandler {#socket-handler}

原生的 socketHandler 通过socket发送的是日志对象，并不是纯文本形式，所以需要自定义 handler 并重写 socketHandler 中的`makePickle(slef,record)`方法。

代码仅供参考：

```python
import logging
import logging.handlers

logger = logging.getLogger("") # 实例化logging

#自定义class 并重写makePickle方法
class PlainTextTcpHandler(logging.handlers.SocketHandler):
    """ Sends plain text log message over TCP channel """

    def makePickle(self, record):
     message = self.formatter.format(record) + "\r\n"
     return message.encode()


def logging_init():
    # 创建文件handler
    fh = logging.FileHandler("test.log", encoding="utf-8")
    #创建自定义handler
    plain = PlainTextTcpHandler("10.200.14.226", 9540)

    # 设置logger日志等级
    logger.setLevel(logging.INFO)

    # 设置输出日志格式
    formatter = logging.Formatter(
        fmt="%(asctime)s - %(filename)s line:%(lineno)d - %(levelname)s: %(message)s"
    )

    # 为handler指定输出格式，注意大小写
    fh.setFormatter(formatter)
    plain.setFormatter(formatter)
  
    
    # 为logger添加的日志处理器
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
