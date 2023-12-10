#!/bin/bash

echo "copy config"
mkdir -p dist/conf
cp conf/config.yaml dist/conf/

echo "copy public resource"
mkdir -p dist/static/public
cp -R static/public dist/static/

echo "build"
go build -o dist/wios_server ./main

echo "done"