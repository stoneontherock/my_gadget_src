package core

//import (
//	"line/client/log"
//	"line/grpcchannel"
//	"context"
//	"io"
//	"os"
//	"path/filepath"
//)
//
//func handleFileUp(pong *grpcchannel.Pong, cc grpcchannel.ChannelClient) {
//	pth:= string(pong.Data)
//	if pth == "" {
//		pth, _ = filepath.Abs(".")
//	}
//	println("fileUp:", pth)
//
//	ctx, cancel := context.WithCancel(context.TODO())
//	defer cancel()
//
//	stream, err := cc.FileUp(ctx)
//	if err != nil {
//		logrus.Errorf("cc.FileUp:%v", err)
//		return
//	}
//
//	fp,err := os.Open(pth)
//	if err != nil {
//		stream.Send(&grpcchannel.FileDataUp{Mid:staticInfo.MachineID,Err:err.Error()})
//		return
//	}
//
//	buf := make([]byte,1<<20)
//	for {
//		n,err := fp.Read(buf)
//		if err == io.EOF {
//			_,err := stream.CloseAndRecv()
//			if err != nil {
//				logrus.Errorf("handleFileUp:stream.CloseAndRecv:%v",err)
//			}
//			break
//		}
//
//		if err != nil {
//			logrus.Errorf("handleFileUp:read:%v",err)
//			break
//		}
//
//		err = stream.Send(&grpcchannel.FileDataUp{Mid:staticInfo.MachineID,Data:buf[:n]})
//		if err != nil {
//			logrus.Errorf("handleFileUp:stream.Send:%v",err)
//		}
//	}
//}
