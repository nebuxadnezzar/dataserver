// parseutil_testing.go
package util

import (
	"fmt"
	"regexp"
	"testing"
)

func TestMapToString(t *testing.T) {
	fmt.Println("parseutil testing")

	m := map[string][]string{
		"a1": []string{"one", "two"},
		"a2": []string{"three"},
	}
	fmt.Printf("MAP: %s\n", CreateKeyValuePairs(m, "\t", true))
	fmt.Printf("MAP: %s\n", CreateKeyValuePairs(m, " ", false))
	fmt.Printf("MAP: %s\n", CreateKeyValuePairs(m, `,`, true))
	fmt.Printf("MAP: %s\n", CreateKeyValuePairs(m, `=`, true))

}

func TestSterilize(t *testing.T) {

	rx := regexp.MustCompile("[;&|]")
	s := "ls -la ; rm -rf | ps -ef"
	ss := Sterilize(s)
	match := rx.MatchString(ss)
	fmt.Printf("Sterilize: %s Match: %v\n", ss, match)
	if match {
		t.Errorf("Failed to clean %s\n", s)
	}

}

func TestConversion(t *testing.T) {

	a := "hello"
	fmt.Printf("A: %s %v\n", a, IsString(a))
	fmt.Printf("A: %s %v\n", a, IsArray(a))
}
