#!/usr/bin/env bash
# Unless explicitly stated otherwise all files in this repository are licensed
# under the MIT License.
# This product includes software developed at Guance Cloud (https://www.guance.com/).
# Copyright 2021-present Guance, Inc.

CMD="py-spy-for-datakit"
WORK_DIR=$(cd "$(dirname "$0")" && pwd)
SHELL_PATH=$WORK_DIR/"$(basename "$0")"
LOCKER_DIR=$WORK_DIR/locks
LOG_DIR=$WORK_DIR/log

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

 DataKit CPython profiling tool for docker/k8s.

 Options:
   -H, --host <host>            Datakit listening host.
   -P, --port <port>            DataKit listening port, default: 9529.
   -S, --service <service>      Your application name, default: unknown.
   -V, --ver  <version>         Your application version (for example, 2.5, 202003181415, 1.3-alpha), default: unknown.
   -E, --env <env>              Your application environment (for example, production, staging), default: unknown.
   -D, --duration <duration>    Run profiling for <duration> seconds, default: infinity.
   -I, --interval <interval>    Send a batch of profiling data to datakit per <interval> seconds, default: 60.
       --hostname <hostname>    The host your application running on.
       --pid <pid>              Your application process PID, default: read from ps command.
   -a, --add-crontab            Add your profiling schedule to Linux crontab, use with "--schedule".
       --schedule <schedule>    Your profiling schedule, use the same format as Linux crontab (for example, "* */5 * * *", "00 02 * * *"), use with "-a/--add-crontab".

   -h, --help                   display this help.
   -v, --version                display version.


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
   DK_PROFILE_SCHEDULE     Your profiling schedule, use setting like crontab (for example, "* */5 * * *", "00 02 * * *").

 Example:
   ./profiling.sh --host 192.168.1.1 --port 9529 --service my-app --ver 1.0.0 --env test --duration 300 --interval 30 --pid 13
   DK_AGENT_HOST=192.168.1.1 DK_AGENT_PORT=9529 DK_PROFILE_PID=13 DK_PROFILE_SERVICE=my-app DK_PROFILE_VERSION=1.0.0 DK_PROFILE_ENV=test ./profiling.sh

EOT
}

if ! opts=$(getopt -o H:P:S:V:E:D:I:hva -l host:,port:,service:,ver:,env:,duration:,interval:,hostname:,pid:,schedule:,help,version,add-crontab -- "$@"); then
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
    -h|--help)
      show_help
      exit 0
      ;;
    -v|--version)
      $CMD --version
      exit 0
      ;;
    -a|--add-crontab)
      add_crontab=1
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

install_crontab() {
  if [ ! -d "$LOG_DIR" ]; then
    mkdir -p "$LOG_DIR"
  fi
  if [ -n "$schedule" ]; then
    echo "${schedule} ${SHELL_PATH} --host $datakit_host --port $datakit_port --service $app_service --env $app_env --ver $app_version --duration $duration --interval $interval --pid $pid --hostname $hostname >> $LOG_DIR/main.log 2>&1" | crontab -u root -
  fi
}

capture_pid() {
  pid=$1

  if [[ ! "$pid" =~ ^[1-9][0-9]*$ ]]; then
    Warn "capture_pid invalid pid: %s" "$pid" >&2
    retrun 1
  fi

    lock_file=$LOCKER_DIR/py-spy-$pid.lock

    exec {lock_fd}> "$lock_file"
    if ! flock -xn "$lock_fd"; then
      Warn "Unable to get lock of file: %s, probably profiling on the target process is already running" "$lock_file" >&2
      exec {lock_fd}>&-
      rm -f "$lock_file"
      return 1
    fi


    $CMD datakit --host "$datakit_host" --port "$datakit_port" --service "$app_service" --version "$app_version" \
    --env "$app_env" --pid "$pid" --duration "$interval" --loop "$duration" --subprocesses

    flock -u "$lock_fd"
    exec {lock_fd}>&-
    rm -f "$lock_file"

}

run_profiling() {

  if [ -z "$datakit_host" ]; then
    Error "datakit_host not set" >&2
    exit 1
  fi

#   if [ "$pid" -le 0 ]; then
#     # shellcheck disable=SC2009
#     pid=$(ps -eH -o pid,comm --no-headers | grep -v grep | grep "python" | head -n 1 | awk '{print $1}')
#     if [ "$pid" -le 0 ]; then
#       Error "Unable to get process pid automatically, please specify it by pass flag '--pid' or env 'DK_PROFILE_PID'" >&2
#       exit 1
#     else
#       Info "find python process: %d" "$pid"
#     fi
#   fi

    python_pids=()
    pids_idx=0

    if [ "$pid" -le 0 ]; then
      # shellcheck disable=SC2009
      mapfile -t processes < <(ps -e -o pid,cmd --no-headers | grep -v grep | grep "python" | head -n 20)

      for ((i=0; i < ${#processes[@]}; i++))
      do
        IFS=" " read -r -a process <<< "${processes[i]}"

        if [ ${#process[@]} -lt 2 ]; then
            continue
        fi

        process_id=${process[0]}
        cmd=${process[1]}

         if [[ ! "$cmd" =~ python([23](\.[0-9]{1,3})?)?$ ]]; then
            continue
         fi

         if [[ ! "$process_id" =~ ^[1-9][0-9]*$ ]]; then
            continue
         fi

         python_pids[$pids_idx]=$process_id
         ((pids_idx++))
      done

    else
      python_pids[$pids_idx]=$pid
    fi

    if [ ${#python_pids[@]} -lt 1 ]; then
        Error "Unable to find python process automatically, please specify it by pass flag '--pid' or env 'DK_PROFILE_PID'" >&2
        exit 1
    else
       Info "find python process: %d" "${python_pids[*]}"
    fi

  if [ ! -d "$LOCKER_DIR" ]; then
    mkdir -p "$LOCKER_DIR"
  fi

  for ((i=0; i<${#python_pids[@]}; i++))
  do
    capture_pid "${python_pids[i]}" &
  done
  wait
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

if ! run_profiling; then
  Error "Unable to run profiling" >&2
  exit 1
fi
