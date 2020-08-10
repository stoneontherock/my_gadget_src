package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"line/common/connection/pb"
	"line/server/db"
	"line/server/grpcserver"
	"line/server/model"
	"time"
)

type delHostIn struct {
	Mid string `form:"mid" binding:"required"`
}

//todo 还有很多需要删，待写, 已经勾住的主机删除后，没有收到fin
func releaseHost(c *gin.Context) {
	var dhi delHostIn
	err := c.ShouldBindWith(&dhi, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	ci := model.ClientInfo{ID: dhi.Mid}
	err = db.DB.First(&ci).Error
	if err != nil {
		respJSAlert(c, 400, "db.First:"+err.Error())
		return
	}

	err = grpcserver.ChangePickup(dhi.Mid,-1)
	if err != nil {
		respJSAlert(c, 500, "ChangePickup:"+err.Error())
		return
	}

	model.CloseAllConnections(dhi.Mid)
	if ci.Pickup >= 1 {
		pongC, ok := model.PongM[dhi.Mid]
		if ok {
			go func() {
				time.Sleep(time.Second * 5)
				pongC <- pb.Pong{Action: "fin"}
				time.Sleep(time.Millisecond * 100) //休息多久？
				delete(model.PongM, dhi.Mid)
			}()
		}
	}

	logrus.Debugf("delHost:删除host:%s成功", dhi.Mid)
	c.Redirect(303, "./list_hosts")
}
