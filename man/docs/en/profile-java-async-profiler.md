# async-profiler

This article will introduce how to collect java applications based on [async-profiler](https://github.com/jvm-profiling-tools/async-profiler#async-profiler), and report the collected data to DataKit, so that it can be analyzed on Guance Cloud platform.

## async-profiler Introduction {#info}

Async-profiler is an open source Java performance analysis tool, based on HotSpot API, which can collect information such as stack and memory allocation in program running.

Async-profiler can collect the following events:

- CPU cycles
- Hardware and software performance counters, such as cache misses, branch misses, page faults, context switches and so on
- Java heap allocation
- Contented lock attempts, including Java object monitors and ReentrantLocks

## async-profiler Installation {#install}

Official website provides downloads of installation packages for different platforms (current version 2.8. 3):

 - Linux x64 (glibc): [async-profiler-2.8.3-linux-x64.tar.gz](https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-x64.tar.gz)
 - Linux x64 (musl): [async-profiler-2.8.3-linux-musl-x64.tar.gz](https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-musl-x64.tar.gz)
 - Linux arm64: [async-profiler-2.8.3-linux-arm64.tar.gz](https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-arm64.tar.gz)
 - macOS x64/arm64: [async-profiler-2.8.3-macos.zip](https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-macos.zip)
 - 不同格式文件转换器: [converter.jar](https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.8.3/converter.jar)  

 Download the corresponding installation package and unzip it.

 Take the Linux x64 (glibc) platform as an example (other platforms are similar):

```shell
$ wget https://github.com/jvm-profiling-tools/async-profiler/releases/download/v2.8.3/async-profiler-2.8.3-linux-x64.tar.gz 
$ tar -zxf async-profiler-2.8.3-linux-x64.tar.gz 
$ cd async-profiler-2.8.3-linux-x64 && ls

  build  CHANGELOG.md  LICENSE  profiler.sh  README.md
```

## async-profiler Usage {#usage}

### Preconditions {#async-requirement} 

- Setting the `perf_events` parameter

After Linux kernel version 4.6, if you need to start `perf_events` in the process with a non-root user, you need to set two system runtime variables, which can be set as follows:

```shell
$ sudo sysctl kernel.perf_event_paranoid=1
$ sudo sysctl kernel.kptr_restrict=0 
```

- Install Debug Symbols (When Collecting Alloc Events)

 If you need to collect alloc-related events, you need to install Debug Symbols. These Symbols are already built into the Oracle JDK, so you can skip this step. OpenJDK needs to be installed, and the installation method is as follows:

Debian / Ubuntu:

```shell
$ sudo apt install openjdk-8-dbg # OpenJDK 8

or

$ sudo apt install openjdk-11-dbg # OpenJDK 11
```

CentOS, RHEL and other RPM versions are available through `debuginfo-install`:

```shell
$ sudo debuginfo-install java-1.8.0-openjdk
```

linux platform can be checked through `gdb` to see if it is installed correctly:
```shell
$ gdb $JAVA_HOME/lib/server/libjvm.so -ex 'info address UseG1GC'
```

If the output contains `Symbol "UseG1GC" is at 0xxxxx` or `No symbol "UseG1GC" in current context`, the installation is successful.

- View java Process ID

Before collecting, you need to check the PID of the java process (you can use the `jps` command).

```shell
$ jps

9234 Jps
8983 Computey
```

### Collect java Process {#collect}

 Select a java process that needs to be collected (such as the 8983 process above), execute `profiler.sh` in the directory and collect data. 

```shell
$ ./profiler.sh -d 10 -f profiling.html 8983 

Profiling for 10 seconds
Done
```

 After about 10 seconds, an html file named `profiling.html` will be generated in the current directory, and you can view the flame map by opening the file through a browser.

## Integrate DataKit with async-profiler {#async-datakit}

### Preparation {#profile-requirement}

- [Prepare DataKit Service](datakit-install.md), DataKit version >= 1.4.3

  The default address for the following operations is `http://localhost:9529`. If not, it needs to be modified to the actual DataKit service address.
- [Open Profile collector](profile.md) 

### Integrate Steps {#steps}

Integration methods can be divided into two types:

- [Automation script (recommended)](#script)
- [manual](#manual) 

#### Automation Script {#script}

Automation scripts can easily integrate async-profiler and DataKit as follows.

**Create Shell Script**

Create a new file in the current directory, named `collect.sh`,  and enter the following:

```shell
set -e

LIBRARY_VERSION=2.8.3

# jfr file size allowed to upload to DataKit (6 M), do not modify
MAX_JFR_FILE_SIZE=6000000

# DataKit service address
datakit_url=http://localhost:9529
if [ -n "$DATAKIT_URL" ]; then
	datakit_url=$DATAKIT_URL
fi

# Full address for uploading profiling data
datakit_profiling_url=$datakit_url/profiling/v1/input

# Applied environment
app_env=dev
if [ -n "$APP_ENV" ]; then
    app_env=$APP_ENV
fi

# Applied version
app_version=0.0.0
if [ -n "$APP_VERSION" ]; then
    app_version=$APP_VERSION
fi

# Host name
host_name=$(hostname)
if [ -n "$HOST_NAME" ]; then
    host_name=$HOST_NAME
fi

# Service name
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

# Collection of java application process ID, here you can customize the need to collect the java process, for example, you can filter according to the process name.
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
  
  process_name=$(jps | grep $process_id | awk '{print $2}')
  
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

# Parallel collection of profiling data
for process_id in $java_process_ids; do
  printf "profiling process %d\n" $process_id
  profile_collect $process_id > $runtime_dir/$process_id.log 2>&1 &
done

# Wait for all tasks to end
wait

# Output task execution log
for process_id in $java_process_ids; do
  log_file=$runtime_dir/$process_id.log
  if [ -f $log_file ]; then
    echo
    cat $log_file
    rm $log_file
  fi
done
```

**Execute script**

```shell
$ bash collect.sh
```

After the script is executed, the collected profiling data will be reported to Guance Cloud platform through DataKit, and can be viewed later in "Application Performance Monitoring"-"Profiling".

The script supports the following environment variables:

- `DATAKIT_URL`: DataKit url address, default to http://localhost:9529
- `APP_ENV`: Current application environment, such as `dev | prod | test` and so on
- `APP_VERSION`: Current application version
- `HOST_NAME`: Host name
- `SERVICE_NAME`: Service name
- `PROFILING_DURATION`: Sampling duration in seconds
- `PROFILING_EVENT`: Collected events such as `cpu,alloc,lock` and so on
- `PROCESS_ID`: Acquired java process IDs, multiple IDs separated by commas, such as `98789,33432`

```shell
$ DATAKIT_URL=http://localhost:9529 APP_ENV=test APP_VERSION=1.0.0 HOST_NAME=datakit PROFILING_EVENT=cpu,alloc PROFILING_DURATION=20 PROCESS_ID=98789,33432 bash collect.sh
```

#### Manual {#manual}

Compared with automated scripts, manual operation has high degree of freedom and can meet the needs of different scenarios.

**Collect Profiling Files (jfr format)**

First use `async-profiler` to collect the profiling information of the java process and generate a file in the **jfr** format.

For example:

```shell
$ ./profiler.sh -d 10 -o jfr -f profiling.jfr jps
```

**Prepare Meta-information Files**

Write a profiling meta-information file, such as event.json:

```json
{
    "tags_profiler": "library_version:2.8.3,library_type:async_profiler,process_id:16718,host:host_name,service:profiling-demo,env:dev,version:1.0.0",
    "start": "2022-10-28T14:30:39.122688553+08:00",
    "end": "2022-10-28T14:32:39.122688553+08:00",
    "family": "java",
    "format": "jfr"
}
```

Field meaning:

- `tags_profiler`: profiling data tag, which can contain custom tags
    - `library_version`: current async-profiler version
    - `library_type`: profiler library type, which is async-profiler
    - `process_id`: java process ID
    - `host`: host name
    - `service`: service name
    - `env`: the type of environment to apply
    - `version`: applied version
    - additional custom tags
- `start`: profiling start time
- `end`: end time of profiling
- `family`: language type
- `format`: file format

**Upload to DataKit**

When both of the above files are ready, `profiling.jfr` and `event.json`, they can be sent to the DataKit via an http POST request as follows:

```shell
$ curl http://localhost:9529/profiling/v1/input \
  -F "main=@profiling.jfr;filename=main.jfr" \
  -F "event=@event.json;filename=event.json;type=application/json"

```

When the above request returns a result in the format `{"content":{"ProfileID":"xxxxxxxx"}}`, it indicates that the upload was successful.
DataKit generates a profile record and saves the jfr file to the appropriate back-end store for subsequent analysis.

