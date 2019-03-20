package webmasq

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
)

const (
	newSiteHtml = `<html><form action="/site" method="get">
<select name="formS"><option value="https://">https</option><option value="http://">http</option></select>
<input type="text" name="formHP" /> 
<input type="submit" value="提交" />
</form></html>`

	loginHtml = `<html><form action="/login" method="post">
<p>用户名: <input type="text" name="uname" /></p>
<p>密码: <input type="text" name="upwd" /></p>
<input type="submit" value="提交" />
</form></html>`

	jumpToForm = `<html><script language='javascript' type='text/javascript'> setTimeout("javascript:location.href='/zzz'", 0); </script></html>`
)

var (
	userAuth []string
	salt     string
)

func AddUser(uname, pstr string) {
	userAuth = append(userAuth, Md5sum(uname+"/"+pstr))
}

func InitLog() {
	fmtr := new(logrus.TextFormatter)
	fmtr.FullTimestamp = true             // 显示完整时间
	fmtr.TimestampFormat = "15:04:05.000" // 时间格式
	fmtr.DisableTimestamp = false         // 禁止显示时间
	fmtr.DisableColors = false            // 禁止颜色显示

	jack := &lumberjack.Logger{
		Filename: "./sp.log",
		MaxSize:  1, //MBytes
		//MaxAge: 1, //day
		MaxBackups: 3,
		LocalTime:  true,
		Compress:   true,
	}
	_ = jack
	//logrus.SetOutput(jack)
	logrus.SetOutput(os.Stdout)

	logrus.SetFormatter(fmtr)
	logrus.SetLevel(logrus.DebugLevel)
}
