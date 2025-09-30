#!/bin/bash
set -e

if [ -f "/var/lib/postgresql/data/PG_VERSION" ]; then
  echo "Database already initialized, replacing config..."
else
  echo "Waiting for database initialization..."
  for i in {1..30}; do
    if [ -f "/var/lib/postgresql/data/PG_VERSION" ]; then
      break
    fi
    sleep 1
  done
fi

cp /tmp/postgresql.conf /var/lib/postgresql/data/postgresql.conf

if [ ! -f "/var/lib/postgresql/data/config_replaced.flag" ]; then
  echo "Restarting PostgreSQL to apply new config..."
  pg_ctl restart -D /var/lib/postgresql/data
  touch /var/lib/postgresql/data/config_replaced.flag
fi
