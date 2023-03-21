# 自动注入 DDTrace-Java Agent

-----

*Author： 宋龙奇*

---

本 Java 工具主要用于将 DDTrace-java agent 注入到当前已经运行的 Java 进程中，无需手动配置和重启宿主 Java 进程。

<div class="grid cards" markdown>

-   [:material-language-java: :material-download:](https://static.guance.com/ddtrace/agent-attach-java.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/agent-attach-java){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/agent-attach-java/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/agent-attach-java/releases){:target="_blank"}

</div>

## 原理 {#principle}

Agent 注入的基本原理是通过 */proc/<Java-PID>*（或者 */tmp/*）目录下的一个文件，注入 `load instrument dd-agent.jar=<params...>`，再给 JVM 一个发送一个 SIGQUIT 信号，然后 JVM 就会读取指定的 agent jar 包。

## 下载并编译 {#download}

建议下载源码并编译:

```shell
git clone https://github.com/GuanceCloud/agent-attach-java
mvn package
```

使用 -h 查看：
```txt
root@q-PC:agent-attach-java$ java -jar target/agent-attach-java-jar-with-dependencies.jar -h
java -jar agent-attach-java.jar [-options <dd options>]
                                [-agent-jar <agent filepath>]
                                [-pid <pid>]
                                [-displayName <service name/displayName>]
                                [-h]
                                [-help]
[-options]:
   this is dd-java-agnet.jar env, example:
       dd.agent.port=9529,dd.agent.host=localhost,dd.service=serviceName,...
[-agent-jar]:
   default is: /usr/local/ddtrace/dd-java-agent.jar
[-pid]:
   service PID String
[-displayName]:
   service name
Note: -pid or -displayName must have a non empty !!!

example command line:
java -jar agent-attach-java.jar -options 'dd.service=test,dd.tag=v1'\
 -displayName tmall.jar \
 -agent-jar /usr/local/ddtrace/dd-java-agent.jar

```

参数说明：
- "-options" ddtrace 参数 ："dd.agent.host=localhost,dd.agent.port=9529,dd.service=mytest ... "
- "-agent-jar" agent 路径 默认为：`/usr/local/ddtrace/dd-java-agent.jar`
- "-pid" 进程 pid , pid 和 displayName 不可以同时为空。使用其中一个即可。
- "-displayName" 进程名称 比如 -displayName tmall.jar
- "-h or -help" 帮助

> 由于从 jdk9 开始就没有 tools.jar 文件。所以在项目目录下带上了tools文件： `lib/tools.jar` 是 jdk1.8 版本的。


## 动态注入 dd-java-agent.jar {#dynamic-inject-ddagent-java}

- 首先下载 dd-java-agent.jar，并放到 */usr/local/ddtrace/* 目录下。

```shell
mkdir -p /usr/local/ddtrace
cd /usr/local/ddtrace
wget https://static.guance.com/ddtrace/dd-java-agent.jar
```

???+ attention

    必须使用[扩展版 DDTrace](ddtrace-ext-java.md)，否则自动注入功能受限（各种 Trace 参数无法设置）。

- 启动 Java 应用（如果 Java 应用已启动，则忽略）

- 启动 agent-attach-java.jar 注入 *dd-trace-java.jar*

```shell
java -jar agent-attach-java.jar \
 -options "dd.agent.port=9529" \
 -displayName "tmall.jar"
 -agent-jar /usr/local/datakit/data/dd-java-agent.jar
```

