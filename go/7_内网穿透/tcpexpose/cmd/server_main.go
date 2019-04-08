package main

import (
	"flag"
	"os"
	"tcpexpose"
	hs "tcpexpose/httpserver"
	ts "tcpexpose/tserver"
)

var addrHTTP = flag.String("http", "", "addr of A side(A side is http server.)")
var addrCtrl = flag.String("ctrl", "", "addr of C side(C side: remote_internal)")

var stdout = flag.Bool("stdout", true, "print log to stdout")
var maxSize = flag.Int("maxsize", 5, "maximum size in MBytes of the log file before it gets rotated")
var maxBackups = flag.Int("maxbackups", 3, "maximum number of old log files to retain")

func main() {
	flag.Parse()
	tcpexpose.InitLog(*stdout, *maxSize, *maxBackups)

	if *addrHTTP == "" || *addrCtrl == "" {
		flag.Usage()
		os.Exit(1)
	}

	//	logrus.Infof("A侧监听%s，C侧监听：%s", *addrA, *addrC)

	go hs.HTTPServe(*addrHTTP)
	ts.CtrlServe(*addrCtrl)
}
