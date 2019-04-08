package tcpexpose

import (
	"log"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func InitLog(stdout bool, maxsize, maxbackups int) error {
	fmtr := new(logrus.TextFormatter)
	fmtr.FullTimestamp = true                   // 显示完整时间
	fmtr.TimestampFormat = "01-02 15:04:05.000" // 时间格式
	fmtr.DisableTimestamp = false               // 禁止显示时间
	fmtr.DisableColors = false                  // 禁止颜色显示

	if stdout {
		logrus.SetOutput(os.Stdout)
	} else {
		d, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return err
		}

		f := filepath.Join(d, filepath.Base(os.Args[0])+".log")
		log.Printf("log file: %s", f)

		jack := &lumberjack.Logger{
			Filename: f,       //如果没目录，它会自己建立
			MaxSize:  maxsize, //MBytes
			//MaxAge: 1, //day
			MaxBackups: maxbackups,
			LocalTime:  true,
			Compress:   true,
		}
		logrus.SetOutput(jack)
	}

	logrus.SetFormatter(fmtr)
	logrus.SetLevel(logrus.DebugLevel)

	return nil
}
