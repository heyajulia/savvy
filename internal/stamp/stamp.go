package stamp

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type Stamp struct {
	dir string
}

func New(directory string) *Stamp {
	return &Stamp{dir: directory}
}

func (s *Stamp) Stamp() error {
	f, err := os.OpenFile(s.today(), os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return fmt.Errorf("stamp: create file %q: %w", s.today(), err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("stamp: close file %q: %w", s.today(), err)
	}

	return nil
}

func (s *Stamp) Exists() (bool, error) {
	_, err := os.Stat(s.today())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		}

		return false, fmt.Errorf("stamp: stat file %q: %w", s.today(), err)
	}

	return true, nil
}

func (s *Stamp) Prune() error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return fmt.Errorf("stamp: read directory %q: %w", s.dir, err)
	}

	keep := date()

	for _, entry := range entries {
		if entry.Name() == keep {
			continue
		}

		path := filepath.Join(s.dir, entry.Name())

		if err := os.Remove(path); err != nil {
			slog.Warn("failed to remove old stamp", "stamp", path, "error", err)
		}
	}

	return nil
}

func (s *Stamp) today() string {
	return filepath.Join(s.dir, date())
}

func date() string {
	return time.Now().Format("2006-01-02")
}
