// +build linux

// Package client 客户端功能
package runcmd

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

// Run 带timeout执行系统命令，超时就杀死子进程
func Run(tmout int, strs ...string) (int, string, string) {
	//logrus.Infof("=== del === cmd:+%v len=%d", cmd, len(cmd))
	for i, str := range strs {
		strs[i] = strings.Replace(str, "\r\n", "\n", -1)
	}

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
		logrus.Errorf("c.Start(%v),%v", strs, err)
		return -1, "", err.Error()
	}

	logrus.Debugf("Start(%v)", strs)

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
			logrus.Errorf("syscall.Wait4 err,%v", err)
			return -2, "", fmt.Sprintf("syscall.Wait4(),err:%v", err)
		}
		if chp == c.Process.Pid {
			logrus.Debugf("Child process, exited code=%d", ws.ExitStatus())
			break
		}
	}

	if i >= max {
		logrus.Errorf("cmd:%v, TIMEOUT", strs)
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
