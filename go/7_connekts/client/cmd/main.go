package main

import (
	"line/client/core"
	"line/client/model"
	"net"
	"os"
	"strconv"
	"time"
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

	interval, _ := strconv.Atoi(os.Getenv("INTERVAL"))
	if interval <= 0 {
		interval = 30 //默认30秒
	}
	core.Reporter(model.ServerTCPAddr, time.Duration(interval)*time.Second)
}
