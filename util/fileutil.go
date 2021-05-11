package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadFileLines(fileName string) []string {

	a := make([]string, 5)
	f, err := os.Open(fileName)

	if err != nil {
		fmt.Printf("error opening file: %v\n", err)
		return nil
	}
	defer f.Close()

	var e error = nil
	var s string

	r := bufio.NewReader(f)

	for e == nil {
		s, e = r.ReadString('\n')
		a = append(a, strings.Trim(s, "\r\n "))
	}
	return a
}

func main() {

	if len(os.Args) < 2 {
		println("ini file name missing")
		os.Exit(-1)
	}
	args := os.Args[1:]
	fn := args[0]

	m := ParseIniFile(ReadFileLines(fn))
	fmt.Printf("%v\n", m)
}
