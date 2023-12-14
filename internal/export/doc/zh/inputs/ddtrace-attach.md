
# 自动注入 DDTrace-Java Agent

-----

本 Java 工具主要用于将 DDTrace-Java agent 注入到当前已经运行的 Java 进程中，无需手动配置和重启宿主 Java 进程。

<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   [:material-language-java: :material-download:](https://static.guance.com/ddtrace/agent-attach-java.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/agent-attach-java){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/agent-attach-java/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/agent-attach-java/releases){:target="_blank"}

</div>
<!-- markdownlint-enable -->

## 原理 {#principle}

Agent 注入的基本原理是通过 */proc/[Java-PID]*（或者 */tmp/*）目录下的一个文件，注入

``` not-set
load instrument dd-agent.jar=<params...>
```

再给 JVM 一个发送一个 SIGQUIT 信号，然后 JVM 就会读取指定的 agent jar 包。

## 下载并编译 {#download}

JDK 版本 1.8 和 1.8 以上不可以交叉使用。

如使用发布版本 请使用相应的 [releases 版本](https://github.com/GuanceCloud/agent-attach-java/releases){:target="_blank"}

***建议下载源码*** 并编译：

```shell
git clone https://github.com/GuanceCloud/agent-attach-java
```

如果是 JDK 1.8 版本，修改配置文件 pom.xml:

```xml
<!--将版本修改为 1.8 -->
    <configuration>
      <source>1.8</source>
      <target>1.8</target>
    </configuration>
    
    <!--将下面的注释放开，并将 tools.jar 注释掉 !!!-->
    <dependency>
      <groupId>io.earcam.wrapped</groupId>
      <artifactId>com.sun.tools.attach</artifactId>
      <version>1.8.0_jdk8u131-b11</version>
      <scope>compile</scope>
      <type>jar</type>
    </dependency>
```

如果版本是 JDK 9、11、17 使用以下配置 pom.xml:

```xml
<!--将目标版本修改为指定的版本-->
    <configuration>
        <source>11</source>
        <target>11</target>
    </configuration>

<!--    <dependency>
      <groupId>io.earcam.wrapped</groupId>
      <artifactId>com.sun.tools.attach</artifactId>
      <version>1.8.0_jdk8u131-b11</version>
      <scope>compile</scope>
      <type>jar</type>
    </dependency>-->
    <dependency>
      <groupId>com.sun</groupId>
      <artifactId>tools</artifactId>
      <version>1.8.0</version>
      <scope>system</scope>
      <systemPath>${project.basedir}/lib/tools.jar</systemPath>
    </dependency>
```

```shell
mvn package
# 使用 target/agent-attach-java-jar-with-dependencies.jar
rm -f target/agent-attach-java.jar
mv target/agent-attach-java-jar-with-dependencies.jar agent-attach-java.jar
```

使用 `-h` 查看：

``` shell
agent-attach-java$ java -jar target/agent-attach-java-jar-with-dependencies.jar -h

java -jar agent-attach-java.jar [-options <dd options>]
                                [-agent-jar <agent filepath>]
                                [-pid <pid>]
                                [-displayName <service name/displayName>]
                                [-h]
                                [-help]
[-options]:
   this is dd-java-agnet.jar env, example:
       dd.agent.port=9529,dd.agent.host=localhost,dd.service.name=serviceName,...
[-agent-jar]:
   default is: /usr/local/ddtrace/dd-java-agent.jar
[-pid]:
   service PID String
[-displayName]:
   service name
Note: -pid or -displayName must have a non empty !!!

Example command line:

java -jar agent-attach-java.jar -options 'dd.service.name=test,dd.tag=v1'\
 -displayName tmall.jar \
 -agent-jar /usr/local/ddtrace/dd-java-agent.jar
```

参数说明：

- `-options` DDTrace 参数 ：`dd.agent.host=localhost,dd.agent.port=9529,dd.service.name=mytest ...`
- `-agent-jar` agent 路径 默认为：`/usr/local/ddtrace/dd-java-agent.jar`
- `-pid` 进程 PID , PID 和 `displayName` 不可以同时为空，使用其中一个即可
- `-displayName` 进程名称 比如 `-displayName tmall.jar`
- `-h or -help` 帮助

注意：由于从 JDK 9 开始就没有 *tools.jar* 文件。所以在项目目录下带上了 *tools* 文件。*lib/tools.jar* 是 JDK 1.8 版本的。

## 动态注入 *dd-java-agent.jar* {#dynamic-inject-ddagent-java}

- 首先下载 *dd-java-agent.jar*，并放到 */usr/local/ddtrace/* 目录下。

```shell
mkdir -p /usr/local/ddtrace
cd /usr/local/ddtrace
wget https://static.guance.com/ddtrace/dd-java-agent.jar
```

<!-- markdownlint-disable MD046 -->
???+ attention

    必须使用[扩展版 DDTrace](ddtrace-ext-java.md)，否则自动注入功能受限（各种 Trace 参数无法设置）。
<!-- markdownlint-enable -->

- 启动 Java 应用（如果 Java 应用已启动，则忽略）

- 启动 *agent-attach-java.jar* 注入 *dd-trace-java.jar*

```shell
java -jar agent-attach-java.jar \
 -options "dd.agent.port=9529" \
 -displayName "tmall.jar" \
 -agent-jar /usr/local/datakit/data/dd-java-agent.jar
```

或者：

```shell
java -jar agent-attach-java.jar \
 -options "dd.agent.port=9529" \
 -pid 7027 \
 -agent-jar /usr/local/datakit/data/dd-java-agent.jar
```

## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: 找不到类 VirtualMachine？ {#NoClassDefFound}
<!-- markdownlint-enable -->

报错信息：

```text
Exception in thread "main" java.lang.NoClassDefFoundError: com/sun/tools/attach/VirtualMachine
        at com.guance.javaagent.JavaAgentLoader.loadAgent(JavaAgentLoader.java:35)
        at com.guance.javaagent.MyMainClass.main(MyMainClass.java:19)
Caused by: java.lang.ClassNotFoundException: com.sun.tools.attach.VirtualMachine
        at java.net.URLClassLoader.findClass(URLClassLoader.java:382)
        at java.lang.ClassLoader.loadClass(ClassLoader.java:424)
        at sun.misc.Launcher$AppClassLoader.loadClass(Launcher.java:349)
        at java.lang.ClassLoader.loadClass(ClassLoader.java:357)
        ... 2 more
```

这是由于缺少 tools.jar 包导致的，使用正确的 pom.xml 配置文件或者下载对应的版本即可。


<!-- markdownlint-disable MD013 -->
### :material-chat-question: 版本不支持？ {#UnsupportedClass}
<!-- markdownlint-enable -->

报错信息：

```text
Error: A JNI error has occurred, please check your installation and try again
Exception in thread "main" java.lang.UnsupportedClassVersionError: com/guance/javaagent/MyMainClass has been compiled by a more recent version of the Java Runtime (class file version 55.0), this version of the Java Runtime only recognizes class file versions up to 52.0
        at java.lang.ClassLoader.defineClass1(Native Method)
        at java.lang.ClassLoader.defineClass(ClassLoader.java:763)
        at java.security.SecureClassLoader.defineClass(SecureClassLoader.java:142)
        at java.net.URLClassLoader.defineClass(URLClassLoader.java:468)
        at java.net.URLClassLoader.access$100(URLClassLoader.java:74)
        at java.net.URLClassLoader$1.run(URLClassLoader.java:369)
        at java.net.URLClassLoader$1.run(URLClassLoader.java:363)
        at java.security.AccessController.doPrivileged(Native Method)
        at java.net.URLClassLoader.findClass(URLClassLoader.java:362)
        at java.lang.ClassLoader.loadClass(ClassLoader.java:424)
        at sun.misc.Launcher$AppClassLoader.loadClass(Launcher.java:349)
        at java.lang.ClassLoader.loadClass(ClassLoader.java:357)
        at sun.launcher.LauncherHelper.checkAndLoadMain(LauncherHelper.java:495)
```

这是因为编译时版本太低，而运行时版本太高导致。更换版本或者使用当前版本从新编译即可。
