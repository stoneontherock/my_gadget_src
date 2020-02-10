package core

import (
	"connekts/client/log"
	gc "connekts/grpcchannel"
	"context"
	"google.golang.org/grpc"
	"time"
)

var staticInfo = static()

func Reporter(addr string, reportInterval time.Duration) {
	i := 0
	for {
		i++
		println(i)
		time.Sleep(reportInterval)
		conn, err := grpc.Dial(addr, grpc.WithInsecure()) //如果需要授权认证或tls加密，则可以使用DialOptions来设置grpc.Dial
		if err != nil {
			log.Errorf("grpc.Dial: %v\n", err)
			continue
		}
		reportDo(conn)
	}
}

func reportDo(conn *grpc.ClientConn) {
	defer conn.Close()
	cc := gc.NewChannelClient(conn) //2.新建一个客户端stub来执行rpc方法

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := cc.Report(ctx, &gc.Ping{Mid: staticInfo.MachineID, Hostname: staticInfo.Hostname, Os: staticInfo.OS})
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

func handlePong(pong gc.Pong, cc gc.ChannelClient) {
	switch pong.Action {
	case "cmd":
		handleCMD(&pong, cc)
	case "rpxy":
		handleRPxy(&pong, cc, "")
	//case "list_file":
	//	handleListFile(&pong,cc)
	//case "file_up":
	//	handleFileUp(&pong,cc)
	case "filesystem":
		handleFilesystem(&pong, cc)
	default:
		println("不支持的acton\n")
	}
}
