package internal

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func Logger() *slog.Logger {
	w := os.Stderr

	return slog.New(
		tint.NewHandler(w, &tint.Options{
			AddSource:  true,
			NoColor:    !isatty.IsTerminal(w.Fd()),
			TimeFormat: time.RFC3339,
		}),
	)
}
