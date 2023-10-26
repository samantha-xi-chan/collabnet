#!/bin/bash

# 设置要统计的文件夹路径

# 使用 find 命令查找所有以 .go 为扩展名的文件，并使用 xargs 传递给 wc 命令统计行数
line_count=$(find "." -name "*.go" -exec cat {} \; | wc -l)

echo "Golang 代码行数：$line_count"

