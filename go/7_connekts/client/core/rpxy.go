package core

import (
	"connekts/client/log"
	"connekts/client/model"
	"connekts/common"
	gc "connekts/grpcchannel"
	"connekts/server/panicerr"
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net"
)

var conn2Pool chan net.Conn

func handleRPxy(pong *gc.Pong, cc gc.ChannelClient, addr3 string ) {
	var reportResp gc.RPxyResp
	err := json.Unmarshal(pong.Data, &reportResp)
	if err != nil {
		log.Errorf("handleRPxy:Unmarshal:pong.Data:%v\n", err)
		return
	}

	log.Infof("收到rpxy的pong: %+v\n", reportResp)
	conn2Pool = make(chan net.Conn, reportResp.NumOfConn2)

	ctx, cancel := context.WithCancel(context.TODO())

	log.Infof("请求rpxy控制端...\n")
	stream, err := cc.RProxyController(ctx, &gc.RPxyReq{Mid: staticInfo.MachineID, Port2: reportResp.Port2, Addr3: reportResp.Addr3, NumOfConn2: reportResp.NumOfConn2})
	if err != nil {
		logrus.Errorf("gc.RProxyController失败,%v", err)
		cancel()
		return
	}
	log.Infof("请求rpxy控制端 done\n")

	for {
		resp, err := stream.Recv()
		if err != nil {
			logrus.Errorf("stream.Recv失败,%v", err)
			cancel()
			break
		}
		log.Infof("收到控制端下发的连接任务:%+v\n", resp)

		rcLen := len(conn2Pool)
		logrus.Infof("rcLen=%d", rcLen)
		nc := int(resp.NumOfConn2)
		if rcLen <= nc/2 {
			genRconn(resp.Port2, nc-rcLen)
		}

		logrus.Infof("收到服务端连接请求,要求建立中转: conn2Addr=%s", resp.Port2)
		if addr3 != "" {
			go bridge(addr3)
		}else{
			go bridge(resp.Addr3)
		}
	}
}

func bridge(addr3 string) {
	conn3, err := net.Dial("tcp", addr3)
	panicerr.Handle(err, "连接近端"+addr3)
	logrus.Infof("近端连接已经建立:%s", conn3.LocalAddr())

	conn2 := <-conn2Pool
	go common.CopyData(conn2, conn3, "2->3", false)
	common.CopyData(conn3, conn2, "2<-3", true)
}

func genRconn(port2 string, cnt int) {
	addr2 := model.ServerIPAddr + port2
	for i := 0; i < cnt; i++ {
		conn2, err := net.Dial("tcp", addr2)
		panicerr.Handle(err, "连接远端:"+addr2)
		logrus.Infof("远端连接已经建立:%s->%s conn2=%p", conn2.LocalAddr(), conn2.RemoteAddr(), conn2)
		conn2Pool <- conn2
		logrus.Infof("conn2已经放入池子 conn2=%p", conn2)
	}
}
