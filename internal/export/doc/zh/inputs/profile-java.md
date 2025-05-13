---
title     : 'Profiling Java'
summary   : 'Java Profiling 集成'
tags:
  - 'JAVA'
  - 'PROFILE'
__int_icon: 'icon/profiling'
---

DataKit 支持两种采集器来采集 Java profiling 数据， [`dd-trace-java`](https://github.com/DataDog/dd-trace-java){:target="_blank"} 和 [async-profiler](https://github.com/async-profiler/async-profiler){:target="_blank"}。

## `dd-trace-java` {#ddtrace}

从页面 [`dd-trace-java`](https://github.com/DataDog/dd-trace-java/releases){:target="_blank"} 下载 `dd-trace-java`.

<!-- markdownlint-disable MD046 -->
???+ note

    DataKit 目前支持 `dd-trace-java 1.47.x` 及以下版本，更高版本未经测试，兼容性未知，如使用上遇到问题，可向我们反馈。
<!-- markdownlint-enable -->

`dd-trace-java` 目前集成了两套分析引擎：[`Datadog Profiler`](https://github.com/datadog/java-profiler){:target="_blank"} 和 JDK 内置的 [`JFR(Java Flight Recorder)`](https://docs.oracle.com/javacomponents/jmc-5-4/jfr-runtime-guide/about.htm){:target="_blank"}，
两种引擎对平台和 JDK 版本都有各自的一些要求，列举如下：

<!-- markdownlint-disable MD046 -->
=== "Datadog Profiler"

    `Datadog Profiler` 目前仅支持 Linux 系统，对 JDK 的版本要求：

    - OpenJDK 8u352+, 11.0.17+, 17.0.5+ (包括 [`Eclipse Adoptium`](https://projects.eclipse.org/projects/adoptium){:target="_blank"}， [`Amazon Corretto`](https://aws.amazon.com/cn/corretto/){:target="_blank"}， [`Azul Zulu`](https://www.azul.com/downloads/?package=jdk#zulu){:target="_blank"} 等构建的相应版本)
    - Oracle JDK 8u352+, 11.0.17+, 17.0.5+
    - OpenJ9 JDK 8u372+, 11.0.18+, 17.0.6+

=== "JFR"

    - OpenJDK 11+
    - Oracle JDK 11+
    - OpenJDK 8 (version 1.8.0.262/8u262+)
    - Oracle JDK 8 (需开启商业特性)
    
    ???+ note

        `JFR` 是 Oracle JDK 8 的商业特性，默认是关闭的，如需启用需在启动项目时增加参数 `-XX:+UnlockCommercialFeatures -XX:+FlightRecorder`，而从 JDK 11 开始，`JFR` 已经成为开源项目且不再是 Oracle JDK 的商业特性。

<!-- markdownlint-enable -->

开启 profiling

```shell
java -javaagent:/<your-path>/dd-java-agent.jar \
    -XX:FlightRecorderOptions=stackdepth=256 \
    -Ddd.agent.host=127.0.0.1 \
    -Ddd.trace.agent.port=9529 \
    -Ddd.service.name=profiling-demo \
    -Ddd.env=dev \
    -Ddd.version=1.2.3  \
    -Ddd.profiling.enabled=true  \
    -Ddd.profiling.ddprof.enabled=true \
    -Ddd.profiling.ddprof.cpu.enabled=true \
    -Ddd.profiling.ddprof.wall.enabled=true \
    -Ddd.profiling.ddprof.alloc.enabled=true \
    -Ddd.profiling.ddprof.liveheap.enabled=true \
    -Ddd.profiling.ddprof.memleak.enabled=true \
    -jar your-app.jar 
```

部分参数说明：

<!-- markdownlint-disable MD038 -->
| 参数名                                    | 对应环境变量                            | 说明                                                                                                                                                                  |
| ---                                       | ---                                     | ---                                                                                                                                                                   |
| `-Ddd.profiling.enabled`                  | `DD_PROFILING_ENABLED                 ` | 是否开启 profiling 功能                                                                                                                                               |
| `-Ddd.profiling.allocation.enabled`       | `DD_PROFILING_ALLOCATION_ENABLED      ` | 是否开启 `JFR` 内存 Allocation 分析，高负载应用可能会对性能产生一定影响，建议 JDK11 及以上版本使用 `Datadog Profiler` Allocation 功能                                 |
| `-Ddd.profiling.heap.enabled`             | `DD_PROFILING_HEAP_ENABLED            ` | 是否开启 `JFR` 内存 Heap 对象采样                                                                                                                                     |
| `-Ddd.profiling.directallocation.enabled` | `DD_PROFILING_DIRECTALLOCATION_ENABLED` | 是否启用 `JFR` JVM 直接内存分配采样                                                                                                                                   |
| `-Ddd.profiling.ddprof.enabled`           | `DD_PROFILING_DDPROF_ENABLED          ` | 是否启用 `Datadog Profiler` 分析引擎                                                                                                                                  |
| `-Ddd.profiling.ddprof.cpu.enabled`       | `DD_PROFILING_DDPROF_CPU_ENABLED      ` | 是否启用 `Datadog Profiler` CPU 分析                                                                                                                                  |
| `-Ddd.profiling.ddprof.wall.enabled`      | `DD_PROFILING_DDPROF_WALL_ENABLED     ` | 是否启用 `Datadog Profiler` Wall time 采集，此选项影响 Trace 和 Profile 之间关联的精确性，建议开启                                                                    |
| `-Ddd.profiling.ddprof.alloc.enabled`     | `DD_PROFILING_DDPROF_ALLOC_ENABLED    ` | 是否启用 `Datadog Profiler` 引擎的内存 Allocation 分析，经验证在 JDK8 上目前无法开启，对于 JDK8 请酌情选用 `-Ddd.profiling.allocation.enabled` 并关注对系统性能的影响 |
| `-Ddd.profiling.ddprof.liveheap.enabled`  | `DD_PROFILING_DDPROF_LIVEHEAP_ENABLED ` | 是否启用 `Datadog Profiler` 引擎当前存活的 Heap 分析                                                                                                                  |
| `-Ddd.profiling.ddprof.memleak.enabled`   | `DD_PROFILING_DDPROF_MEMLEAK_ENABLED  ` | 是否启用 `Datadog Profiler` 引擎内存泄漏分析                                                                                                                          |


程序运行后，约 1 分钟后即可在<<<custom_key.brand_name>>>平台查看相关数据。

### 生成性能指标 {#metrics}

DataKit 自 [:octicons-tag-24: Version-1.39.0](../datakit/changelog.md#cl-1.39.0) 开始支持从 `dd-trace-java` 的输出信息中抽取一组 JVM 运行时的相关指标，该组指标被置于 `profiling_metrics` 指标集下，下面列举其中部分指标加以说明：

| 指标名称                              | 说明                                                                                       | 单位       |
| ---                                   | ---                                                                                        | ---        |
| `prof_jvm_cpu_cores                 ` | 应用程序消耗的 CPU 总核数                                                                  | core       |
| `prof_jvm_alloc_bytes_per_sec       ` | 程序每秒分配内存总大小                                                                     | byte       |
| `prof_jvm_allocs_per_sec            ` | 程序每秒分配内存次数                                                                       | count      |
| `prof_jvm_alloc_bytes_total         ` | 单次 profiling 期间分配的总内存大小                                                        | byte       |
| `prof_jvm_class_loads_per_sec       ` | 程序每秒执行类加载的次数                                                                   | count      |
| `prof_jvm_compilation_time          ` | 单次 profiling 持续期间（ dd-trace 默认以 60 秒为一个采集周期，下同）执行 JIT 编译的总时间 | nanosecond |
| `prof_jvm_context_switches_per_sec  ` | 每秒线程上下文切换次数                                                                     | count      |
| `prof_jvm_direct_alloc_bytes_per_sec` | 每秒分配直接内存的大小                                                                     | byte       |
| `prof_jvm_throws_per_sec            ` | 每秒抛出异常次数                                                                           | count      |
| `prof_jvm_throws_total              ` | 单次 profiling 持续期间抛出异常总次数                                                      | count      |
| `prof_jvm_file_io_max_read_bytes    ` | 单次 profiling 持续期间一次文件读写读取的最大字节数                                        | byte       |
| `prof_jvm_file_io_max_read_time     ` | 单次 profiling 持续期间一次文件读持续的最长时间                                            | nanosecond |
| `prof_jvm_file_io_max_write_bytes   ` | 单次 profiling 持续期间读一次文件操作的最大字节数                                          | byte       |
| `prof_jvm_file_io_max_write_time    ` | 单次 profiling 持续期间写一次文件花费的最长时间                                            | nanosecond |
| `prof_jvm_file_io_read_bytes        ` | 单次 profiling 持续期间读取的文件总字节数                                                  | byte       |
| `prof_jvm_file_io_time              ` | 单次 profiling 持续期间执行文件 IO 总耗时                                                  | nanosecond |
| `prof_jvm_file_io_read_time         ` | 单次 profiling 持续期间执行文件读取总耗时                                                  | nanosecond |
| `prof_jvm_file_io_write_time        ` | 单次 profiling 持续期间执行文件写入总耗时                                                  | nanosecond |
| `prof_jvm_avg_gc_pause_time         ` | 每次 GC 导致的程序中断平均持续时间                                                         | nanosecond |
| `prof_jvm_max_gc_pause_time         ` | 单次 profiling 持续期间 GC 导致的最大程序中断时间                                          | nanosecond |
| `prof_jvm_gc_pauses_per_sec         ` | 每秒因 GC 导致程序中断的次数                                                               | count      |
| `prof_jvm_gc_pause_time             ` | 单次 profiling 持续期间 GC 导致程序中断持续时间总和                                        | nanosecond |
| `prof_jvm_lifetime_heap_bytes       ` | 活跃的堆内对象占用内存总大小                                                               | byte       |
| `prof_jvm_lifetime_heap_objects     ` | 活跃的堆内对象总数                                                                         | count      |
| `prof_jvm_locks_max_wait_time       ` | 单次 profiling 持续期间锁争用导致的最长等待时间                                            | nanosecond |
| `prof_jvm_locks_per_sec             ` | 每秒出现锁争用次数                                                                         | count      |
| `prof_jvm_socket_io_max_read_time   ` | 单次 profiling 持续期间 socket 单次读取数据消耗最长时间                                    | nanosecond |
| `prof_jvm_socket_io_max_write_bytes ` | 单次 profiling 持续期间 socket 单次最大发送字节数                                          | byte       |
| `prof_jvm_socket_io_max_write_time  ` | 单次 profiling 持续期间 socket 单次发送数据消耗的最大时间                                  | nanosecond |
| `prof_jvm_socket_io_read_bytes      ` | 单次 profiling 持续期间 socket 收取的总字节数                                              | byte       |
| `prof_jvm_socket_io_read_time       ` | 单次 profiling 持续期间 socket 用于读取数据的时间总消耗                                    | nanosecond |
| `prof_jvm_socket_io_write_time      ` | 单次 profiling 持续期间 socket 用于发送数据的时间总消耗                                    | nanosecond |
| `prof_jvm_socket_io_write_bytes     ` | 单次 profiling 持续期间 socket 发送数据总字节数                                            | byte       |
| `prof_jvm_threads_created_per_sec   ` | 每秒线程创建次数                                                                           | count      |
| `prof_jvm_threads_deadlocked        ` | 处于死锁状态的线程数                                                                       | count      |
| `prof_jvm_uptime_nanoseconds        ` | 程序已启动时长                                                                             | nanosecond |
<!-- markdownlint-enable -->

<!-- markdownlint-disable MD046 -->
???+ note

    该功能默认开启，如果不需要可以通过修改采集器的配置文件 `<DATAKIT_INSTALL_DIR>/conf.d/profile/profile.conf` 把其中的配置项 `generate_metrics` 置为 false 并重启 DataKit.

    ```toml
    [[inputs.profile]]

    ## set false to stop generating apm metrics from ddtrace output.
    generate_metrics = false
    ```
<!-- markdownlint-enable -->

## Async Profiler {#async-profiler}

async-profiler 是一款开源的 Java 性能分析工具，基于 HotSpot 的 API，可以收集程序运行中的堆栈和内存分配等信息。

async-profiler 可以收集以下几种事件：

- CPU cycles
- 硬件和软件性能计数器，如 Cache Misses, Branch Misses, Page Faults, Context Switches 等
- Java 堆的分配
- Contented Lock Attempts, 包括 Java object monitors 和 ReentrantLocks

### Async Profiler 安装 {#install}

<!-- markdownlint-disable MD046 -->
???+ node "版本要求"

    DataKit 目前支持 `async-profiler v2.9` 及以下版本，更高版本未经测试，兼容性未知。
<!-- markdownlint-enable -->

官网提供了不同平台的安装包的下载（以 v2.8.3 为例）：

- Linux x64 (glibc): [async-profiler-2.8.3-linux-x64.tar.gz](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-x64.tar.gz){:target="_blank"}
- Linux x64 (musl): [async-profiler-2.8.3-linux-musl-x64.tar.gz](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-musl-x64.tar.gz){:target="_blank"}
- Linux arm64: [async-profiler-2.8.3-linux-arm64.tar.gz](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-arm64.tar.gz){:target="_blank"}
- macOS x64/arm64: [async-profiler-2.8.3-macos.zip](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-macos.zip){:target="_blank"}
- 不同格式文件转换器：[converter.jar](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/converter.jar){:target="_blank"}

下载相应的安装包，并解压。下面以 Linux x64（glibc）平台为例（其他平台类似）：

```shell
$ wget https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-x64.tar.gz 
$ tar -zxf async-profiler-2.8.3-linux-x64.tar.gz 
$ cd async-profiler-2.8.3-linux-x64 && ls

  build  CHANGELOG.md  LICENSE  profiler.sh  README.md
```

### Async-Profiler 使用 {#usage}

- 设置 `perf_events` 参数

Linux 内核版本为 4.6 以后的，如果需要使用非 root 用户启动进程中的 `perf_events`，需要设置两个系统运行时变量，可通过如下方式设置：

```shell
sudo sysctl kernel.perf_event_paranoid=1
sudo sysctl kernel.kptr_restrict=0 
```

- 安装 Debug Symbols（采集内存分配事件）

如果需要采集内存分配（`alloc`）相关事件，则要求安装 Debug Symbols 。Oracle JDK 已经内置这些 Symbols，可跳过此步骤。而 OpenJDK 则需要安装，安装方式参考如下：

<!-- markdownlint-disable MD046 -->
=== "Debian/Ubuntu"

    ```shell
    sudo apt install openjdk-8-dbg # OpenJDK 8
    # Or
    sudo apt install openjdk-11-dbg # OpenJDK 11
    ```

=== "CentOS/RHEL"

    ```shell
    sudo debuginfo-install java-1.8.0-openjdk
    ```
<!-- markdownlint-enable -->

Linux 平台可以通过 `gdb` 查看是否正确安装：

```shell
gdb $JAVA_HOME/lib/server/libjvm.so -ex 'info address UseG1GC'
```

输出结果如果包含 `Symbol "UseG1GC" is at 0xxxxx` 或 `No symbol "UseG1GC" in current context`，则表明安装成功。

- 查看 Java 进程 ID

采集之前，需要查看 Java 进程的 PID（可以使用 `jps` 命令）

```shell
$ jps

9234 Jps
8983 Computey
```

- 采集 Java 进程

选定一个需要采集的 Java 进程 (如上面的 8983 进程)， 执行目录下的 `profiler.sh`，采集数据：

```shell
./profiler.sh -d 10 -f profiling.html 8983 

Profiling for 10 seconds
Done
```

约 10 秒后，会在当前目录下生成一个名为 `profiling.html` 的 html 文件，通过浏览器打开该文件，就可以查看火焰图。

### 整合 DataKit 和 async-profiler {#async-datakit}

准备工作

- [准备 DataKit 服务](../datakit/datakit-install.md)，版本 DataKit >= 1.4.3

以下操作默认地址为 `http://localhost:9529`。如果不是，需要修改为实际的 DataKit 服务地址。

- [开启 Profile 采集器](profile.md)

- Java 应用注入服务名称 (`service`)（可选）

默认会自动获取程序名称作为 `service` 上报<<<custom_key.brand_name>>>，如果需要自定义，可以程序启动时注入 service 名称：

```shell
java -Ddk.service=<service-name> ... -jar <your-jar>
```

整合方式，可以分为两种：

- [自动化脚本（推荐）](profile-java.md#script)
- [手动操作](profile-java.md#manual)
- [k8s 环境下使用](../datakit/datakit-operator.md#inject-async-profiler)

#### 自动化脚本 {#script}

自动化脚本可以方便地整合 async-profiler 和 DataKit，使用方法如下。

- 创建 shell 脚本

在当前目录下新建一个文件，命名为 `collect.sh`，输入以下内容：

<!-- markdownlint-disable MD046 -->
???- note "collect.sh（单击点开）"

    ```shell
    set -e
    
    LIBRARY_VERSION=2.8.3
    
    # 允许上传至 DataKit 的 jfr 文件大小 (6 M)，请勿修改
    MAX_JFR_FILE_SIZE=6000000
    
    # DataKit 服务地址
    datakit_url=http://localhost:9529
    if [ -n "$DATAKIT_URL" ]; then
        datakit_url=$DATAKIT_URL
    fi
    
    # 上传 profiling 数据的完整地址
    datakit_profiling_url=$datakit_url/profiling/v1/input
    
    # 应用的环境
    app_env=dev
    if [ -n "$APP_ENV" ]; then
        app_env=$APP_ENV
    fi
    
    # 应用的版本
    app_version=0.0.0
    if [ -n "$APP_VERSION" ]; then
        app_version=$APP_VERSION
    fi
    
    # 主机名称
    host_name=$(hostname)
    if [ -n "$HOST_NAME" ]; then
        host_name=$HOST_NAME
    fi
    
    # 服务名称
    service_name=
    if [ -n "$SERVICE_NAME" ]; then
        service_name=$SERVICE_NAME
    fi
    
    # profiling duration, in seconds
    profiling_duration=10
    if [ -n "$PROFILING_DURATION" ]; then
        profiling_duration=$PROFILING_DURATION
    fi
    
    # profiling event
    profiling_event=cpu
    if [ -n "$PROFILING_EVENT" ]; then
        profiling_event=$PROFILING_EVENT
    fi
    
    # 采集的 java 应用进程 ID，此处可以自定义需要采集的 java 进程，比如可以根据进程名称过滤
    java_process_ids=$(jps -q -J-XX:+PerfDisableSharedMem)
    if [ -n "$PROCESS_ID" ]; then
        java_process_ids=`echo $PROCESS_ID | tr "," " "`
    fi
    
    if [[ $java_process_ids == "" ]]; then
        printf "Warning: no java program found, exit now\n"
        exit 1
    fi
    
    is_valid_process_id() {
        if [ -n "$1" ]; then
            if [[ $1 =~ ^[0-9]+$ ]]; then
                return 1
            fi
        fi
        return 0
    }
    
    profile_collect() {
        # disable -e
        set +e
    
        process_id=$1
        is_valid_process_id $process_id
        if [[ $? == 0 ]]; then
            printf "Warning: invalid process_id: $process_id, ignore"
            return 1
        fi
    
        uuid=$(uuidgen)
        jfr_file=$runtime_dir/profiler_$uuid.jfr
        event_json_file=$runtime_dir/event_$uuid.json
    
        arr=($(jps -v | grep "^$process_id"))
    
        process_name="default"
    
        for (( i = 0; i < ${#arr[@]}; i++ ))
        do
            value=${arr[$i]}
            if [ $i == 1 ]; then
                process_name=$value
            elif [[ $value =~ "-Ddk.service=" ]]; then
                service_name=${value/-Ddk.service=/}
            fi
        done
    
        start_time=$(date +%FT%T.%N%:z)
        ./profiler.sh -d $profiling_duration --fdtransfer -e $profiling_event -o jfr -f $jfr_file $process_id
        end_time=$(date +%FT%T.%N%:z)
    
        if [ ! -f $jfr_file ]; then
            printf "Warning: generating profiling file failed for %s, pid %d\n" $process_name $process_id
            return
        else
            printf "generate profiling file successfully for %s, pid %d\n" $process_name $process_id
        fi
    
        jfr_zip_file=$jfr_file.gz
    
        gzip -qc $jfr_file > $jfr_zip_file
    
        zip_file_size=`ls -la $jfr_zip_file | awk '{print $5}'`
    
        if [ -z "$service_name" ]; then
            service_name=$process_name
        fi
    
        if [ $zip_file_size -gt $MAX_JFR_FILE_SIZE ]; then
            printf "Warning: the size of the jfr file generated is bigger than $MAX_JFR_FILE_SIZE bytes, now is $zip_file_size bytes\n"
        else
            tags="library_version:$LIBRARY_VERSION,library_type:async_profiler,process_id:$process_id,process_name:$process_name,service:$service_name,host:$host_name,env:$app_env,version:$app_version"
            if [ -n "$PROFILING_TAGS" ]; then
              tags="$tags,$PROFILING_TAGS"
            fi
            cat >$event_json_file <<END
    {
            "tags_profiler": "$tags",
            "start": "$start_time",
            "end": "$end_time",
            "family": "java",
            "format": "jfr"
    }
    END
    
            res=$(curl -i $datakit_profiling_url \
                -F "main=@$jfr_zip_file;filename=main.jfr" \
                -F "event=@$event_json_file;filename=event.json;type=application/json" | head -n 1 )
    
            if [[ ! $res =~ 2[0-9][0-9] ]]; then
                printf "Warning: send profile file to datakit failed, %s\n" "$res"
                printf "$res"
            else
                printf "Info: send profile file to datakit successfully\n"
                rm -rf $event_json_file $jfr_file $jfr_zip_file
            fi
        fi
    
        set -e
    }
    
    runtime_dir=runtime
    if [ ! -d $runtime_dir ]; then
        mkdir $runtime_dir
    fi
    
    # 并行采集 profiling 数据
    for process_id in $java_process_ids; do
        printf "profiling process %d\n" $process_id
        profile_collect $process_id > $runtime_dir/$process_id.log 2>&1 &
    done
    
    # 等待所有任务结束
    wait
    
    # 输出任务执行日志
    for process_id in $java_process_ids; do
        log_file=$runtime_dir/$process_id.log
        if [ -f $log_file ]; then
            echo
            cat $log_file
            rm $log_file
        fi
    done
    ```
<!-- markdownlint-enable -->

- 执行脚本

```shell
bash collect.sh
```

脚本执行完毕后，采集的 profiling 数据会通过 DataKit 上报给<<<custom_key.brand_name>>>平台，稍后可在"应用性能监测"-"Profile" 查看。

脚本支持如下环境变量：

- `DATAKIT_URL`        ：DataKit URL 地址，默认为 `http://localhost : 9529`
- `APP_ENV`            ：当前应用环境，如 `dev/prod/test` 等
- `APP_VERSION`        ：当前应用版本
- `HOST_NAME`          ：主机名称
- `SERVICE_NAME`       ：服务名称
- `PROFILING_DURATION` ：采样持续时间，单位为秒
- `PROFILING_EVENT`    ：采集的事件，如 `cpu/alloc/lock` 等
- `PROFILING_TAGS`     ：设置其他的自定义 Tag，英文逗号分隔的键值对，如 `key1:value1,key2:value2`
- `PROCESS_ID`         ：采集的 Java 进程 ID, 多个 ID 以逗号分割，如 `98789,33432`

```shell
DATAKIT_URL=http://localhost:9529 APP_ENV=test APP_VERSION=1.0.0 HOST_NAME=datakit PROFILING_EVENT=cpu,alloc PROFILING_DURATION=60 PROFILING_TAGS="tag1:val1,tag2:val2" PROCESS_ID=98789,33432 bash collect.sh
```

#### 手动操作 {#manual}

相比自动化脚本，手动操作自由度高，可满足不同的场景需求。

- 采集 profiling 文件（*jfr* 格式）

首先使用 `async-profiler` 收集 Java 进程的 profiling 信息，并生成 *jfr* 格式的文件。如：

```shell
./profiler.sh -d 10 -o jfr -f profiling.jfr jps
```

- 准备元信息文件

编写 profiling 元信息文件，如 `event.json`：

```json
{
    "tags_profiler": "library_version:2.8.3,library_type:async_profiler,process_id:16718,host:host_name,service:profiling-demo,env:dev,version:1.0.0",
    "start": "2022-10-28T14:30:39.122688553+08:00",
    "end": "2022-10-28T14:32:39.122688553+08:00",
    "family": "java",
    "format": "jfr"
}
```

字段含义：

- `tags_profiler`: profiling 数据标签，可包含自定义标签
    - `library_version`: 当前 `async-profiler` 版本
    - `library_type`: profiler 库类型， 即 `async-profiler`
    - `process_id`: Java 进程 ID
    - `host`: 主机名称
    - `service`: 服务名称
    - `env`: 应用的环境类型
    - `version`: 应用的版本
    - 其他自定义标签
- `start`: profiling 开始时间
- `end`: profiling 结束时间
- `family`: 语言种类
- `format`: 文件格式

- 上传至 DataKit

上述的两种文件都准备完毕，即 `profiling.jfr` 和 `event.json`，就可以通过 http POST 请求发送至 DataKit，方式如下：

```shell
$ curl http://localhost:9529/profiling/v1/input \
  -F "main=@profiling.jfr;filename=main.jfr" \
  -F "event=@event.json;filename=event.json;type=application/json"
```

当上述请求返回结果格式为 `{"content":{"ProfileID":"xxxxxxxx"}}` 时，表明上传成功。DataKit 会产生一条 profiling 记录，并将 jfr 文件保存至相应的后端存储，便于后续分析使用。

#### Kubernetes 环境下使用 {#under-k8s}

请参考 [使用 `datakit-operator` 注入 `async-profiler`](../datakit/datakit-operator.md#inject-async-profiler){:target="_blank"}。
