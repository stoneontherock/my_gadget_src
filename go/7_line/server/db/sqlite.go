package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"line/server"
	"line/server/model"
	"line/server/panicerr"
	oslog "log"
	"os"
	"time"
)

var DB *gorm.DB

func InitSQLite() {
	var err error
	DB, err = gorm.Open("sqlite3", server.BinDir+"/data.sqlite3")
	//defer DB.Close()

	DB.AutoMigrate(&model.ClientInfo{})

	DB.SetLogger(oslog.New(os.Stdout, "", oslog.LstdFlags))
	//DB.LogMode(true) //todo :Debug, 发布后改为false

	err = DB.DB().Ping()
	panicerr.Handle(err, "InitSQLite:Ping()")

	go checkAlive()
}

func checkAlive() {
	for {
		time.Sleep(time.Second * time.Duration(server.CheckAliveInterval))

		var cis []model.ClientInfo
		err := DB.Model(&model.ClientInfo{}).Find(&cis).Error
		if err != nil {
			continue
		}

		now := int32(time.Now().Unix())
		for _, ci := range cis {
			if now-ci.LastReport < int32(server.CheckAliveInterval) {
				continue
			}
			DB.Delete(&model.ClientInfo{}, `id = ?`, ci.ID)
		}
	}
}
