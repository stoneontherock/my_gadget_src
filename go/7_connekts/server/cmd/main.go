package main

import (
	"line/server/db"
	"line/server/grpcserver"
	"line/server/httpserver"
	"line/server/log"
	"fmt"
)

func main() {
	fmt.Printf("提示：环境变量LOGTO=stdout可以把日志答应到stdout上\n")
	log.InitLog()
	db.InitSQLite()
	go grpcserver.Serve()

	httpserver.Serve("0.0.0.0:65080")
}
