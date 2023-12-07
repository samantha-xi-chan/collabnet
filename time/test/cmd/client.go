package main

import (
	"collab-net-v2/time/api_time/pb_pkg"
	"collab-net-v2/time/internal/config"
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

func main() {
	conn, err := grpc.Dial(fmt.Sprintf("%s%s", config.GrpcIpAddr, config.GrpcListenPort), grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	client := pb_pkg.NewUserServiceClient(conn)
	req := &pb_pkg.UserRequest{
		Id: 1,
	}
	response, e := client.GetUser(context.Background(), req)
	if e != nil {
		handleGRPCError(e)
		return
	}

	resp, err := json.Marshal(response)
	fmt.Printf("%s", resp)
}

func handleGRPCError(err error) {
	// 从 gRPC 错误中获取状态
	st, ok := status.FromError(err)
	if !ok {
		// 不是 gRPC 错误，可能是其他类型的错误
		log.Fatalf("Unexpected error: %v", err)
	}

	// 根据 gRPC 状态处理错误
	switch st.Code() {
	case codes.NotFound:
		log.Println("Resource not found:", st.Message())
	case codes.InvalidArgument:
		log.Println("Invalid argument:", st.Message())
	case codes.Internal:
		log.Println("Internal server error:", st.Message())
	default:
		log.Println("Unknown error:", st.Code(), st.Message())
	}
}
