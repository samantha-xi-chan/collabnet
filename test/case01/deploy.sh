

CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build cmd/server.go; file server
scp -r config 36101_root:/root/
scp ./server  36101_root:/root/


CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build cmd/node_manager.go; file node_manager
scp -r config 36102_root:/root/
scp ./node_manager  36102_root:/root/


CGO_ENABLED=0 GOOS=linux GOARCH=amd64  go build cmd/plugin.go; file plugin
scp ./plugin  36102_root:/root/


