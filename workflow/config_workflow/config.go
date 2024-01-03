package config_workflow

import "fmt"

const (
	//RepoMySQLDsn = "root:gzn%zkTJ8x!gGZO6@tcp(192.168.31.6:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local"
	RepoLogLevel = 2
	RepoSlowMs   = 200
)

const (
	LISTEN_PORT = ":8081"

	UrlPathSetting  = "/api/v1/setting"
	UrlPathTask     = "/api/v1/task"
	UrlPathWorkflow = "/api/v1/workflow"

	UrlPathWorkflowV2 = "/api/v2/workflow"
)

const (
	EVT_TIMEOUT_WORKFLOW = 70017759
)

const MINIO_API_URL = "minio-service:9000"
const MINIO_AK = "admin"
const MINIO_SK = "password"

const MINIO_BUCKET_NAME_INTERTASK = "workflowintertask"
const MINIO_BUCKET_NAME_WF = "workflowshare"

const DockerGroupPref = "/mnt/sss"
const SSSDefaultObjectName = "server"
const DefaultServerSignPath = "/mnt/sss/server"

const (
	SHELL_PATH      = "/bin/sh"
	VOL_TOOL        = "base_tool"
	CMD_DIR         = "/in_docker"
	SCRIPT_FILENAME = "cmd.sh"
	HOSTS_BIND      = "/etc/hosts:/etc/hosts:ro"
)

var DockerPathBind = fmt.Sprintf("base_tool:%s:ro", CMD_DIR)

const (
	SCRIPT_CONTENT = `#/bin/sh

cleanup() {
    rm -r "$temp_dir"
}

trap cleanup EXIT

temp_dir=$(mktemp -d)
cd "$temp_dir"

if [ "${1%%:*}" = "http" ]; then
  curl -s -o remote.sh --fail "$1"
  if [ $? -ne 0 ]; then
    echo "Failed to download the remote script."
    exit 5
  fi
elif [ "${1%%:*}" = "base64" ]; then
  encoded="${1#*:}"
  decoded=$(printf "%s" "$encoded" | base64 -d)
  printf "%s" "$decoded" > remote.sh
else
    echo "unsupported format "
fi

shift

sh remote.sh "$@"
remote_script_exit_code=$?

if [ $remote_script_exit_code -ne 0 ]; then
#  echo "remote_script_exit_code: $remote_script_exit_code."
  exit $remote_script_exit_code
fi

exit 0
`
)
