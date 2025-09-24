#!/bin/bash
set -e

echo "Waiting for master to be ready..."
sleep 10  # 给主库更多时间完成初始化
until pg_isready -h $MASTER_HOST -p $MASTER_PORT -U postgres; do
  echo "Still waiting for master..."
  sleep 3
done

mkdir -p $PGDATA

if [ -z "$(ls -A "$PGDATA")" ]; then
  echo "Starting base backup from master..."
  rm -rf $PGDATA/*
  PGPASSWORD=$REPLICATION_PASSWORD pg_basebackup -h $MASTER_HOST -p $MASTER_PORT -U $REPLICATION_USER -D $PGDATA -Fp -Xs -P -R --no-password
  
  echo "hot_standby = on" >> "$PGDATA/postgresql.conf"
  
  if [ ! -f "$PGDATA/standby.signal" ]; then
    touch "$PGDATA/standby.signal"
  fi
  
  echo "Replication setup completed."
  exit 0
else
  echo "Data directory is not empty, skipping base backup."
  exit 0
fi
