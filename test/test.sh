

OUT=out
rm -rf $OUT
mkdir -p $OUT
go build -o $OUT/scheduler cmd/scheduler.go
go build -o $OUT/node_manager cmd/node_manager.go
