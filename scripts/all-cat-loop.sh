#!/bin/bash

# 检查是否提供了参数
if [ $# -eq 0 ]; then
    echo "Usage: $0 <number_of_iterations>"
    exit 1
fi

# 读取命令行参数作为循环次数
file=$1
host=$2
num_iterations=$3

# 使用for循环执行指定次数的迭代
for ((i=1; i<=num_iterations; i++)); do
		curl -H "Content-Type: application/protobuf; proto=com.guance.Point"  "http://$host/v1/write/metric" --data-binary "@$file"  
		curl -H "Content-Type: application/protobuf; proto=com.guance.Point"  "http://$host/v1/write/logging" --data-binary "@$file" 
		curl -H "Content-Type: application/protobuf; proto=com.guance.Point"  "http://$host/v1/write/network" --data-binary "@$file" 
		curl -H "Content-Type: application/protobuf; proto=com.guance.Point"  "http://$host/v1/write/object" --data-binary "@$file"  
		curl -H "Content-Type: application/protobuf; proto=com.guance.Point"  "http://$host/v1/write/tracing" --data-binary "@$file" 
		sleep 0.1 # to avoid HTTP 429
done
