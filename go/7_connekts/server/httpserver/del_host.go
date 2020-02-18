package httpserver

import (
	"line/grpcchannel"
	"line/server/db"
	"line/server/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"time"
)

type delHostIn struct {
	MID string `form:"mid" binding:"required"`
}

//todo 还有很多需要删，待写
func delHost(c *gin.Context) {
	var dhi delHostIn
	err := c.ShouldBindWith(&dhi, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	err = db.DB.Delete(&model.ClientInfo{ID: dhi.MID}).Error
	if err != nil {
		respJSAlert(c, 400, "db.Find.Count:"+err.Error())
		return
	}

	closeConnection("", dhi.MID)

	pongC, ok := model.PongM[dhi.MID]
	if ok {
		pongC <- grpcchannel.Pong{Action: "fin"}
		time.Sleep(time.Millisecond * 10)
		delete(model.PongM, dhi.MID)
	}

	logrus.Debugf("delHost:删除host:%s成功", dhi.MID)
	c.Data(200, "text/html", []byte(LIST_HOSTS))
}

const LIST_HOSTS = `
<html>
  <script language='javascript' type='text/javascript'> 
     setTimeout("javascript:location.href='./list_hosts'", 1000); 
  </script>
</html>
`
