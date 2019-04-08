package tserver

import (
	"context"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"

	te "tcpexpose"
)

var connChan = make(map[string]chan net.Conn) //map的Key是a端的ip地址-c端的IP地址+d端的端口号

func newDataServer(ctx context.Context, bPort int, cHost string, ch chan<- net.Conn) {
	bp := strconv.Itoa(bPort)
	dl, err := net.Listen("tcp", "0.0.0.0:"+bp)
	if err != nil {
		logrus.Fatalf("DataServe:net.Listen,Data: ERR:%v", err)
	}

	for {
		logrus.Debugf("newDataServer: waiting accept")
		bcconn, err := dl.Accept()
		if err != nil {
			logrus.Errorf("DataServe:dl.Accept: ERR: %v", err)
			continue
		}
		h, _, _ := net.SplitHostPort(bcconn.RemoteAddr().String())
		if h != cHost {
			logrus.Warnf("client not cHost,close it")
			bcconn.Close()
		}
		logrus.Debugf("newDataServer: cHost comming, %s, sending conn to channel...", h)
		ch <- bcconn
		logrus.Debugf("newDataServer: conn has been sent")
	}
}

func requestC2BConn(cHost string, bPort, dPort int) error {
	logrus.Debugf("requestC2BConn: getting ctrlConn...")
	cc := getCtrlConn(cHost)
	logrus.Debugf("requestC2BConn: ctrlConn got")
	cc.Wr.Lock()
	defer cc.Wr.Unlock()
	err := te.NewDataChannel(cc.Conn, bPort, dPort)
	logrus.Debugf("requestC2BConn: new data channel CMD has sent, ERR:%v", err)
	return err
}
