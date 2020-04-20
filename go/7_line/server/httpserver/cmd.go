package httpserver

import (
	"encoding/json"
	"fmt"
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

	pongC, ok := model.PongM[ci.Mid]
	if !ok {
		respJSAlert(c, 400, fmt.Sprintf("key=%s对应PongM映射没有值,主机不在活动状态"+ci.Mid))
		return
	}
	ch, ok := model.CmdOutM[ci.Mid]

	logrus.Debugf("command:发送pongC, cmd=%s ...", ci.Cmd)
	data, err := json.Marshal(&sharedmodel.CmdPong{Cmd: ci.Cmd, InShell: ci.InShell, Timeout: ci.Timeout})
	if err != nil {
		respJSAlert(c, 400, "json.Marshal:"+err.Error())
		return
	}
	pongC <- pb.Pong{Action: "cmd", Data: data}
	logrus.Debugf("command:发送pongC done, cmdout ch addr:%p, ok:%t", ch, ok)
	if _, ok := model.CmdOutM[ci.Mid]; !ok {
		model.CmdOutM[ci.Mid] = make(chan pb.CmdOutput)
	}

	tk := time.NewTicker(time.Second * time.Duration(ci.Timeout+5))
	select {
	case <-tk.C:
		respJSAlert(c, 400, "等待执行结果超时")
		tk.Stop()
	case out := <-model.CmdOutM[ci.Mid]:
		storeToDB(&ci)
		cmdOutTmpl.Execute(c.Writer, &cmdOutHTTPResp{Mid: ci.Mid, Stdout: out.Stdout, Stderr: out.Stderr, CmdHistory: getCmdHistory(ci.Mid)})
	}
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
		for i := len(chl) - 1; i >= MAXCMDHISTORY; i-- {
			db.DB.Delete(&chl[i])
		}
		chl = chl[:MAXCMDHISTORY]
	}

	logrus.Debugf("getCmdHistory:%v", chl)
	return chl
}
