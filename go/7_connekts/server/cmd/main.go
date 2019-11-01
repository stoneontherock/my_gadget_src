package main

import (
	"connekts/server/db"
	"connekts/server/grpcserver"
	"connekts/server/httpserver"
	"connekts/server/log"
)

func main() {
	log.InitLog()
	db.InitSQLite()
	go grpcserver.Serve()

	httpserver.Serve("0.0.0.0:32768")
}
