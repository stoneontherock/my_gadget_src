package core

//import (
//	"connekts/client/log"
//	gc "connekts/grpcchannel"
//	"context"
//	"io"
//	"os"
//	"path/filepath"
//)
//
//func handleFileUp(pong *gc.Pong, cc gc.ChannelClient) {
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
//		log.Errorf("cc.FileUp:%v\n", err)
//		return
//	}
//
//	fp,err := os.Open(pth)
//	if err != nil {
//		stream.Send(&gc.FileDataUp{Mid:staticInfo.MachineID,Err:err.Error()})
//		return
//	}
//
//	buf := make([]byte,1<<20)
//	for {
//		n,err := fp.Read(buf)
//		if err == io.EOF {
//			_,err := stream.CloseAndRecv()
//			if err != nil {
//				log.Errorf("handleFileUp:stream.CloseAndRecv:%v\n",err)
//			}
//			break
//		}
//
//		if err != nil {
//			log.Errorf("handleFileUp:read:%v\n",err)
//			break
//		}
//
//		err = stream.Send(&gc.FileDataUp{Mid:staticInfo.MachineID,Data:buf[:n]})
//		if err != nil {
//			log.Errorf("handleFileUp:stream.Send:%v",err)
//		}
//	}
//}
