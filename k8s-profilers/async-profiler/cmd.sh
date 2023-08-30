#!/usr/bin/env bash
mv -f /app/async-profiler/* /app/datakit-profiler/; ./profiling.sh --add-crontab; cron -f
