package grpcserver

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"line/grpcchannel"
	"line/server/db"
	"line/server/model"
	"time"
)

func (s *grpcServer) Report(ping *grpcchannel.Ping, stream grpcchannel.Channel_ReportServer) error {
	wanIP := getClientIPAddr(stream.Context())

	needFin := true
	defer func() {
		if needFin {
			sendFin(stream) //关闭当前stream
		}
	}()

	ci := model.ClientInfo{ID: ping.Mid}
	err := db.DB.First(&ci).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			err := db.DB.Create(&model.ClientInfo{
				ID:         ping.Mid,
				WanIP:      wanIP,
				Kernel:     ping.Kernel,
				OsInfo:     ping.OsInfo,
				Interval:   ping.Interval,
				StartAt:    ping.StartAt,
				LastReport: int32(time.Now().Unix()),
			}).Error
			if err != nil {
				logrus.Errorf("Report:Create:%v", err)
				return err
			}
		} else {
			logrus.Errorf("Report:First:%v", err)
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
		return nil
	}

	if ci.StartAt != ping.StartAt {
		err := db.DB.Delete(&model.ClientInfo{ID: ping.Mid}).Error
		if err != nil {
			logrus.Errorf("Report:删除:%v", err)
		}

		if ci.Pickup == 2 {
			go func() {
				model.CloseAllConnections(ping.Mid)
				pongC <- grpcchannel.Pong{Action: "fin"} //这里的sendFin是为了关闭已经失效的stream
			}()
		}

		return errors.New("startAt标记已经变更")
	}

	if ci.Pickup == 1 {
		ChangePickup(ping.Mid, 2)
		needFin = false
	}

	logrus.Debugf("ci:%+v", ci)

	now := time.Now()
	deadline, _ := time.ParseInLocation("2006-01-02 15:04:05", ci.Timeout, now.Location())
	tmout := deadline.Sub(now)
	if tmout <= time.Second*60 {
		tmout = time.Second * 60
	}

	tk := time.NewTicker(tmout)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			logrus.Infof("Report: %s超时,发fin", ping.Mid)
			ChangePickup(ping.Mid, -1)
			model.CloseAllConnections(ping.Mid)
			needFin = true
			return nil
		case pong, ok := <-pongC:
			if !ok || pong.Action == "fin" {
				logrus.Debugf("pongC通道关闭或者收到fin,ok=%v, action=%s", ok, pong.Action)
				needFin = true
				return nil
			}

			logrus.Infof("Report: id:%s收到pong, action:%s", ping.Mid, pong.Action)
			err = stream.Send(&pong)
			if err != nil {
				logrus.Warnf("Report:stream.Send:%v", err)
				needFin = false
				return nil
			}

			//cmd的pong特殊处理
			if pong.Action == "cmd" {
				logrus.Debugf("Report:创建cmd响应chan...")
				if _, ok := model.CmdOutM[ping.Mid]; !ok {
					model.CmdOutM[ping.Mid] = make(chan grpcchannel.CmdOutput)
					logrus.Debugf("Report:创建cmd响应chan...done")
				}
			}
		}
	}
}

func ChangePickup(mid string, pickup int) error {
	err := db.DB.Model(&model.ClientInfo{ID: mid}).Update("pickup", pickup).Error
	if err != nil {
		logrus.Errorf("Report:更新Pickup值:%v,mid=%s", err, mid)
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
