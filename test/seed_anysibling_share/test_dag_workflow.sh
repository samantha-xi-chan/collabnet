#!/bin/sh

#API_IP="192.168.31.6"
API_IP="192.168.36.102"
API_IP="192.168.34.179"
API_IP="192.168.36.5"
API_IP="192.168.31.45"

echo $API_IP

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

# 创建 workflow
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

sleep 3

# 查询 workflow
API_URL="http://$API_IP:$API_PORT/api/v1/task?workflow_id=$wfId"
echo $API_URL
resultGetTask=$(curl -X GET "$API_URL"  -H "request-id: $RANDOM")
if [ $? -ne 0 ]; then
  echo "Error: Failed to post the HTTP request"
  exit 1
fi
echo $resultGetTask

# 测试业务角度故意杀灭某个任务
secondTaskId=$(echo $resultGetTask | jq -r '.data.task[0].id')
echo $secondTaskId
sleep 3

curl -X GET "http://$API_IP:$API_PORT/api/v1/task/$secondTaskId"


#curl -X GET "http://192.168.31.45:32080/api/v1/task/$secondTaskId"



# 关闭其中的某个 task
#patchTaskResp=$(curl -X PATCH "http://localhost:8081/api/v1/task/$taskId" )
#echo $patchTaskResp
#echo "patchTaskResp: "$patchTaskResp


# 关闭 workflow
#API_URL="http://$API_IP:$API_PORT/api/v1/workflow/$wfId"
#echo $API_URL
#result=$(curl -X PATCH "$API_URL"  -H "request-id: $RANDOM")
#
#echo "PATCH result: "$result