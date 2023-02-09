# Sample Socket Log Access
---

This article focuses on how the Java Go Python logging framework configures socket output to the datakit socket log collector.

> File collection and socket are mutually exclusive. Please close file collection before opening socket. Please configure `logging.conf` [specific configuration instructions](logging.md).  

## Java {#java}

When configuring log4j, it should be noted that log4j v1 is configured by default using *. properties file; log4j v2 is currently configured using the *. xml file.

Although there are differences in file names, when log4j looks for configuration files, it always goes to the classpath directory. According to the specification, v1 is configured in resources/log4j. properties, and v2 is configured in resources/log4j. xml.

### log4j(v2) {#log4j-v2}

Import the log4j 2. x jar package in the configuration of maven:
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

Configure log4j. xml in resources and add socket appender: 

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
 
Java code example:

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

Import log4j 1. x jar package in maven configuration

``` xml
 <dependency>
    <groupId>org.apache.logging.log4j</groupId>
    <artifactId>log4j-api</artifactId>
    <version>1.2.17</version>
 </dependency>
```

Create the log4j. properties file in the resources directory

``` text
log4j.rootLogger=INFO,server
# ... other configurations

log4j.appender.server=org.apache.log4j.net.SocketAppender
log4j.appender.server.Port=<dk socket port>
log4j.appender.server.RemoteHost=<dk socket ip>
log4j.appender.server.ReconnectionDelay=10000

# Configurable to json format
# log4j.appender.server.layout=net.logstash.log4j.JSONEventLayout
...
```

### logback {#logback}

`SocketAppender` in logback cannot send plain text to socket [doc](https://logback.qos.ch/manual/appenders.html#SocketAppender){:target="_blank"}

> The problem is that the SocketAppender sends serialized Java objects instead of plain text. You can use log4j for input, but I do not recommend replacing the logging component. Rather, I rewrite an Appender that sends log data as plain text, and you use it with JSON formatting.

Datakit also supports logging from files [logging from text](logging.md), which is the best solution when socket collection is not available.

## Golang {#golang}

### zap {#zap}

Most commonly used in Golang is uber's zap open source logging framework, which supports custom output injection

Customize the log exporter and inject it into `zap.core`

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

The native socketHandler sends a log object through the socket, not plain text, so you need to customize the handler and override the `makePickle(slef,record)` method in the socketHandler.

The code is for reference only:

```python
import logging
import logging.handlers

logger = logging.getLogger("") # Instantiate logging

#Customize the class and override the makePickle method
class PlainTextTcpHandler(logging.handlers.SocketHandler):
    """ Sends plain text log message over TCP channel """

    def makePickle(self, record):
     message = self.formatter.format(record) + "\r\n"
     return message.encode()


def logging_init():
    # Creat file handler
    fh = logging.FileHandler("test.log", encoding="utf-8")
    #Creat custom handler
    plain = PlainTextTcpHandler("10.200.14.226", 9540)

    # Setting logger log levels
    logger.setLevel(logging.INFO)

    # Setting output log format
    formatter = logging.Formatter(
        fmt="%(asctime)s - %(filename)s line:%(lineno)d - %(levelname)s: %(message)s"
    )

    # Specify the output format for the handler, paying attention to the case
    fh.setFormatter(formatter)
    plain.setFormatter(formatter)
  
    
    # Log handler added for logger
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

TODO: The log framework of other languages will be supplemented later to use socket to send logs to DataKit. 
