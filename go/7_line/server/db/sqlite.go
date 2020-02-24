package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"line/server"
	"line/server/model"
	"line/server/panicerr"
	oslog "log"
	"os"
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
}
