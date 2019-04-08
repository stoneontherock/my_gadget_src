package tserver

import (
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	te "tcpexpose"
)

type ctrlConn struct {
	net.Conn `json:"-"`
	Rd       sync.Mutex `json:"-"`
	Wr       sync.Mutex `json:"-"`
	TTL      int
}

type ctrlChan struct {
	M map[string]*ctrlConn //Key是C侧的ip地址
	sync.RWMutex
}

var CtrlMap ctrlChan

func init() {
	CtrlMap.M = make(map[string]*ctrlConn)
}

func CtrlServe(addr string) {
	cl, err := net.Listen("tcp", addr)
	if err != nil {
		logrus.Fatalf("net.Listen,控制, ERR:%v", err)
	}

	go stateCheck() //fixme: lifeCheck
	for {
		conn, err := cl.Accept()
		if err != nil {
			logrus.Errorf("cl.Accept,控制,ERR: %v", err)
			continue
		}
		logrus.Debugf("CtrlServe: cHost connected, %s", conn.RemoteAddr().String())
		addToMap(conn)
	}
}

func addToMap(c net.Conn) {
	h, _, _ := net.SplitHostPort(c.RemoteAddr().String())
	CtrlMap.Lock()
	if cc, ok := CtrlMap.M[h]; ok {
		logrus.Warnf("addToMap: old ctrl channel is exist, replace it and close old one！！")
		CtrlMap.M[h].Conn = c
		cc.Close()
	} else {
		safeC := ctrlConn{Conn: c, TTL: te.CtrlConnTTL}
		logrus.Debugf("addToMap: cHost ctrl conn add to MAP")
		CtrlMap.M[h] = &safeC
	}
	CtrlMap.Unlock()
}

func stateCheck() {
	tk := time.Tick(time.Second * te.LifeLong)
	for {
		<-tk
		CtrlMap.RLock()
		for h, cc := range CtrlMap.M {
			if cc != nil {
				go stateCheckSingle(h, cc)
			}
		}
		CtrlMap.RUnlock()
	}
}

func stateCheckSingle(host string, cc *ctrlConn) {
	logrus.Debugf("rcving hb... %s", host)
	cc.Conn.SetReadDeadline(time.Now().Add(time.Second * te.LifeLong))

	hdr, _, err := te.ReadProto(cc.Conn)
	if err != nil {
		logrus.Errorf("stateCheckSingle: te.ReadProto:%v", err)
		cc.Conn.Close()
		cutLife(host)
	}

	if hdr.CMD != te.HB {
		logrus.Errorf("stateCheckSingle: Not HB preamble")
		cc.Conn.Close()
		cutLife(host)
	}
	logrus.Debugf("stateCheckSingle: %s alive", host)

}

func getCtrlConn(host string) ctrlConn {
	CtrlMap.RLock()
	cc := *(CtrlMap.M[host])
	CtrlMap.RUnlock()
	return cc
}

func cutLife(host string) {
	CtrlMap.Lock()
	CtrlMap.M[host].TTL--
	if CtrlMap.M[host].TTL == 0 {
		logrus.Debugf("delete %s in CtrlMap", host)
		delete(CtrlMap.M, host)
	}
	CtrlMap.Unlock()
}

func fullLife(host string) {
	CtrlMap.Lock()
	CtrlMap.M[host].TTL = te.CtrlConnTTL
	CtrlMap.Unlock()
}
