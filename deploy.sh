#!/bin/bash

go build -ldflags "-w -s"
mkdir -p /hiprice/admin
cp -f hiprice-chatbot /hiprice/chatbot
cp -f conf.yaml /hiprice/
cp -rf assets/. /hiprice/
go clean

cd admin
rm -rf dist/
yarn run build
cp -rf dist/. /hiprice/admin/

cd /hiprice
nohup ./chatbot > /dev/null 2>&1 &