package db

import (
	"line/server/model"
	"line/server/panicerr"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"os"
)

var DB *gorm.DB

func InitSQLite() {
	var err error
	DB, err = gorm.Open("sqlite3", "/tmp/gorm.db")
	//defer DB.Close()

	DB.AutoMigrate(&model.ClientInfo{})

	DB.SetLogger(log.New(os.Stdout, "", log.LstdFlags))
	//DB.LogMode(true) //todo :Debug, 发布后改为false

	err = DB.DB().Ping()
	panicerr.Handle(err, "InitSQLite:Ping()")
}
