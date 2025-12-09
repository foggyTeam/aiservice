package log

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
)

type JsonFormatter struct {
	w io.Writer
}

func (j JsonFormatter) Write(p []byte) (int, error) {
	var v any
	if err := json.Unmarshal(p, &v); err != nil {
		return j.w.Write(p)
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return j.w.Write(p)
	}
	out = append(out, '\n')
	return j.w.Write(out)
}

func NewJsonFormatter() JsonFormatter {
	return JsonFormatter{w: os.Stdout}
}

func SetupJsonLogger() *slog.Logger {
	logger := slog.New(slog.NewJSONHandler(NewJsonFormatter(), nil))
	slog.SetDefault(logger)
	return logger
}
