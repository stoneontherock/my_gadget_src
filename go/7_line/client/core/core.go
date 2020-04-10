package core

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"line/client/model"
	"line/common/connection/pb"
	"strconv"
	"time"
)

var staticInfo = static()
var startAt = int32(time.Now().Unix())

func Reporter(addr string) {
	logrus.Infof("报告间隔%d秒", model.ReportInterval)
	i := 0
	for {
		i++
		logrus.Debugf("report: i=%d", i)
		time.Sleep(time.Duration(model.ReportInterval) * time.Second)
		conn, err := grpc.Dial(addr, grpc.WithInsecure()) //如果需要授权认证或tls加密，则可以使用DialOptions来设置grpc.Dial
		if err != nil {
			logrus.Errorf("grpc.Dial: %v", err)
			continue
		}
		reportDo(conn)
		conn.Close()
	}
}

func reportDo(conn *grpc.ClientConn) {
	cc := pb.NewChannelClient(conn) //2.新建一个客户端stub来执行rpc方法

	mainCtx, mainCancelF := context.WithCancel(context.Background())
	defer mainCancelF()

	stream, err := cc.Report(mainCtx, &pb.Ping{
		Mid:      staticInfo.MachineID,
		Kernel:   staticInfo.Kernel,
		OsInfo:   staticInfo.OsInfo,
		Interval: int32(model.ReportInterval),
		StartAt:  startAt,
	})

	if err != nil {
		logrus.Errorf("reportDo:c.Report: %v", err)
		return
	}

	finCtx, loopCancelF := context.WithCancel(context.TODO())
	for {
		pong, err := stream.Recv()
		if err != nil {
			logrus.Errorf("reportDo:stream.Recv: %v", err)
			return
		}

		if pong.Action == "lifetime" {
			lt, _ := strconv.Atoi(string(pong.Data))
			if lt <= 0 {
				lt = 1
			}
			dur := time.Duration(lt) * time.Second
			logrus.Infof("客户端将于%s释放, 持续%d秒, c0=%p", time.Now().Add(dur).Format("01-02 15:04:05"), lt, finCtx)
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
				logrus.Infof("释放，c0=%p,原因%s", c0, cause)
			}(finCtx)
			continue
		}

		if pong.Action == "fin" {
			logrus.Infof("收到fin, c0=%p", finCtx)
			loopCancelF()
			return
		}

		logrus.Infof("收到被捡起的pong: %v", pong)
		go handlePong(*pong, cc)
	}
}

func handlePong(pong pb.Pong, cc pb.ChannelClient) {
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
		logrus.Errorf("不支持的action")
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
			logrus.Infof("关闭conn %p -> port2=%s", conn, port2)
			conn.Close()
		}
		delete(port2ConnM, port2)
	}
}
