package grpcserver

//import (
//	gc "connekts/grpcchannel"
//	"connekts/server/model"
//	"context"
//	"github.com/sirupsen/logrus"
//)
//
//func (s *server) ListFile(ctx context.Context, fl *gc.FileList) (*gc.EmptyResp, error) {
//	logrus.Debugf("ListFile: len=%d", len(fl.Fs))
//	flC, ok := model.ListFileM[fl.Mid]
//	if !ok {
//		logrus.Errorf("server.Run:CmdOut channel未就绪")
//		return &gc.EmptyResp{}, nil
//	}
//
//	flC <- fl
//	logrus.Debugf("ListFile:send Fs -> flC done")
//	return &gc.EmptyResp{}, nil
//}
