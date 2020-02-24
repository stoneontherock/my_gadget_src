package log

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"line/server"
	"line/server/panicerr"
	"log"
	"os"
	"path/filepath"
)

func InitLog() {
	fmtr := new(logrus.TextFormatter)
	fmtr.FullTimestamp = true                   // 显示完整时间
	fmtr.TimestampFormat = "01-02 15:04:05.000" // 时间格式
	fmtr.DisableTimestamp = false               // 禁止显示时间
	fmtr.DisableColors = true                   // 禁止颜色显示

	var err error

	dir := server.BinDir + "/log"
	err = os.MkdirAll(dir, 0700)
	panicerr.Handle(err, "创建日志目录失败")

	f := filepath.Join(dir, filepath.Base(os.Args[0])+".log")

	log.Printf("log file: %s", f)

	jack := &lumberjack.Logger{
		Filename: f, //如果没目录，它会自己建立
		MaxSize:  5, //MBytes
		//MaxAge: 1, //day
		MaxBackups: 50,
		LocalTime:  true,
		Compress:   true,
	}

	logrus.SetOutput(jack)
	if server.Debug == "on" {
		logrus.SetLevel(logrus.Level(5 - 0)) //debug
	} else {
		logrus.SetLevel(logrus.Level(5 - 1)) //info
	}

	logrus.SetFormatter(fmtr)

	return
}
