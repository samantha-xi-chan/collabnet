#./$1

current_dir="$(cd "$(dirname "$0")" && pwd)"
echo "\$current_dir：$current_dir"
cd $current_dir
pwd
nohup sh -c "while true; do sh ./test_dag_workflow.sh; done" > ./$(date -u +'%Y-%m-%dT%H:%M:%SZ').log &



