package main

import (
	"github.com/sirupsen/logrus"
	"line/server/db"
	"line/server/grpcserver"
	"line/server/httpserver"
	"line/server/log"
	"os"
)

func main() {
	getEnv(&log.Debug, "LINE_STDOUT_DEBUG")
	log.InitLog()

	db.InitSQLite()

	getEnv(&grpcserver.GRPCServer.ListenAddr, "LINE_GRPC_LISTEN_ADDR")
	go grpcserver.Serve()
	logrus.Infof("GRPC服务端监听地址%s", grpcserver.GRPCServer.ListenAddr)

	getEnv(&httpserver.AdminName, "LINE_HTTP_ADMIN")
	getEnv(&httpserver.AdminPv, "LINE_HTTP_PASSWD")

	var addr = ":65080"
	getEnv(&addr, "LINE_HTTP_LISTEN_ADDR")
	logrus.Infof("HTTP服务端监听地址%s", addr)
	httpserver.Serve(addr)
}

func getEnv(value *string, envKey string) {
	tmp := os.Getenv(envKey)
	if tmp == "" {
		return
	}

	*value = tmp
}
