#!/bin/bash

docker network create chrome-vnc-network 2>/dev/null || true

docker build ./ --tag chrome-vnc:latest
docker run -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock --network chrome-vnc-network chrome-vnc:latest