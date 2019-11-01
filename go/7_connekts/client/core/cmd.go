package core

import (
	"connekts/client/log"
	"connekts/client/runcmd"
	"connekts/common"
	gc "connekts/grpcchannel"
	"context"
	"encoding/json"
	"time"
)


func handleCMD(pong *gc.Pong, cc gc.ChannelClient) {
	println("cmd:", string(pong.Data))
	var cmd common.CmdPong
	err := json.Unmarshal(pong.Data, &cmd)
	if err != nil {
		log.Errorf("Unmarshal:%v\n", err)
		return
	}
	rc, stdout, stderr := runcmd.Run(cmd.Cmd,cmd.Timeout)

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*60)
	defer cancel()

	_, err = cc.CmdResult(ctx, &gc.CmdOutput{ReturnCode: int32(rc), Stdout: stdout, Stderr: stderr, Mid: staticInfo.MachineID})
	if err != nil {
		log.Errorf("cc.CmdResult:%v\n", err)
		return
	}
}
