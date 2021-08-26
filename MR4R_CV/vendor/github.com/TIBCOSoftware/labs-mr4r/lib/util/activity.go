/*
 * Copyright © 2019. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/P-f1/LC/flogo-lib/core/activity"
)

func ActivityId(ctx activity.Context) string {
	return fmt.Sprintf("%s_%s", ctx.FlowDetails().Name(), ctx.TaskName())
}

// Process monitor
type ProcessMonitor struct {
	CmdName *string
	CmdArgs *[]string
	Process *os.Process
	Cmd     *exec.Cmd
	Output  *[]byte
	Err     error
}

// Process state listener interface
type ProcessStateListener interface {
	OnComplete(processMonitor *ProcessMonitor)
	OnError(processMonitor *ProcessMonitor, err error)
}

// Method to fork a process for given command
// and return ProcessMonitor
func Fork(processStateListener ProcessStateListener, cmdName string, cmdArgs ...string) {
	go func() {
		processMonitor := &ProcessMonitor{}
		args := strings.Join(cmdArgs, ",")
		command := exec.Command(cmdName, args)
		output, err := command.Output()
		if err != nil {
			processMonitor.Err = err
			processStateListener.OnError(processMonitor, err)
		}
		processMonitor.Output = &output
		processStateListener.OnComplete(processMonitor)
	}()
}
