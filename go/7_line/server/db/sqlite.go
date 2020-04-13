package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/sirupsen/logrus"
	"line/common/connection/pb"
	"line/common/log"
	"line/common/panicerr"
	"line/server/model"
	oslog "log"
	"os"
	"time"
)

var DB *gorm.DB

func InitSQLite() {
	var err error
	DB, err = gorm.Open("sqlite3", log.BinDir+"/data.sqlite3")
	//defer DB.Close()

	DB.AutoMigrate(model.ClientInfo{}, model.CmdHistory{})

	if model.LogLevel == "debug" {
		DB.SetLogger(oslog.New(os.Stdout, "", oslog.LstdFlags))
		DB.LogMode(true)
	}

	err = DB.DB().Ping()
	panicerr.Handle(err, "InitSQLite:Ping()")

	DB.Model(&model.ClientInfo{}).Update("pickup", -1)
	go checkAlive()
}

func checkAlive() {
	for {
		time.Sleep(time.Second * time.Duration(model.CheckAliveInterval))

		var cis []model.ClientInfo
		err := DB.Model(&model.ClientInfo{}).Find(&cis).Error
		if err != nil {
			continue
		}

		now := time.Now().Unix()
		for _, ci := range cis {
			if now-int64(ci.LastReport) < model.CheckAliveInterval || ci.Pickup > 0 {
				continue
			}

			DB.Delete(&model.ClientInfo{}, `id = ?`, ci.ID)
			model.CloseAllConnections(ci.ID)
			if ci.Pickup >= 1 {
				pongC, ok := model.PongM[ci.ID]
				if ok {
					go func() {
						time.Sleep(time.Second * 5)
						pongC <- pb.Pong{Action: "fin"}
						time.Sleep(time.Millisecond * 100) //休息多久？
						delete(model.PongM, ci.ID)
					}()
				}
			}

			logrus.Debugf("%s寿终", ci.ID)
		}
	}
}
