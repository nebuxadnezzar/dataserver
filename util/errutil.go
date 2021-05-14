//errutil.go
package util

type Error struct {
	code int
	msg  string
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Code() int {
	return e.code
}

func (e *Error) Set(c int, s string) {
	e.msg = s
	e.code = c
}
