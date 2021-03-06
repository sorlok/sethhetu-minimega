// Copyright (2012) Sandia Corporation.
// Under the terms of Contract DE-AC04-94AL85000 with Sandia Corporation,
// the U.S. Government retains certain rights in this software.

package main

import (
	"fmt"
	log "minilog"
	"os/exec"
)

var externalProcesses = map[string]string{
	"qemu":     "kvm",
	"ip":       "ip",
	"ovs":      "ovs-vsctl",
	"dnsmasq":  "dnsmasq",
	"kill":     "kill",
	"dhcp":     "dhclient",
	"openflow": "ovs-ofctl",
}

// check for the presence of each of the external processes we may call,
// and error if any aren't in our path
func externalCheck(c cliCommand) cliResponse {
	if len(c.Args) != 0 {
		return cliResponse{
			Error: "check does not take any arguments",
		}
	}
	for _, i := range externalProcesses {
		path, err := exec.LookPath(i)
		if err != nil {
			e := fmt.Sprintf("%v not found", i)
			return cliResponse{
				Error: e,
			}
		} else {
			log.Info("%v found at: %v", i, path)
		}
	}
	return cliResponse{}
}

func process(p string) string {
	path, err := exec.LookPath(externalProcesses[p])
	if err != nil {
		log.Errorln(err)
		return ""
	}
	return path
}
