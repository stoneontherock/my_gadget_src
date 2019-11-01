package grpcserver

import (
	gc "connekts/grpcchannel"
	"connekts/server/model"
	"context"
	"github.com/sirupsen/logrus"
)

func (s *server) CmdResult(ctx context.Context, output *gc.CmdOutput) (*gc.EmptyResp, error) {
	logrus.Debugf("Run:output: %+v", output)
	outputC, ok := model.CmdOutM[output.Mid]
	if !ok {
		logrus.Errorf("server.Run:CmdOut channel未就绪")
		return &gc.EmptyResp{}, nil
	}

	outputC <- gc.CmdOutput{ReturnCode: output.ReturnCode, Stdout: output.Stdout, Stderr: output.Stderr}
	logrus.Debugf("Run:send cmdout -> outputC done, chan addr:%p", outputC)
	return &gc.EmptyResp{}, nil
}
