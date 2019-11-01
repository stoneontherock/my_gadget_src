package grpcserver

import (
	"connekts/common"
	gc "connekts/grpcchannel"
	"connekts/server/model"
	"connekts/server/panicerr"
	"github.com/sirupsen/logrus"
	"net"
)

func (s *server) RProxyController(req *gc.RPxyReq, stream gc.Channel_RProxyControllerServer) error {
	logrus.Infof("客户端%s报到", req.Mid)

	for {
		conn1 := <-model.RPxyConn1M[req.Mid]
		logrus.Infof("conn1来了 %s -> %s", conn1.LocalAddr().String(), conn1.RemoteAddr().String())

		conn2C := model.RPxyConn2M[req.Mid]
		c2len := len(conn2C)
		logrus.Infof("c2len:%d", c2len)

		err := stream.Send(&gc.RPxyResp{Port2: req.Port2, Addr3: req.Addr3, NumOfConn2: req.NumOfConn2})
		if err != nil {
			logrus.Errorf("GRPC控制:发送2侧的addr到客户端失败")
			conn1.Close()
			return err
		}
		logrus.Infof("下发命令，要求3端连到2端")

		conn2 := <-conn2C
		logrus.Infof("conn2:%p  %s -> %s", conn2, conn2.LocalAddr().String(), conn2.RemoteAddr().String())

		go common.CopyData(conn1, conn2, "1->2", false)
		go common.CopyData(conn2, conn1, "1<-2", true)
	}
}

func RProxyListen(mid, port1, port2 string, numOfConn2 int) error {
	connC1 := make(chan *net.TCPConn)
	conn2Pool := make(chan *net.TCPConn, numOfConn2)

	model.RPxyConn1M[mid] = connC1
	model.RPxyConn2M[mid] = conn2Pool

	taddr1, err := net.ResolveTCPAddr("tcp", port1)
	if err != nil {
		return err
	}

	taddr2, err := net.ResolveTCPAddr("tcp", port2)
	if err != nil {
		return err
	}

	go listen(taddr1, connC1)
	go listen(taddr2, conn2Pool)

	return nil
}

func listen(addr *net.TCPAddr, connC chan *net.TCPConn) {
	lis, err := net.ListenTCP("tcp", addr)
	panicerr.Handle(err, "监听1侧失败:"+addr.String())

	for {
		conn, err := lis.AcceptTCP()
		if err != nil {
			logrus.Warnf("lis.Accept:addr=%s err:%v", addr, err)
			continue
		}
		connC <- conn
	}
}
