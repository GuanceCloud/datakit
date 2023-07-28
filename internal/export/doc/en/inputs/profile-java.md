
# Java profiling

---

Datakit now supports two Java profiling tools: [dd-trace-java](https://github.com/DataDog/dd-trace-java){:target="_blank"} and [async-profiler](https://github.com/async-profiler/async-profiler){:target="_blank"}.

## dd-trace-java {#ddtrace}

Firstly Download dd-trace-java from [https://github.com/DataDog/dd-trace-java/releases](https://github.com/DataDog/dd-trace-java/releases) .

<!-- markdownlint-disable MD046 -->
???+ Note "Requirements"


    Datakit is now compatible with dd-trace-java 1.15.x and below, the compatibility with higher version is unknown. please feel free to send your feedback to us if you encounter any incompatibility.
    
    Required JDK version:

    - OpenJDK 11.0.17+, 17.0.5+
    - Oracle JDK 11.0.17+, 17.0.5+
    - OpenJDK 8 version 8u352+
<!-- markdownlint-enable -->

Run Java Code

```shell
java -javaagent:/<your-path>/dd-java-agent.jar \
    -Ddd.service=profiling-demo \
    -Ddd.env=dev \
    -Ddd.version=1.2.3  \
    -Ddd.profiling.enabled=true  \
    -XX:FlightRecorderOptions=stackdepth=256 \
    -Ddd.trace.agent.port=9529 \
    -jar your-app.jar 
```

After a minute or two, you can visualize your profiles on the [profile](https://console.guance.com/tracing/profile).

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

Prior to JDK 11, the allocation profiler required HotSpot debug symbols. Oracle JDK already has them embedded in libjvm.so, but in OpenJDK builds they are typically shipped in a separate package, you can install it as follows:

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

The gdb tool can be used to verify if the debug symbols are properly installed for the libjvm library. For example on Linux:

```shell
gdb $JAVA_HOME/lib/server/libjvm.so -ex 'info address UseG1GC'
```

This command's output will either contain Symbol "UseG1GC" is at 0xxxxx or No symbol "UseG1GC" in current context.

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

After about 10s, there will generate a file named `profiling.html` in current workdir, you can use browser to open it.

### Combine DataKit with async-profiler {#async-datakit}

Requirements:

- [Install DataKit](datakit-install.md).

- [Enable Profile Inputs](profile.md)

- Set your service name（optional）

By default, the program name will be automatically obtained as a 'service' to report the observation cloud. If customization is needed, the service name can be injected when the program starts:

```shell
java -Ddk.service=<service-name> ... -jar <your-jar>
```

There are two integration methods：

- [automate（Recommend）](profile-java.md#script)
- [manual](profile-java.md#manual)

#### automate by script {#script}

Automated scripts can easily integrate async profiler and DataKit, use as follows:

- create shell script

Create a file named "collect.sh" in current workdir, type follow text：

<!-- markdownlint-disable MD046 -->
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
    
    # 采集的 java 应用进程 ID, 此处可以自定义需要采集的 java 进程, 比如可以根据进程名称过滤
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
    
        jfr_zip_file=$jfr_file.zip
    
        zip -q $jfr_zip_file $jfr_file
    
        zip_file_size=`ls -la $jfr_zip_file | awk '{print $5}'`
    
        if [ -z "$service_name" ]; then
            service_name=$process_name
        fi
    
        if [ $zip_file_size -gt $MAX_JFR_FILE_SIZE ]; then
            printf "Warning: the size of the jfr file generated is bigger than $MAX_JFR_FILE_SIZE bytes, now is $zip_file_size bytes\n"
        else
            cat >$event_json_file <<END
    {
            "tags_profiler": "library_version:$LIBRARY_VERSION,library_type:async_profiler,process_id:$process_id,process_name:$process_name,service:$service_name,host:$host_name,env:$app_env,version:$app_version",
            "start": "$start_time",
            "end": "$end_time",
            "family": "java",
            "format": "jfr"
    }
    END
    
            res=$(curl $datakit_profiling_url \
                -F "main=@$jfr_zip_file;filename=main.jfr" \
                -F "event=@$event_json_file;filename=event.json;type=application/json"  )
    
            if [[ $res != *ProfileID* ]]; then
                printf "Warning: send profile file to datakit failed\n"
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

After the script is executed, the collected profiling data will be reported to the Observation Cloud platform through DataKit, which can be viewed later in the "APM" - "Profile" page.

available env：

- `DATAKIT_URL`        ：DataKit URL address, default: `http://localhost:9529`
- `APP_ENV`            ：current env, for example: `dev/prod/test`
- `APP_VERSION`        ：your application version
- `HOST_NAME`          ：hostname
- `SERVICE_NAME`       ：your service name
- `PROFILING_DURATION` ：duration, in seconds
- `PROFILING_EVENT`    ：events, for example: `cpu/alloc/lock` 
- `PROCESS_ID`         ：target process PID, for example: `98789,33432`

```shell
DATAKIT_URL=http://localhost:9529 APP_ENV=test APP_VERSION=1.0.0 HOST_NAME=datakit PROFILING_EVENT=cpu,alloc PROFILING_DURATION=20 PROCESS_ID=98789,33432 bash collect.sh
```

## manually collect {#manual}

Compared to automated scripts, manual operations have higher degrees of freedom and can meet the needs of different scenarios

- generate profiling file, format in "jfr"

```shell
./profiler.sh -d 10 -o jfr -f profiling.jfr jps
```

- prepare "event.json" file

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
