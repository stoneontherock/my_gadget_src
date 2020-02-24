package grpcserver

import (
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"line/grpcchannel"
	"line/server"
	"line/server/db"
	"line/server/model"
	"time"
)

//todo 1-pickup 2-stop client 3-start client 4-del host 然后client就卡住了
func (s *grpcServer) Report(ping *grpcchannel.Ping, stream grpcchannel.Channel_ReportServer) error {
	wanIP := getClientIPAddr(stream.Context())

	ci := model.ClientInfo{ID: ping.Mid}
	err := db.DB.First(&ci).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err := db.DB.Create(&model.ClientInfo{
				ID:       ping.Mid,
				WanIP:    wanIP,
				Hostname: ping.Hostname,
				OS:       ping.Os,
				Interval: ping.Interval,
				StartAt:  ping.StartAt,
			}).Error
			if err != nil {
				logrus.Errorf("Report:Create:%v", err)
				return err
			}
		} else {
			logrus.Errorf("Report:First:%v", err)
			sendFin(stream)
			return err
		}
	}

	//不存在chan, 就初始化 pong chan
	pongC, ok := model.PongM[ping.Mid]
	if !ok {
		pongC = make(chan grpcchannel.Pong)
		model.PongM[ping.Mid] = pongC
	}

	if ci.Pickup <= 0 {
		logrus.Debugf("Report:丢弃%s", ping.Mid)
		sendFin(stream)
		return nil
	}

	if ci.Pickup >= 1 && ci.StartAt != ping.StartAt {
		err := db.DB.Delete(&model.ClientInfo{ID: ping.Mid}).Error
		if err != nil {
			logrus.Errorf("Report:更新会话ID:%v", err)
			return err
		}

		if ci.Pickup == 2 {
			pongC <- grpcchannel.Pong{Action: "fin"} //这里的sendFin是为了关闭已经失效的stream
			model.CloseConnection("", ping.Mid)
		}

		sendFin(stream) //关闭当前的连接
		return nil
	}

	if ci.Pickup == 1 {
		err := ChangePickup(ping.Mid, 2)
		if err != nil {
			logrus.Errorf("Report:Set pickup->2:%v", err)
			return nil
		}
	}

	logrus.Debugf("ci:%+v", ci)

	tk := time.NewTicker(time.Second * server.GRPCPongTimeout)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			logrus.Infof("Report: %s超时,pickup->-1", ping.Mid)
			sendFin(stream)
			return nil
		case pong, ok := <-pongC:
			if !ok || pong.Action == "fin" {
				logrus.Debugf("pongC通道关闭或者收到fin")
				sendFin(stream)
				return nil
			}

			logrus.Infof("Report: id:%s收到pong, action:%s", ping.Mid, pong.Action)
			err = stream.Send(&pong)
			if err != nil {
				logrus.Warnf("Report:stream.Send:%v", err)
				return nil
			}

			if pong.Action == "cmd" {
				logrus.Debugf("Report:创建响应chan...")
				if _, ok := model.CmdOutM[ping.Mid]; !ok {
					model.CmdOutM[ping.Mid] = make(chan grpcchannel.CmdOutput)
					logrus.Debugf("Report:创建响应chan...done")
				}
			}
		}
	}
}

func ChangePickup(mid string, pickup int) error {
	err := db.DB.Model(&model.ClientInfo{ID: mid}).Update("pickup", pickup).Error
	if err != nil {
		logrus.Errorf("Report:更新Pickup值:%v", err)
		return err
	}

	return nil
}

func sendFin(stream grpcchannel.Channel_ReportServer) {
	err := stream.Send(&grpcchannel.Pong{Action: "fin"})
	if err != nil {
		logrus.Warnf("Report:send fin:%v", err)
	}
}
