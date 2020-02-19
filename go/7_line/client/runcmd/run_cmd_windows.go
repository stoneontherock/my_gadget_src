// +build windows

package runcmd

import (
	"bytes"
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
	} else {
		log.Infof("Start(%s)\n", cmd)
	}

	waitErr := make(chan string, 1)
	go func() {
		err := c.Wait()
		if err != nil {
			waitErr <- err.Error()
			return
		}
		waitErr <- ""
	}()

	defer close(waitErr)
	tch := time.After(time.Second * time.Duration(tmout))
	select {
	case we := <-waitErr:
		if we != "" { //执行命令发生错误
			return -2, "", we
		}
		var exitcode int
		if !c.ProcessState.Success() {
			exitcode = 1
		}

		return exitcode, legalUTF8Str(stdout.Bytes()), legalUTF8Str(stderr.Bytes())
	case <-tch:
		err = c.Process.Kill()
		if err != nil {
			log.Errorf("kill process failed, cmd = %v\n", cmd)
			return -3, "", fmt.Sprintf("wait result timeout(%d).kill process failed,%v", tmout, err)
		}
		return -3, "", fmt.Sprintf("wait result timeout(%d).kill process successfully", tmout)
	}

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
