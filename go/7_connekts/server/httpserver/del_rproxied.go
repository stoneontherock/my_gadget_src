package httpserver

import (
	"connekts/server/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type delRpxyIn struct {
	MID   string `form:"mid"`
	Label string `form:"label"`
}

func del_rproxied(c *gin.Context) {
	var di delRpxyIn
	err := c.ShouldBindWith(&di, binding.FormPost)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	for lab, ls := range model.RPxyListenerM[di.MID] {
		if lab != di.Label {
			continue
		}

		for _, l := range ls {
			l.Close()
		}
		delete(model.RPxyListenerM[di.MID], lab)
		if len(model.RPxyListenerM[di.MID]) == 0 {
			delete(model.RPxyListenerM, di.MID)
		}
	}

	c.Redirect(303, "./list_rproxied")
}
