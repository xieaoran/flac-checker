#!/bin/sh

dir="." # 默认当前目录

if [ "${1}" != "" ]; then
    dir=${1}
fi

find "${dir}" -name "*.go" -exec gocmt -t "..." -i {} \;
