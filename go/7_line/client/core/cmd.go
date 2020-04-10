package core

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"line/client/runcmd"
	"line/common/connection/pb"
	"line/common/sharedmodel"
	"regexp"
	"time"
)

var repRegex = regexp.MustCompile(`[ \r\t]+`)

func handleCMD(pong *pb.Pong, cc pb.ChannelClient) {
	logrus.Debugf("cmd: %s", string(pong.Data))
	var cmd sharedmodel.CmdPong
	err := json.Unmarshal(pong.Data, &cmd)
	if err != nil {
		logrus.Errorf("Unmarshal:%v", err)
		return
	}

	var rc int
	var stdout, stderr string
	if cmd.Cmd == "" {
		rc = 0
		stderr = "命令不能为空"
	} else {
		var strs []string
		if cmd.InShell {
			strs = []string{cmd.Cmd}
		} else {
			strs = repRegex.Split(cmd.Cmd, -1)
		}
		rc, stdout, stderr = runcmd.Run(cmd.Timeout, strs...)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
	defer cancel()

	_, err = cc.CmdResult(ctx, &pb.CmdOutput{ReturnCode: int32(rc), Stdout: stdout, Stderr: stderr, Mid: staticInfo.MachineID})
	if err != nil {
		logrus.Errorf("cc.CmdResult:%v", err)
		return
	}
}
