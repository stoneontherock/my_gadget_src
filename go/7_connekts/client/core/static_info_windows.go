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

	//os名
	_, caption, _ := runcmd.Run(10, "wmic", "os", "get", "caption", "/value")
	cp := getValue(caption)[0]
	//codeSet
	_, codeSet, _ := runcmd.Run(10, "wmic", "os", "get", "codeSet", "/value")
	cs := getValue(codeSet)[0]
	//主机名
	hostName, _ := os.Hostname()
	//非guest，非管理员用户: wmic useraccount WHERE (Status='ok' AND Name!='guest' AND Name!='Administrator') get name /value
	_, users, _ := runcmd.Run(10, `wmic`, `useraccount`, `WHERE`, `(Status='ok' AND Name!='guest' AND Name!='Administrator')`, `get`, `name`, `/value`)
	us := strings.Join(getValue(users), ";")

	si.OS = strings.Replace(cp, "Microsoft ", "", -1) + "(codeSet=" + cs + ")"
	si.Hostname = hostName + "(user=" + us + ")"

	//windows专有信息
	model.CodeSet, _ = strconv.Atoi(cs)
	_, logicDisks, _ := runcmd.Run(10, "wmic", "logicaldisk", "where", "drivetype=3", "get", "name", "/value")
	disks := getValue(logicDisks)
	if len(disks) == 0 {
		model.WinDiskList = []string{"C:"}
	} else {
		model.WinDiskList = disks
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
