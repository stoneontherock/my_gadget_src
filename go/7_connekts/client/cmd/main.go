package main

import (
	"connekts/client/core"
	"connekts/client/model"
	"net"
	"os"
)

func main() {
	var err error
	model.ServerTCPAddr = os.Getenv("KSERVER")
	model.ServerIPAddr, _, err = net.SplitHostPort(model.ServerTCPAddr)
	if err != nil {
		println("invalid KSERVER")
		os.Exit(127)
	}

	core.Reporter(model.ServerTCPAddr)
}
