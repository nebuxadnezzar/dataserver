// parseutil.go
package util

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

func ParseIniFile(records []string) map[string]map[string]string {

	mp := make(map[string]map[string]string)
	var m map[string]string
	re := regexp.MustCompile("\\[(.*)\\]")
	rb := regexp.MustCompile("(\\[|\\])")

	for _, v := range records {

		if re.MatchString(v) {
			//fmt.Printf( "HEADER %s\n", rb.ReplaceAllString( v, "") )
			m = make(map[string]string)
			mp[rb.ReplaceAllString(v, "")] = m
			continue
		}
		//fmt.Println( v )

		if strings.HasPrefix(v, ";") {
			continue
		}

		if ss := strings.SplitN(v, "=", 2); len(ss) > 1 {
			m[ss[0]] = ss[1]
		}
	}
	return mp
}

func CreateKeyValuePairs(m map[string][]string, sep string, quoted bool) string {
	var q string = ``
	if quoted {
		q = `"`
	}

	//b := new(bytes.Buffer) // same as below but might allocate on the heap
	b := &bytes.Buffer{}
	for key, value := range m {
		var s string = ``
		s = strings.Join(value, `,`)
		fmt.Fprintf(b, "%s%s%s%s%s%s%s ", q, key, q, sep, q, s, q)
	}
	return b.String()
}

/*
const MAX_PARAM_VALS int = 2
func ParseRequestForm(form string) map[string][]string {

	m := make(map[string][]string)

	if ss := strings.Split(form, "&"); len(ss) > 0 {

		for i, v := range ss {
			if zz := strings.SplitN(v, "=", 2); len(zz) == 2 {
				if m[zz[0]] == nil {
					m[zz[0]] = make([]string, MAX_PARAM_VALS)
				}
				a := m[zz[0]]
				fmt.Printf("%d %v %v\n", i, a, zz[1])
				fmt.Printf("%d %d %v\n", len(a), cap(a), a)
				if i < len(a) {
					a[i] = zz[1]
				} else {
					m[zz[0]] = append(a, zz[1])
				}
			}
		}
	}
	return m
}
*/
