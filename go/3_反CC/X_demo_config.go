package cc

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
)

func InitLog() {
	log.Printf("init logrus...\n")
	fmtr := new(logrus.TextFormatter)
	fmtr.FullTimestamp = true         // 显示完整时间
	fmtr.TimestampFormat = "15:04:05" // 时间格式
	fmtr.DisableTimestamp = false     // 禁止显示时间
	fmtr.DisableColors = false        // 禁止颜色显示

	jack := &lumberjack.Logger{
		Filename:   "./acc.log",
		MaxSize:    1, // unit: MBytes
		MaxAge:     1, // unit: day
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true,
	}

	logrus.SetFormatter(fmtr)
	logrus.SetOutput(jack)
	logrus.SetLevel(logrus.DebugLevel)
}
