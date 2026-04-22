#!/usr/bin/env sh
set -eu

output=""

while [ "$#" -gt 0 ]; do
  case "$1" in
    -f)
      shift
      output="${1:-}"
      ;;
  esac
  shift || true
done

if [ -z "$output" ]; then
  echo "missing output path" >&2
  exit 1
fi

mkdir -p "$(dirname "$output")"
printf 'mock jfr generated at %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" > "$output"
exit 0
