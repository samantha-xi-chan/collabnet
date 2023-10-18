

OUT=out
rm -rf $OUT
mkdir -p $OUT
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $OUT/scheduler cmd/server.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $OUT/node_manager cmd/node_manager.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $OUT/plugin cmd/plugin.go

file $OUT/*

# curl http://localhost:8080/api/v1/link; echo