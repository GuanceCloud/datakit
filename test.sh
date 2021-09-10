#!/bin/bash
all=(
	./...
)

# truncate 可能要单独安装(linux 一般自带)
# Mac: brew install truncate
truncate -s 0 test.output

for pkg in "${all[@]}"
do
	echo "testing $pkg" | tee -a test.output
	GO111MODULE=off CGO_ENABLED=0 go test -cover $pkg |tee -a test.output
done
