OutDir=./
protoc ./proto/time.proto --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=$OutDir -I .
protoc ./proto/time.proto --go_out=$OutDir