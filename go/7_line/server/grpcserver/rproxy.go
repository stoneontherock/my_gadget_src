package grpcserver

import (
	"github.com/sirupsen/logrus"
	"line/common"
	"line/grpcchannel"
	"line/server/model"
)

//todo 从连接池中取出来也不可用
func (s *grpcServer) RProxyController(req *grpcchannel.RPxyReq, stream grpcchannel.Channel_RProxyControllerServer) error {
	logrus.Infof("客户端%s报到", req.Mid)

	for {
		conn1 := <-model.RPxyConn1M[req.Mid]
		logrus.Infof("conn1来了 %s -> %s", conn1.LocalAddr().String(), conn1.RemoteAddr().String())

		conn2Ch := model.RPxyConn2M[req.Mid]
		c2len := len(conn2Ch)
		logrus.Infof("c2len:%d", c2len)

		logrus.Debugf("下发命令，要求3端连到2端...")
		err := stream.Send(&grpcchannel.RPxyResp{Port2: req.Port2, Addr3: req.Addr3, NumOfConn2: req.NumOfConn2})
		if err != nil {
			if c2len == 0 {
				logrus.Errorf("GRPC控制:发送2侧的addr到客户端失败,连接池空")
				conn1.Close()
				return err
			}
			logrus.Warnf("GRPC控制:发送2侧的addr到客户端失败，从连接池中取，连接池剩余连接数%d", c2len)
		} else {
			logrus.Debugf("下发命令，要求3端连到2端...Done")
		}

		conn2 := <-conn2Ch
		conn1cp := conn1
		logrus.Infof("从连接池取的conn2:%p, conn1:%p  %s -> %s", conn2, conn1cp, conn2.RemoteAddr().String(), conn2.LocalAddr().String())

		go common.CopyData(conn1cp, conn2, "1->2", false)
		go common.CopyData(conn2, conn1cp, "1<-2", true)
	}
}
