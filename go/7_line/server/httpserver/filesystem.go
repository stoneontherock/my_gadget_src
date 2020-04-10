package httpserver

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"line/common/connection"
	"line/common/connection/pb"
	"line/server/model"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type fsIn struct {
	MID string `form:"mid" binding:"required"`
}

var regPatt = regexp.MustCompile(`^(.*)(:[0-9]+)$`)

func filesystem(c *gin.Context) {
	var fi fsIn
	err := c.ShouldBindQuery(&fi)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	if !isHostPickedUp(fi.MID) {
		respJSAlert(c, 500, "主机未勾住")
		return
	}

	host := regPatt.ReplaceAllString(c.Request.Host, "$1")
	//如果已经存在文件系统反代，就重定向
	logrus.Debugf("RPxyConnResM[%s]=%+v", fi.MID, model.RPxyConnResM[fi.MID])
	for pLabel, _ := range model.RPxyConnResM[fi.MID] {
		if strings.HasPrefix(pLabel, "filesystem") {
			ss := strings.Split(pLabel, ":")
			if len(ss) != 2 {
				continue
			}
			c.Redirect(303, "http://"+host+":"+ss[1]+"/filesystem")
			return
		}
	}

	port1 := ":" + strconv.Itoa(int(connection.RandomAvaliblePort()))
	port2 := ":" + strconv.Itoa(int(connection.RandomAvaliblePort()))

	pongC, ok := model.PongM[fi.MID]
	if !ok {
		respJSAlert(c, 400, "主机不在活动状态,id:"+fi.MID)
		return
	}

	err = listen2Side(fi.MID, "filesystem", port1, port2, 3)
	if err != nil {
		respJSAlert(c, 500, "创建bridge listener 失败:"+err.Error())
		return
	}

	rpr := pb.RPxyResp{Port2: port2, NumOfConn2: 3}
	data, err := json.Marshal(&rpr)
	if err != nil {
		respJSAlert(c, 500, "序列化到pong data失败:"+err.Error())
		return
	}

	pongC <- pb.Pong{Action: "filesystem", Data: data}

	dm, port, _ := net.SplitHostPort(c.Request.Host)
	scheme := "http://"
	if c.Request.TLS != nil {
		scheme = "https://"
	}
	home := "/filesystem?home=" + url.QueryEscape(scheme+dm+":"+port+"/line/list_hosts")
	c.Redirect(303, "http://"+host+port1+home)
}
