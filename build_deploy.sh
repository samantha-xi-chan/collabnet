#set -e
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/server        -ldflags "-X main.Version=v1.8-dev -X main.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.GitCommit=$(git rev-parse --short HEAD)"   cmd/version.go cmd/server.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/node_manager  -ldflags "-X main.Version=v1.8-dev -X main.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.GitCommit=$(git rev-parse --short HEAD)"   cmd/version.go cmd/node_manager.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/plugin        -ldflags "-X main.Version=v1.8-dev -X main.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.GitCommit=$(git rev-parse --short HEAD)"   cmd/version.go cmd/plugin.go ;

# HOST="36107_root"
# ssh $HOST "killall node_manager"
# scp release/node_manager  $HOST:/root/;
# scp -r ./config           $HOST:/root/ ;