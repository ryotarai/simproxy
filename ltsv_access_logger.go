package simproxy

import (
	"fmt"
	"strings"

	"io"

	"github.com/ryotarai/simproxy/handler"
)

type LTSVAccessLogger struct {
	w      io.Writer
	Fields []string
}

func (l *LTSVAccessLogger) Log(r handler.LogRecord) error {
	a := []string{}
	for _, f := range l.Fields {
		a = append(a, fmt.Sprintf("%s:%s", f, r[f]))
	}
	line := strings.Join(a, "\t")
	_, err := fmt.Fprintf(l.w, "%s\n", line)
	if err != nil {
		return err
	}
	return nil
}
