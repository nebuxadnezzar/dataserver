// processutil.go
package util

import (
	"io"
	"os/exec"
	"strings"
)

func RunCgi(w io.Writer, cmdPath string, args ...string) error {
	//fmt.Printf("run cgi called %s %s\n", cmdPath, strings.Join(args, " "))
	cmd := exec.Command(cmdPath, strings.Join(args, " "))
	var err error
	var b []byte
	if b, err = cmd.Output(); err == nil {
		//fmt.Printf("cgi -> %s", string(b))
		w.Write(b)
	}

	return err
}
