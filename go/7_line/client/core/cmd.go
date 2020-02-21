package core

import (
	"context"
	"encoding/json"
	"line/client/log"
	"line/client/runcmd"
	"line/common"
	"line/grpcchannel"
	"regexp"
	"time"
)

var repRegex = regexp.MustCompile(`[ \r\t]+`)

func handleCMD(pong *grpcchannel.Pong, cc grpcchannel.ChannelClient) {
	println("cmd:", string(pong.Data))
	var cmd common.CmdPong
	err := json.Unmarshal(pong.Data, &cmd)
	if err != nil {
		log.Errorf("Unmarshal:%v\n", err)
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

	_, err = cc.CmdResult(ctx, &grpcchannel.CmdOutput{ReturnCode: int32(rc), Stdout: stdout, Stderr: stderr, Mid: staticInfo.MachineID})
	if err != nil {
		log.Errorf("cc.CmdResult:%v\n", err)
		return
	}
}
