package commands

import "os/exec"

func StartBackgroundCommand(cmd *exec.Cmd) error {
	return cmd.Start()
}
