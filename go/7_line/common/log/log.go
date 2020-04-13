package log

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"line/common/panicerr"
	"os"
	"path/filepath"
)

var BinDir string

func init() {
	var err error
	BinDir, err = filepath.Abs(filepath.Dir(os.Args[0]))
	panicerr.Handle(err, "获取可执行文件所在路径的绝对路径失败")
}

func InitLog(level string) {
	//级别
	lvl, err := logrus.ParseLevel(level)
	panicerr.Handle(err)
	logrus.SetLevel(lvl)

	//格式
	fmtr := new(logrus.TextFormatter)
	fmtr.FullTimestamp = true                   // 显示完整时间
	fmtr.TimestampFormat = "01-02 15:04:05.000" // 时间格式
	fmtr.DisableTimestamp = false               // 禁止显示时间
	fmtr.DisableColors = true                   // 禁止颜色显示
	logrus.SetFormatter(fmtr)

	jack := &lumberjack.Logger{
		Filename: filepath.Join(BinDir+"/log", filepath.Base(os.Args[0])+".log"),
		MaxSize:  5, //MBytes
		//MaxAge: 1, //day
		MaxBackups: 50,
		LocalTime:  true,
		Compress:   true,
	}
	logrus.SetOutput(jack)
}
