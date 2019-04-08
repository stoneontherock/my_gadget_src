package tserver

import (
	"context"
	"net"
	"strconv"

	"github.com/sirupsen/logrus"

	te "tcpexpose"
)

func SideAServe(ctx context.Context, aHost, cHost string, dPort, portOffset int) {
	aPort := (dPort + portOffset) % 65536
	aAddr := "0.0.0.0:" + strconv.Itoa(aPort)
	al, err := net.Listen("tcp", aAddr)
	if err != nil {
		logrus.Fatalf("SideAServe:net.Listen: %v", err)
	}

	bPort := aPort + 1
	ch := make(chan net.Conn)
	logrus.Debugf("SideAServe: new data serve...")
	go newDataServer(ctx, bPort, cHost, ch)

	for {
		aconn, err := al.Accept()
		if err != nil {
			logrus.Infof("SideAServe:al.Accept: %v", err)
			continue
		}
		h, _, _ := net.SplitHostPort(aconn.RemoteAddr().String())
		if h != aHost {
			logrus.Warnf("SideAServe: client is not aHost")
			continue
		}

		err = requestC2BConn(cHost, bPort, dPort) //Todo : 准备多个连接，以便快速连接
		if err != nil {
			logrus.Errorf("SideAServe:requestC2BConn:%v", err)
			continue
		}

		logrus.Debugf("SideAServe:waiting B->C conn...")
		bcconn := <-ch
		logrus.Debugf("SideAServe:B-C conn recieved.")
		go roundTrip(aconn, bcconn)
	}
}

func roundTrip(abConn, bcConn net.Conn) {
	go te.NetCopy(abConn, bcConn, "A->C")
	te.NetCopy(bcConn, abConn, "C->A")
}

// func fromYtoX(b, a net.Conn) {
// 	logrus.Debugf("Y->X...")
// 	n, err := io.Copy(a, b)
// 	if err != nil {
// 		logrus.Errorf("Y->X, %dByte,err:%v", n, err)
// 		return
// 	}
// 	logrus.Debugf("Y->X..Done, %dBytes", n)
// }
