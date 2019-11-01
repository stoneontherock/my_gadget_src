// +build linux

// Package client 客户端功能
package runcmd

import (
	"bytes"
	"connekts/client/log"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Run 带timeout执行系统命令，超时就杀死子进程
func Run(cmd string,tmout int) (int, string, string) {
	log.Infof("=== del === cmd:+%v len=%d\n",cmd,len(cmd))

	var c *exec.Cmd

	if strings.Contains(cmd,"...") {
		cmds := strings.Split(cmd,"...")
		c = exec.Command(cmds[0], cmds[1:]...)
	}else{
		c = exec.Command("/bin/bash","-c", cmd)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Start()
	if err != nil {
		log.Errorf("c.Start(%v),%v\n", cmd, err)
		return -1, "", err.Error()
	}

	log.Infof("Start(%s)\n", cmd)

	var (
		ws  syscall.WaitStatus
		chp int
		i   int
	)

	timeout := time.Second * time.Duration(tmout) //超过X秒子进程没有结束，就不等待子进程了(杀死子进程)。
	interval := time.Millisecond * 50
	max := int(timeout / interval)
	for i = 0; i < max; i++ {
		time.Sleep(interval)
		chp, err = syscall.Wait4(c.Process.Pid, &ws, syscall.WNOHANG, nil)
		if err != nil {
			log.Errorf("syscall.Wait4 err,%v\n", err)
			return -2, "", fmt.Sprintf("syscall.Wait4(),err:%v", err)
		}
		if chp == c.Process.Pid {
			log.Infof("Child process, exited code=%d\n", ws.ExitStatus())
			break
		}
	}

	if i >= max {
		log.Errorf("cmd:%v, TIMEOUT\n", cmd)
		c.Process.Kill()
		return -3, "", fmt.Sprintf("timeout,loop=%d,output=%s", i, stdout.String()+stderr.String())
	}

	return ws.ExitStatus(), stdout.String(), stderr.String()
}
