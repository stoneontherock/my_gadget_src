package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type delRpxyIn struct {
	MID   string `form:"mid"`
	Label string `form:"label"`  //label:port
}

func del_rproxied(c *gin.Context) {
	var di delRpxyIn
	err := c.ShouldBindWith(&di, binding.FormPost)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	closeConnection(di.Label,di.MID)

	c.Redirect(303, "./list_rproxied?mid="+di.MID)
}
