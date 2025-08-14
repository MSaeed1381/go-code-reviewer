#!/bin/bash
set -e

services=(
  "kafka:9092"
  "chroma_db:8000"
  "prometheus:9090"
  "grafana:3000"
)

echo "Waiting for all services to be healthy..."

for svc in "${services[@]}"; do
  host=$(echo $svc | cut -d: -f1)
  port=$(echo $svc | cut -d: -f2)
  echo "Checking $host:$port ..."
  while ! nc -z $host $port; do
    echo "  $host:$port not ready yet. Sleeping 2s..."
    sleep 2
  done
  echo "  $host:$port is up!"
done

echo "All services are ready. Starting application..."
exec "$@"
