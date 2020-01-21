package grpcserver

import (
	"connekts/common"
	gc "connekts/grpcchannel"
	"connekts/server/model"
	"github.com/sirupsen/logrus"
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
