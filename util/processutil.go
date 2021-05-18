// processutil.go
package util

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	//"strings"
)

func Spawn(buf *bytes.Buffer, cmdStr string, args ...string) error {

	procAttr := &os.ProcAttr{Files: []*os.File{nil, os.Stdout, os.Stderr},
		Env: os.Environ(),
		Dir: "./",
	}
	var path string
	var err error

	if path, err = exec.LookPath(cmdStr); err != nil {
		fmt.Errorf("%s\n", err.Error())
		return err
	}
	r, w, _ := os.Pipe()
	procAttr.Files[1] = w
	//fmt.Printf("CGI ARGS: %s\n", strings.Join(args, ` `))
	if _, err = os.StartProcess(path, args, procAttr); err != nil {
		fmt.Errorf("%s\n", err.Error())
		return err
	}
	w.Close()
	io.Copy(buf, r)

	return nil
}
