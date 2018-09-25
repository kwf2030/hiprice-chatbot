#!/bin/bash

go build -ldflags "-w -s"
mkdir -p /hiprice/admin
cp -f hiprice-chatbot /hiprice/chatbot
cp -f conf.yaml /hiprice/
cp -f assets/welcome1.png /hiprice/
cp -f assets/welcome2.png /hiprice/
cp -f assets/welcome.mp4 /hiprice/
go clean

cd admin
rm -rf dist/
yarn run build
cp -rf dist/. /hiprice/admin/

cd /hiprice
nohup ./chatbot > /dev/null 2>&1 &