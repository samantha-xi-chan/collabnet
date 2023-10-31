set -e
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/server       cmd/server.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/node_manager cmd/node_manager.go ;
CGO_ENABLED=0 GOOS=linux  GOARCH=amd64 go build -o release/plugin       cmd/plugin.go ;

#scp release/node_manager root@192.168.34.179:/root/;
#exit 2

#HOST="36108_root"
#scp release/node_manager $HOST:/root/;
#HOST="36109_root"
#scp release/node_manager $HOST:/root/;
#HOST="36110_root"
#scp release/node_manager $HOST:/root/;

HOST="36109_root"
# scp release/* $HOST:/root/;  scp -r ./config $HOST:/root/ ;
HOST="36110_root"
scp release/* $HOST:/root/;  scp -r ./config $HOST:/root/ ;

