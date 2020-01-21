package httpserver

import (
	"connekts/common"
	gc "connekts/grpcchannel"
	"connekts/server/model"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"regexp"
	"strconv"
)

type fsIn struct {
	MID string `form:"mid"`
}

var regPatt = regexp.MustCompile(`^(.*)(:[0-9]+)$`)

func filesystem(c *gin.Context) {
	var fi fsIn
	err := c.ShouldBindQuery(&fi)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	port1 := ":" + strconv.Itoa(int(common.RandomAvaliblePort()))
	port2 := ":" + strconv.Itoa(int(common.RandomAvaliblePort()))

	pongC, ok := model.PongM[fi.MID]
	if !ok {
		respJSAlert(c, 400, "主机不在活动状态,id:"+fi.MID)
		return
	}

	err = listen2Side(fi.MID,"filesystem", port1, port2, 6)
	if err != nil {
		respJSAlert(c, 500, "创建bridge listener 失败:"+err.Error())
		return
	}

	rpr := gc.RPxyResp{Port2: port2, NumOfConn2: 6}
	data, err := json.Marshal(&rpr)
	if err != nil {
		respJSAlert(c, 500, "序列化到pong data失败:"+err.Error())
		return
	}

	pongC <- gc.Pong{Action: "filesystem", Data: data}

	host := regPatt.ReplaceAllString(c.Request.Host, "$1")
	//c.Writer.WriteString(host+port1)
	c.Redirect(307, "http://"+host+port1+"/")
}
