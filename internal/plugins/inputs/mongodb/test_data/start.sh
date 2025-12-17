#!/bin/bash
set -e

if [ "$1" != "single" ] && [ "$1" != "replica-set" ]; then
  echo "Usage: $0 <single|replica-set>"
  exit 1
fi

MODE=$1
echo "Starting $MODE mode..."

cd "$MODE" || { echo "Directory $MODE not found"; exit 1; }
docker-compose up --build -d

cd ..