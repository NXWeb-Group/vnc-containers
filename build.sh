#!/bin/bash

set -e 

rm -rf output

mkdir output
mkdir output/docker

cp -r ./docker/chrome ./output/docker
cp -r ./public ./output
cp ./docker/Dockerfile ./output
cp ./start.sh ./output
cp ./cleanup.sh ./output
cp ./README.md ./output

go build -o ./output/vnc ./main.go
