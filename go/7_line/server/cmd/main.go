package main

import (
	"github.com/sirupsen/logrus"
	"line/server"
	"line/server/db"
	"line/server/grpcserver"
	"line/server/httpserver"
	"line/server/log"
)

func main() {
	log.InitLog()

	db.InitSQLite()

	go grpcserver.Serve()
	logrus.Infof("GRPC服务端监听地址%s", server.GRPCListenAddr)

	logrus.Infof("HTTP服务端监听地址%s", server.HTTPListenAddr)
	httpserver.Serve()
}
