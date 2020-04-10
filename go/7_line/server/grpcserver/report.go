package grpcserver

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"line/common/connection/pb"
	"line/server/db"
	"line/server/model"
	"strconv"
	"time"
)

func (s *grpcServer) Report(ping *pb.Ping, stream pb.Channel_ReportServer) error {
	needFin := true
	defer func() {
		if needFin {
			sendFin(stream) //关闭当前stream
		}
	}()

	//不存在chan, 就初始化 pong chan
	pongC, ok := model.PongM[ping.Mid]
	if !ok {
		pongC = make(chan pb.Pong)
		model.PongM[ping.Mid] = pongC
	}

	ci := model.ClientInfo{ID: ping.Mid}
	err := db.DB.First(&ci).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("Report:First:%v", err)
		return err
	}

	logrus.Debugf("ci:%+v", ci)

	if ci.LastReport == 0 {
		//无则创建
		ci = model.ClientInfo{
			ID:         ping.Mid,
			WanIP:      getClientIPAddr(stream.Context()),
			Kernel:     ping.Kernel,
			OsInfo:     ping.OsInfo,
			Interval:   ping.Interval,
			StartAt:    ping.StartAt,
			LastReport: int32(time.Now().Unix()),
		}
		err := db.DB.Create(&ci).Error
		if err != nil {
			logrus.Errorf("Report:Create:%v", err)
			return err
		}
	} else {
		err = db.DB.Model(&model.ClientInfo{ID: ping.Mid}).Update("last_report", int32(time.Now().Unix())).Error
		if err != nil {
			logrus.Errorf("Report:更新LastReport值失败:%v", err)
			return err
		}
	}

	if ci.StartAt != ping.StartAt {
		err := db.DB.Delete(&model.ClientInfo{ID: ping.Mid}).Error
		if err != nil {
			logrus.Errorf("Report:删除:%v", err)
		}

		if ci.Pickup == 2 {
			go func() {
				model.CloseAllConnections(ping.Mid)
				pongC <- pb.Pong{Action: "fin"} //这里的sendFin是为了关闭已经失效的stream
			}()
		}

		return errors.New("startAt标记已经变更")
	}

	if ci.Pickup <= 0 {
		logrus.Debugf("Report:丢弃%s", ping.Mid)
		return nil
	}

	now := time.Now()
	lifetime := time.Unix(int64(ci.Lifetime), 0)
	tmout := lifetime.Sub(now)
	if tmout <= time.Second*60 {
		tmout = time.Second * 60
	}

	if ci.Pickup == 1 {
		ChangePickup(ping.Mid, 2)
		clientLifetime := int32(tmout/1e9) + ci.Interval
		logrus.Debugf("tmout=%d clientLiftime=%d", tmout, clientLifetime)
		stream.Send(&pb.Pong{Action: "lifetime", Data: []byte(strconv.Itoa(int(clientLifetime)))})
		needFin = false
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
					model.CmdOutM[ping.Mid] = make(chan pb.CmdOutput)
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

func sendFin(stream pb.Channel_ReportServer) {
	err := stream.Send(&pb.Pong{Action: "fin"})
	if err != nil {
		logrus.Warnf("Report:send fin:%v", err)
	}
}
