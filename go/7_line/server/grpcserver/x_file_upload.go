package grpcserver

//import (
//	"line/grpcchannel"
//	"line/server/model"
//	"github.com/sirupsen/logrus"
//	"io"
//)
//
//func (s *server) FileUp(stream grpcchannel.Channel_FileUpServer) error {
//	var dataC chan *grpcchannel.FileDataUp
//	var ok bool
//
//	for {
//		fu, err := stream.Recv()
//		if err == io.EOF {
//			logrus.Info("FileUp:Recv:文件收完了")
//			if dataC != nil {
//				close(dataC)
//			}
//			return nil
//		}
//
//		if err != nil {
//			logrus.Infof("FileUp:fus.Recv:%v\n", err)
//			return err
//		}
//
//		dataC, ok = model.FileUpDataM[fu.Mid]
//		if !ok {
//			logrus.Errorf("FileUp:FileUP结果通道未就绪")
//			return err
//		}
//
//		dataC <- fu
//	}
//}
