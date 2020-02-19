package httpserver

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"line/common"
	"line/grpcchannel"
	"line/server/model"
	"time"
)

type cmdFormIn struct {
	MID     string `form:"mid" binding:"required"`
	Cmd     string `form:"cmd"`
	InShell bool   `form:"inShell"`
	Timeout int    `form:"timeout"`
}

type cmdOutHTTPResp struct {
	MID    string
	Stdout string
	Stderr string
}

func command(c *gin.Context) {
	var ci cmdFormIn
	err := c.ShouldBindWith(&ci, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	if ci.Cmd == "" {
		cmdOutTmpl.Execute(c.Writer, &cmdOutHTTPResp{MID: ci.MID})
		return
	}

	data, err := json.Marshal(&common.CmdPong{Cmd: ci.Cmd, InShell: ci.InShell, Timeout: ci.Timeout})
	if err != nil {
		respJSAlert(c, 400, "json.Marshal:"+err.Error())
		return
	}

	pongC, ok := model.PongM[ci.MID]
	if !ok {
		respJSAlert(c, 400, "主机不在活动状态,id:"+ci.MID)
		return
	}

	ch, ok := model.CmdOutM[ci.MID]

	logrus.Debugf("command:发送pongC...")
	pongC <- grpcchannel.Pong{Action: "cmd", Data: data}
	logrus.Debugf("command:发送pongC done, cmdout ch addr:%p, ok:%t", ch, ok)
	//time.Sleep(time.Millisecond)

	var cmdOutC chan grpcchannel.CmdOutput
	for i := 0; i < ci.Timeout*100; i++ { //这里的100和下面的毫秒数相关
		time.Sleep(time.Millisecond * 10)
		cmdOutC, ok = model.CmdOutM[ci.MID]
		if ok {
			tk := time.NewTicker(time.Second * time.Duration(ci.Timeout))
			select {
			case <-tk.C:
				respJSAlert(c, 400, "等待执行结果超时")
				tk.Stop()
			case out := <-cmdOutC:
				cmdOutTmpl.Execute(c.Writer, &cmdOutHTTPResp{MID: ci.MID, Stdout: out.Stdout, Stderr: out.Stderr})
			}
			return
		}
	}

	respJSAlert(c, 400, "等待cmdOutC超时")
}
