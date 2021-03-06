package httpserver

//import (
//	"line/grpcchannel"
//	"line/server/model"
//	"github.com/gin-gonic/gin"
//	"github.com/sirupsen/logrus"
//	"path/filepath"
//	"strconv"
//	"time"
//)
//
//type fileUpIn struct {
//	Mid      string `form:"mid"`
//	FilePath string `form:"path"`
//	Size     int    `form:"size"`
//}
//
//func fileUpload(c *gin.Context) {
//	var fi fileUpIn
//	err := c.BindQuery(&fi)
//	if err != nil {
//		respJSAlert(c, 400, "参数错误:"+err.Error())
//		return
//	}
//
//	pongC, ok := model.PongM[fi.Mid]
//	if !ok {
//		respJSAlert(c, 400, "主机不在活动状态,id:"+fi.Mid)
//		return
//	}
//
//	c.Header("Content-Disposition", "attachment;filename="+filepath.Base(fi.FilePath))
//	c.Header("Content-Type","application/octet-stream")
//	if fi.Size > 0 {
//		c.Header("Content-Length",strconv.Itoa(fi.Size))
//	}
//
//	logrus.Debugf("content-length=%s, fi.Size=%d",c.GetHeader("Content-Length"),fi.Size)
//
//
//	model.FileUpDataM[fi.Mid] = make(chan *grpcchannel.FileDataUp)
//	pth := []byte(fi.FilePath)
//	pongC <- grpcchannel.Pong{Action: "file_up", Data: []byte(pth)}
//
//	for i := 0; i < 100; i++ {
//		time.Sleep(time.Millisecond * 10)
//		dataC, ok := model.FileUpDataM[fi.Mid]
//		if !ok {
//			continue
//		}
//
//		tkC := time.After(time.Second * 30)
//		for {
//			select {
//			case <-tkC:
//				respJSAlert(c, 500, "等待cmdout超时")
//				return
//			case data, ok := <-dataC:
//				if !ok {
//					return
//				}
//				if data.Err != "" {
//					logrus.Errorf("fileUpload: fileUpload失败,%v",err)
//					c.Header("Content-Type","text/html")
//					c.Header("Content-Disposition","")
//					respJSAlert(c, 500, data.Err)
//					return
//				}
//				c.Writer.Write(data.Data)
//			}
//		}
//	}
//
//	return
//}
