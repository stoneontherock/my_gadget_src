package httpserver

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"line/server/model"
)

type delRpxyIn struct {
	Mid               string `form:"mid"`
	Label             string `form:"label"` //label
	Port              string `form:"port"`
	RedirectListHosts bool   `form:"redirect_list_hosts"`
}

func del_rproxied(c *gin.Context) {
	var di delRpxyIn
	err := c.ShouldBindWith(&di, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	model.CloseConnections(di.Label+":"+di.Port, di.Mid)

	if di.RedirectListHosts {
		c.Redirect(303, "./list_hosts")
		return
	}
	c.Redirect(303, "./rpxy?mid="+di.Mid)
}
