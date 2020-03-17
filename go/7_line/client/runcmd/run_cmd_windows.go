// +build windows

package runcmd

import (
	"bytes"
	"context"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"line/client/log"
	"line/client/model"
	"os/exec"
	"regexp"
	"time"
	"unicode/utf8"
)

var newLineRegex = regexp.MustCompile(`\r*\n`)

func Run(tmout int, cmd ...string) (int, string, string) {
	var c *exec.Cmd
	if len(cmd) == 1 {
		c = exec.Command("cmd", "/C", cmd[0])
	} else {
		c = exec.Command(cmd[0], cmd[1:]...)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Start()
	if err != nil {
		log.Errorf("exec.Cmd.Start(%v),%v\n", cmd, err)
		return -1, "", err.Error()
	}
	log.Infof("Start(%s)\n", cmd)

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*time.Duration(tmout))

	go func() {
		err := c.Wait()
		if err != nil {
			log.Errorf("等待结果失败, err=%v\n", err)
		}
		cancelFunc()
	}()

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		err = c.Process.Kill()
		if err != nil {
			log.Errorf("杀死子进程失败, cmd = %v, err=%v\n", cmd, err)
			return -4, "", fmt.Sprintf("等待子进程超时,cmd=%v, err=%v", cmd, err)
		}
		return -3, "", fmt.Sprintf("等待结果超时, 正常结束子进程。 cmd=%v", cmd)
	}

	var exitcode int
	if !c.ProcessState.Success() {
		exitcode = 1
	}

	return exitcode, legalUTF8Str(stdout.Bytes()), legalUTF8Str(stderr.Bytes())
}

func legalUTF8Str(bs []byte) string {
	bs = newLineRegex.ReplaceAll(bs, []byte{'\n'})

	if utf8.Valid(bs) {
		return string(bs)
	}

	if model.CodeSet == 936 {
		ustr, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader(bs), simplifiedchinese.GBK.NewDecoder()))
		if err == nil {
			return string(ustr)
		}
	}

	str := string(bs)
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
