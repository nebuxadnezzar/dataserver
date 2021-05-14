// processutil.go
package util

import (
	"io"
	"os/exec"
	"strings"
)

func RunCgi(w io.Writer, cmdPath string, args []string) error {

	cmd := exec.Command(cmdPath, strings.Join(args, " "))
	var err error
	var b []byte
	if b, err = cmd.Output(); err == nil {
		w.Write(b)
	}

	return err
}
