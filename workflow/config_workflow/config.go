package config_workflow

const (
	//RepoMySQLDsn = "root:gzn%zkTJ8x!gGZO6@tcp(192.168.31.6:3306)/biz?charset=utf8mb4&parseTime=True&loc=Local"
	RepoLogLevel = 2
	RepoSlowMs   = 200
)

const (
	LISTEN_PORT = ":8081"
)

const MINIO_API_URL = "minio-service:9000"
const MINIO_AK = "admin"
const MINIO_SK = "password"
const MINIO_BUCKET_NAME = "bucket001"

const DockerGroupPref = "/mnt/sss"

const (
	VOL_TOOL         = "base_tool"
	SCRIPT_FILENAME  = "cmd.sh"
	DOCKER_PATH_BIND = "base_tool:/path/in/docker:ro"
	HOSTS_BIND       = "/etc/hosts:/etc/hosts:ro"
	SCRIPT_CONTENT   = `#/bin/sh

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
