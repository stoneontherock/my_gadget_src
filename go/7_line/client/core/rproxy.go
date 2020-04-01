package core

import (
	"context"
	"encoding/json"
	"line/client/log"
	"line/client/model"
	"line/common"
	"line/grpcchannel"
	"net"
)

var conn2Pool chan net.Conn
var port2ConnM = make(map[string][]net.Conn)

//todo 多次连接后， conn很多。conn的重用未实现
func handleRPxy(pong *grpcchannel.Pong, cc grpcchannel.ChannelClient, fsAddr3 string) {
	var reportResp grpcchannel.RPxyResp
	err := json.Unmarshal(pong.Data, &reportResp)
	if err != nil {
		log.Errorf("handleRPxy:Unmarshal:pong.Data:%v\n", err)
		return
	}

	log.Infof("收到rpxy的pong: %+v\n", reportResp)
	conn2Pool = make(chan net.Conn, reportResp.NumOfConn2)

	ctx, cancel := context.WithCancel(context.TODO())

	log.Infof("请求rpxy控制端...\n")
	stream, err := cc.RProxyController(ctx, &grpcchannel.RPxyReq{Mid: staticInfo.MachineID, Port2: reportResp.Port2, Addr3: reportResp.Addr3, NumOfConn2: reportResp.NumOfConn2})
	if err != nil {
		log.Errorf("grpcchannel.RProxyController失败,%v\n", err)
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
		if rcLen < nc {
			genRconn(resp.Port2, nc-rcLen)
		}

		log.Infof("收到服务端连接请求,要求建立中转: conn2Addr=%s\n", resp.Port2)
		if fsAddr3 != "" {
			filesystemServer.port2 = resp.Port2
			go bridge(fsAddr3, resp.Port2)
		} else {
			go bridge(resp.Addr3, resp.Port2)
		}
	}
}

func bridge(addr3, port2 string) {
	conn3, err := net.Dial("tcp", addr3)
	if err != nil {
		log.Errorf("连接近端失败,addr3:%s\n", addr3)
		return
	}
	log.Infof("近端连接已经建立:%s\n", conn3.LocalAddr())
	port2ConnM[port2] = append(port2ConnM[port2], conn3)

	conn2 := <-conn2Pool
	go common.CopyData(conn2, conn3, "2->3", false)
	common.CopyData(conn3, conn2, "2<-3", true)
}

func genRconn(port2 string, cnt int) {
	addr2 := model.ServerIPAddr + port2
	for i := 0; i < cnt; i++ {
		conn2, err := net.Dial("tcp", addr2)
		if err != nil {
			log.Errorf("连接远端失败,addr2:%s\n", addr2)
			return
		}
		log.Infof("远端连接已经建立:%s->%s conn2=%p\n", conn2.LocalAddr(), conn2.RemoteAddr(), conn2)
		port2ConnM[port2] = append(port2ConnM[port2], conn2)
		conn2Pool <- conn2
		log.Infof("conn2已经放入池子 conn2=%p\n", conn2)
	}
}

func handleCloseConnections(pong *grpcchannel.Pong) {
	port2 := string(pong.Data)
	if filesystemServer.port2 == port2 {
		if filesystemServer.server != nil {
			filesystemServer.server.Shutdown(context.TODO())
		}
		filesystemServer.server = nil
		filesystemServer.port2 = ""
	}

	//log.Infof("**** port2ConnM=%v pong.Data=%v\n", port2ConnM, port2)
	connList, ok := port2ConnM[port2]
	if !ok {
		return
	}
	for _, conn := range connList {
		log.Infof("关闭conn %p\n", conn)
		conn.Close()
	}

	delete(port2ConnM, port2)
}
