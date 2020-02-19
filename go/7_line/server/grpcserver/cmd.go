package grpcserver

import (
	"context"
	"github.com/sirupsen/logrus"
	"line/grpcchannel"
	"line/server/model"
)

func (s *grpcServer) CmdResult(ctx context.Context, output *grpcchannel.CmdOutput) (*grpcchannel.EmptyResp, error) {
	logrus.Debugf("Run:output: %+v", output)
	outputC, ok := model.CmdOutM[output.Mid]
	if !ok {
		logrus.Errorf("server.Run:CmdOut channel未就绪")
		return &grpcchannel.EmptyResp{}, nil
	}

	outputC <- grpcchannel.CmdOutput{ReturnCode: output.ReturnCode, Stdout: output.Stdout, Stderr: output.Stderr}
	logrus.Debugf("Run:send cmdout -> outputC done, chan addr:%p", outputC)
	return &grpcchannel.EmptyResp{}, nil
}
