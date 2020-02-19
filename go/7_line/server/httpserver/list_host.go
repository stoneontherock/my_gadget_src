package httpserver

import (
	"github.com/gin-gonic/gin"
	"line/server/db"
	"line/server/model"
)

type lahIn struct {
	ID     int    `form:"id"`
	Sort   string `form:"sort"`
	Order  string `form:"order"`
	Offset int    `form:"offset"`
	Limit  int    `form:"limit"`
}

func listHosts(c *gin.Context) {
	li := lahIn{Order: "asc"}
	err := c.BindQuery(&li)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	q := db.DB //.Where("update_at > ? OR pickup > 0", time.Now().Unix()-300) //5分钟内更新过的

	if li.ID > 0 {
		q = q.Where("id = ?", li.ID)
	}

	if li.Sort != "" {
		q = q.Order(li.Sort + " " + li.Order)
	}

	if li.Offset > 0 {
		q = q.Offset(li.Offset)
	}

	if li.Limit > 0 {
		q = q.Limit(li.Limit)
	}

	var cis []model.ClientInfo
	var total int
	err = q.Find(&cis).Offset(-1).Limit(-1).Count(&total).Error
	if err != nil {
		respJSAlert(c, 400, "db.Find.Count:"+err.Error())
		return
	}

	err = listHostsTmpl.Execute(c.Writer, &cis)
	if err != nil {
		respJSAlert(c, 400, "模板渲染出错"+err.Error())
		return
	}
}
