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
	"net/http"
	"os/exec"
	"strconv"
	"time"
)

func main() {
	execCMD("systemctl stop ssh")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/ssh/on", enableSSHD)
	http.HandleFunc("/ssh/off", disableSSHD)
	http.HandleFunc("/ssh/status", statusSSHD)
	http.ListenAndServe(":3389", nil)
}

func enableSSHD(wr http.ResponseWriter, req *http.Request) {
	tmout := req.FormValue("timeout")
	dur, err := strconv.Atoi(tmout)
	if err != nil {
		http.Error(wr, "`timeout` required, type int", 400)
		return
	}

	execCMD("systemctl start ssh")
	go func() {
		tk := time.Tick(time.Second * time.Duration(dur))
		<-tk
		execCMD("systemctl stop ssh")
	}()
	statusSSHD(wr, req)
}

func disableSSHD(wr http.ResponseWriter, req *http.Request) {
	execCMD("systemctl stop ssh")
	statusSSHD(wr, req)
}

func statusSSHD(wr http.ResponseWriter, req *http.Request) {
	wr.Write([]byte(execCMD("systemctl status ssh")))
}

func execCMD(c string) string {
	cmd := exec.Command("/bin/bash", "-c", c+"; exit 0")
	bs, err := cmd.CombinedOutput()
	if err != nil {
		return err.Error()
	}

	return string(bs)
}
