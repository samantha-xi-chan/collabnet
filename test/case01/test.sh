#!/bin/sh


current_dir="$(cd "$(dirname "$0")" && pwd)"
echo "$current_dir：$current_dir"
cd $current_dir

# 生成 数据实例
templateFile="template.json"
instanceFile="new_task.json"
cp -rf $templateFile $instanceFile

# 注册退出信号的处理函数
trap cleanup EXIT
cleanup() {
    exit_code=$?
    echo "脚本即将退出，执行清理操作，退出代码为: $exit_code"
    rm -rf $instanceFile
}

echo "link: "
linkResp=$(curl -X GET "http://localhost:8080/api/v1/link")
echo $linkResp
first_id=$(echo "$linkResp" | jq -r '.data[0].id')
echo $first_id

if [ -z "$first_id" ]; then
    echo "first_id 为空，退出脚本"
    exit 1
fi
if [ "$first_id" == "null" ]; then
    echo "\$first_id 为 null，退出脚本"
    exit 1
fi

search_string="co"
replace_string="$first_id"
sed -i '' "s/$search_string/$replace_string/g" "$instanceFile"

# task
DAG=$current_dir"/"$instanceFile
req=$(cat $DAG)
echo "$req"

curl -X POST "http://localhost:8081/api/v1/task" -d "$req"

sleep 2
curl -X GET "http://localhost:8081/api/v1/task"; echo ;
sleep 5
curl -X GET "http://localhost:8081/api/v1/task";  echo "\n\n";
