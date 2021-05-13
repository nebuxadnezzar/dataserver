// processutil.go
package util

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func Dummy() {
	fmt.Println("this is dummy")
}

func RunCgi(w io.Writer, cmdPath string, args []string) error {

	cmd := exec.Command(cmdPath, strings.Join(args, " "))
	var err error
	var b []byte
	if b, err = cmd.Output(); err == nil {
		w.Write(b)
	}

	return err
}