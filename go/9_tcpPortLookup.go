// cmd:   ./scan  -a 1.1.1.1 -p 22-30 -l 500
package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

var (
	ip        = flag.String("a", "", "ip addr")
	portRange = flag.String("p", "", "port range, example: 22-1024")
	limit     = flag.Int("l", 20, "go routine limit")
	tmout     = flag.Int("t", 3000, "tcp dial time out(ms)")
)

func main() {
	flag.Parse()
	if *ip == "" || *portRange == "" {
		flag.Usage()
		return
	}

	ss := strings.Split(*portRange, "-")
	portMin, err := strconv.Atoi(ss[0])
	if err != nil {
		log.Fatal("port range is illegal")
	}

	portMax, err := strconv.Atoi(ss[1])
	if err != nil {
		log.Fatal("port range is illegal")
	}

	maxRoutineCh := make(chan struct{}, *limit)
	for i := portMin; i <= portMax; i++ {
		go scanner(*ip, i, *tmout, maxRoutineCh)
	}

	for range maxRoutineCh {
		if len(maxRoutineCh) == 0 {
			break
		}
	}
	println("----END----")
}

func scanner(ip string, port, tmout int, limit chan<- struct{}) {
	conn, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(port), time.Millisecond*time.Duration(tmout))
	if err != nil {
		limit<- struct{}{}
		return
	}

	println(port)
	conn.Close()
	limit<- struct{}{}
}
