// cmd:   ./scan  -a 1.1.1.1 -p 22-30 -l 500
package main

import (
	"flag"
	"log"
	"net"
	"os"
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
		go goroutineControl(i, maxRoutineCh, &wg)
	}

	wg.Wait()
	println("----END----")
}

func goroutineControl(port int, limit <-chan struct{}, wg *sync.WaitGroup) {
	err := scanner(*ip, port, time.Duration(*tmout)*time.Millisecond)
	if err != nil && *debug {
		errStr := err.Error()
		if !strings.HasSuffix(errStr, "connection refused") {
			println(err.Error())
		}
	}
	<-limit
	wg.Done()
}

func scanner(ip string, port int, tmout time.Duration) error {
	p := strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", ip+":"+p, tmout)
	if err != nil {
		return err
	}

	os.Stdout.Write([]byte(p + "\n")) //避免重复打印
	conn.Close()

	return nil
}
