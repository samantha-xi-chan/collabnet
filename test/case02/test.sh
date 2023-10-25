#!/bin/sh


current_dir="$(cd "$(dirname "$0")" && pwd)"
echo "$current_dir：$current_dir"
cd $current_dir
pwd

# 生成 数据实例
templateFile="prototype.json"
instanceFile="new_task.json"
cp -rf $templateFile $instanceFile

# 注册退出信号的处理函数
trap cleanup EXIT
cleanup() {
    exit_code=$?
    echo "脚本即将退出，执行清理操作，退出代码为: $exit_code"
    rm -rf $instanceFile
}

serverIp=192.168.36.5 #localhost #192.168.36.101

linkEndpoint=localhost:1080
taskEndpoint=localhost:2080

echo "link: "
linkResp=$(curl -X GET "http://$linkEndpoint/api/v1/link")
echo $linkResp
first_id=$(echo "$linkResp" | jq -r '.data[0].id')
echo $first_id

if [ -z "$first_id" ]; then
    echo "first_id 为空，退出脚本"
    exit 1
fi
#if [ "$first_id" == "null" ]; then
#    echo "\$first_id 为 null，退出脚本"
#    exit 1
#fi

search_string="co"
replace_string="$first_id"

os_name=$(uname -s)
case "$os_name" in
    "Darwin")
        sed -i '' "s/$search_string/$replace_string/g" "$instanceFile"
        ;;
    "Linux")
        sed -i "s/$search_string/$replace_string/g" "$instanceFile"
        ;;
    *)
        # 在其他操作系统上执行的代码
        ;;
esac

# task
DAG=$current_dir"/"$instanceFile
req=$(cat $DAG)
echo "$req"

postTaskResp=$(curl -X POST "http://$taskEndpoint/api/v1/task" -d "$req")
echo $postTaskResp
taskId=$(echo "$postTaskResp" | jq -r '.data.id')
echo "taskId: "$taskId

sleep 1
#patchTaskResp=$(curl -X PATCH "http://$taskEndpoint/api/v1/task/$taskId" )
#echo $patchTaskResp
#echo "patchTaskResp: "$patchTaskResp

#sleep 5; echo "get tasks: "
#curl -X GET "http://$serverIp:8081/api/v1/task"; echo ;
#sleep 5; echo "get tasks: "
#curl -X GET "http://$serverIp:8081/api/v1/task";  echo "\n\n";
