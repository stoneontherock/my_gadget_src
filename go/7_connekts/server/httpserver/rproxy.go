package httpserver

import (
	"connekts/common"
	gc "connekts/grpcchannel"
	"connekts/server/grpcserver"
	"connekts/server/model"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/square/go-jose.v2/json"
	"strconv"
)

type rproxyIn struct {
	MID        string `form:"mid"` // binding:"hexadecimal"`
	NumOfConn2 int32  `form:"num_of_conn2"`
	Port1      string `form:"port1"` // binding:"numeric"`
	Addr3      string `form:"addr3"` //` binding:"numeric"`
}

func rProxy(c *gin.Context) {
	var ri rproxyIn
	err := c.ShouldBindWith(&ri, binding.FormPost)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	if ri.Port1 == "" {
		rPxyTmpl.Execute(c.Writer, ri.MID)
		return
	}

	ri.Port1 = ":" + ri.Port1
	port2 := ":" + strconv.Itoa(int(common.RandomAvaliblePort()))

	pongC, ok := model.PongM[ri.MID]
	if !ok {
		respJSAlert(c, 400, "主机不在活动状态,id:"+ri.MID)
		return
	}

	err = grpcserver.RProxyListen(ri.MID, ri.Port1, port2, int(ri.NumOfConn2))
	if err != nil {
		respJSAlert(c, 500, "创建bridge listener 失败:"+err.Error())
		return
	}

	rpr := gc.RPxyResp{Port2: port2, Addr3: ri.Addr3, NumOfConn2: ri.NumOfConn2}
	data, err := json.Marshal(&rpr)
	if err != nil {
		respJSAlert(c, 500, "序列化到pong data失败:"+err.Error())
		return
	}

	pongC <- gc.Pong{Action: "rpxy", Data: data}

	c.JSON(200, "done")
}
