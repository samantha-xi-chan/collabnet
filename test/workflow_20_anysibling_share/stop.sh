#!/bin/bash

KeyWord="sh ./test_dag_workflow.sh;"

# 获取进程ID
pid=$(pgrep -f "$KeyWord")

if [ -n "$pid" ]; then
    # 尝试优雅地关闭进程
    if kill $pid 2>/dev/null; then
        echo "进程已关闭，PID: $pid"
    else
        # 如果优雅关闭失败，使用强制终止
        if kill -9 $pid 2>/dev/null; then
            echo "进程已强制关闭，PID: $pid"
        else
            echo "无法关闭进程，可能不存在或没有足够权限"
        fi
    fi
else
    echo "没有找到运行中的进程"
fi
