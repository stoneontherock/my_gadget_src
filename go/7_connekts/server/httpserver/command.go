package httpserver

import (
	"connekts/common"
	gc "connekts/grpcchannel"
	"connekts/server/model"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"time"
)

type cmdFormIn struct {
	MID     string `form:"mid" binding:"required"`
	Cmd     string `form:"cmd"`
	Timeout int    `form:"timeout"`
}

type cmdOutHTTPResp struct {
	MID string
	Stdout string
	Stderr string
}

func command(c *gin.Context) {
	var ci cmdFormIn
	err := c.ShouldBindWith(&ci,binding.FormPost)
	if err != nil {
		respJSAlert(c,400,"参数错误:" + err.Error())
		return
	}

	if ci.Cmd == "" {
		cmdOutTmpl.Execute(c.Writer,&cmdOutHTTPResp{MID:ci.MID})
		return
	}


	data, err := json.Marshal(&common.CmdPong{Cmd:ci.Cmd,Timeout:ci.Timeout})
	if err != nil {
		respJSAlert(c,400,"json.Marshal:" + err.Error())
		return
	}

	pongC, ok := model.PongM[ci.MID]
	if !ok {
		respJSAlert(c,400,"主机不在活动状态,id:" + ci.MID)
		return
	}

	ch,ok := model.CmdOutM[ci.MID]

	logrus.Debugf("command:发送pongC...")
	pongC <- gc.Pong{Action: "cmd", Data: data}
	logrus.Debugf("command:发送pongC done, cmdout ch addr:%p, ok:%t",ch,ok)
	//time.Sleep(time.Millisecond)

	var cmdOutC chan gc.CmdOutput
	for i:=0; i<ci.Timeout*1000;i++ {
		time.Sleep(time.Millisecond)
		cmdOutC,ok = model.CmdOutM[ci.MID]
     	if ok {
     		out:= <-cmdOutC
			cmdOutTmpl.Execute(c.Writer,&cmdOutHTTPResp{MID:ci.MID,Stdout:out.Stdout,Stderr:out.Stderr})
			return
		}
	 }

	respJSAlert(c,400,"等待cmdout超时")
}

