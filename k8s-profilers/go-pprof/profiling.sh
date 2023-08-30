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
PPROF_VERSION=unknown
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
       --procname <keyword>     Your keyword of application process name, it will use this keyword to find process PID (pgrep -of "keyword")
                                if "--pid" option not set.
       --events <events>        Specify one or more events to profile, available events: cpu|heap|mutex|block|goroutine, default: "cpu,heap".
   -a, --add-crontab            Add your profiling schedule to Linux crontab, use with "--schedule".
       --schedule <schedule>    Your profiling schedule, use the same format as Linux crontab (for example, "* */5 * * *", "00 02 * * *"),
                                use with "-a/--add-crontab".
       --pprof_port <port>      The go package "net/http/pprof" listening at.

   -h, --help                   display this help.
   -v, --version                display version.
   -s, --show-all-events        show all profiling events supported by this tool.


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
   DK_PROFILE_PROCNAME     Your golang application process name keyword, using this keyword to search process PID (pgrep -of "keyword") if DK_PROFILE_PID not set.
   DK_PROFILE_EVENT        Supported profiling event: cpu|heap|mutex|block|goroutine etc, default: cpu,heap.
   DK_PROFILE_SCHEDULE     Your profiling schedule, use setting like crontab (for example, "* */5 * * *", "00 02 * * *").
   DK_PROFILE_PPROF_PORT   Listening port of go pprof package "net/http/pprof", default: find from 'netstat -ltnop'.

 Example:
   ./profiling.sh --host 192.168.1.1 --service my-app --ver 1.0.0 --env test --duration 300 --interval 30 --pid 13 --events cpu,alloc
   DK_AGENT_HOST=192.168.1.1 DK_PROFILE_PID=13 DK_PROFILE_EVENT="cpu,heap" DK_PROFILE_SERVICE=my-app DK_PROFILE_VERSION=1.0.0 DK_PROFILE_ENV=test ./profiling.sh

EOT
}


if ! opts=$(getopt -o H:P:S:V:E:D:I:hvas -l host:,port:,service:,ver:,env:,duration:,interval:,hostname:,pid:,procname:,pprof-port:,schedule:,events:,help,version,add-crontab,show-all-events -- "$@"); then
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
procname=""
profiling_events="cpu,heap"
pprof_port=0


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
    --procname)
      shift
      procname="$1"
      ;;
    --pprof_port)
      shift
      pprof_port="$1"
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

if [ -n "$DK_PROFILE_PROCNAME" ]; then
  procname="$DK_PROFILE_PROCNAME"
fi


if [ -n "$DK_PROFILE_PID" ]; then
  pid="$DK_PROFILE_PID"
fi

if [ -n "$DK_PROFILE_PPROF_PORT" ]; then
  pprof_port=$DK_PROFILE_PPROF_PORT
fi

if [ -n "$DK_PROFILE_SCHEDULE" ]; then
  schedule="$DK_PROFILE_SCHEDULE"
fi

if [ -n "$DK_PROFILE_EVENT" ]; then
    profiling_events=$DK_PROFILE_EVENT
fi

profiling_events=${profiling_events//,/ }
profiling_events="${profiling_events} "

if ! [[ $profiling_events =~ ^([[:blank:]]*(cpu|heap|mutex|block|goroutine)[[:blank:]]+)+$ ]]; then
  Error "invalid profiling_events: %s, only one or more of 'cpu heap mutex block goroutine' supported" "$profiling_events" >&2
  exit 1
fi

fetch_cpu_event=""

if [[ "$profiling_events" =~ cpu ]]; then
  fetch_cpu_event="cpu"
  profiling_events_exclude_cpu=${profiling_events//cpu/}
fi


# 允许上传至 DataKit 的 jfr 文件大小 (6 M)，请勿修改
# shellcheck disable=SC2034
MAX_JFR_FILE_SIZE=6000000

if [ -z "$datakit_host" ]; then
  Error "datakit_host not set" >&2
  exit 1
fi

# DataKit 服务地址
datakit_url=http://"$datakit_host":"$datakit_port"

# 上传 profiling 数据的完整地址
datakit_profiling_url=$datakit_url/profiling/v1/input

PPROF_URL="http://127.0.0.1:%d/debug/pprof/"

if [ "$pprof_port" -eq 0 ]; then
  mapfile -t networks < <(ss -lntpH)

  for ((i=0; i<${#networks[@]}; i++))
  do
  network=${networks[$i]}
  port=$(echo "$network" | awk '{print $4}')
  port=$(echo "${port##*:}" | grep -oE '[0-9]+')
  pprof_pid=$(echo "$network" | awk '{print $6}' | grep -oE 'pid=[0-9]+,' | grep -oE '[0-9]+')
  if ! [[ $port =~ ^[1-9][0-9]*$ ]]; then
    continue
  fi
  # shellcheck disable=SC2059
  pprof_url=$(printf "$PPROF_URL" "$port")
  if ! match=$(curl -sL "$pprof_url" 2>/dev/null | grep -i 'goroutine' -); then
    continue
  fi
  if [ ${#match} -ge 9 ]; then
    pprof_port=$port
    if [ "$pid" -eq 0 ] && [ "$pprof_pid" -gt 0 ]; then
      pid=$pprof_pid
    fi
    break
  fi
  done
fi

if [ "$pprof_port" -le 0 ]; then
  Error "unable to find listening port of net/http/pprof" >&2
  exit 1
fi

# shellcheck disable=SC2059
pprof_url_prefix=$(printf "$PPROF_URL" "$pprof_port")

cpu_fetch_url="${pprof_url_prefix}profile?seconds=${interval}"
heap_fetch_url="${pprof_url_prefix}heap?gc=1"
mutex_fetch_url="${pprof_url_prefix}mutex"
block_fetch_url="${pprof_url_prefix}block"
goroutine_fetch_url="${pprof_url_prefix}goroutine"

if [ -n "$procname" ] && [ "$pid" -eq 0 ]; then
  pid=$(pgrep -of "$procname" 2>/dev/null)
fi

# 采集的进程 ID，此处可以自定义需要采集的进程，比如可以根据进程名称过滤
if [[ "$pid" =~ ^[0-9]+$ ]] && [ "$pid" -gt 0 ]; then
  go_process_id=$pid
else
  Error "invalid process PID: %s" "$pid" >&2
  exit 1
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
    echo "${schedule} ${SHELL_PATH} --host $datakit_host --port $datakit_port --service $app_service --env $app_env --ver $app_version --duration $duration --interval $interval --pid $pid --pprof_port $pprof_port --hostname $hostname >> $LOG_DIR/main.log 2>&1 " | crontab -u root -
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

get_delta_pporf() {
  event=$1
  index=$2
  delta_file=$3

  pre_filename=$(get_file_name "$event" $((index-1)))
  if ! stat --printf=%n "$pre_filename" >/dev/null 2>&1; then
    Warn "previous profiling file not found: %s" "$pre_filename" 2>&1
    return 1
  fi

  base_events=""

  case $event in
  heap)
    base_events="alloc_objects,alloc_space"
    ;;
  mutex|block)
    base_events="contentions,delay"
    ;;
  esac

  if ! pprof -proto -output "$delta_file" -base "$pre_filename" -base_events "$base_events" "$filename" 1>&2; then
    Warn "Unable to generate delta profiling file: %s" "$delta_file" >&2
    return 2
  fi

  return 0
}

get_delta_filename() {
    event=$1
    index=$2
    echo "delta_${event}_${index}.pb.gz"
}

get_event_filename() {
    idx=$1
    echo "event_${idx}.json"
}

get_file_name() {
    event=$1
    index=$2
    echo "${event}_${index}.pb.gz"
}

get_file_path() {
    out_dir=$1
    event=$2
    index=$3
    echo "${out_dir}/${event}_${index}.pb.gz"
}

get_form_file_name() {
    event=$1
    case $event in
    cpu)
      echo "cpu.pprof"
      ;;
    goroutine)
      echo "goroutines.pprof"
      ;;
    heap|block|mutex)
      echo "delta-${event}.pprof"
      ;;
    esac
}

remove_unused_files() {
  index=$1

  del_files_index=0
  for event in $profiling_events
  do
    delete_files[$del_files_index]=$(get_file_name "$event" "$index")
    ((del_files_index+=1))
    delete_files[$del_files_index]=$(get_delta_filename "$event" "$index")
    ((del_files_index+=1))
  done
  delete_files[$del_files_index]=$(get_event_filename "$index")
  ((del_files_index+=1))

  rm -f "${delete_files[@]}"
}



upload_profiling() {
  pprof_out_dir=$1
  start_time=$2
  process_id=$3
  process_name=$4
  index=$5


  if ! cd "$pprof_out_dir"; then
    Error "Unable to enter dir %s" "$pprof_out_dir" >&2
    return 1
  fi

  max_mtime=0
  array_idx=0
  for event in $profiling_events
  do
    filename=$(get_file_name "$event" "$index")
    if ! file_mtime=$(stat --printf=%Y "$filename" >/dev/null 2>&1); then
      Warn "file not found: %s" "$filename" >&2
      continue
    fi

    delta_file=$(get_delta_filename "$event" "$index")
    if [[ $event =~ ^heap|mutex|block$ ]]; then

      if ! get_delta_pporf "$event" "$index" "$delta_file" ; then
        Warn "unable to generate delta pprof file for event: %s and index: %s" "$event" "$index" >&2
        continue
      fi

    else
      if ! mv -f "$filename" "$delta_file"; then
        Warn "Unable to rename file %s to %s" "$filename" "$delta_file" >&2
        continue
      fi
    fi
    Info "succeed to generate delta file: %s" "$delta_file"
    succeed_events[$array_idx]="$event"
    succeed_files[$array_idx]="$delta_file"
    if [ "$file_mtime" -gt $max_mtime ]; then
      max_mtime=$file_mtime
    fi
    ((array_idx+=1))
  done

  if [ "$array_idx" -eq 0 ]; then
    Error "no available profiling files to upload" >&2
    return 1
  fi
  end_time=$(date --date="@$max_mtime" +%FT%T.%N%:z)

  attachments=""
  param_idx=0
  for ((i=0;i<${#succeed_events[@]};i++))
  do
    event=${succeed_events[$i]}
    filename=${succeed_files[$i]}
    form_file_name=$(get_form_file_name "$event")

    curl_params[$param_idx]="-F"
    ((param_idx+=1))
    curl_params[$param_idx]="$form_file_name=@$filename;filename=${form_file_name}"
    ((param_idx+=1))

    if [ -z "$attachments" ]; then
      attachments="\"${form_file_name}\""
    else
      attachments="$attachments,\"${form_file_name}\""
    fi
  done

  event_json_file=$(get_event_filename "$index")

  cat >"$event_json_file" <<END
{
  "attachments": [$attachments],
  "tags_profiler": "library_version:$PPROF_VERSION,library_type:pprof,process_id:$process_id,process_name:$process_name,service:$app_service,host:$hostname,env:$app_env,version:$app_version",
  "start": "$start_time",
  "end": "$end_time",
  "family": "golang",
  "format": "pprof"
}
END
  Info "start to upload profile file to DataKit"
  res=$(curl -m 120 -s "$datakit_profiling_url" "${curl_params[@]}" -F "event=@$event_json_file;filename=event.json;type=application/json")

  if [[ $res != *ProfileID* ]]; then
      Error "send profile file to datakit failed" >&2
      echo "$res" >&2
  else
      Info "Successfully send profile file to datakit"
  fi

  remove_unused_files $((index-1))
}

fetch_profiling_file() {
  event=$1
  filename=$2

  case $event in
  cpu)
    url=$cpu_fetch_url
    ;;
  heap)
    url=$heap_fetch_url
    ;;
  mutex)
    url=$mutex_fetch_url
    ;;
  block)
    url=$block_fetch_url
    ;;
  goroutine)
    url=$goroutine_fetch_url
    ;;
  *)
    Error "event unsupported: %s" "$event" >&2
    return 1
    ;;
  esac

  if ! curl -sfL --max-time 600 -o "$filename" "$url" 2>/dev/null; then
    Error "Unable to fetch profiling file for event: %s and filename: %s " "$event" "$filename" >&2
    return 2
  fi

  return 0
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

    lock_file=$LOCKER_DIR/pprof-$process_id.lock
    exec 888> "$lock_file"
    if ! flock -xn 888; then
      Warn "Unable to get lock of file: %s, probably profiling on the target process is already running" "$lock_file" >&2
      exec 888>&-
      rm -f "$lock_file"
      return 1
    fi

    pprof_out_dir=$OUTPUT_DIR/output-$process_id-$(date +%s%N)
    if [ ! -d "$pprof_out_dir" ]; then
      mkdir -p "$pprof_out_dir"
    fi

    if ! cd "$pprof_out_dir"; then
      Error "unable to cd %s" "$pprof_out_dir"
      exit 1
    fi

    PPROF_TMPDIR=/tmp/datakit_pprof
    if [ ! -d $PPROF_TMPDIR ]; then
      mkdir -p $PPROF_TMPDIR
    fi

    process_name=$procname
    if [ -z "$procname" ]; then
      process_name=$(ps -p "$process_id" -o comm=)
    fi

    process_start_time=$(date +%s)
    file_index=0

    # create initial profiling file
    for event in $profiling_events_exclude_cpu
    do
      filename=$(get_file_path "$pprof_out_dir" "$event" $file_index)
      fetch_profiling_file "$event" "$filename" &
    done
    wait

    ((file_index+=1))

    while ((1))
    do
      start_time=$(date +%FT%T.%N%:z)
      if [ -n "$fetch_cpu_event" ]; then
        filename=$(get_file_path "$pprof_out_dir" $fetch_cpu_event "$file_index")
        fetch_profiling_file $fetch_cpu_event "$filename" &
      fi
      sleep "$interval"

      for event in $profiling_events_exclude_cpu
      do
        filename=$(get_file_path "$pprof_out_dir" "$event" "$file_index")
        fetch_profiling_file "$event" "$filename" &
      done
      wait

      upload_profiling "$pprof_out_dir" "$start_time" "$process_id" "$process_name" "$file_index" &

      ((file_index+=1))
      now_unix_time=$(date +%s)
      if [ $((now_unix_time - process_start_time)) -ge "$duration" ]; then
        break
      fi
    done

    wait
    flock -u 888
    exec 888>&-
    rm -f "$lock_file"
    rm -rf "$pprof_out_dir"
}

if [ ! -d "$OUTPUT_DIR" ]; then
  mkdir -p "$OUTPUT_DIR"
fi

Info "profiling process %d..." "$process_id"
bootstrap_profiling "$go_process_id"

# 等待所有任务结束
wait
