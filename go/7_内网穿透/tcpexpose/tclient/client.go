package tclient

import (
	"context"
	"encoding/binary"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	te "tcpexpose"
)

type safeNetConn struct {
	net.Conn
	Rd sync.Mutex
	Wr sync.Mutex
}

var ctrlConn safeNetConn //key是baddr
var BIPAddr string

func ConnectToControler(baddr string) {
	BIPAddr = baddr[:strings.LastIndex(baddr, ":")]
	for {
		bc, err := net.Dial("tcp", baddr)
		if err != nil {
			logrus.Errorf("net.Dial CTRL: %v", err)
			time.Sleep(time.Second * 20)
			continue
		}
		//接收信息
		worker(bc)
	}
}

func worker(bc net.Conn) {
	defer bc.Close()

	ctrlConn.Conn = bc
	ctx, cancel := context.WithCancel(context.Background())

	go rcvCMD(ctx)
	heartBeat(ctx)
	cancel()
}

func heartBeat(ctx context.Context) {
	for {
		begin := time.Now()

		logrus.Debugf("sending hb")
		ctrlConn.Wr.Lock()
		err := te.WriteProto(ctrlConn.Conn, nil, te.HB)
		ctrlConn.Wr.Unlock()
		if err != nil {
			logrus.Errorf("heartBeat: write HB to ctrlChannel failed, %v", err)
			return
		}

		sleep := time.Second*te.LifeLong - time.Now().Sub(begin)
		logrus.Debugf("hb sent. sleep %f", sleep.Seconds())
		time.Sleep(sleep)
	}
}

func rcvCMD(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		ctrlConn.Rd.Lock()
		hdr, body, err := te.ReadProto(ctrlConn.Conn)
		if err != nil {
			ctrlConn.Rd.Unlock()
			logrus.Warnf("rcvCMD: te.ReadProto: %v", err)
			continue
		}
		ctrlConn.Rd.Unlock()

		switch hdr.CMD {
		case te.NDC:
			logrus.Debugf("rcvCMD: NDC CMD comming")
			newDataConnection(body)
		default:
			logrus.Errorf("rcvCMD: unknown CMD, %d", hdr.CMD)
		}
	}
}

func newDataConnection(buf []byte) {
	bPort := strconv.Itoa(int(binary.BigEndian.Uint16(buf[:2])))
	dPort := strconv.Itoa(int(binary.BigEndian.Uint16(buf[2:])))

	bdataAddr := net.JoinHostPort(BIPAddr, bPort)
	bconn, err := net.Dial("tcp", bdataAddr)
	if err != nil {
		logrus.Errorf("net.Dial到Data: %v", err)
		return
	}
	logrus.Debugf("newDataConnection: connected to bDataAddr %s", bdataAddr)

	dconn, err := net.Dial("tcp", ":"+dPort)
	if err != nil {
		logrus.Errorf("net.Dial到D: %v", err)
		return
	}

	logrus.Debugf("newDataConnection: connected to dAddr %s", ":"+dPort)
	go roundTrip(bconn, dconn)
}

func roundTrip(bconn, dconn net.Conn) {
	go te.NetCopy(bconn, dconn, "B->C")
	te.NetCopy(dconn, bconn, "C->B")
}
