#!/bin/bash
set -e

if [ "$1" != "single" ] && [ "$1" != "replica-set" ]; then
  echo "Usage: $0 <single|replica-set>"
  exit 1
fi

MODE=$1
echo "Cleaning $MODE mode..."

./stop.sh "$MODE"

docker images --filter "label=custom.project=datakit-mongodb-tool" -q | xargs -r docker rmi

cd "$MODE" && docker volume prune -f && cd ..

echo "Clean $MODE mode completed"