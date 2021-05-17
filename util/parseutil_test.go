// parseutil_testing.go
package util

import (
	"fmt"
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

func TestConversion(t *testing.T) {

	a := "hello"
	fmt.Printf("A: %s %v\n", a, IsString(a))
	fmt.Printf("A: %s %v\n", a, IsArray(a))
}
