//通过http管理sshd服务
/*
----debain:/lib/systemd/system/mad_sshd.service -----
[Unit]
Description= Server Daemon
After=multi-user.target

[Service]
ExecStart=/root/ssh_ctl
Restart=on-failure

[Install]
WantedBy=multi-user.target
------------------------------
*/

package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	simpleRunCMD("systemctl stop ssh")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/ssh/on", enableSSHD)
	http.HandleFunc("/ssh/off", disableSSHD)
	http.HandleFunc("/ssh/status", statusSSHD)
	http.HandleFunc("/ssh/cmd", sshCMD)
	http.ListenAndServe(":3389", nil)
}

func enableSSHD(wr http.ResponseWriter, req *http.Request) {
	tmout := req.FormValue("timeout")
	dur, err := strconv.Atoi(tmout)
	if err != nil {
		http.Error(wr, "`timeout` required, type int", 400)
		return
	}

	simpleRunCMD("systemctl start ssh")
	go func() {
		tk := time.Tick(time.Second * time.Duration(dur))
		<-tk
		simpleRunCMD("systemctl stop ssh")
	}()
	statusSSHD(wr, req)
}

func disableSSHD(wr http.ResponseWriter, req *http.Request) {
	simpleRunCMD("systemctl stop ssh")
	statusSSHD(wr, req)
}

func statusSSHD(wr http.ResponseWriter, req *http.Request) {
	wr.Write([]byte(simpleRunCMD("systemctl status ssh")))
}

func sshCMD(wr http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		wr.Write([]byte(SHELL_HTML))
		return
	}

	cmd := strings.Replace(req.PostFormValue("cmd"), "\r\n", "\n", -1)
	stdout, stderr := runCMD(30, cmd)
	fmt.Fprintln(wr, stdout+"----------------------------------------\n"+stderr)
}

func simpleRunCMD(c string) string {
	cmd := exec.Command("/bin/bash", "-c", c+"; exit 0")
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error()
	}

	return string(bs)
}

func runCMD(tmout int, cmd string) (string, string) {
	var c *exec.Cmd
	c = exec.Command("/bin/bash", "-c", cmd)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	c.Stdout = &stdout
	c.Stderr = &stderr

	err := c.Start()
	if err != nil {
		log.Printf("c.Start(%v),%v", cmd, err)
		return "", err.Error()
	}
	log.Printf("Start(%s)", cmd)

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
			log.Printf("syscall.Wait4 err,%v\n", err)
			return "", fmt.Sprintf("syscall.Wait4(),err:%v", err)
		}
		if chp == c.Process.Pid {
			log.Printf("Child process, exited code=%d\n", ws.ExitStatus())
			break
		}
	}

	if i >= max {
		log.Printf("cmd:%v, TIMEOUT\n", cmd)
		c.Process.Kill()
		return "", fmt.Sprintf("timeout,loop=%d,output=%s", i, stdout.String()+stderr.String())
	}

	return stdout.String(), stderr.String()
}

const SHELL_HTML = `
<!doctype html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport"
          content="width=device-width, user-scalable=no, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <title>Shell</title>
</head>
<body>
<form name= "form1" action="/ssh/cmd" method='POST'> 
         <textarea  rows="5" cols="100" name="cmd"></textarea> <br />
         <input type="submit" value="执行">
</form>
</body>
</html>
`
