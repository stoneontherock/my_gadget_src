package grpcserver

//import (
//	"line/grpcchannel"
//	"line/server/model"
//	"context"
//	"github.com/sirupsen/logrus"
//)
//
//func (s *grpcServer) ListFile(ctx context.Context, fl *grpcchannel.FileList) (*grpcchannel.EmptyResp, error) {
//	logrus.Debugf("ListFile: len=%d", len(fl.Fs))
//	flC, ok := model.ListFileM[fl.Mid]
//	if !ok {
//		logrus.Errorf("server.Run:CmdOut channel未就绪")
//		return &grpcchannel.EmptyResp{}, nil
//	}
//
//	flC <- fl
//	logrus.Debugf("ListFile:send Fs -> flC done")
//	return &grpcchannel.EmptyResp{}, nil
//}
