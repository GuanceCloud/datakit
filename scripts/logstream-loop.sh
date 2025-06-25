#!/bin/bash

# 检查是否提供了参数
if [ $# -eq 0 ]; then
    echo "Usage: $0 <number_of_iterations>"
    exit 1
fi

# 读取命令行参数作为循环次数
num_iterations=$1
file=$2

# 使用for循环执行指定次数的迭代
for ((i=1; i<=num_iterations; i++)); do
		curl -v "http://localhost:19529/v1/write/logstreaming?source=drop-testing&storage_index=index_logstream" --data-binary "@$2"
		sleep 0.1 # to avoid HTTP 429
done
