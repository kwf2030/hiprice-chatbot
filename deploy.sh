#!/bin/bash

cd admin
rm -rf dist/
npm run build

cd ..
go clean
go build -ldflags "-w -s"

target=/var/hiprice/hiprice-chatbot/

mkdir -p $target/admin/
cp -rf admin/dist/. ${target}/admin/
cp -f hiprice-chatbot $target
cp -f conf.yaml $target
cp -f assets/welcome1.png $target
cp -f assets/welcome2.png $target
cp -f assets/welcome.mp4 $target

cd $target
nohup ./hiprice-chatbot > /dev/null 2>&1 &
