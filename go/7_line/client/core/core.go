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

	mainCtx, mainCancelF := context.WithCancel(context.Background())
	defer mainCancelF()

	stream, err := cc.Report(mainCtx, &grpcchannel.Ping{
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

	finCtx, loopCancelF := context.WithCancel(context.TODO())
	for {
		pong, err := stream.Recv()
		if err != nil {
			log.Errorf("reportDo:stream.Recv: %v\n", err)
			return
		}

		if pong.Action == "lifetime" {
			lt, _ := strconv.Atoi(string(pong.Data))
			if lt <= 0 {
				lt = 1
			}
			dur := time.Duration(lt)
			log.Infof("客户端将于%s释放, 持续%.0f秒 c0=%p\n", time.Now().Add(dur).Format("01-02 15:04:05"), dur.Seconds(), finCtx)
			go func(c0 context.Context) {
				cause := ""
				c1, _ := context.WithTimeout(c0, dur)
				select {
				case <-c1.Done():
					cause = c1.Err().Error()
				case <-c0.Done():
					cause = c0.Err().Error()
				}

				closeAllConnection()
				mainCancelF()
				log.Infof("释放，c0=%p,原因%s\n", c0, cause)
			}(finCtx)
			continue
		}

		if pong.Action == "fin" {
			log.Infof("收到fin, c0=%p\n", finCtx)
			loopCancelF()
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
