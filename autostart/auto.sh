#!/usr/bin/bash

# exporter
#
# chkconfig: 2345 60 80
# description: node exporter service

### BEGIN INIT INFO
# Provides: ftagent
# Required-Start: $network
# Required-Stop: $network
# Default-Start: 2 3 4 5
# Default-Stop: 0 1 6
# Short-Description: collect basic metrics
# Description: The Corsair exporter collect basic metrics such as
#   1. os info
#   2. custom env info
### END INIT INFO

SERVICE=ftcollector
INSTALL_DIR="/usr/local/cloudcare/forethought/ftcollector"
BINARY="${INSTALL_DIR}/${SERVICE}"
YAML_CFG="${INSTALL_DIR}/cfg.yml"
PID="${INSTALL_DIR}/${SERVICE}".pid
LOG="${INSTALL_DIR}/${SERVICE}".log

start() {

    printf "$SERVICE starting... "
    (${BINARY} --cfg "${YAML_CFG}" &) # run in backend

    for i in {1..5}
    do
        if [ ! -f ${PID} ]; then
            sleep 1
            continue
        fi

        pid=`cat "${PID}"`

        # test process running
        if kill -0 "${pid}" 2>/dev/null; then
            printf "ok (Pid: %d)\n" "${pid}"
            return
        else
            sleep 1
            continue
        fi
    done

    printf "failed (see %s for more details)\n" "${LOG}"
}

stop() {
    printf 'stoping %s... ' "${SERVICE}"
    if ps -ef | pgrep "${SERVICE}" | xargs kill -2 &>/dev/null ; then
        printf 'ok\n'
    fi
}

status() {
    if [ -n "$(ps -ef | pgrep "${SERVICE}")" ]; then
        echo "$SERVICE is running"
    else
        echo "$SERVICE stopped"
    fi
}

restart() {
    stop
    status
    printf "$SERVICE restarting..."
    sleep 1
    start
}

case "$1" in
    start|stop|restart|status)
        $1
        ;;
    *)
        echo $"Usage: $0 {start|stop|restart|status}"
        exit 1
    ;;
esac
