package main

import (
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

	getEnv(&httpserver.AdminName, "LINE_HTTP_ADMIN")
	getEnv(&httpserver.AdminPv, "LINE_HTTP_PASSWD")

	var addr = ":65080"
	getEnv(&addr, "LINE_HTTP_LISTEN_ADDR")
	httpserver.Serve(addr)
}

func getEnv(value *string, envKey string) {
	tmp := os.Getenv(envKey)
	if httpserver.AdminPv == "" {
		return
	}

	*value = tmp
}
