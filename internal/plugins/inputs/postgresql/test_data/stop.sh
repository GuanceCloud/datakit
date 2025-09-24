#!/bin/bash
set -e

if [ "$1" != "single" ] && [ "$1" != "master-slave" ]; then
  echo "Usage: $0 <single|master-slave>"
  exit 1
fi

MODE=$1
echo "Stopping $MODE mode..."

cd "$MODE" || { echo "Directory $MODE not found"; exit 1; }
docker-compose down

cd ..