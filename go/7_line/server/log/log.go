package log

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"line/server/panicerr"
	"log"
	"os"
	"path/filepath"
)

var Debug string
var BinDir string

func InitLog() {
	fmtr := new(logrus.TextFormatter)
	fmtr.FullTimestamp = true                   // 显示完整时间
	fmtr.TimestampFormat = "01-02 15:04:05.000" // 时间格式
	fmtr.DisableTimestamp = false               // 禁止显示时间
	fmtr.DisableColors = true                   // 禁止颜色显示

	var err error
	BinDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	panicerr.Handle(err, "获取可执行文件所在路径的绝对路径失败")

	dir := BinDir + "/log"
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

	if Debug == "on" {
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.Level(5 - 0)) // debug
	} else {
		logrus.SetOutput(jack)
		logrus.SetLevel(logrus.Level(5 - 1)) //info
	}

	logrus.SetFormatter(fmtr)

	return
}
