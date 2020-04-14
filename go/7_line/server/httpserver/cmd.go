package httpserver

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
	"html/template"
	"line/common/connection/pb"
	"line/common/sharedmodel"
	"line/server/db"
	"line/server/model"
	"net/url"
	"strconv"
	"time"
)

type cmdFormIn struct {
	Mid     string `form:"mid" binding:"required"`
	Cmd     string `form:"cmd"`
	InShell bool   `form:"inShell"`
	Timeout int    `form:"timeout"`
}

type cmdOutHTTPResp struct {
	Mid        string
	Stdout     string
	Stderr     string
	CmdHistory []model.CmdHistory
}

const MAXCMDHISTORY = 20

func command(c *gin.Context) {
	var ci cmdFormIn
	err := c.ShouldBindWith(&ci, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	if !isHostPickedUp(ci.Mid) {
		respJSAlert(c, 500, "主机未勾住")
		return
	}

	if ci.Cmd == "" {
		cmdOutTmpl.Execute(c.Writer, &cmdOutHTTPResp{Mid: ci.Mid, CmdHistory: getCmdHistory(ci.Mid)})
		return
	}

	data, err := json.Marshal(&sharedmodel.CmdPong{Cmd: ci.Cmd, InShell: ci.InShell, Timeout: ci.Timeout})
	if err != nil {
		respJSAlert(c, 400, "json.Marshal:"+err.Error())
		return
	}

	pongC, ok := model.PongM[ci.Mid]
	if !ok {
		respJSAlert(c, 400, "主机不在活动状态,id:"+ci.Mid)
		return
	}

	ch, ok := model.CmdOutM[ci.Mid]

	logrus.Debugf("command:发送pongC...")
	pongC <- pb.Pong{Action: "cmd", Data: data}
	logrus.Debugf("command:发送pongC done, cmdout ch addr:%p, ok:%t", ch, ok)
	//time.Sleep(time.Millisecond)

	var cmdOutC chan pb.CmdOutput
	for i := 0; i < ci.Timeout*100; i++ { //这里的100和下面的毫秒数相关
		time.Sleep(time.Millisecond * 10)
		cmdOutC, ok = model.CmdOutM[ci.Mid]
		if ok {
			tk := time.NewTicker(time.Second * time.Duration(ci.Timeout+5))
			select {
			case <-tk.C:
				respJSAlert(c, 400, "等待执行结果超时")
				tk.Stop()
			case out := <-cmdOutC:
				storeToDB(&ci)
				cmdOutTmpl.Execute(c.Writer, &cmdOutHTTPResp{Mid: ci.Mid, Stdout: out.Stdout, Stderr: out.Stderr, CmdHistory: getCmdHistory(ci.Mid)})
			}
			return
		}
	}

	respJSAlert(c, 400, "等待cmdOutC超时")
}

func storeToDB(ci *cmdFormIn) {
	var cmdHis model.CmdHistory
	err := db.DB.First(&cmdHis, "cmd = ?", template.HTML(ci.Cmd)).Error
	if cmdHis.ID > 0 {
		db.DB.Model(&cmdHis).Update("update_at", int32(time.Now().Unix()))
		return
	}

	u := url.Values{}
	u.Set("cmd", ci.Cmd)
	u.Set("inShell", strconv.FormatBool(ci.InShell))
	u.Set("timeout", strconv.Itoa(ci.Timeout))
	err = db.DB.Create(&model.CmdHistory{Mid: ci.Mid, Cmd: template.HTML(ci.Cmd), QueryString: u.Encode(), UpdateAt: int32(time.Now().Unix())}).Error
	if err != nil {
		logrus.Errorf("创建cmd历史记录失败")
	}
	//logrus.Debugf("cmd=%s已经存在",cmdHis.Cmd)
}

func getCmdHistory(mid string) []model.CmdHistory {
	var chl []model.CmdHistory
	err := db.DB.Model(&model.CmdHistory{}).Order("update_at desc").Find(&chl, "mid = ?", mid).Error
	if err != nil {
		logrus.Errorf("查询cmd历史记录失败, %v", err)
	}

	if len(chl) > MAXCMDHISTORY {
		for i := len(chl); i >= MAXCMDHISTORY; i-- {
			db.DB.Delete(&chl[i])
		}
		chl = chl[:MAXCMDHISTORY]
	}

	logrus.Debugf("getCmdHistory:%v", chl)
	return chl
}


func delCmdHistory(c *gin.Context) {
	id := struct { Id int32 `form:id binding:"required"` }
	err := c.ShouldBindQuery(&id, binding.Form)
	if err != nil {
		respJSAlert(c, 400, "参数错误:"+err.Error())
		return
	}

	err = db.DB.Delete(&model.CmdHistory{ID:id.Id}).Error
	if err != nil {
		respJSAlert(c, 400, "删除命令记录失败:"+err.Error())
		return
	}

	c.Status(200)
}
