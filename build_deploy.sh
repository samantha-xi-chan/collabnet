set -e
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/server        -ldflags "-X main.Version={{.Version}} -X main.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.GitCommit=$(git rev-parse --short HEAD)"   cmd/version.go cmd/server.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/node_manager  -ldflags "-X main.Version={{.Version}} -X main.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.GitCommit=$(git rev-parse --short HEAD)"   cmd/version.go cmd/node_manager.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/plugin        -ldflags "-X main.Version={{.Version}} -X main.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.GitCommit=$(git rev-parse --short HEAD)"   cmd/version.go cmd/plugin.go ;

# go run   -ldflags "-X main.Version={{.Version}} -X main.BuildTime=$(date -u +'%Y-%m-%dT%H:%M:%SZ') -X main.GitCommit=$(git rev-parse --short HEAD)"   cmd/version.go cmd/node_manager.go ;

# scp release/node_manager 34179_root:/root/;

#HOST="36108_root"
#scp release/node_manager $HOST:/root/;
#HOST="36109_root"
#scp release/node_manager $HOST:/root/;
#HOST="36110_root"
#scp release/node_manager $HOST:/root/;
#
#HOST="36109_root"
## scp release/* $HOST:/root/;  scp -r ./config $HOST:/root/ ;
#HOST="36110_root"
#scp release/* $HOST:/root/;  scp -r ./config $HOST:/root/ ;

