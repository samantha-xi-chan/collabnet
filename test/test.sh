

OUT=out
rm -rf $OUT
mkdir -p $OUT
go build -o $OUT/scheduler cmd/server.go
go build -o $OUT/node_manager cmd/node_manager.go


curl http://localhost:8080/api/v1/link; echo