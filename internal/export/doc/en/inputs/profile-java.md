---
title     : 'Profiling Java'
summary   : 'Profling Java applications'
tags:
  - 'JAVA'
  - 'PROFILE'
__int_icon: 'icon/profiling'
---


Datakit now supports two Java profiling tools: [dd-trace-java](https://github.com/DataDog/dd-trace-java){:target="_blank"} and [async-profiler](https://github.com/async-profiler/async-profiler){:target="_blank"}.

## dd-trace-Java {#ddtrace}

Download `dd-trace-java` from the page [`dd-trace-java`](https://github.com/DataDog/dd-trace-java/releases){:target="_blank"}.

<!-- markdownlint-disable MD046 -->
???+ Note

    Datakit currently supports `dd-trace-java 1.47.x` and lower versions. Higher versions have not been tested, and their compatibility is unknown. If you encounter any issues during use, please feel free to provide feedback to us.
<!-- markdownlint-enable -->

Currently, `dd-trace-java` integrates two sets of analysis engines: [`Datadog Profiler`](https://github.com/datadog/java-profiler){:target="_blank"} and the built - in [`JFR (Java Flight Recorder)`](https://docs.oracle.com/javacomponents/jmc-5-4/jfr-runtime-guide/about.htm){:target="_blank"} in the JDK.
Both engines have their own requirements for the platform and JDK version, which are listed as follows:

<!-- markdownlint-disable MD046 -->
=== "Datadog Profiler"

    The `Datadog Profiler` currently only supports the Linux system, and has the following requirements for the JDK version:

    - OpenJDK 8u352+, 11.0.17+, 17.0.5+ (including the corresponding versions built by [`Eclipse Adoptium`](https://projects.eclipse.org/projects/adoptium){:target="_blank"}, [`Amazon Corretto`](https://aws.amazon.com/cn/corretto/){:target="_blank"}, [`Azul Zulu`](https://www.azul.com/downloads/?package=jdk#zulu){:target="_blank"}, etc.)
    - Oracle JDK 8u352+, 11.0.17+, 17.0.5+
    - OpenJ9 JDK 8u372+, 11.0.18+, 17.0.6+

=== "JFR"

    - OpenJDK 11+
    - Oracle JDK 11+
    - OpenJDK 8 (version 1.8.0.262/8u262+)
    - Oracle JDK 8 (commercial features need to be enabled)

    ???+ Note

        `JFR` is a commercial feature of Oracle JDK 8 and is disabled by default. If you need to enable it, you need to add the parameters `-XX:+UnlockCommercialFeatures -XX:+FlightRecorder` when starting the project. Since JDK 11, `JFR` has become an open-source project and is no longer a commercial feature of Oracle JDK.
<!-- markdownlint-enable -->

Run Java Code

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

Explanation of some parameters:

| Parameter Name                            | Corresponding Environment Variable    | Explanation                                                                                                                                                                                                                                                                           |
|-------------------------------------------|---------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `-Ddd.profiling.enabled`                  | DD_PROFILING_ENABLED                  | Whether to enable the profiling function.                                                                                                                                                                                                                                             |
| `-Ddd.profiling.allocation.enabled`       | DD_PROFILING_ALLOCATION_ENABLED       | Whether to enable the `JFR` memory Allocation analysis. High-load applications may have a certain impact on performance. It is recommended to use the `Datadog Profiler` Allocation function for JDK 11 and above versions.                                                           |
| `-Ddd.profiling.heap.enabled`             | DD_PROFILING_HEAP_ENABLED             | Whether to enable the sampling of `JFR` memory Heap objects.                                                                                                                                                                                                                          |
| `-Ddd.profiling.directallocation.enabled` | DD_PROFILING_DIRECTALLOCATION_ENABLED | Whether to enable the sampling of `JFR` JVM direct memory allocation.                                                                                                                                                                                                                 |
| `-Ddd.profiling.ddprof.enabled`           | DD_PROFILING_DDPROF_ENABLED           | Whether to enable the `Datadog Profiler` analysis engine.                                                                                                                                                                                                                             |
| `-Ddd.profiling.ddprof.cpu.enabled`       | DD_PROFILING_DDPROF_CPU_ENABLED       | Whether to enable the `Datadog Profiler` CPU analysis.                                                                                                                                                                                                                                |
| `-Ddd.profiling.ddprof.wall.enabled`      | DD_PROFILING_DDPROF_WALL_ENABLED      | Whether to enable the collection of `Datadog Profiler` Wall time. This option affects the accuracy of the association between Trace and Profile, and it is recommended to enable it.                                                                                                  |
| `-Ddd.profiling.ddprof.alloc.enabled`     | DD_PROFILING_DDPROF_ALLOC_ENABLED     | Whether to enable the memory Allocation analysis of the `Datadog Profiler` engine. It has been verified that it cannot be enabled on JDK 8 currently. For JDK 8, please use `-Ddd.profiling.allocation.enabled` as appropriate and pay attention to the impact on system performance. |
| `-Ddd.profiling.ddprof.liveheap.enabled`  | DD_PROFILING_DDPROF_LIVEHEAP_ENABLED  | Whether to enable the analysis of the currently live Heap by the `Datadog Profiler` engine.                                                                                                                                                                                           |
| `-Ddd.profiling.ddprof.memleak.enabled`   | DD_PROFILING_DDPROF_MEMLEAK_ENABLED   | Whether to enable the memory leak analysis of the `Datadog Profiler` engine.                                                                                                                                                                                                          |

After a minute or two, you can visualize your profiles on the [profile](https://console.guance.com/tracing/profile){:target="_blank"}.

## Async Profiler {#async-profiler}

async-profiler is an open source Java profiler Based on HotSpot API, it can collect information such as stack and memory allocation during program operation.

async-profiler can trace the following kinds of events:

- CPU cycles
- Hardware and Software performance counters like cache misses, branch misses, page faults, context switches etc.
- Allocations in Java Heap
- Contented lock attempts, including both Java object monitors and ReentrantLocks

### Install async-profiler {#install}

???+ node "Requirements"
    Datakit is now compatible with async-profiler v2.9 and below, higher version compatibility is unknown.

The official website provides download for different platform binaries：

- Linux x64 (glibc): [async-profiler-2.8.3-linux-x64.tar.gz](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-x64.tar.gz){:target="_blank"}
- Linux x64 (musl): [async-profiler-2.8.3-linux-musl-x64.tar.gz](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-musl-x64.tar.gz){:target="_blank"}
- Linux arm64: [async-profiler-2.8.3-linux-arm64.tar.gz](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-arm64.tar.gz){:target="_blank"}
- macOS x64/arm64: [async-profiler-2.8.3-macos.zip](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-macos.zip){:target="_blank"}
- format converter：[converter.jar](https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/converter.jar){:target="_blank"}

Download archive and extract as below(Linux x64):

```shell
$ wget https://github.com/async-profiler/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-x64.tar.gz 
$ tar -zxf async-profiler-2.8.3-linux-x64.tar.gz 
$ cd async-profiler-2.8.3-linux-x64 && ls

  build  CHANGELOG.md  LICENSE  profiler.sh  README.md
```

### Use async-profiler {#usage}

- Set Linux kernel option `perf_events`

As of Linux 4.6, capturing kernel call stacks using perf_events from a non-root process requires setting two runtime variables. You can set them using sysctl or as follows:

```shell
sudo sysctl kernel.perf_event_paranoid=1
sudo sysctl kernel.kptr_restrict=0 
```

- Install Debug Symbols
<!-- markdownlint-disable MD046 -->
If memory allocation (` allocate `) related events need to be collected, it is required to install Debug Symbols. Oracle JDK already has these symbols built-in, so this step can be skipped. OpenJDK needs to be installed, and the installation method is as follows:

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

The gdb tool can be used to verify if the debug symbols are properly installed . For example on Linux:

```shell
gdb $JAVA_HOME/lib/server/libjvm.so -ex 'info address UseG1GC'
```

This command's output will either contain Symbol "UseG1GC" is at 0xxxxx or No symbol "UseG1GC" in current context.
<!-- markdownlint-enable -->
- Check Java process PID

Before collection, you need to know the Java process's PID（use `jps` command）

```shell
$ jps

9234 Jps
8983 Computey
```

- Profile Java process

Run `profiler.sh` and specify Java process PID：

```shell
./profiler.sh -d 10 -f profiling.html 8983 

Profiling for 10 seconds
Done
```
<!-- markdownlint-disable MD046 -->
After about 10s, there will generate a file named `profiling.html` in current dir, you can use browser to open it.
<!-- markdownlint-enable -->
### Combine DataKit with async-profiler {#async-datakit}

Requirements:

- [Install DataKit](../datakit/datakit-install.md).

- [Enable Profile Inputs](profile.md)

- Set your service name（optional）

By default, the program name will be automatically obtained as a 'service' to report the Guance Cloud. If customization is needed, the service name can be injected when the program starts:

```shell
java -Ddk.service=<service-name> ... -jar <your-jar>
```

There are two integration methods：

- [automate（Recommend）](profile-java.md#script)
- [manual](profile-java.md#manual)

#### automate by script {#script}

Automated scripts can easily integrate async profiler and DataKit, use as follows:

- create shell script
<!-- markdownlint-disable MD046 -->
Create a file named "collect.sh" in current dir, type follow text：


???- note "collect.sh"(click to expand)
    ```shell
    set -e
    LIBRARY_VERSION=2.8.3

    MAX_JFR_FILE_SIZE=6000000

    datakit_url=http://localhost:9529
    if [ -n "$DATAKIT_URL" ]; then
        datakit_url=$DATAKIT_URL
    fi

    datakit_profiling_url=$datakit_url/profiling/v1/input
    

    app_env=dev
    if [ -n "$APP_ENV" ]; then
        app_env=$APP_ENV
    fi

    app_version=0.0.0
    if [ -n "$APP_VERSION" ]; then
        app_version=$APP_VERSION
    fi

    host_name=$(hostname)
    if [ -n "$HOST_NAME" ]; then
        host_name=$HOST_NAME
    fi

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
    
    # 采集的 java 应用进程 ID, 此处可以自定义需要采集的 java 进程，比如可以根据进程名称过滤
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

    for process_id in $java_process_ids; do
        printf "profiling process %d\n" $process_id
        profile_collect $process_id > $runtime_dir/$process_id.log 2>&1 &
    done

    wait

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

- Execute script

```shell
bash collect.sh
```

After the script is executed, the collected profiling data will be reported to the GuanceCloud platform through DataKit, which can be viewed later in the "APM" - "Profile" page.

available env：

- `DATAKIT_URL`        ：DataKit URL address, default: `http://localhost:9529`
- `APP_ENV`            ：current env, for example: `dev/prod/test`
- `APP_VERSION`        ：your application version
- `HOST_NAME`          ：hostname
- `SERVICE_NAME`       ：your service name
- `PROFILING_DURATION` ：duration, in seconds
- `PROFILING_EVENT`    ：events, for example: `cpu/alloc/lock`
- `PROFILING_TAGS`     ：set custom tags, split by comma if multiples, e.g., `key1:value1,key2:value2`
- `PROCESS_ID`         ：target process PID, for example: `98789,33432`

```shell
DATAKIT_URL=http://localhost:9529 APP_ENV=test APP_VERSION=1.0.0 HOST_NAME=datakit PROFILING_EVENT=cpu,alloc PROFILING_DURATION=60 PROFILING_TAGS="tag1:val1,tag2:val2" PROCESS_ID=98789,33432 bash collect.sh
```

## manually collect {#manual}

Compared to automated scripts, manual operations have higher degrees of freedom and can meet the needs of different scenarios

- generate profiling file, format in "jfr"

```shell
./profiler.sh -d 10 -o jfr -f profiling.jfr jps
```
<!-- markdownlint-disable MD046 -->
- prepare "event.JSON" file
<!-- markdownlint-enable -->
```json
{
    "tags_profiler": "library_version:2.8.3,library_type:async_profiler,process_id:16718,host:host_name,service:profiling-demo,env:dev,version:1.0.0",
    "start": "2022-10-28T14:30:39.122688553+08:00",
    "end": "2022-10-28T14:32:39.122688553+08:00",
    "family": "java",
    "format": "jfr"
}
```

fields：

- `tags_profiler`: profiling tags,
    - `library_version`:  `async-profiler` version
    - `library_type`: profiler name
    - `process_id`: Java process PID
    - `host`: hostname
    - `service`: your service name
    - `env`: your service env
    - `version`: your app version
    - others
- `start`: profiling start time
- `end`: profiling end time
- `family`: language
- `format`: format

- upload to DataKit

```shell
$ curl http://localhost:9529/profiling/v1/input \
  -F "main=@profiling.jfr;filename=main.jfr" \
  -F "event=@event.json;filename=event.json;type=application/json"
```

If the http response body contains `{"content":{"ProfileID":"xxxxxxxx"}}` indicate successfully uploading.
