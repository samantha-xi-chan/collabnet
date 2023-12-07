# 以下文件中的 斜杠符号 加 dollar符 是为了屏蔽自动解析变量



current_dir="$(cd "$(dirname "$0")" && pwd)"
echo "$current_dir：$current_dir"
cd $current_dir

rm -rf pb_pkg
sleep 1

OutDir=./
protoc ./proto/user.proto --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=$OutDir -I .
protoc ./proto/user.proto --go_out=$OutDir

cd -
