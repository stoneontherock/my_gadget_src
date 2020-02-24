package grpcserver

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"line/grpcchannel"
	"line/server"
	"line/server/panicerr"
	"net"
)

func getClientIPAddr(ctx context.Context) string {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "0.0.0.0"
	}

	h, _, err := net.SplitHostPort(pr.Addr.String())
	if err != nil {
		return "0.0.0.0"
	}

	return h
}

type grpcServer struct{}

func Serve() {
	lis, err := net.Listen("tcp", server.GRPCListenAddr) //1.指定监听地址:端口号
	panicerr.Handle(err, "grpcserver:Serve:net.Listen")

	s := grpc.NewServer()                               //2.新建gRPC实例
	grpcchannel.RegisterChannelServer(s, &grpcServer{}) //3.在gRPC服务器注册我们的服务实现。参数2是接口(满足服务定义的方法)。在.pb.go文件中搜索Register关键字即可找到这个函数签名
	err = s.Serve(lis)
	panicerr.Handle(err, "grpcserver:Serve:s.Serve")
}
