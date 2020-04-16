package grpcserver

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"line/common/connection/pb"
	"line/server/db"
	"line/server/model"
	"strconv"
	"time"
)

func (s *grpcServer) Report(ping *pb.Ping, stream pb.Channel_ReportServer) error {
	defer func() {
		model.CloseAllConnections(ping.Mid)
		sendFin(stream) //关闭当前stream
	}()

	ci := model.ClientInfo{ID: ping.Mid}
	err := db.DB.First(&ci).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		logrus.Errorf("Report:First:%v", err)
		return err
	}

	logrus.Debugf("Report: ci=%+v", ci)

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

	//不存在chan, 就初始化 pong chan
	pongC, ok := model.PongM[ping.Mid]
	if !ok {
		pongC = make(chan pb.Pong)
		model.PongM[ping.Mid] = pongC
	}

	if ci.StartAt != ping.StartAt {
		err := db.DB.Delete(&model.ClientInfo{ID: ping.Mid}).Error
		if err != nil {
			logrus.Errorf("Report:删除ClientInfo id=%s, err=%v", ping.Mid, err)
		}
		return errors.New("Report:服务端检测到startAt标记已经变更")
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
		logrus.Debugf("id=%s tmout=%d clientLiftime=%d", ping.Mid, tmout, clientLifetime)
		stream.Send(&pb.Pong{Action: "lifetime", Data: []byte(strconv.Itoa(int(clientLifetime)))})
	}

	tk := time.NewTicker(tmout)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			logrus.Infof("Report: %s超时,发fin", ping.Mid)
			ChangePickup(ping.Mid, -1)
			return nil
		case pong, ok := <-pongC:
			if !ok || pong.Action == "fin" {
				logrus.Debugf("Report: pongC通道关闭或者收到fin, id=%s, ok=%v, action=%s", ping.Mid, ok, pong.Action)
				return nil
			}

			logrus.Infof("Report: id:%s收到pong, action:%s", ping.Mid, pong.Action)
			err = stream.Send(&pong)
			if err != nil {
				logrus.Warnf("Report:stream.Send: id:%s err=%v", ping.Mid, err)
				pongC <- pb.Pong{Action: "fin"} //使用sendFin是不行了，因为stream已经断开，所以需要发fin的新的goroutine
				return fmt.Errorf("服务端Report:stream.Send(&pong) %v", err)
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
