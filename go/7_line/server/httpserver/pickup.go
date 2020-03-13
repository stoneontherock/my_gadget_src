package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"line/server/db"
	"line/server/grpcserver"
	"line/server/model"
	"time"
)

type pickupIn struct {
	MID     string `form:"mid" binding:"required"`
	Pickup  int    `form:"pickup" binding:"required"`
	Timeout int64  `form:"timeout"`
}

func pickup(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")             //允许访问所有域
	c.Header("Access-Control-Allow-Headers", "Content-Type") //header的类型
	c.Header("content-type", "text/html")                    //返回数据格式是json

	var pi pickupIn
	err := c.ShouldBindWith(&pi, binding.Form)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	err = grpcserver.ChangePickup(pi.MID, pi.Pickup)
	if err != nil {
		c.String(500, "修改pickup失败:"+err.Error())
		return
	}

	hms := time.Now().Add(time.Duration(pi.Timeout) * time.Minute).Format("2006-01-02 15:04:05")
	err = db.DB.Model(&model.ClientInfo{ID: pi.MID}).Update("timeout", hms).Error
	if err != nil {
		c.String(500, "修改timeout失败:"+err.Error())
		return
	}

	c.String(200, "ok")
}
