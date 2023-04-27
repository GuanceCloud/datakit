# Auto Attach DDTrace-Java Agent

-----

*Author： 宋龙奇*

---

This Java tool is mainly used to inject DDTrace-java agent into the currently running Java process without manually configuring and restarting the host Java process.

<div class="grid cards" markdown>

-   [:material-language-java: :material-download:](https://static.guance.com/ddtrace/agent-attach-java.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/agent-attach-java){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/agent-attach-java/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/agent-attach-java/releases){:target="_blank"}

</div>

## principle {#principle}

Attach after the JVM is started and load through the Attach API. This method will execute the agentmain method after the agent is loaded.

## download {#download}

build from source code:

```shell
git clone https://github.com/GuanceCloud/agent-attach-java
mvn package
```

use -h ：
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

Parameter Description:
- "- options" ddtrace parameter: "dd. agent. host=localhost, dd. agent. port=9529, dd. service=mytest..."
- "- agent jar" The default agent path is: `/usr/local/ddtrace/dd java agent jar`
- "- pid "process pid, pid, and displayName cannot both be empty.". Just use one of them.
- "- displayName" Process name such as - displayName tmall.jar
- "- h or - help" Help

> Since there has been no tools. jar file since jdk9. So I brought the tools file under the project directory: 'lib/tools. jar' is from the jdk1.8 version.

java run：

```shell
java -jar agent-attach-java.jar
```

## Auto attach dd-java-agent.jar {#dynamic-inject-ddagent-java}

- download dd-java-agent.jar，and put to */usr/local/ddtrace/*.

```shell
mkdir -p /usr/local/ddtrace
cd /usr/local/ddtrace
wget https://static.guance.com/ddtrace/dd-java-agent.jar
```

???+ attention

You must use [Extended DDTrace] (ddtrace-ext-java.md), otherwise the automatic injection function is limited (various Trace parameters cannot be set).

- Start Java application (ignored if Java application is started)

- use agent-attach-java.jar attach to *dd-trace-java.jar*

```shell
java -jar agent-attach-java.jar \
 -options "dd.agent.port=9529" \
 -displayName "tmall.jar"
 -agent-jar /usr/local/datakit/data/dd-java-agent.jar
```