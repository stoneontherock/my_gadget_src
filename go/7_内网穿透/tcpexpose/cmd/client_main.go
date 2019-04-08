package main

import (
	"flag"
	"os"
	"tcpexpose"
	"tcpexpose/tclient"
)

func main() {
	var baddr = flag.String("ctrl", "", "addr of B side(B side: middle server)")
	var stdout = flag.Bool("stdout", true, "print log to stdout")
	var maxSize = flag.Int("maxsize", 5, "maximum size in MBytes of the log file before it gets rotated")
	var maxBackups = flag.Int("maxbackups", 3, "maximum number of old log files to retain")

	flag.Parse()
	tcpexpose.InitLog(*stdout, *maxSize, *maxBackups)

	if *baddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	tclient.ConnectToControler(*baddr)
}
