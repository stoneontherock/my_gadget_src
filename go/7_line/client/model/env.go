package model

import (
	"fmt"
	"line/common/panicerr"
	"net"
	"os"
	"strconv"
)

var (
	LogLevel          string
	GRPCServerTCPAddr string
	GRPCServerIPaddr  string
	ReportInterval    uint64
)

func init() {
	LogLevel = os.Getenv("LINE_LOG_LEVEL")
	if LogLevel == "" {
		LogLevel = "panic"
	}

	GRPCServerTCPAddr = os.Getenv("LINE_GRPC_SERVER")
	if GRPCServerTCPAddr == "" {
		panicerr.Handle(fmt.Errorf("server空"))
	}

	ReportInterval, _ = strconv.ParseUint(os.Getenv("LINE_REPORT_INTERVAL"), 10, 0)
	if ReportInterval < 30 {
		ReportInterval = 30
	}

	var err error
	GRPCServerIPaddr, _, err = net.SplitHostPort(GRPCServerTCPAddr)
	panicerr.Handle(err)
}