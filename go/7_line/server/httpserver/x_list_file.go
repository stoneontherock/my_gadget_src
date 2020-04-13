package httpserver

//
//import (
//	"line/grpcchannel"
//	"line/server/model"
//	"github.com/gin-gonic/gin"
//	"github.com/gin-gonic/gin/binding"
//	"time"
//)
//
//type lfIn struct {
//	Mid     string    `form:"mid"`
//	Path    string  `form:"path"`
//}
//
//func listFile(c *gin.Context){
//	var li lfIn
//	c.ShouldBindWith(&li,binding.Query)
//
//	pongC, ok := model.PongM[li.Mid]
//	if !ok {
//		respJSAlert(c,400,"主机不在活动状态,id:" +li.Mid)
//		return
//	}
//
//	pongC <- grpcchannel.Pong{Action:"list_file",Data:[]byte(li.Path)}
//
//	var flC chan *grpcchannel.FileList
//	for i:=0; i<=20;i++ {
//		flC,ok = model.ListFileM[li.Mid]
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
