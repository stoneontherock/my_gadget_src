package httpserver

//重复开发，新建反代的地方就已经有该功能

//import (
//	"line/server/model"
//	"github.com/gin-gonic/gin"
//)
//
//type lrpIn struct {
//	Mid string `form:"mid"`
//}
//
//func list_rproxied(c *gin.Context) {
//	var li lrpIn
//	err := c.BindQuery(&li)
//	if err != nil {
//		respJSAlert(c, 400, "参数错误:"+err.Error())
//		return
//	}
//
//	data := make(map[string][]string)
//	if li.Mid != "" {
//		for label, _ := range model.RPxyConnResM[li.Mid] {
//			data[li.Mid] = append(data[li.Mid], label)
//		}
//	} else {
//		for mid, tmpMap := range model.RPxyConnResM {
//			for label, _ := range tmpMap {
//				data[mid] = append(data[mid], label)
//			}
//		}
//	}
//
//	err = listRProxiedTmpl.Execute(c.Writer, &data)
//	if err != nil {
//		respJSAlert(c, 500, "模板渲染出错"+err.Error())
//		return
//	}
//}
