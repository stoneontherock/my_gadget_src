package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"line/server/model"
)

type delRpxyIn struct {
	MID   string `form:"mid"`
	Label string `form:"label"` //label:port
}

func del_rproxied(c *gin.Context) {
	var di delRpxyIn
	err := c.ShouldBindWith(&di, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	model.CloseConnections(di.Label, di.MID)

	c.Redirect(303, "./rpxy?mid="+di.MID)
}
