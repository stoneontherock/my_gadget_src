package main

import (
	"line/client/core"
	"line/client/model"
	"net"
	"os"
	"strconv"
)

func main() {
	var err error
	model.ServerTCPAddr = os.Getenv("SERVER")
	if model.ServerTCPAddr == "" {
		model.ServerTCPAddr = "32521746.xyz:65000"
	}
	model.ServerIPAddr, _, err = net.SplitHostPort(model.ServerTCPAddr)
	if err != nil {
		println(err.Error())
		os.Exit(127)
	}

	core.ReportInterval, _ = strconv.Atoi(os.Getenv("INTERVAL"))
	if core.ReportInterval <= 0 {
		core.ReportInterval = 30 //默认30秒
	}
	core.Reporter(model.ServerTCPAddr)
}
