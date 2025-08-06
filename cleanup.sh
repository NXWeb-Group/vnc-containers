#!/bin/bash

echo "Stopping and removing chrome-instance containers..."
docker ps -q --filter "name=chrome-instance-" | xargs -r docker stop
docker ps -aq --filter "name=chrome-instance-" | xargs -r docker rm

echo "Stopping main chrome-vnc container..."
docker stop chrome-vnc-main 2>/dev/null || true
docker rm chrome-vnc-main 2>/dev/null || true

echo "Removing chrome-vnc-network..."
docker network rm chrome-vnc-network 2>/dev/null || true

echo "Cleanup completed!"
