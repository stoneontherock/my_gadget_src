package main

import (
	"github.com/sirupsen/logrus"
	"line/common/log"
	"line/server/db"
	"line/server/grpcserver"
	"line/server/httpserver"
	"line/server/model"
)

func main() {
	log.InitLog(model.LogLevel)

	db.InitSQLite()

	go grpcserver.Serve()
	logrus.Infof("GRPC服务端监听地址%s", model.GRPCListenAddr)

	logrus.Infof("HTTP服务端监听地址%s", model.HTTPListenAddr)
	httpserver.Serve()
}
