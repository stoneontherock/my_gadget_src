package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"line/server/db"
	"line/server/model"
)

func getHostState(c *gin.Context) {
	var in struct {
		Mid string `form:"mid" binding:"required"`
	}
	err := c.BindQuery(&in)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	var rec model.ClientInfo
	err = db.DB.First(&rec, "id = ?", in.Mid).Error
	if err != nil {
		c.String(500, err.Error())
		logrus.Errorf("getHostState:First:" + err.Error())
		return
	}

	c.String(200, "%t", rec.Pickup == 2)
}
