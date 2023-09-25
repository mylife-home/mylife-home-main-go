package tools

import (
	"fmt"

	"github.com/pkg/errors"
)

func GetStackTrace(err error) errors.StackTrace {
	if st, ok := err.(stackTracer); ok {
		return st.StackTrace()
	} else {
		return nil
	}
}

func GetStackTraceStr(err error) string {
	if st, ok := err.(stackTracer); ok {
		return fmt.Sprintf("%+v", st.StackTrace())
	} else {
		return ""
	}
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
