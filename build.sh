#!/bin/bash
out=diff
dir=package

if [ ! -d $dir ]; then
  mkdir -p $dir
fi

echo "building for mac"
go build -o $dir/mac-$out main.go
echo "building for linux"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $dir/linux-$out main.go
echo "building for windows"
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $dir/windows-$out.exe main.go

#压缩
which upx > /dev/null 2>&1
if [ $? -eq 0 ]; then
  upx $dir/mac-$out > /dev/null
  upx $dir/linux-$out > /dev/null
  upx $dir/windows-$out.exe > /dev/null
fi

cp config.json $dir/
cp mysql.json $dir/
cp mysql-compare.json $dir/
cp README.md $dir/
echo > $dir/debug.log

tar zcf package.tar.gz package

echo "success"