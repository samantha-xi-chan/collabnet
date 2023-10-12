#!/bin/sh

current_dir="$(cd "$(dirname "$0")" && pwd)"
echo "$current_dir：$current_dir"
DAG=$current_dir"/new_task.json"
req=$(cat $DAG)
echo "$req"

curl -X POST "http://localhost:8081/api/v1/task" -d "$req"


while true; do sh test/case01/test.sh; sleep 30; done