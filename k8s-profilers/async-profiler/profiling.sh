#!/usr/bin/env bash
# Unless explicitly stated otherwise all files in this repository are licensed
# under the MIT License.
# This product includes software developed at Guance Cloud (https://www.guance.com/).
# Copyright 2021-present Guance, Inc.

WORK_DIR=$(cd "$(dirname "$0")" && pwd)
CMD="$WORK_DIR"/profiler.sh
OUTPUT_DIR="$WORK_DIR"/output
LOG_DIR=$WORK_DIR/log
LOCKER_DIR=$WORK_DIR/locks
ASYNC_PROFILER_VERSION=$($CMD -v | head -n 1)
MAX_WAIT_TIME=200
SHELL_PATH="$WORK_DIR"/"$(basename "$0")"

Log() {
  level=$1
  fmt=$2
  shift 2
  printf "%s  [%s]  $fmt\n" "$(date '+%F %T.%3N%:z')" "$level" "$@"
}

Info() {
  Log "INFO" "$@"
}

Warn() {
  Log "WARN" "$@"
}

Error() {
  Log "ERROR" "$@"
}

Fatal() {
  Log "FATAL" "$@"
  exit 1
}



show_help() {
  cat <<EOT
Usage: [env1=val1 env2=val2 ...] $0 [options]

 DataKit JVM profiling tool for docker/k8s.

 Options:
   -H, --host <host>            Datakit listening host.
   -P, --port <port>            DataKit listening port, default: 9529.
   -S, --service <service>      Your application name, default: unknown.
   -V, --ver  <version>         Your application version (for example, 2.5, 202003181415, 1.3-alpha), default: unknown.
   -E, --env <env>              Your application environment (for example, production, staging), default: unknown.
   -D, --duration <duration>    Run profiling for <duration> seconds.
   -I, --interval <interval>    Send a batch of profiling data to datakit per <interval> seconds, default: 60.
       --hostname <hostname>    The host your application running on.
       --pid <pid>              Your application process PID, default: read from ps command.
       --events <events>        Profiling event: cpu|alloc|lock|cache-misses etc, default: cpu,alloc,lock.
   -a, --add-crontab            Add your profiling schedule to Linux crontab, use with "--schedule".
       --schedule <schedule>    Your profiling schedule, use the same format as Linux crontab (for example, "* */5 * * *", "00 02 * * *"),
                                use with "-a/--add-crontab".

   -h, --help                   display this help.
   -v, --version                display version.
   -s, --show-all-events        show all profiling events supported by the target JVM.


 Environment variable for a correlative option is also available, when both are set, the environment variable takes precedence.
   DK_AGENT_HOST           Datakit listening host.
   DK_AGENT_PORT           DataKit listening port, default: 9529.
   DK_PROFILE_SERVICE      Your application name, default: unknown.
   DK_PROFILE_VERSION      Your application version (for example, 2.5, 202003181415, 1.3-alpha), default: unknown.
   DK_PROFILE_ENV          Your application environment (for example, production, staging), default: unknown.
   DK_PROFILE_DURATION     Run profiling for <duration> seconds, default: infinity.
   DK_PROFILE_INTERVAL     Send a bit of profiling data to datakit per <interval> seconds, default: 60.
   DK_PROFILE_HOSTNAME     The host your application running on.
   DK_PROFILE_PID          Your application process PID, default: read from ps command.
   DK_PROFILE_EVENT        Profiling event: cpu|alloc|lock|cache-misses etc, default: cpu,alloc,lock.
   DK_PROFILE_SCHEDULE     Your profiling schedule, use setting like crontab (for example, "* */5 * * *", "00 02 * * *").

 Example:
   ./profiling.sh --host 192.168.1.1 --service my-app --ver 1.0.0 --env test --duration 300 --interval 30 --pid 13 --events cpu,alloc
   DK_AGENT_HOST=192.168.1.1 DK_PROFILE_PID=13 DK_PROFILE_EVENT="cpu,alloc" DK_PROFILE_SERVICE=my-app DK_PROFILE_VERSION=1.0.0 DK_PROFILE_ENV=test ./profiling.sh

EOT
}


if ! opts=$(getopt -o H:P:S:V:E:D:I:hvas -l host:,port:,service:,ver:,env:,duration:,interval:,hostname:,pid:,schedule:,events:,help,version,add-crontab,show-all-events -- "$@"); then
  exit 1
fi

eval set -- "$opts"

add_crontab=0
datakit_host=""
datakit_port="9529"
app_service=$(hostname)
app_version="unknown"
app_env="unknown"
duration=0
interval=60
hostname=$(hostname)
pid=0
profiling_events="cpu,alloc,lock"


while true; do
    case $1 in
    -H|--host)
      shift
      datakit_host="$1"
      ;;
    -P|--port)
      shift
      datakit_port="$1"
      ;;
    -S|--service)
      shift
      app_service="$1"
      ;;
    -V|--ver)
      shift
      app_version="$1"
      ;;
    -E|--env)
      shift
      app_env="$1"
      ;;
    -D|--duration)
      shift
      duration="$1"
      ;;
    -I|--interval)
      shift
      interval="$1"
      ;;
    --hostname)
      shift
      hostname="$1"
      ;;
    --pid)
      shift
      pid="$1"
      ;;
    --schedule)
      shift
      schedule="$1"
      ;;
    --events)
      shift
      profiling_events="$1"
      ;;
    -h|--help)
      show_help
      exit 0
      ;;
    -v|--version)
      $CMD -v
      exit 0
      ;;
    -a|--add-crontab)
      add_crontab=1
      ;;
    -s|--show-all-events)
      $CMD list
      exit 0
      ;;
    --)
      shift
      break
      ;;
    esac
    shift
done

if [ -n "$DK_AGENT_HOST" ]; then
  datakit_host="$DK_AGENT_HOST"
fi

if [ -n "$DK_AGENT_PORT" ]; then
  datakit_port="$DK_AGENT_PORT"
fi

if [ -n "$DK_PROFILE_SERVICE" ]; then
  app_service="$DK_PROFILE_SERVICE"
fi

if [ -n "$DK_PROFILE_VERSION" ]; then
  app_version="$DK_PROFILE_VERSION"
fi

if [ -n "$DK_PROFILE_ENV" ]; then
  app_env="$DK_PROFILE_ENV"
fi

if [ -n "$DK_PROFILE_DURATION" ]; then
  duration="$DK_PROFILE_DURATION"
fi


if [ -n "$DK_PROFILE_INTERVAL" ]; then
  interval="$DK_PROFILE_INTERVAL"
fi


if [ -n "$DK_PROFILE_HOSTNAME" ]; then
  hostname="$DK_PROFILE_HOSTNAME"
fi


if [ -n "$DK_PROFILE_PID" ]; then
  pid="$DK_PROFILE_PID"
fi


if [ -n "$DK_PROFILE_SCHEDULE" ]; then
  schedule="$DK_PROFILE_SCHEDULE"
fi

if [ -n "$DK_PROFILE_EVENT" ]; then
    profiling_events=$DK_PROFILE_EVENT
fi


# 允许上传至 DataKit 的 jfr 文件大小 (6 M)，请勿修改
MAX_JFR_FILE_SIZE=6000000

if [ -z "$datakit_host" ]; then
  Error "datakit_host not set" >&2
  exit 1
fi

# DataKit 服务地址
datakit_url=http://"$datakit_host":"$datakit_port"

# 上传 profiling 数据的完整地址
datakit_profiling_url=$datakit_url/profiling/v1/input


# 采集的 java 应用进程 ID，此处可以自定义需要采集的 java 进程，比如可以根据进程名称过滤
if { [[ "$pid" =~ ^[0-9]+$ ]] && [ "$pid" -gt 0 ]; } || [[ "$pid" =~ ^([1-9][0-9]*[[:blank:]]+)+[1-9][0-9]*$ ]]; then
  java_process_id=$pid
else
  java_process_id=$(jps -q -J-XX:+PerfDisableSharedMem | head -n 20)
fi

is_valid_process_id() {
    if [ -n "$1" ]; then
        if [[ $1 =~ ^[0-9]+$ ]]; then
            return 0
        fi
    fi
    return 1
}


install_crontab() {
  if [ ! -d "$LOG_DIR" ]; then
    mkdir -p "$LOG_DIR"
  fi
  if [ -n "$schedule" ]; then
    echo "${schedule} ${SHELL_PATH} --host $datakit_host --port $datakit_port --service $app_service --env $app_env --ver $app_version --duration $duration --interval $interval --pid $pid --hostname $hostname >> $LOG_DIR/main.log 2>&1 " | crontab -u root -
  fi
}

if [ $add_crontab -gt 0 ]; then
  schedule="${schedule#"${schedule%%[![:blank:]]*}"}"
  schedule="${schedule%"${schedule##*[![:blank:]]}"}"
  if [ -z "$schedule" ]; then
    Error "empty schedule parameter, please use crontab format like '* * * * *' " >&2
    exit 1
  fi
  if [[ ! "$schedule" =~  ^([0-9*,/-]+[[:blank:]]+){4}[0-9*,/-]+$ ]]; then
    Error "invalid schedule setting %s, please use crontab format like '* * * * *' " "$schedule" >&2
    exit 1
  else
    install_crontab
    exit 0
  fi
fi

#MAX_WAIT_TIME=$((interval * 3))

upload_profiling() {
  jfr_output_dir=$1
  start_time=$2
  process_id=$3
  process_name=$4

  index=0
  total_wait_time=0

  if ! cd "$jfr_output_dir"; then
    Error "Unable to enter dir %s" "$jfr_output_dir" >&2
    return 1
  fi
  while ((1))
   do
       if [ "$total_wait_time" -ge $MAX_WAIT_TIME ]; then
         break
       fi
       sleep 10
       ((total_wait_time += 10))
      # shellcheck disable=SC2012
      mapfile -t files < <(ls -tr1 --time=creation -- *.jfr 2>/dev/null | head -n 20)
      for ((i=0; i < ${#files[@]}; i++))
      do
        file=${files[i]}
        if [[ $file =~ ^[0-9]+\.jfr$ ]]; then
          filename=${file%.*}
          if [ "$filename" -ge $index ]; then
                Info "find next jfr file %s" "$file"
                start_time=$(date --date="$(stat --printf "%w" "$file")" +%FT%T.%N%:z) # file create time
                start_time_seconds=$(stat --printf "%W" "$file")
                end_time=$(date --date="$(stat --printf "%y" "$file")" +%FT%T.%N%:z) # file last modified time
                end_time_seconds=$(stat --printf "%Y" "$file")
                if [ "$duration" -gt 0 ]; then
                  least_time=0
                  if [ "$interval" -gt 2 ]; then
                    least_time=$((interval-2))
                  fi
                  if [ $(($(date +%s) - start_time_seconds)) -lt $((interval+5)) ] || [ $((end_time_seconds - start_time_seconds)) -le $least_time ]; then
                    continue 2
                  fi
                fi
                jfr_gzip_file="$file".gz
                event_json_file="$file".json
                gzip -9ck "$file" > "$jfr_gzip_file"
                gzip_file_size=$(stat -L --printf "%s" "$jfr_gzip_file")

                if [ "$gzip_file_size" -gt $MAX_JFR_FILE_SIZE ]; then
                    Error "the size of the jfr file generated is bigger than %d bytes, now is %d bytes" "$MAX_JFR_FILE_SIZE" "$gzip_file_size"
                else
                    cat >"$event_json_file" <<END
            {
                "attachments": ["main.jfr"],
                "tags_profiler": "library_version:$ASYNC_PROFILER_VERSION,library_type:async_profiler,process_id:$process_id,process_name:$process_name,service:$app_service,host:$hostname,env:$app_env,version:$app_version",
                "start": "$start_time",
                "end": "$end_time",
                "family": "java",
                "format": "jfr"
            }
END
                  Info "start to upload profile file %s to DataKit" "$file"
                  res=$(curl -s "$datakit_profiling_url" \
                      -F "main=@$jfr_gzip_file;filename=main.jfr" \
                      -F "event=@$event_json_file;filename=event.json;type=application/json"  )

                  if [[ $res != *ProfileID* ]]; then
                      Error "send profile file to datakit failed" >&2
                      echo "$res" >&2
                  else
                      Info "Successfully send profile file to datakit"
                  fi
                  rm -rf "$event_json_file" "$file" "$jfr_gzip_file"
              fi
              index=$((filename + 1))
              start_time=$end_time
              total_wait_time=0
              break
          fi
        fi
      done
  done
}

bootstrap_profiling() {
    process_id=$1
    if ! is_valid_process_id "$process_id"; then
      Warn "invalid process_id: %s, ignore" "$process_id" >&2
      return 1
    fi

    if [ ! -d "$LOCKER_DIR" ]; then
      mkdir -p "$LOCKER_DIR"
    fi

    lock_file=$LOCKER_DIR/async-profiler-$process_id.lock
    exec {locker_fd}> "$lock_file"
    if ! flock -xn "$locker_fd"; then
      Warn "Unable to get lock of file: %s, probably profiling on the target process is already running" "$lock_file" >&2
      exec {locker_fd}>&-
      rm -f "$lock_file"
      return 1
    fi

    jfr_output_dir=$OUTPUT_DIR/output-$process_id-$(date +%s%N)
    if [ ! -d "$jfr_output_dir" ]; then
      mkdir -p "$jfr_output_dir"
    fi
    jfr_filename="${jfr_output_dir}/%n{100000000}.jfr"

  mapfile -t arr < <(jps -v | grep "^${process_id} ")

  process_name="java"

  for (( i = 0; i < ${#arr[@]}; i++ ))
  do
    value=$(echo "${arr[$i]}" | awk '{print $2}')
    process_name=$value
  done

    if [ -z "$app_service" ]; then
      app_service=$process_name
    fi

    start_time=$(date +%FT%T.%N%:z)
    if [ "$duration" -gt 0 ]; then
      if ! $CMD --fdtransfer --loop "$interval"s --ttl "$duration"s -e "$profiling_events" -o jfr -f "$jfr_filename" "$process_id"; then
        Error "Unable to start async profiler with loop" >&2
        return 2
      fi
      sleep "$duration" &
    else
      if ! $CMD --timeout "$interval"s --fdtransfer -e "$profiling_events" -o jfr -f "$jfr_filename" "$process_id"; then
        Error "Unable to run async_profiler with timeout" >&2
        return 2
      fi
      sleep "$interval" &
    fi

    trap '$CMD stop ${process_id};rm -rf ${jfr_output_dir};exit 0;' SIGHUP SIGINT SIGTERM
    upload_profiling "$jfr_output_dir" "$start_time" "$process_id" "$process_name"
    wait

    flock -u "$locker_fd"
    exec {locker_fd}>&-
    rm -f "$lock_file"
    rm -rf "$jfr_output_dir"
}

if [ ! -d "$OUTPUT_DIR" ]; then
  mkdir -p "$OUTPUT_DIR"
fi

# 并行采集 profiling 数据
for process_id in $java_process_id; do
  if [ "$duration" -gt 0 ]; then
    time_span=$duration
  else
    time_span=$interval
  fi
  Info "profiling process %d for %ds..." "$process_id" "$time_span"
  bootstrap_profiling "$process_id" &
done

# 等待所有任务结束
wait
