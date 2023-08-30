#!/usr/bin/env bash
mv -f /app/py-spy/* /app/datakit-profiler/; ./profiling.sh --add-crontab; cron -f
