#!/usr/bin/env bash

set -euo pipefail

DCA_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${DCA_DIR}/.." && pwd)"
RUN_DIR="${DCA_DIR}/.tmp/dev"
BACKEND_PID="${RUN_DIR}/backend.pid"
FRONTEND_PID="${RUN_DIR}/frontend.pid"
BACKEND_LOG="${RUN_DIR}/backend.log"
FRONTEND_LOG="${RUN_DIR}/frontend.log"

DCA_HTTP_PORT="${DCA_HTTP_PORT:-7001}"
DCA_FRONTEND_PORT="${DCA_FRONTEND_PORT:-3000}"
DCA_PROM_LISTEN="${DCA_PROM_LISTEN:-localhost:9090}"
DCA_LOG_PATH="${DCA_LOG_PATH:-stdout}"
DCA_LOG_LEVEL="${DCA_LOG_LEVEL:-debug}"
DCA_BRAND_DOMAIN="${DCA_BRAND_DOMAIN:-guance.com}"
DCA_CONSOLE_API_URL="${DCA_CONSOLE_API_URL:-https://console-api.${DCA_BRAND_DOMAIN}}"
DCA_CONSOLE_WEB_URL="${DCA_CONSOLE_WEB_URL:-https://console.${DCA_BRAND_DOMAIN}}"
DCA_STATIC_BASE_URL="${DCA_STATIC_BASE_URL:-https://static.${DCA_BRAND_DOMAIN}}"

mkdir -p "${RUN_DIR}"

show_help() {
	cat <<EOF
Usage:
  $0 {start|run|stop|status|logs|help} [backend|frontend|all]
  $0 {backend|frontend} {start|run|stop|status|logs}

Commands:
  start     Start service in the background, default: all
  run       Run service in the foreground
  stop      Stop background service, default: all
  status    Show service status, default: all
  logs      Show recent background service logs, default: all
  help      Show this help

Service:
  backend   DCA backend only
  frontend  DCA frontend only
  all       DCA backend and frontend, default

Examples:
  $0 start backend
  $0 start frontend
  $0 start all
  $0 backend start
  $0 frontend start

Environment:
  DCA_HTTP_PORT    Backend HTTP port, default: ${DCA_HTTP_PORT}
  DCA_FRONTEND_PORT Frontend HTTP port, default: ${DCA_FRONTEND_PORT}
  DCA_PROM_LISTEN  Prometheus listen address, default: ${DCA_PROM_LISTEN}
  DCA_LOG_PATH     Backend log path, default: ${DCA_LOG_PATH}
  DCA_LOG_LEVEL    Backend log level, default: ${DCA_LOG_LEVEL}
  DCA_BRAND_DOMAIN Brand domain used to build default Console URLs, default: ${DCA_BRAND_DOMAIN}
  DCA_CONSOLE_API_URL Console API URL, default: ${DCA_CONSOLE_API_URL}
  DCA_CONSOLE_WEB_URL Console web URL, default: ${DCA_CONSOLE_WEB_URL}
  DCA_STATIC_BASE_URL Static base URL, default: ${DCA_STATIC_BASE_URL}

Runtime files:
  ${RUN_DIR}
EOF
}

is_running() {
	local pid_file="$1"
	[[ -f "${pid_file}" ]] || return 1
	local pid
	pid="$(cat "${pid_file}")"
	[[ -n "${pid}" ]] || return 1
	kill -0 "${pid}" >/dev/null 2>&1
}

port_owner() {
	local port="$1"
	if command -v lsof >/dev/null 2>&1; then
		lsof -nP -iTCP:"${port}" -sTCP:LISTEN 2>/dev/null | awk 'NR==2 {print $1 " " $2}' || true
	fi
}

ensure_port_free() {
	local name="$1"
	local port="$2"
	local pid_file="$3"
	local owner
	owner="$(port_owner "${port}")"
	[[ -n "${owner}" ]] || return 0

	if is_running "${pid_file}" && [[ "${owner}" == *" $(cat "${pid_file}")" ]]; then
		return 0
	fi

	echo "${name} port ${port} is already used by: ${owner}"
	echo "Stop that process or choose another port."
	return 1
}

start_backend() {
	if is_running "${BACKEND_PID}"; then
		echo "DCA backend already running: $(cat "${BACKEND_PID}")"
		return
	fi
	ensure_port_free "DCA backend" "${DCA_HTTP_PORT}" "${BACKEND_PID}"

	echo "Starting DCA backend on http://localhost:${DCA_HTTP_PORT}"
	echo "Using Console API: ${DCA_CONSOLE_API_URL}"
	echo "Using Console Web: ${DCA_CONSOLE_WEB_URL}"
	DCA_HTTP_PORT="${DCA_HTTP_PORT}" \
	DCA_PROM_LISTEN="${DCA_PROM_LISTEN}" \
	DCA_LOG_PATH="${DCA_LOG_PATH}" \
	DCA_LOG_LEVEL="${DCA_LOG_LEVEL}" \
	DCA_CONSOLE_API_URL="${DCA_CONSOLE_API_URL}" \
	DCA_CONSOLE_WEB_URL="${DCA_CONSOLE_WEB_URL}" \
	DCA_STATIC_BASE_URL="${DCA_STATIC_BASE_URL}" \
		nohup sh -c 'cd "$1" && exec go run ./cmd/dca' sh "${ROOT_DIR}" >"${BACKEND_LOG}" 2>&1 &
	echo "$!" >"${BACKEND_PID}"
}

start_frontend() {
	if is_running "${FRONTEND_PID}"; then
		echo "DCA frontend already running: $(cat "${FRONTEND_PID}")"
		return
	fi
	ensure_port_free "DCA frontend" "${DCA_FRONTEND_PORT}" "${FRONTEND_PID}"

	echo "Starting DCA frontend on http://localhost:${DCA_FRONTEND_PORT}"
	PORT="${DCA_FRONTEND_PORT}" \
		nohup sh -c 'cd "$1" && exec npm start' sh "${DCA_DIR}/web" >"${FRONTEND_LOG}" 2>&1 &
	echo "$!" >"${FRONTEND_PID}"
}

run_backend_foreground() {
	ensure_port_free "DCA backend" "${DCA_HTTP_PORT}" "${BACKEND_PID}"
	echo "Running DCA backend on http://localhost:${DCA_HTTP_PORT}"
	echo "Using Console API: ${DCA_CONSOLE_API_URL}"
	echo "Using Console Web: ${DCA_CONSOLE_WEB_URL}"

	cd "${ROOT_DIR}"
	DCA_HTTP_PORT="${DCA_HTTP_PORT}" \
	DCA_PROM_LISTEN="${DCA_PROM_LISTEN}" \
	DCA_LOG_PATH="${DCA_LOG_PATH}" \
	DCA_LOG_LEVEL="${DCA_LOG_LEVEL}" \
	DCA_CONSOLE_API_URL="${DCA_CONSOLE_API_URL}" \
	DCA_CONSOLE_WEB_URL="${DCA_CONSOLE_WEB_URL}" \
	DCA_STATIC_BASE_URL="${DCA_STATIC_BASE_URL}" \
	go run ./cmd/dca
}

run_frontend_foreground() {
	ensure_port_free "DCA frontend" "${DCA_FRONTEND_PORT}" "${FRONTEND_PID}"
	echo "Running DCA frontend on http://localhost:${DCA_FRONTEND_PORT}"

	cd "${DCA_DIR}/web"
	PORT="${DCA_FRONTEND_PORT}" npm start
}

run_all_foreground() {
	ensure_port_free "DCA backend" "${DCA_HTTP_PORT}" "${BACKEND_PID}"
	ensure_port_free "DCA frontend" "${DCA_FRONTEND_PORT}" "${FRONTEND_PID}"

	echo "Running DCA backend on http://localhost:${DCA_HTTP_PORT}"
	echo "Using Console API: ${DCA_CONSOLE_API_URL}"
	echo "Using Console Web: ${DCA_CONSOLE_WEB_URL}"
	echo "Running DCA frontend on http://localhost:${DCA_FRONTEND_PORT}"
	echo "Press Ctrl-C to stop both services."

	run_backend_foreground &
	local backend_pid="$!"

	run_frontend_foreground &
	local frontend_pid="$!"

	trap 'kill "${backend_pid}" "${frontend_pid}" >/dev/null 2>&1 || true; wait "${backend_pid}" "${frontend_pid}" 2>/dev/null || true' INT TERM EXIT
	wait "${backend_pid}" "${frontend_pid}"
}

stop_one() {
	local name="$1"
	local pid_file="$2"

	if ! is_running "${pid_file}"; then
		echo "${name} is not running"
		rm -f "${pid_file}"
		return
	fi

	local pid
	pid="$(cat "${pid_file}")"
	echo "Stopping ${name}: ${pid}"
	kill "${pid}" >/dev/null 2>&1 || true
	rm -f "${pid_file}"
}

status_port() {
	local name="$1"
	local port="$2"
	local owner
	owner="$(port_owner "${port}")"
	if [[ -n "${owner}" ]]; then
		echo "${name} port ${port}: used by ${owner}"
	else
		echo "${name} port ${port}: free"
	fi
}

status_one() {
	local name="$1"
	local pid_file="$2"
	local port="$3"

	if is_running "${pid_file}"; then
		echo "${name}: running ($(cat "${pid_file}"))"
		return
	fi

	local owner
	owner="$(port_owner "${port}")"
	if [[ -n "${owner}" ]]; then
		echo "${name}: running (${owner}, pid file not managed)"
	else
		echo "${name}: stopped"
	fi
}

show_logs() {
	local service="${1:-all}"

	case "${service}" in
		backend)
			echo "==> ${BACKEND_LOG}"
			tail -n 80 "${BACKEND_LOG}" 2>/dev/null || true
			;;
		frontend)
			echo "==> ${FRONTEND_LOG}"
			tail -n 80 "${FRONTEND_LOG}" 2>/dev/null || true
			;;
		all)
			show_logs backend
			echo
			show_logs frontend
			;;
		*)
			echo "Unknown service: ${service}"
			show_help
			exit 2
			;;
	esac
}

run_service() {
	local service="${1:-all}"

	case "${service}" in
		backend)
			run_backend_foreground
			;;
		frontend)
			run_frontend_foreground
			;;
		all)
			run_all_foreground
			;;
		*)
			echo "Unknown service: ${service}"
			show_help
			exit 2
			;;
	esac
}

start_service() {
	local service="${1:-all}"

	case "${service}" in
		backend)
			start_backend
			;;
		frontend)
			start_frontend
			;;
		all)
			start_backend
			start_frontend
			;;
		*)
			echo "Unknown service: ${service}"
			show_help
			exit 2
			;;
	esac
}

stop_service() {
	local service="${1:-all}"

	case "${service}" in
		backend)
			stop_one "DCA backend" "${BACKEND_PID}"
			;;
		frontend)
			stop_one "DCA frontend" "${FRONTEND_PID}"
			;;
		all)
			stop_one "DCA frontend" "${FRONTEND_PID}"
			stop_one "DCA backend" "${BACKEND_PID}"
			;;
		*)
			echo "Unknown service: ${service}"
			show_help
			exit 2
			;;
	esac
}

status_service() {
	local service="${1:-all}"

	case "${service}" in
		backend)
			status_one "DCA backend" "${BACKEND_PID}" "${DCA_HTTP_PORT}"
			status_port "DCA backend" "${DCA_HTTP_PORT}"
			;;
		frontend)
			status_one "DCA frontend" "${FRONTEND_PID}" "${DCA_FRONTEND_PORT}"
			status_port "DCA frontend" "${DCA_FRONTEND_PORT}"
			;;
		all)
			status_service backend
			status_service frontend
			;;
		*)
			echo "Unknown service: ${service}"
			show_help
			exit 2
			;;
	esac
}

normalize_args() {
	local first="${1:-start}"
	local second="${2:-all}"

	case "${first}" in
		backend|frontend|all)
			SERVICE="${first}"
			COMMAND="${second}"
			;;
		*)
			COMMAND="${first}"
			SERVICE="${second}"
			;;
	esac
}

normalize_args "${1:-start}" "${2:-all}"

case "${COMMAND}" in
	start)
		start_service "${SERVICE}"
		;;
	run)
		run_service "${SERVICE}"
		;;
	stop)
		stop_service "${SERVICE}"
		;;
	status)
		status_service "${SERVICE}"
		;;
	logs)
		show_logs "${SERVICE}"
		;;
	help|-h|--help)
		show_help
		;;
	*)
		show_help
		exit 2
		;;
esac
