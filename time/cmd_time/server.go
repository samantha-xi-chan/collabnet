package cmd_time

import (
	"collab-net-v2/time/api_time/pb_pkg"
	"collab-net-v2/time/internal/config"
	"collab-net-v2/time/service_time"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
)

type UserServiceInterface interface {
	GetUser(ctx context.Context, req *pb_pkg.UserRequest) (*pb_pkg.UserResponse, error)
	CreateTime(ctx context.Context, req *pb_pkg.CreateTimeRequest) (*pb_pkg.CreateTimeResponse, error)
}

type UserServiceStruct struct {
}

func NewUserService() UserServiceInterface {
	return &UserServiceStruct{}
}

func (userService *UserServiceStruct) GetUser(ctx context.Context, req *pb_pkg.UserRequest) (*pb_pkg.UserResponse, error) {
	response := &pb_pkg.UserResponse{
		Id:   req.Id,
		Name: "Hello World",
	}

	return response, nil
}

func (userService *UserServiceStruct) CreateTime(ctx context.Context, req *pb_pkg.CreateTimeRequest) (*pb_pkg.CreateTimeResponse, error) {
	log.Println("req: ", req)

	id, e := service_time.NewTimer(
		int(req.Timeout),
		int(req.Type),
		req.Holder,
		req.Desc,
		req.CallbackAddr,
	)
	if e != nil {
		log.Println("service_time.NewTimer: e = ", e)
		return nil, e
	}

	response := &pb_pkg.CreateTimeResponse{
		Code: 0,
		Id:   id,
	}

	return response, nil
}

func StartGRPC() {
	l, err := net.Listen("tcp", config.GrpcListenPort)
	if err != nil {
		panic(err)
	}

	fmt.Println("listen on:  ", config.GrpcListenPort)

	grpcServer := grpc.NewServer()
	var userService UserServiceInterface
	userService = NewUserService()
	pb_pkg.RegisterUserServiceServer(grpcServer, userService)

	err = grpcServer.Serve(l)
	if err != nil {
		println(err)
	}
}
