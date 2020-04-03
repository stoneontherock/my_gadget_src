package core

import (
	"context"
	"google.golang.org/grpc"
	"line/client/log"
	"line/grpcchannel"
	"strconv"
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
			closeAllConnection()
			return
		}

		if pong.Action == "lifetime" {
			lt, _ := strconv.Atoi(string(pong.Data))
			if lt <= 0 {
				lt = 1
			}
			dur := time.Duration(lt)
			tk := time.NewTicker(dur)
			go func(tk *time.Ticker) {
				log.Infof("本次任务将于%s终止, 持续%.0f秒\n", time.Now().Add(dur).Format("01-02 15:04:05"), dur.Seconds())
				<-tk.C
				closeAllConnection()
				cancel()
				log.Infof("任务终止\n")
				tk.Stop()
			}(tk)
			continue
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
		println("不支持的action\n")
	}
}

func closeAllConnection() {
	if filesystemServer.server != nil {
		filesystemServer.server.Shutdown(context.TODO())
	}
	filesystemServer.server = nil
	filesystemServer.port2 = ""

	for port2, connList := range port2ConnM {
		for _, conn := range connList {
			log.Infof("关闭conn %p -> port2=%s\n", conn, port2)
			conn.Close()
		}
		delete(port2ConnM, port2)
	}
}
