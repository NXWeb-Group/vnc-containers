#!/bin/bash

set -e 

rm -rf output

mkdir output
mkdir output/docker
mkdir output/frontend

cd frontend
npm install
npm run build
cd ..

cp -r ./docker/chrome ./output/docker
cp -r ./frontend/dist ./output/frontend
cp ./docker/Dockerfile ./output
cp ./start.sh ./output
cp ./cleanup.sh ./output
cp ./README.md ./output

go build -o ./output/vnc ./main.go
