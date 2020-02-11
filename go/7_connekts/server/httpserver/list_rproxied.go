package httpserver

import (
	"connekts/server/model"
	"github.com/gin-gonic/gin"
)

type lrpIn struct {
	MID string `form:"mid"`
}

func list_rproxied(c *gin.Context) {
	var li lrpIn
	err := c.BindQuery(&li)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	data := make(map[string][]string)
	if li.MID != "" {
		for label, _ := range model.RPxyListenerM[li.MID] {
			data[li.MID] = append(data[li.MID], label)
		}
	} else {
		for mid, tmpMap := range model.RPxyListenerM {
			for label, _ := range tmpMap {
				data[mid] = append(data[mid], label)
			}
		}
	}

	err = listRProxiedTmpl.Execute(c.Writer, &data)
	if err != nil {
		respJSAlert(c, 500, "模板渲染出错"+err.Error())
		return
	}
}