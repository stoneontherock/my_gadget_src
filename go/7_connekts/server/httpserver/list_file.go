package httpserver
//
//import (
//	gc "connekts/grpcchannel"
//	"connekts/server/model"
//	"github.com/gin-gonic/gin"
//	"github.com/gin-gonic/gin/binding"
//	"time"
//)
//
//type lfIn struct {
//	MID     string    `form:"mid"`
//	Path    string  `form:"path"`
//}
//
//func listFile(c *gin.Context){
//	var li lfIn
//	c.ShouldBindWith(&li,binding.Query)
//
//	pongC, ok := model.PongM[li.MID]
//	if !ok {
//		respJSAlert(c,400,"主机不在活动状态,id:" +li.MID)
//		return
//	}
//
//	pongC <- gc.Pong{Action:"list_file",Data:[]byte(li.Path)}
//
//	var flC chan *gc.FileList
//	for i:=0; i<=20;i++ {
//		flC,ok = model.ListFileM[li.MID]
//		if ok {
//			flist := <-flC
//			listFileTmpl.Execute(c.Writer,*flist)
//			return
//		}
//		time.Sleep(time.Millisecond*100)
//	}
//
//	respJSAlert(c,400,"等待listFile超时")
//}
