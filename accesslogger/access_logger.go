package accesslogger

import (
	"fmt"
	"io"

	"github.com/ryotarai/simproxy/handler"
)

func New(format string, w io.Writer, fields []string) (handler.AccessLogger, error) {
	switch format {
	case "ltsv":
		return &LTSVAccessLogger{
			w:      w,
			Fields: fields,
		}, nil
	}
	return nil, fmt.Errorf("%s is not valid format", format)
}
