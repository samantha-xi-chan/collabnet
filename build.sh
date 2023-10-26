set -e
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/node_manager cmd/node_manager.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/server       cmd/server.go ;

HOST="36110_root"
scp release/* $HOST:/root/;  scp -r ./config $HOST:/root/ ;