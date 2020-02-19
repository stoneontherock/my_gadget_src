// +build linux

// Package client 客户端功能
package runcmd

import (
	"bytes"
	"fmt"
	"line/client/log"
	"os/exec"
	"syscall"
	"time"
	"unicode/utf8"
)

// Run 带timeout执行系统命令，超时就杀死子进程
func Run(tmout int, strs ...string) (int, string, string) {
	//log.Infof("=== del === cmd:+%v len=%d\n", cmd, len(cmd))

	var c *exec.Cmd

	if len(strs) > 1 {
		c = exec.Command(strs[0], strs[1:]...)
	} else {
		c = exec.Command("/bin/bash", "-c", strs[0])
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Start()
	if err != nil {
		log.Errorf("c.Start(%v),%v\n", strs, err)
		return -1, "", err.Error()
	}

	log.Infof("Start(%v)\n", strs)

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
		log.Errorf("cmd:%v, TIMEOUT\n", strs)
		c.Process.Kill()
		return -3, "", fmt.Sprintf("timeout,loop=%d,output=%s", i, stdout.String()+stderr.String())
	}

	return ws.ExitStatus(), legalUTF8Str(stdout.String()), legalUTF8Str(stderr.String())
}

func legalUTF8Str(str string) string {
	if utf8.ValidString(str) {
		return str
	}

	okStr := make([]rune, len(str))
	for i, r := range str {
		if r == utf8.RuneError {
			okStr[i] = '\u2662' //使用扑克的方块替代
		} else {
			okStr[i] = r
		}
	}

	return string(okStr)
}
