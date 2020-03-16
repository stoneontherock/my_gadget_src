// +build windows

package core

import (
	"line/client/machineid"
	"line/client/model"
	"line/client/runcmd"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var kvRegex = regexp.MustCompile(`[^\r\n]+`)

func static() model.StaticInfo {
	var si model.StaticInfo
	si.MachineID, _ = machineid.ID()
	if si.MachineID == "" {
		si.MachineID = "sample_machine_id"
	}

	//nt版本
	_, k, _ := runcmd.Run(10, "wmic", "os", "get", "version", "/value")
	kernel := getValue(k)[0]
	//主机名
	hostName, _ := os.Hostname()
	//非guest，非管理员用户: wmic useraccount WHERE (Status='ok' AND Name!='guest' AND Name!='Administrator') get name /value
	_, users, _ := runcmd.Run(10, `wmic`, `useraccount`, `WHERE`, `(Status='ok' AND Name!='guest' AND Name!='Administrator')`, `get`, `name`, `/value`)
	us := strings.Join(getValue(users), ",")

	ks := strings.Split(kernel, ".")
	if len(ks) >= 2 {
		kernel = strings.Join(ks[:2], ".")
	}
	si.Kernel = "NT " + kernel
	si.OsInfo = "主机名:" + hostName + ";普通用户:" + us

	//windows专有信息
	//codeSet
	_, codeSet, _ := runcmd.Run(10, "wmic", "os", "get", "codeSet", "/value")
	cs := getValue(codeSet)[0]
	model.CodeSet, _ = strconv.Atoi(cs)
	_, logicDisks, _ := runcmd.Run(10, "wmic", "logicaldisk", "where", "drivetype=3", "get", "name", "/value")
	disks := getValue(logicDisks)
	if len(disks) == 0 {
		model.WinDiskList = []string{"C:"}
	} else {
		model.WinDiskList = disks
		si.OsInfo += ";磁盘分区:" + strings.Join(disks, ",")
	}
	return si
}

func getValue(s string) []string {
	ss := kvRegex.FindAllString(s, -1)
	for i, s := range ss {
		kv := strings.SplitN(s, "=", 2)
		if len(kv) == 2 {
			ss[i] = kv[1]
		}
	}
	if len(ss) == 0 {
		return []string{"空内容"}
	}

	return ss
}
