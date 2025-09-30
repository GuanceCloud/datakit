#!/bin/bash
set -e

until pg_isready -U postgres; do
  sleep 1
done

sleep 5

echo "Creating replication user..."
psql -U postgres -c "CREATE USER replicator WITH REPLICATION ENCRYPTED PASSWORD 'replicatorpassword';" || true

echo "Configuring pg_hba.conf..."
cp $PGDATA/pg_hba.conf $PGDATA/pg_hba.conf.backup
sed -i '/^host.*replication/d' $PGDATA/pg_hba.conf || true
echo "host    replication     replicator      0.0.0.0/0               trust" >> $PGDATA/pg_hba.conf
echo "host    replication     replicator      172.18.0.0/16           trust" >> $PGDATA/pg_hba.conf

echo "Configuring postgresql.conf..."
cp $PGDATA/postgresql.conf $PGDATA/postgresql.conf.backup
sed -i '/^wal_level/d' $PGDATA/postgresql.conf || true
sed -i '/^max_wal_senders/d' $PGDATA/postgresql.conf || true
sed -i '/^wal_keep_size/d' $PGDATA/postgresql.conf || true
sed -i '/^hot_standby/d' $PGDATA/postgresql.conf || true
sed -i '/^listen_addresses/d' $PGDATA/postgresql.conf || true
sed -i '/^ssl/d' $PGDATA/postgresql.conf || true

echo "wal_level = replica" >> $PGDATA/postgresql.conf
echo "max_wal_senders = 10" >> $PGDATA/postgresql.conf
echo "wal_keep_size = 16MB" >> $PGDATA/postgresql.conf
echo "hot_standby = on" >> $PGDATA/postgresql.conf
echo "listen_addresses = '*'" >> $PGDATA/postgresql.conf
echo "ssl = off" >> $PGDATA/postgresql.conf

echo "Reloading PostgreSQL configuration..."
psql -U postgres -c "SELECT pg_reload_conf();" || true
sleep 2

echo "Master configuration completed!"
