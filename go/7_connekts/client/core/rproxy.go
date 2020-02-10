package core

import (
	"connekts/client/log"
	"connekts/client/model"
	"connekts/common"
	gc "connekts/grpcchannel"
	"context"
	"encoding/json"
	"net"
)

var conn2Pool chan net.Conn

func handleRPxy(pong *gc.Pong, cc gc.ChannelClient, addr3 string) {
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
		log.Errorf("gc.RProxyController失败,%v\n", err)
		cancel()
		return
	}
	log.Infof("请求rpxy控制端 done\n")

	for {
		resp, err := stream.Recv()
		if err != nil {
			log.Errorf("stream.Recv失败,%v\n", err)
			cancel()
			break
		}
		log.Infof("收到控制端下发的连接任务:%+v\n", resp)

		rcLen := len(conn2Pool)
		log.Infof("rcLen=%d\n", rcLen)
		nc := int(resp.NumOfConn2)
		if rcLen <= nc/2 { //todo 注意地板除的情况
			genRconn(resp.Port2, nc-rcLen)
		}

		log.Infof("收到服务端连接请求,要求建立中转: conn2Addr=%s\n", resp.Port2)
		if addr3 != "" {
			go bridge(addr3)
		} else {
			go bridge(resp.Addr3)
		}
	}
}

func bridge(addr3 string) {
	conn3, err := net.Dial("tcp", addr3)
	if err != nil {
		log.Errorf("连接近端失败,addr3:" + addr3)
		return
	}
	log.Infof("近端连接已经建立:%s\n", conn3.LocalAddr())

	conn2 := <-conn2Pool
	go common.CopyData(conn2, conn3, "2->3", false)
	common.CopyData(conn3, conn2, "2<-3", true)
}

func genRconn(port2 string, cnt int) {
	addr2 := model.ServerIPAddr + port2
	for i := 0; i < cnt; i++ {
		conn2, err := net.Dial("tcp", addr2)
		if err != nil {
			log.Errorf("连接远端失败,addr2:" + addr2)
			return
		}
		log.Infof("远端连接已经建立:%s->%s conn2=%p\n", conn2.LocalAddr(), conn2.RemoteAddr(), conn2)
		conn2Pool <- conn2
		log.Infof("conn2已经放入池子 conn2=%p\n", conn2)
	}
}
