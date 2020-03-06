// cmd:   ./scan  -a 1.1.1.1 -p 22-30 -l 500
package main

import (
	"flag"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ip        = flag.String("a", "", "ip addr")
	portRange = flag.String("p", "", "port range, example: 22-1024")
	limit     = flag.Int("l", 20, "go routine limit")
	tmout     = flag.Int64("t", 3000, "tcp dial time out(ms)")
	debug     = flag.Bool("d", false, "print dial error or not")
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
	wg := sync.WaitGroup{}
	wg.Add(portMax - portMin + 1)
	for i := portMin; i <= portMax; i++ {
		maxRoutineCh <- struct{}{}
		go func() {
			err := scanner(*ip, i, time.Duration(*tmout)*time.Millisecond)
			if err != nil && *debug {
				println(err.Error())
			}
			<-maxRoutineCh
			wg.Done()
		}()
	}

	wg.Wait()
	println("----END----")
}

func scanner(ip string, port int, tmout time.Duration) error {
	conn, err := net.DialTimeout("tcp", ip+":"+strconv.Itoa(port), tmout)
	if err != nil {
		return err
	}

	println(port)
	conn.Close()

	return nil
}
