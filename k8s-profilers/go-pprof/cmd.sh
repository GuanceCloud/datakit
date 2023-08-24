#!/usr/bin/env bash
mv -f /app/go-pprof/* /app/datakit-profiler/; ./profiling.sh --add-crontab; cron -f
