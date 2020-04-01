// +build linux

package core

import (
	"line/client/machineid"
	"line/client/model"
	"line/client/runcmd"
	"strings"
)

func static() model.StaticInfo {
	var si model.StaticInfo
	si.MachineID, _ = machineid.ID()
	if si.MachineID == "" {
		si.MachineID = "sample_machine_id"
	}

	_, kernel, _ := runcmd.Run(10, "uname", "-r")
	_, hostname, _ := runcmd.Run(10, "hostname")

	ss := strings.Split(kernel, ".")
	if len(ss) > 1 {
		si.Kernel = "linux " + strings.Join(ss[0:2], ".")
	} else {
		si.Kernel = "linux"
	}
	si.OsInfo = "主机名:" + strings.Replace(hostname, "\n", "", -1)
	return si
}
