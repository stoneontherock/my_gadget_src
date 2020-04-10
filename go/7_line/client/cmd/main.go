package main

import (
	"line/client/core"
	"line/client/model"
	"line/common/log"
)

func main() {
	log.InitLog(model.LogLevel)
	core.Reporter(model.GRPCServerTCPAddr)
}
