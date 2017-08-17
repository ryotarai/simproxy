package accesslogger

import (
	"fmt"
	"io"
	"strings"

	"github.com/ryotarai/simproxy/handler"
)

type LTSVAccessLogger struct {
	w      io.Writer
	Fields []string
}

func (l *LTSVAccessLogger) Log(r handler.LogRecord) error {
	a := make([]string, len(l.Fields))
	for i, f := range l.Fields {
		a[i] = fmt.Sprintf("%s:%s", f, r[f])
	}
	line := strings.Join(a, "\t")
	_, err := fmt.Fprintf(l.w, "%s\n", line)
	if err != nil {
		return err
	}
	return nil
}
