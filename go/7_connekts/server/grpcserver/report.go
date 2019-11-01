package grpcserver

import (
	gc "connekts/grpcchannel"
	"connekts/server/db"
	"connekts/server/model"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"time"
)

func (s *server) Report(ping *gc.Ping, stream gc.Channel_ReportServer) error {
	wanIP := getClientIPAddr(stream.Context())

	ci := model.ClientInfo{ID: ping.Mid}
	err := db.DB.First(&ci).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			logrus.Errorf("Report:First:%v\n", err)
			return err
		}
		err := db.DB.Create(&model.ClientInfo{ID: ping.Mid, WanIP: wanIP, Hostname: ping.Hostname, OS: ping.Os}).Error
		if err != nil {
			logrus.Errorf("Report:Create:%v\n", err)
			return err
		}
	}

	//不存在chan, 就初始化 pong chan
	pongC, ok := model.PongM[ping.Mid]
	if !ok {
		pongC = make(chan gc.Pong)
		model.PongM[ping.Mid] = pongC
	}

	if ci.Pickup <= 0 {
		logrus.Debugf("Report:丢弃")
		sendFin(stream)
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

	tk := time.NewTicker(time.Second * 3600)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			logrus.Infof("Report:超时,pickup->-1")
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
			logrus.Debugf("Report:创建响应chan...")
			createRespChan(pong.Action, ping.Mid)
		}
	}
}

func createRespChan(typ, mid string) {
	switch typ {
	case "cmd": //cmd需要反馈到前端，所以需要创建map
		if _, ok := model.CmdOutM[mid]; !ok {
			model.CmdOutM[mid] = make(chan gc.CmdOutput)
		}
		//case "list_file":
		//	if _,ok := model.ListFileM[mid]; !ok {
		//		model.ListFileM[mid] = make(chan *gc.FileList)
		//	}
		//case "file_up":
		//	if _,ok := model.FileUpDataM[mid]; !ok {
		//		model.FileUpDataM[mid] = make(chan []byte)
		//	}
	}

	logrus.Debugf("Report:创建响应chan done")
}

func ChangePickup(mid string, pickup int) error {
	err := db.DB.Model(&model.ClientInfo{ID: mid}).Update("pickup", pickup).Error
	if err != nil {
		logrus.Errorf("Report:Update:%v", err)
		return err
	}

	return nil
}

func sendFin(stream gc.Channel_ReportServer) {
	err := stream.Send(&gc.Pong{Action: "fin"})
	if err != nil {
		logrus.Warnf("Report:send fin:%v", err)
	}
}
