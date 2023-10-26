#!/bin/sh

#API_IP="192.168.31.6"
API_IP="192.168.36.102"
API_PORT="30181"
DOWNLOAD_HOST="8_root"

current_dir="$(cd "$(dirname "$0")" && pwd)"
echo "$current_dir：$current_dir"
DAG=$current_dir"/user_provide/dag.json"

os_name=$(uname -s)
case "$os_name" in
    "Darwin")
        scp -r $current_dir/user_provide/* $DOWNLOAD_HOST:/usr/http_download/static
        ;;
    "Linux")
        # 在Linux上执行的代码
        ;;
    *)
        # 在其他操作系统上执行的代码
        ;;
esac

API_URL="http://$API_IP:$API_PORT/api/v1/workflow"

req=$(cat $DAG)
#echo "$req"

# curl -X POST "http://192.168.31.8:8081/api/v1/workflow"  -H "request-id: $RANDOM" -d "$req"
result=$(curl -X POST "$API_URL"  -H "request-id: $RANDOM" -d "$req")
if [ $? -ne 0 ]; then
  echo "Error: Failed to post the HTTP request"
  exit 1
fi

echo "result: "$result

wfId=$(echo "$result" | jq -r '.data.id')
if [ -z "$wfId" ]; then
  echo "Error: Failed to extract workflow ID from the response"
  exit 1
fi

echo
echo "wfId: " $wfId

sleep 2

API_URL="http://$API_IP:$API_PORT/api/v1/task?workflow_id=$wfId"
resultGetTask=$(curl -X GET "$API_URL"  -H "request-id: $RANDOM")
if [ $? -ne 0 ]; then
  echo "Error: Failed to post the HTTP request"
  exit 1
fi
echo $resultGetTask


#    "import_obj_id":"task_1694855605076icxk",
#    "import_obj_as":"/docker/imported/",
