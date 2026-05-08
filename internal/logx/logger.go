package logx

import (
	"fmt"
	"io"
)

type Logger struct {
	out io.Writer
}

func New(out io.Writer) Logger {
	return Logger{out: out}
}

func (l Logger) Printf(format string, args ...any) {
	if l.out == nil {
		return
	}
	fmt.Fprintf(l.out, "[dmp] "+format+"\n", args...)
}
