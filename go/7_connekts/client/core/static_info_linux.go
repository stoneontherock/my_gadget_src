// +build linux

package core

import (
	"connekts/client/machineid"
	"connekts/client/model"
	"connekts/client/runcmd"
	"strings"
)

func static() model.StaticInfo {
	var si model.StaticInfo
	si.MachineID, _ = machineid.ID()
	if si.MachineID == "" {
		si.MachineID = "sample_machine_id"
	}

	_, kernel, _ := runcmd.Run("uname...-r", 10)
	_, hostname, _ := runcmd.Run("hostname", 10)

	si.OS = "linux " + strings.ReplaceAll(kernel, "\n", "")
	si.Hostname = strings.ReplaceAll(hostname, "\n", "")
	return si
}
