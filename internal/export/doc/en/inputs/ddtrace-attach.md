# Auto Attach DDTrace-Java Agent

This Java tool is mainly used to inject DDTrace-Java agent into the currently running Java process without manually configuring and restarting the host Java process.
<!-- markdownlint-disable MD046 MD030 -->
<div class="grid cards" markdown>

-   [:material-language-java: :material-download:](https://static.guance.com/ddtrace/agent-attach-java.jar){:target="_blank"} ·
    [:material-github:](https://github.com/GuanceCloud/agent-attach-java){:target="_blank"} ·
    [Issue](https://github.com/GuanceCloud/agent-attach-java/issues/new){:target="_blank"} ·
    [:octicons-history-16:](https://github.com/GuanceCloud/agent-attach-java/releases){:target="_blank"}

</div>
<!-- markdownlint-enable -->
## principle {#principle}

Attach after the JVM is started and load through the Attach API. This method will execute the agent main method after the agent is loaded.

## download {#download}

JDK versions 1.8 and above cannot be used interchangeably.

If using a released version, please use the corresponding [releases version](https://github.com/GuanceCloud/agent-attach-java/releases){:target="_blank"}

Download source and build：

```shell
git clone https://github.com/GuanceCloud/agent-attach-java
```

If it is JDK version 1.8, modify the configuration file pom.xml:

```xml
    <configuration>
      <source>1.8</source>
      <target>1.8</target>
    </configuration>

    <dependency>
      <groupId>io.earcam.wrapped</groupId>
      <artifactId>com.sun.tools.attach</artifactId>
      <version>1.8.0_jdk8u131-b11</version>
      <scope>compile</scope>
      <type>jar</type>
    </dependency>
```

If the version is JDK 9, 11, 17, use the following configuration pom.xml:

```xml
<!--Modify the target version to the specified version-->
    <configuration>
        <source>11</source>
        <target>11</target>
    </configuration>

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
       dd.agent.port=9529,dd.agent.host=localhost,dd.service.name=serviceName,...
[-agent-jar]:
   default is: /usr/local/ddtrace/dd-java-agent.jar
[-pid]:
   service PID String
[-displayName]:
   service name
Note: -pid or -displayName must have a non empty !!!

example command line:
java -jar agent-attach-java.jar -options 'dd.service.name=test,dd.tag=v1'\
 -displayName tmall.jar \
 -agent-jar /usr/local/ddtrace/dd-java-agent.jar
```

Parameter Description:

- `-options` ddtrace parameter: `dd.agent.host=localhost,dd.agent.port=9529,dd.service.name=mytest ...`
- "-agent-jar" The default agent path is: `/usr/local/ddtrace/dd-java-agent.jar`
- "-pid "process PID, PID and `displayName` cannot both be empty. Just use one of them
- "-displayName" Process name such as `-displayName tmall.jar`
- "-h or -help" Help

> Since there has been no tools. jar file since jdk9. So I brought the tools file under the project directory: 'lib/tools. jar' is from the jdk1.8 version.

Java run：

```shell
java -jar agent-attach-java.jar
```

## Auto attach `dd-java-agent.jar`` {#dynamic-inject-ddagent-java}

- download `dd-java-agent.jar`, and put to */usr/local/ddtrace/*.

```shell
mkdir -p /usr/local/ddtrace
cd /usr/local/ddtrace
wget https://static.guance.com/ddtrace/dd-java-agent.jar
```

???+ attention

You must use [Extended DDTrace](ddtrace-ext-java.md), otherwise the automatic injection function is limited (various Trace parameters cannot be set).

- Start Java application (ignored if Java application is started)

- use `agent-attach-java.jar` attach to `dd-trace-java.jar`

```shell
java -jar agent-attach-java.jar \
 -options "dd.agent.port=9529" \
 -displayName "tmall.jar" \
 -agent-jar /usr/local/datakit/data/dd-java-agent.jar
```


## FAQ {#faq}

<!-- markdownlint-disable MD013 -->
### :material-chat-question: NoClassDefFoundError VirtualMachine？ {#NoClassDefFound}
<!-- markdownlint-enable -->

Error messages:

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

This is due to the lack of tools.jar package. Use the correct pom.xml configuration file or download the corresponding version.


<!-- markdownlint-disable MD013 -->
### :material-chat-question: UnsupportedClassVersionError？ {#UnsupportedClass}
<!-- markdownlint-enable -->

Error messages:

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

This is because the compile time version is too low and the runtime version is too high.
Replace the version or use the current version to recompile.
